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

type usedModel struct{ Var, ID string }

// usedModels sind die per Umgebung konfigurierten Modell-/Profil-IDs.
func usedModels(env driven.Environment) []usedModel {
	var out []usedModel
	for _, v := range usedModelVars {
		if id, ok := env.Lookup(v); ok && id != "" {
			out = append(out, usedModel{v, id})
		}
	}
	return out
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
	for _, u := range usedModels(s.Env) {
		if Classify(u.ID, cfg) == model.ModelBlocked {
			usedBlocked[u.ID] = true
			out = append(out, model.Finding{RuleID: IDBedrockBlocked, Severity: model.FAIL, Reason: model.ReasonModelBlocked,
				Subject: u.Var, Message: u.Var + "=" + u.ID + " ist nicht erlaubt",
				Recommendation: "EU-Profil-ID (eu.*) verwenden"})
		}
	}
	visible := map[string]bool{}
	for _, r := range s.Bedrock.Regions {
		for _, p := range r.Profiles {
			if Classify(p.ID, cfg) == model.ModelBlocked && !usedBlocked[p.ID] {
				visible[p.ID] = true
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
			Message:        fmt.Sprintf("%d nicht erlaubte Profile sichtbar, aber nicht verwendet, z. B. %s", len(ids), ids[0]),
			Recommendation: "Per IAM/SCP einschränken, damit sie nicht aktiv genutzt werden können"})
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
	used := usedModels(s.Env)
	if len(used) == 0 {
		return nil
	}
	var out []model.Finding
	for _, u := range used {
		if !profileAllowed(u.ID, cfg) {
			out = append(out, model.Finding{RuleID: IDBedrockAllowlist, Severity: model.FAIL, Reason: model.ReasonProfileNotListed,
				Subject: u.Var, Message: u.Var + "=" + u.ID + " ist nicht durch die Allowlist abgedeckt",
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
	used := usedModels(s.Env)
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
		f := model.Finding{RuleID: IDBedrockAvailability, Subject: u.Var}
		base, ok := baseModelID(u.ID)
		switch {
		case rd == nil || (rd.ModelsErr != nil && rd.ProfilesErr != nil) || !ok:
			f.Severity, f.Reason = model.WARN, model.ReasonModelUnknown
			f.Message = u.Var + "=" + u.ID + ": Verfügbarkeit nicht eindeutig ermittelbar"
		case availableIn(rd, u.ID, base):
			f.Severity, f.Reason = model.PASS, model.ReasonModelAvailable
			f.Message = u.Var + "=" + u.ID + " ist in " + rd.Region + " verfügbar"
		case rd.ModelsErr != nil || rd.ProfilesErr != nil:
			f.Severity, f.Reason = model.WARN, model.ReasonModelUnknown
			f.Message = u.Var + "=" + u.ID + ": Verfügbarkeit nicht eindeutig ermittelbar"
		default:
			f.Severity, f.Reason = model.FAIL, model.ReasonModelUnavailable
			f.Message = u.Var + "=" + u.ID + " ist in " + rd.Region + " nicht verfügbar"
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
