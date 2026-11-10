// Package fakeconfig ist ein Test-Double für driven.ConfigLoader.
package fakeconfig

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Loader liefert eine vorbereitete Konfiguration oder einen Fehler.
type Loader struct {
	Config *model.PartialConfig
	Err    error
}

// Load implementiert driven.ConfigLoader.
func (l Loader) Load() (*model.PartialConfig, error) { return l.Config, l.Err }
