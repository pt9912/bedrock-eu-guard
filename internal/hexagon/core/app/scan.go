package app

import (
	"errors"
	"sort"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// Scan ist der Use-Case des Modus `scan` (Claude-Settings). Dateifehler
// werden zu Findings (nie zu Exit 2, SPEC-030 bleibt unberührt).
type Scan struct {
	Settings driven.ClaudeSettingsReader
}

// Run lädt die Settings einmal und wertet alle Scan-Regeln aus; Findings
// stabil nach RuleID (SPEC-001), innerhalb einer Regel in Ladereihenfolge.
func (u Scan) Run(cfg model.Config) ([]model.Finding, error) {
	sc := rules.ScanContext{User: load(rules.UserSettingsPath, u.Settings.ReadUser)}
	for _, rel := range rules.ProjectSettingsFiles {
		rel := rel
		sc.Projects = append(sc.Projects, load(rel, func() (*model.ClaudeSettings, error) {
			return u.Settings.ReadProject(rel)
		}))
	}
	var out []model.Finding
	for _, r := range rules.ScanRules() {
		out = append(out, r(sc, cfg)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out, nil
}

func load(path string, read func() (*model.ClaudeSettings, error)) rules.SettingsFile {
	s, err := read()
	f := rules.SettingsFile{Path: path, S: s}
	if err != nil {
		var fe *model.FileError
		if !errors.As(err, &fe) {
			fe = &model.FileError{Path: path, Reason: model.ReasonFileUnreadable}
		}
		f.S, f.Err = nil, fe
	}
	return f
}
