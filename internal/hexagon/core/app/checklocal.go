// Package app enthält die Use-Cases und die Aggregation (ARC-002).
package app

import (
	"sort"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// CheckLocal ist der Use-Case des Modus `local`.
type CheckLocal struct{}

// Run wertet alle LENV-Regeln aus; Findings stabil nach RuleID (SPEC-001).
func (CheckLocal) Run(env driven.Environment, cfg model.Config) ([]model.Finding, error) {
	var out []model.Finding
	for _, r := range rules.LocalRules() {
		out = append(out, r(env, cfg)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out, nil
}

// ExitCode bildet Findings/Fehler auf SPEC-030 ab: Fehler → 2,
// mindestens ein FAIL → 1, sonst 0.
func ExitCode(findings []model.Finding, err error) int {
	if err != nil {
		return 2
	}
	if model.MaxSeverity(findings) == model.FAIL {
		return 1
	}
	return 0
}
