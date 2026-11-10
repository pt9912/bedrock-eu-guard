package rules

import (
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

const mantleVar = "CLAUDE_CODE_USE_MANTLE"

// CheckMantle prüft CLAUDE_CODE_USE_MANTLE (LH-FA-LENV-005).
//
// Nicht aktiv (nicht "1") → kein Finding. Aktiv: PASS, wenn mindestens eine
// Default-Modellvariable gesetzt ist und alle gesetzten Modell-IDs laut
// SPEC-010 erlaubt sind; sonst WARN (EU-Verfügbarkeit nicht prüfbar).
// Blockierte IDs meldet zusätzlich LENV-006 als FAIL.
func CheckMantle(env driven.Environment, cfg model.Config) []model.Finding {
	if v, ok := env.Lookup(mantleVar); !ok || v != "1" {
		return nil
	}
	f := model.Finding{RuleID: IDMantle, Subject: mantleVar}
	configured, allAllowed := 0, true
	for _, name := range defaultModelVars {
		id, ok := env.Lookup(name)
		if !ok || id == "" {
			continue
		}
		configured++
		if Classify(id, cfg) != model.ModelAllowed {
			allAllowed = false
		}
	}
	if configured > 0 && allAllowed {
		f.Severity, f.Reason = model.PASS, model.ReasonMantleVerified
		f.Message = mantleVar + "=1, alle konfigurierten Modell-IDs sind erlaubt"
		return []model.Finding{f}
	}
	f.Severity, f.Reason = model.WARN, model.ReasonMantleUnverified
	f.Message = mantleVar + "=1: EU-Verfügbarkeit manuell prüfen"
	return []model.Finding{f}
}
