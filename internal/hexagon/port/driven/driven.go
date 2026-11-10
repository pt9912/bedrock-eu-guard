// Package driven enthält die Ports, die der Kern von außen braucht (ARC-005).
package driven

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Environment liest Umgebungsvariablen (Semantik wie os.LookupEnv).
type Environment interface {
	Lookup(key string) (value string, ok bool)
}

// Reporter gibt Findings aus. Er erhält nur Findings, nie Secret-Werte.
type Reporter interface {
	Report(findings []model.Finding) error
}
