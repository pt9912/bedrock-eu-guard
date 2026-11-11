package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// Regel-IDs der Bedrock-Regeln.
const (
	IDBedrockProfiles     = "LH-FA-BEDROCK-001"
	IDBedrockBlocked      = "LH-FA-BEDROCK-002"
	IDBedrockAllowlist    = "LH-FA-BEDROCK-003"
	IDBedrockAvailability = "LH-FA-BEDROCK-004"
)

// RegionData sind die gelesenen Bedrock-Daten einer Region.
type RegionData struct {
	Region      string
	Profiles    []model.InferenceProfile
	ProfilesErr *model.AWSError
	Models      []model.FoundationModel
	ModelsErr   *model.AWSError
	// VisibilityOnly kennzeichnet eine Zusatzregion (visibilityRegions): dort
	// werden nur Profile gelesen und nur die Sichtbarkeit bewertet.
	VisibilityOnly bool
}

// BedrockScan bündelt die Bedrock-Daten je abgefragter Region (in
// Abfragereihenfolge). Primary ist die Region für die Verfügbarkeitsprüfung.
type BedrockScan struct {
	Primary string
	Regions []RegionData
}

// QueryRegions liefert die abzufragenden Regionen: allowedRegions (stabil,
// ohne Duplikate). Zusätzliche Nicht-EU-Regionen werden bewusst nicht
// abgefragt (siehe Plan slice-009).
func QueryRegions(cfg model.Config) []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range cfg.AllowedRegions {
		if r != "" && !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

// VisibilityRegions liefert die Zusatzregionen (stabil, ohne Duplikate und
// ohne erlaubte Regionen), in denen nur die Sichtbarkeit geprüft wird.
func VisibilityRegions(cfg model.Config) []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range cfg.VisibilityRegions {
		if r != "" && !seen[r] && !cfg.RegionAllowed(r) {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

// PrimaryRegion ist AWS_REGION, wenn erlaubt, sonst die erste erlaubte Region.
func PrimaryRegion(env driven.Environment, cfg model.Config) string {
	if r, ok := env.Lookup("AWS_REGION"); ok && r != "" && cfg.RegionAllowed(r) {
		return r
	}
	if q := QueryRegions(cfg); len(q) > 0 {
		return q[0]
	}
	return ""
}

var usedModelVars = []string{"ANTHROPIC_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL", "ANTHROPIC_DEFAULT_OPUS_MODEL", "ANTHROPIC_SMALL_FAST_MODEL"}

// SettingsModel ist eine in Claude-Settings konfigurierte Modell-/Profil-ID
// (Source = Dateipfad, Var = "model" bzw. Name der env-Variable).
type SettingsModel struct{ Source, Var, ID string }

// UsedModelVars liefert die Umgebungsvariablen, die Modell-/Profil-IDs tragen.
func UsedModelVars() []string { return append([]string(nil), usedModelVars...) }

type usedModel struct{ Var, ID, Source string }

// usedModels sind die konfigurierten Modell-/Profil-IDs: zuerst die der
// Umgebung, dann die aus Claude-Settings (Duplikate gleicher Variable und ID
// entfallen).
func usedModels(s AWSScan) []usedModel {
	var out []usedModel
	seen := map[string]bool{}
	add := func(u usedModel) {
		k := u.Var + "\x00" + u.ID
		if !seen[k] {
			seen[k] = true
			out = append(out, u)
		}
	}
	for _, v := range usedModelVars {
		if id, ok := s.Env.Lookup(v); ok && id != "" {
			add(usedModel{v, id, ""})
		}
	}
	for _, m := range s.Settings {
		add(usedModel{m.Var, m.ID, m.Source})
	}
	return out
}

// label und finding-Hilfen: Herkunft aus Settings sichtbar machen.
func (u usedModel) where() string {
	if u.Source == "" {
		return ""
	}
	return " (in " + u.Source + ")"
}

// CheckBedrockProfiles listet die Inference-Profile je Region
// (LH-FA-BEDROCK-001): PASS bei gelesenen Profilen, WARN ohne sichtbare
// Profile, FAIL bei Fehler der AWS-API.
func CheckBedrockProfiles(s AWSScan, _ model.Config) []model.Finding {
	if s.Bedrock == nil {
		return nil
	}
	var out []model.Finding
	for _, r := range s.Bedrock.Regions {
		if r.VisibilityOnly {
			continue
		}
		f := model.Finding{RuleID: IDBedrockProfiles, Subject: r.Region}
		switch {
		case r.ProfilesErr != nil:
			f = awsErrorFinding(IDBedrockProfiles, r.Region, r.ProfilesErr)
			f.Message += " [" + r.Region + "]"
		case len(r.Profiles) == 0:
			f.Severity, f.Reason = model.WARN, model.ReasonProfilesNone
			f.Message = "Keine Inference-Profile sichtbar in " + r.Region
		default:
			f.Severity, f.Reason = model.PASS, model.ReasonProfilesListed
			f.Message = fmt.Sprintf("%d Inference-Profile in %s gelesen", len(r.Profiles), r.Region)
		}
		out = append(out, f)
	}
	return out
}

// CheckBedrockBlocked erkennt nicht erlaubte Profile (LH-FA-BEDROCK-002):
// FAIL, wenn eine konfigurierte (verwendete) ID blockiert ist; WARN, wenn
// nicht erlaubte Profile nur sichtbar sind; sonst PASS.
func CheckBedrockBlocked(s AWSScan, cfg model.Config) []model.Finding {
	if s.Bedrock == nil {
		return nil
	}
	var out []model.Finding
	usedBlocked := map[string]bool{}
	for _, u := range usedModels(s) {
		if Classify(u.ID, cfg) == model.ModelBlocked {
			usedBlocked[u.ID] = true
			out = append(out, model.Finding{RuleID: IDBedrockBlocked, Severity: model.FAIL, Reason: model.ReasonModelBlocked,
				Subject: u.Var, File: u.Source, Message: u.Var + "=" + u.ID + " ist nicht erlaubt" + u.where(),
				Recommendation: "EU-Profil-ID (eu.*) verwenden"})
		}
	}
	visible := map[string]string{} // Profil-ID → erste Region, in der es sichtbar ist
	var unknown []string
	for _, r := range s.Bedrock.Regions {
		if r.VisibilityOnly && r.ProfilesErr != nil {
			unknown = append(unknown, r.Region)
			continue
		}
		for _, p := range r.Profiles {
			if Classify(p.ID, cfg) == model.ModelBlocked && !usedBlocked[p.ID] {
				if _, ok := visible[p.ID]; !ok {
					visible[p.ID] = r.Region
				}
			}
		}
	}
	if len(visible) > 0 {
		ids := make([]string, 0, len(visible))
		for id := range visible {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		out = append(out, model.Finding{RuleID: IDBedrockBlocked, Severity: model.WARN, Reason: model.ReasonProfilesVisible,
			Subject:        "inference-profiles",
			Message:        fmt.Sprintf("%d nicht erlaubte Profile sichtbar, aber nicht verwendet, z. B. %s (in %s)", len(ids), ids[0], visible[ids[0]]),
			Recommendation: "Per IAM/SCP einschränken, damit sie nicht aktiv genutzt werden können"})
	}
	sort.Strings(unknown)
	for _, region := range unknown {
		out = append(out, model.Finding{RuleID: IDBedrockBlocked, Severity: model.WARN, Reason: model.ReasonVisibilityUnknown,
			Subject: region, Message: "Sichtbarkeit nicht erlaubter Profile in " + region + " nicht ermittelbar",
			Recommendation: "bedrock:ListInferenceProfiles in dieser Region erlauben oder die Region aus visibilityRegions entfernen"})
	}
	if len(out) == 0 {
		out = append(out, model.Finding{RuleID: IDBedrockBlocked, Severity: model.PASS, Reason: model.ReasonOK,
			Subject: "inference-profiles", Message: "Keine nicht erlaubten Profile verwendet oder sichtbar"})
	}
	return out
}

// profileAllowed: ID passt vollständig auf ein Muster aus allowedProfiles
// (anchored) oder ist laut SPEC-010 erlaubt.
func profileAllowed(id string, cfg model.Config) bool {
	for _, p := range cfg.AllowedProfiles {
		re, err := regexp.Compile("^(?:" + p + ")$")
		if err == nil && re.MatchString(id) {
			return true
		}
	}
	return Classify(id, cfg) == model.ModelAllowed
}

// CheckBedrockAllowlist prüft verwendete Profile gegen die Allowlist
// (LH-FA-BEDROCK-003): PASS, wenn alle verwendeten abgedeckt sind, FAIL für
// jede nicht abgedeckte; ohne verwendete Profile kein Finding.
func CheckBedrockAllowlist(s AWSScan, cfg model.Config) []model.Finding {
	if s.Bedrock == nil {
		return nil
	}
	used := usedModels(s)
	if len(used) == 0 {
		return nil
	}
	var out []model.Finding
	for _, u := range used {
		if !profileAllowed(u.ID, cfg) {
			out = append(out, model.Finding{RuleID: IDBedrockAllowlist, Severity: model.FAIL, Reason: model.ReasonProfileNotListed,
				Subject: u.Var, File: u.Source, Message: u.Var + "=" + u.ID + " ist nicht durch die Allowlist abgedeckt" + u.where(),
				Recommendation: "allowedProfiles ergänzen oder ein erlaubtes Profil verwenden"})
		}
	}
	if len(out) == 0 {
		out = append(out, model.Finding{RuleID: IDBedrockAllowlist, Severity: model.PASS, Reason: model.ReasonOK,
			Subject: "inference-profiles", Message: "Alle verwendeten Profile sind durch die Allowlist abgedeckt"})
	}
	return out
}

// baseModelID entfernt ein Geo-Präfix ("eu.", "us.", "global.", …) von einer
// Profil-ID; ARNs und IDs ohne Punkt-Präfix bleiben unverändert.
func baseModelID(id string) (string, bool) {
	if strings.HasPrefix(id, "arn:") {
		return "", false
	}
	for _, p := range []string{"eu.", "us.", "apac.", "global.", "us-gov."} {
		if strings.HasPrefix(id, p) {
			return id[len(p):], true
		}
	}
	return id, true
}

// CheckBedrockAvailability prüft, ob die konfigurierten Modelle in der
// Primärregion verfügbar sind (LH-FA-BEDROCK-004): PASS verfügbar, FAIL nicht
// verfügbar (Liste lesbar), WARN nicht eindeutig (Listenfehler, ARN,
// unbekannte Form). Ohne konfigurierte Modelle kein Finding.
func CheckBedrockAvailability(s AWSScan, _ model.Config) []model.Finding {
	if s.Bedrock == nil {
		return nil
	}
	used := usedModels(s)
	if len(used) == 0 {
		return nil
	}
	var rd *RegionData
	for i := range s.Bedrock.Regions {
		if s.Bedrock.Regions[i].Region == s.Bedrock.Primary {
			rd = &s.Bedrock.Regions[i]
		}
	}
	var out []model.Finding
	for _, u := range used {
		f := model.Finding{RuleID: IDBedrockAvailability, Subject: u.Var, File: u.Source}
		base, ok := baseModelID(u.ID)
		switch {
		case rd == nil || (rd.ModelsErr != nil && rd.ProfilesErr != nil) || !ok:
			f.Severity, f.Reason = model.WARN, model.ReasonModelUnknown
			f.Message = u.Var + "=" + u.ID + ": Verfügbarkeit nicht eindeutig ermittelbar" + u.where()
		case availableIn(rd, u.ID, base):
			f.Severity, f.Reason = model.PASS, model.ReasonModelAvailable
			f.Message = u.Var + "=" + u.ID + " ist in " + rd.Region + " verfügbar" + u.where()
		case rd.ModelsErr != nil || rd.ProfilesErr != nil:
			f.Severity, f.Reason = model.WARN, model.ReasonModelUnknown
			f.Message = u.Var + "=" + u.ID + ": Verfügbarkeit nicht eindeutig ermittelbar" + u.where()
		default:
			f.Severity, f.Reason = model.FAIL, model.ReasonModelUnavailable
			f.Message = u.Var + "=" + u.ID + " ist in " + rd.Region + " nicht verfügbar" + u.where()
			f.Recommendation = "Modell in der Region aktivieren oder eine andere Region/ein anderes Modell wählen"
		}
		out = append(out, f)
	}
	return out
}

func availableIn(rd *RegionData, id, base string) bool {
	for _, p := range rd.Profiles {
		if p.ID == id {
			return true
		}
	}
	for _, m := range rd.Models {
		if m.ID == base || m.ID == id {
			return true
		}
	}
	return false
}
