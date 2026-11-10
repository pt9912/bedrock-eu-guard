package rules

import (
	"regexp"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Classify klassifiziert eine Modell-/Profil-ID (SPEC-010): Block-Muster
// haben Vorrang vor Allow-Mustern, sonst unclassified. Muster, die keine
// gültigen regulären Ausdrücke sind, treffen nie (Config-Validierung
// gehört in den Config-Adapter).
func Classify(id string, cfg model.Config) model.ModelClass {
	if matchAny(cfg.BlockedModelPatterns, id) {
		return model.ModelBlocked
	}
	if matchAny(cfg.AllowedModelPatterns, id) {
		return model.ModelAllowed
	}
	return model.ModelUnclassified
}

func matchAny(patterns []string, s string) bool {
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
