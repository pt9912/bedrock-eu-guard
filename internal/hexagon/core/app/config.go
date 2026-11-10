package app

import (
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// LoadConfig lädt und validiert die Konfiguration. Fehlt die Datei, gelten
// die Defaults. Jeder Fehler (Datei kaputt, ungültiger Wert) ist ein
// Tool-Fehler: ExitCode(nil, err) == 2 (SPEC-030, LH-FA-EXIT-003) — anders
// als bei geprüften Dateien, wo Fehler zu Findings werden.
func LoadConfig(l driven.ConfigLoader) (model.Config, error) {
	p, err := l.Load()
	if err != nil {
		return model.Config{}, err
	}
	if p == nil {
		return model.DefaultConfig(), nil
	}
	return model.Resolve(*p)
}
