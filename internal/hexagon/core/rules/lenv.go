// Package rules enthält die Regel-Module (ARC-004). Regeln sind rein und
// kennen nur model und port/driven.
package rules

import (
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// Regel-IDs (Lastenheft-Anker).
const (
	IDRegion        = "LH-FA-LENV-001"
	IDDefaultRegion = "LH-FA-LENV-002"
	IDUseBedrock    = "LH-FA-LENV-003"
	IDAnthropicKey  = "LH-FA-LENV-004"
	IDDefaultModels = "LH-FA-LENV-006"
)

// Rule wertet die Umgebung gegen die Config aus (SPEC-002).
type Rule func(env driven.Environment, cfg model.Config) []model.Finding

// LocalRules liefert die LENV-Regeln in stabiler Reihenfolge.
func LocalRules() []Rule {
	return []Rule{CheckAWSRegion, CheckAWSDefaultRegion, CheckUseBedrock, CheckAnthropicKey, CheckDefaultModels}
}

// CheckAWSRegion prüft AWS_REGION (LH-FA-LENV-001, SPEC-011). Fehlt
// AWS_REGION, aber AWS_DEFAULT_REGION ist gesetzt, übernimmt LENV-002.
func CheckAWSRegion(env driven.Environment, cfg model.Config) []model.Finding {
	region, ok := env.Lookup("AWS_REGION")
	if !ok || region == "" {
		if d, dok := env.Lookup("AWS_DEFAULT_REGION"); dok && d != "" {
			return nil
		}
		return []model.Finding{{RuleID: IDRegion, Severity: model.FAIL, Reason: model.ReasonRegionMissing,
			Subject: "AWS_REGION", Message: "AWS_REGION (und AWS_DEFAULT_REGION) nicht gesetzt"}}
	}
	if !cfg.RegionAllowed(region) {
		return []model.Finding{{RuleID: IDRegion, Severity: model.FAIL, Reason: model.ReasonRegionNotAllowed,
			Subject: "AWS_REGION", Message: "AWS_REGION=" + region + " ist keine erlaubte Region"}}
	}
	return []model.Finding{{RuleID: IDRegion, Severity: model.PASS, Reason: model.ReasonOK,
		Subject: "AWS_REGION", Message: "AWS_REGION=" + region + " ist erlaubt"}}
}

// CheckAWSDefaultRegion prüft AWS_DEFAULT_REGION als Fallback (LH-FA-LENV-002),
// nur wenn AWS_REGION fehlt. Erlaubte Region → WARN (Fallback in Benutzung),
// nicht erlaubte → FAIL.
func CheckAWSDefaultRegion(env driven.Environment, cfg model.Config) []model.Finding {
	if r, ok := env.Lookup("AWS_REGION"); ok && r != "" {
		return nil
	}
	d, ok := env.Lookup("AWS_DEFAULT_REGION")
	if !ok || d == "" {
		return nil // beide fehlen: LENV-001 meldet FAIL
	}
	if !cfg.RegionAllowed(d) {
		return []model.Finding{{RuleID: IDDefaultRegion, Severity: model.FAIL, Reason: model.ReasonRegionNotAllowed,
			Subject: "AWS_DEFAULT_REGION", Message: "AWS_DEFAULT_REGION=" + d + " ist keine erlaubte Region"}}
	}
	return []model.Finding{{RuleID: IDDefaultRegion, Severity: model.WARN, Reason: model.ReasonRegionFallback,
		Subject: "AWS_DEFAULT_REGION", Message: "AWS_REGION fehlt; AWS_DEFAULT_REGION=" + d + " wird als Fallback genutzt"}}
}

// CheckUseBedrock prüft CLAUDE_CODE_USE_BEDROCK=1 (LH-FA-LENV-003).
func CheckUseBedrock(env driven.Environment, _ model.Config) []model.Finding {
	if v, ok := env.Lookup("CLAUDE_CODE_USE_BEDROCK"); ok && v == "1" {
		return []model.Finding{{RuleID: IDUseBedrock, Severity: model.PASS, Reason: model.ReasonOK,
			Subject: "CLAUDE_CODE_USE_BEDROCK", Message: "CLAUDE_CODE_USE_BEDROCK=1"}}
	}
	return []model.Finding{{RuleID: IDUseBedrock, Severity: model.FAIL, Reason: model.ReasonBedrockNotEnabled,
		Subject: "CLAUDE_CODE_USE_BEDROCK", Message: "CLAUDE_CODE_USE_BEDROCK ist nicht auf 1 gesetzt"}}
}

// CheckAnthropicKey prüft ANTHROPIC_API_KEY (LH-FA-LENV-004). Der Wert
// wird nie ausgegeben, nur "set"/"not set" (LH-NF-004).
func CheckAnthropicKey(env driven.Environment, _ model.Config) []model.Finding {
	if _, ok := env.Lookup("ANTHROPIC_API_KEY"); ok {
		return []model.Finding{{RuleID: IDAnthropicKey, Severity: model.FAIL, Reason: model.ReasonAnthropicKeySet,
			Subject: "ANTHROPIC_API_KEY", Message: "ANTHROPIC_API_KEY: set"}}
	}
	return []model.Finding{{RuleID: IDAnthropicKey, Severity: model.PASS, Reason: model.ReasonAnthropicKeyNotSet,
		Subject: "ANTHROPIC_API_KEY", Message: "ANTHROPIC_API_KEY: not set"}}
}
