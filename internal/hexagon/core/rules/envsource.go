package rules

import (
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Regel-IDs der Env-Quellen-Regeln.
const (
	IDDevcontainer = "LH-FA-DEVCON-001"
	IDCompose      = "LH-FA-DOCKER-001"
	IDDockerfile   = "LH-FA-DOCKER-002"
	IDDotenv       = "LH-FA-LENV-007"
)

// Such-Muster je Quelle (Glob, relativ zum Scan-Root).
var (
	DevcontainerPatterns = []string{".devcontainer/devcontainer.json", ".devcontainer.json"}
	ComposePatterns      = []string{"docker-compose*.yml", "docker-compose*.yaml", "compose*.yml", "compose*.yaml"}
	DockerfilePatterns   = []string{"Dockerfile*"}
)

// EnvSourceFile ist ein geladener Dateizustand einer Env-Quelle.
type EnvSourceFile struct {
	Path string
	F    *model.EnvFile
	Err  *model.FileError
}

var regionVars = []string{"AWS_REGION", "AWS_DEFAULT_REGION"}

// envFacts wertet die gemeinsamen Env-Sachverhalte einer Datei aus.
type envFacts struct {
	keyEntry     *model.EnvEntry // ANTHROPIC_API_KEY / ANTHROPIC_AUTH_TOKEN / BASE_URL auf Anthropic
	keyWhat      string
	badRegion    *model.EnvEntry
	regions      int
	bedrockOn    bool
	bedrockAny   bool
	anyClaudeEnv bool
}

func factsOf(f *model.EnvFile, cfg model.Config) envFacts {
	var x envFacts
	for _, k := range []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"} {
		if es := f.All(k); len(es) > 0 && x.keyEntry == nil {
			x.keyEntry, x.keyWhat = &es[0], k+" gesetzt"
		}
	}
	for _, e := range f.All("ANTHROPIC_BASE_URL") {
		if x.keyEntry == nil && strings.Contains(strings.ToLower(e.Value), model.HitAnthropicHost) {
			e := e
			x.keyEntry, x.keyWhat = &e, "ANTHROPIC_BASE_URL zeigt auf api.anthropic.com"
		}
	}
	for _, k := range regionVars {
		for _, e := range f.All(k) {
			e := e
			x.regions++
			if !cfg.RegionAllowed(e.Value) && x.badRegion == nil {
				x.badRegion = &e
			}
		}
	}
	for _, e := range f.All("CLAUDE_CODE_USE_BEDROCK") {
		x.bedrockAny = true
		if e.Value == "1" {
			x.bedrockOn = true
		}
	}
	x.anyClaudeEnv = x.bedrockAny || x.regions > 0 || x.keyEntry != nil
	return x
}

func envBase(id string, f EnvSourceFile) model.Finding {
	return model.Finding{RuleID: id, Subject: f.Path, File: f.Path}
}

func envErr(id string, f EnvSourceFile) model.Finding {
	return fileErrorFinding(id, SettingsFile{Path: f.Path, Err: f.Err})
}

// failFacts liefert ein FAIL-Finding bei Anthropic-Zugang oder nicht
// erlaubter Region, sonst false.
func failFacts(fd model.Finding, x envFacts) (model.Finding, bool) {
	switch {
	case x.keyEntry != nil:
		fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonDirectAnthropic, x.keyEntry.Line
		fd.Message = "Direkte Anthropic-Konfiguration: " + x.keyWhat
		fd.Recommendation = "ANTHROPIC_API_KEY entfernen; Bedrock mit IAM-Rolle bzw. Credential-Provider nutzen"
	case x.badRegion != nil:
		fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonRegionNotAllowed, x.badRegion.Line
		fd.Message = x.badRegion.Key + "=" + x.badRegion.Value + " ist keine erlaubte Region"
		fd.Recommendation = "Region auf eine erlaubte EU-Region setzen"
	default:
		return fd, false
	}
	return fd, true
}

// CheckDevcontainer prüft devcontainer.json (LH-FA-DEVCON-001): FAIL bei
// gesetztem Anthropic-Zugang oder nicht erlaubter Region; PASS bei
// AWS_REGION (erlaubt) und CLAUDE_CODE_USE_BEDROCK=1; WARN ohne
// Claude-/Bedrock-Konfiguration oder bei unvollständiger Konfiguration.
func CheckDevcontainer(sc ScanContext, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range sc.Devcontainers {
		if f.Err != nil {
			out = append(out, envErr(IDDevcontainer, f))
			continue
		}
		if f.F == nil {
			continue
		}
		x := factsOf(f.F, cfg)
		fd := envBase(IDDevcontainer, f)
		if r, bad := failFacts(fd, x); bad {
			out = append(out, r)
			continue
		}
		switch {
		case !x.anyClaudeEnv:
			fd.Severity, fd.Reason = model.WARN, model.ReasonNoClaudeConfig
			fd.Message = "Keine Claude-/Bedrock-Konfiguration im Devcontainer"
			fd.Recommendation = "containerEnv/remoteEnv: CLAUDE_CODE_USE_BEDROCK=1 und AWS_REGION setzen"
		case x.bedrockOn && x.regions > 0:
			fd.Severity, fd.Reason = model.PASS, model.ReasonOK
			fd.Message = "Bedrock aktiviert, Region erlaubt"
		default:
			fd.Severity, fd.Reason = model.WARN, model.ReasonEnvIncomplete
			fd.Message = "Unvollständige Konfiguration: CLAUDE_CODE_USE_BEDROCK=1 und AWS_REGION erwartet"
			fd.Recommendation = "Fehlende Variable ergänzen"
		}
		out = append(out, fd)
	}
	return out
}

// CheckCompose prüft Compose-Dateien (LH-FA-DOCKER-001): FAIL bei
// Anthropic-Zugang oder nicht erlaubter Region, PASS wenn nur erlaubte
// Regionen gesetzt sind; ohne relevante Variablen kein Finding.
func CheckCompose(sc ScanContext, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range sc.Composes {
		if f.Err != nil {
			out = append(out, envErr(IDCompose, f))
			continue
		}
		if f.F == nil {
			continue
		}
		x := factsOf(f.F, cfg)
		fd := envBase(IDCompose, f)
		if r, bad := failFacts(fd, x); bad {
			out = append(out, r)
			continue
		}
		if x.regions == 0 {
			continue
		}
		fd.Severity, fd.Reason = model.PASS, model.ReasonOK
		fd.Message = "Nur erlaubte Regionen gesetzt"
		out = append(out, fd)
	}
	return out
}

// CheckDockerfile prüft Dockerfiles (LH-FA-DOCKER-002): FAIL bei
// Anthropic-Zugang oder nicht erlaubter Region, WARN ohne Region, sonst PASS.
func CheckDockerfile(sc ScanContext, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range sc.Dockerfiles {
		if f.Err != nil {
			out = append(out, envErr(IDDockerfile, f))
			continue
		}
		if f.F == nil {
			continue
		}
		x := factsOf(f.F, cfg)
		fd := envBase(IDDockerfile, f)
		if r, bad := failFacts(fd, x); bad {
			out = append(out, r)
			continue
		}
		if x.regions == 0 {
			fd.Severity, fd.Reason = model.WARN, model.ReasonRegionNotSetEnv
			fd.Message = "Keine AWS-Region per ENV gesetzt"
			fd.Recommendation = "ENV AWS_REGION auf eine erlaubte EU-Region setzen"
		} else {
			fd.Severity, fd.Reason = model.PASS, model.ReasonOK
			fd.Message = "Region erlaubt, kein Anthropic-Zugang"
		}
		out = append(out, fd)
	}
	return out
}

// CheckDotenv prüft .env-Dateien (LH-FA-LENV-007): FAIL bei Anthropic-Zugang
// oder nicht erlaubter Region, PASS wenn mindestens eine Region gesetzt ist
// und alle erlaubt sind; ohne relevante Variablen kein Finding.
func CheckDotenv(sc ScanContext, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range sc.Dotenvs {
		if f.Err != nil {
			out = append(out, envErr(IDDotenv, f))
			continue
		}
		if f.F == nil {
			continue
		}
		x := factsOf(f.F, cfg)
		fd := envBase(IDDotenv, f)
		if r, bad := failFacts(fd, x); bad {
			out = append(out, r)
			continue
		}
		if x.regions == 0 {
			continue
		}
		fd.Severity, fd.Reason = model.PASS, model.ReasonOK
		fd.Message = "Nur erlaubte Regionen gesetzt, kein Anthropic-Zugang"
		out = append(out, fd)
	}
	return out
}
