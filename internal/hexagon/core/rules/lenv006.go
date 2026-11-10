package rules

import (
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// defaultModelVars sind die geprüften Variablen in fester Reihenfolge.
var defaultModelVars = []string{
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
}

// CheckDefaultModels prüft die Claude-Default-Modellvariablen
// (LH-FA-LENV-006, SPEC-010): blocked → FAIL, allowed → PASS,
// unclassified → WARN. Nicht gesetzte Variablen erzeugen kein Finding.
func CheckDefaultModels(env driven.Environment, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, name := range defaultModelVars {
		id, ok := env.Lookup(name)
		if !ok || id == "" {
			continue
		}
		f := model.Finding{RuleID: IDDefaultModels, Subject: name}
		switch Classify(id, cfg) {
		case model.ModelBlocked:
			f.Severity, f.Reason = model.FAIL, model.ReasonModelBlocked
			f.Message = name + "=" + id + " ist nicht erlaubt"
		case model.ModelAllowed:
			f.Severity, f.Reason = model.PASS, model.ReasonModelAllowed
			f.Message = name + "=" + id + " ist erlaubt"
		default:
			f.Severity, f.Reason = model.WARN, model.ReasonModelUnclassified
			f.Message = name + "=" + id + " ist nicht klassifizierbar"
		}
		out = append(out, f)
	}
	return out
}
