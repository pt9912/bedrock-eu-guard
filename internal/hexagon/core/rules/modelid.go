package rules

import (
	"regexp"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// regionInID findet AWS-Regionen in einer ID (z. B. eu-central-1,
// ap-southeast-1, us-gov-west-1), auch in ARNs. Präfixe wie "eu." und
// Modellnamen wie "claude-sonnet-4-6" treffen nicht.
var regionInID = regexp.MustCompile(`(?i)(?:^|[^a-z0-9])([a-z]{2}(?:-[a-z]+)+-[0-9])(?:[^a-z0-9]|$)`)

// Classify klassifiziert eine Modell-/Profil-ID (SPEC-010): Block-Muster
// haben Vorrang, dann eine nicht erlaubte Region in der ID (blocked),
// dann Allow-Muster, sonst unclassified. Muster, die keine gültigen
// regulären Ausdrücke sind, treffen nie (Config-Validierung gehört in den
// Config-Adapter).
func Classify(id string, cfg model.Config) model.ModelClass {
	if matchAny(cfg.BlockedModelPatterns, id) {
		return model.ModelBlocked
	}
	if _, bad := DisallowedRegionInID(id, cfg); bad {
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

// DisallowedRegionInID liefert die erste Region in id, die nicht in
// cfg.AllowedRegions steht (Vergleich ohne Groß-/Kleinschreibung).
func DisallowedRegionInID(id string, cfg model.Config) (string, bool) {
	// Manuell iterieren, damit das Trennzeichen zwischen zwei direkt
	// aufeinanderfolgenden Regionen nicht vom ersten Treffer verbraucht wird.
	for rest := id; ; {
		loc := regionInID.FindStringSubmatchIndex(rest)
		if loc == nil {
			break
		}
		region := strings.ToLower(rest[loc[2]:loc[3]])
		rest = rest[loc[3]:]
		allowed := false
		for _, a := range cfg.AllowedRegions {
			if strings.EqualFold(a, region) {
				allowed = true
				break
			}
		}
		if !allowed {
			return region, true
		}
	}
	return "", false
}
