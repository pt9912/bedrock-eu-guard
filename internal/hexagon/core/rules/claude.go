package rules

import (
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Regel-IDs der Claude-Settings-Regeln.
const (
	IDClaudeUser      = "LH-FA-CLAUDE-001"
	IDClaudeProject   = "LH-FA-CLAUDE-002"
	IDClaudeEndpoints = "LH-FA-CLAUDE-003"
	IDClaudeWizard    = "LH-FA-CLAUDE-004"
)

// Anzeigepfade der geprüften Dateien.
const (
	UserSettingsPath = "~/.claude/settings.json"
)

// ProjectSettingsFiles sind die projektlokalen Settings (relativ zum Scan-Root).
var ProjectSettingsFiles = []string{".claude/settings.json", ".claude/settings.local.json"}

// SettingsFile ist ein geladener Dateizustand: fehlt (S==nil, Err==nil),
// kaputt (Err!=nil) oder vorhanden (S!=nil).
type SettingsFile struct {
	Path string
	S    *model.ClaudeSettings
	Err  *model.FileError
}

// ScanContext bündelt die einmal geladenen Settings für alle Scan-Regeln.
type ScanContext struct {
	User          SettingsFile
	Projects      []SettingsFile
	Devcontainers []EnvSourceFile
	Composes      []EnvSourceFile
	Dockerfiles   []EnvSourceFile
	Terraform     TerraformScan
	// SkipClaude überspringt die vier Claude-Settings-Regeln (kein Reader
	// konfiguriert); der Nullwert behält das bisherige Verhalten.
	SkipClaude bool
}

// ScanRule wertet die geladenen Settings gegen die Config aus.
type ScanRule func(sc ScanContext, cfg model.Config) []model.Finding

// ScanRules liefert die Scan-Regeln (Claude-Settings, Devcontainer, Compose,
// Dockerfile) in stabiler Reihenfolge.
func ScanRules() []ScanRule {
	return []ScanRule{CheckClaudeUser, CheckClaudeProject, CheckClaudeEndpoints, CheckClaudeWizard,
		CheckDevcontainer, CheckCompose, CheckDockerfile,
		CheckTerraformFiles, CheckTerraformRegion, CheckTerraformIAM, CheckTerraformModels}
}

func fileErrorFinding(id string, f SettingsFile) model.Finding {
	msg := "Datei nicht lesbar"
	if f.Err.Reason == model.ReasonFileInvalid {
		msg = "Datei ungültig"
	}
	return model.Finding{RuleID: id, Severity: model.FAIL, Reason: f.Err.Reason, Subject: f.Path,
		File: f.Path, Line: f.Err.Line, Message: msg,
		Recommendation: "Datei prüfen und reparieren; Inhalt wird aus Sicherheitsgründen nicht ausgegeben"}
}

// directAnthropic meldet aktive direkte Anthropic-Konfiguration (LH-FA-CLAUDE-003)
// samt Zeile des ersten Belegs.
func directAnthropic(s *model.ClaudeSettings) (what string, line int, ok bool) {
	for _, k := range []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"} {
		if e, found := s.EnvEntry(k); found {
			return k + " gesetzt", e.Line, true
		}
	}
	if e, found := s.EnvEntry("ANTHROPIC_BASE_URL"); found && hostIsAnthropic(e.Value) {
		return "ANTHROPIC_BASE_URL zeigt auf api.anthropic.com", e.Line, true
	}
	for _, h := range s.Hits {
		if h.Pattern == model.HitAnthropicHost {
			return "api.anthropic.com referenziert", h.Line, true
		}
	}
	return "", 0, false
}

func hostIsAnthropic(v string) bool {
	return strings.Contains(strings.ToLower(v), model.HitAnthropicHost)
}

func bedrockEnabled(s *model.ClaudeSettings) bool {
	e, ok := s.EnvEntry("CLAUDE_CODE_USE_BEDROCK")
	return ok && e.Value == "1"
}

// CheckClaudeUser prüft ~/.claude/settings.json (LH-FA-CLAUDE-001).
func CheckClaudeUser(sc ScanContext, _ model.Config) []model.Finding {
	if sc.SkipClaude {
		return nil
	}
	f := sc.User
	switch {
	case f.Err != nil:
		return []model.Finding{fileErrorFinding(IDClaudeUser, f)}
	case f.S == nil:
		return []model.Finding{{RuleID: IDClaudeUser, Severity: model.WARN, Reason: model.ReasonFileMissing,
			Subject: f.Path, File: f.Path, Message: "Benutzer-Settings nicht vorhanden"}}
	}
	if what, line, ok := directAnthropic(f.S); ok {
		return []model.Finding{{RuleID: IDClaudeUser, Severity: model.FAIL, Reason: model.ReasonDirectAnthropic,
			Subject: f.Path, File: f.Path, Line: line, Message: "Direkte Anthropic-Konfiguration: " + what,
			Recommendation: "Direkte Anthropic-Konfiguration entfernen und Bedrock verwenden"}}
	}
	if !bedrockEnabled(f.S) {
		return []model.Finding{{RuleID: IDClaudeUser, Severity: model.WARN, Reason: model.ReasonBedrockNotEnabled,
			Subject: f.Path, File: f.Path, Message: "CLAUDE_CODE_USE_BEDROCK=1 ist in den Benutzer-Settings nicht gesetzt",
			Recommendation: "env.CLAUDE_CODE_USE_BEDROCK auf \"1\" setzen"}}
	}
	return []model.Finding{{RuleID: IDClaudeUser, Severity: model.PASS, Reason: model.ReasonOK,
		Subject: f.Path, File: f.Path, Message: "Bedrock aktiviert, keine direkte Anthropic-Konfiguration"}}
}

// modelValues sind die Modell-IDs einer Settings-Datei (nicht geheim).
func modelValues(s *model.ClaudeSettings) []string {
	var out []string
	if s.Model != "" {
		out = append(out, s.Model)
	}
	for _, k := range []string{"ANTHROPIC_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL", "ANTHROPIC_DEFAULT_OPUS_MODEL"} {
		if e, ok := s.EnvEntry(k); ok && !e.Secret && e.Value != "" {
			out = append(out, e.Value)
		}
	}
	return out
}

// CheckClaudeProject prüft .claude/settings.json und settings.local.json
// einzeln (LH-FA-CLAUDE-002); fehlende Dateien sind optional.
func CheckClaudeProject(sc ScanContext, cfg model.Config) []model.Finding {
	if sc.SkipClaude {
		return nil
	}
	var out []model.Finding
	for _, f := range sc.Projects {
		if f.Err != nil {
			out = append(out, fileErrorFinding(IDClaudeProject, f))
			continue
		}
		if f.S == nil {
			continue
		}
		out = append(out, projectFinding(f, cfg))
	}
	return out
}

func projectFinding(f SettingsFile, cfg model.Config) model.Finding {
	fd := model.Finding{RuleID: IDClaudeProject, Subject: f.Path, File: f.Path}
	var fails, warns []string
	line := 0
	note := func(l int) {
		if line == 0 {
			line = l
		}
	}
	if e, ok := f.S.EnvEntry("ANTHROPIC_API_KEY"); ok {
		fails = append(fails, "ANTHROPIC_API_KEY gesetzt")
		note(e.Line)
	}
	if e, ok := f.S.EnvEntry("AWS_REGION"); ok && !cfg.RegionAllowed(e.Value) {
		fails = append(fails, "AWS_REGION="+e.Value+" ist keine erlaubte Region")
		note(e.Line)
	}
	for _, id := range modelValues(f.S) {
		switch Classify(id, cfg) {
		case model.ModelBlocked:
			fails = append(fails, "Modell "+id+" nicht erlaubt")
		case model.ModelUnclassified:
			warns = append(warns, "Modell "+id+" nicht klassifizierbar")
		}
	}
	fd.Line = line
	switch {
	case len(fails) > 0:
		fd.Severity, fd.Reason = model.FAIL, model.ReasonSettingsUnsafe
		fd.Message = strings.Join(append(fails, warns...), "; ")
		fd.Recommendation = "Nicht erlaubte Modelle, Regionen und API-Keys aus den Projekt-Settings entfernen"
	case len(warns) > 0:
		fd.Severity, fd.Reason = model.WARN, model.ReasonSettingsReview
		fd.Message = strings.Join(warns, "; ")
	default:
		fd.Severity, fd.Reason = model.PASS, model.ReasonOK
		fd.Message = "Projekt-Settings sind EU-safe"
	}
	return fd
}

// CheckClaudeEndpoints sucht direkte Anthropic-Endpunkte in allen
// vorhandenen Settings-Dateien (LH-FA-CLAUDE-003). FAIL nur bei aktiver
// Konfiguration; reine Textfunde und fremde Hosts sind WARN.
func CheckClaudeEndpoints(sc ScanContext, _ model.Config) []model.Finding {
	if sc.SkipClaude {
		return nil
	}
	files := append([]SettingsFile{sc.User}, sc.Projects...)
	var out []model.Finding
	for _, f := range files {
		if f.Err != nil || f.S == nil {
			continue
		}
		fd := model.Finding{RuleID: IDClaudeEndpoints, Subject: f.Path, File: f.Path}
		if what, line, ok := directAnthropic(f.S); ok {
			fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonDirectAnthropic, line
			fd.Message = "Direkter Anthropic-Endpunkt: " + what
			fd.Recommendation = "Direkte Anthropic-Konfiguration entfernen"
		} else if f.S.APIKeyHelperLine > 0 {
			fd.Severity, fd.Reason, fd.Line = model.WARN, model.ReasonAPIKeyHelper, f.S.APIKeyHelperLine
			fd.Message = "apiKeyHelper konfiguriert: liefert einen API-Key und kann eine direkte Anthropic-Verbindung ermöglichen; manuell prüfen"
			fd.Recommendation = "apiKeyHelper entfernen, wenn Claude Code über Bedrock laufen soll"
		} else if e, ok := f.S.EnvEntry("ANTHROPIC_BASE_URL"); ok && e.Value != "" {
			fd.Severity, fd.Reason, fd.Line = model.WARN, model.ReasonEndpointForeign, e.Line
			fd.Message = "ANTHROPIC_BASE_URL zeigt auf einen anderen Host; manuell prüfen"
		} else if h, ok := textOnlyHit(f.S); ok {
			fd.Severity, fd.Reason, fd.Line = model.WARN, model.ReasonEndpointTextOnly, h.Line
			fd.Message = h.Pattern + " im Text erwähnt (nicht als env gesetzt); manuell prüfen"
		} else {
			fd.Severity, fd.Reason = model.PASS, model.ReasonOK
			fd.Message = "Keine direkten Anthropic-Endpunkte"
		}
		out = append(out, fd)
	}
	return out
}

func textOnlyHit(s *model.ClaudeSettings) (model.Hit, bool) {
	for _, h := range s.Hits {
		if h.Pattern == model.HitAPIKey || h.Pattern == model.HitBaseURL {
			return h, true
		}
	}
	return model.Hit{}, false
}

// CheckClaudeWizard prüft eine vom Bedrock-Login-Wizard geschriebene
// Konfiguration in den Benutzer-Settings (LH-FA-CLAUDE-004). Die Regel
// greift nur, wenn CLAUDE_CODE_USE_BEDROCK=1 dort gesetzt ist; sonst
// meldet CLAUDE-001.
func CheckClaudeWizard(sc ScanContext, cfg model.Config) []model.Finding {
	if sc.SkipClaude {
		return nil
	}
	f := sc.User
	if f.Err != nil || f.S == nil || !bedrockEnabled(f.S) {
		return nil
	}
	fd := model.Finding{RuleID: IDClaudeWizard, Subject: f.Path, File: f.Path}
	if what, line, ok := directAnthropic(f.S); ok {
		fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonDirectAnthropic, line
		fd.Message = "Bedrock aktiviert, aber parallel direkte Anthropic-Konfiguration: " + what
		fd.Recommendation = "Direkte Anthropic-Konfiguration entfernen"
		return []model.Finding{fd}
	}
	e, ok := f.S.EnvEntry("AWS_REGION")
	switch {
	case !ok || e.Value == "":
		fd.Severity, fd.Reason = model.FAIL, model.ReasonRegionMissing
		fd.Message = "env.AWS_REGION fehlt in den Benutzer-Settings"
		fd.Recommendation = "env.AWS_REGION auf eine erlaubte EU-Region setzen"
	case !cfg.RegionAllowed(e.Value):
		fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonRegionNotAllowed, e.Line
		fd.Message = "env.AWS_REGION=" + e.Value + " ist keine erlaubte Region"
		fd.Recommendation = "env.AWS_REGION auf eine erlaubte EU-Region setzen"
	default:
		fd.Severity, fd.Reason, fd.Line = model.PASS, model.ReasonOK, e.Line
		fd.Message = "Wizard-Konfiguration: Bedrock aktiviert, AWS_REGION=" + e.Value
	}
	return []model.Finding{fd}
}
