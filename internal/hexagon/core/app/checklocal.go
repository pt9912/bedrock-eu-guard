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
// mindestens ein FAIL → 1, sonst 0 (Schwelle FAIL, Standardverhalten).
func ExitCode(findings []model.Finding, err error) int {
	return ExitCodeAt(findings, err, model.FAIL)
}

// ExitCodeAt wie ExitCode, aber mit wählbarer Schwelle (--fail-on): Fehler → 2
// (hat Vorrang vor der Schwelle), höchster Schweregrad >= failOn → 1, sonst 0.
// Zulässig sind model.WARN und model.FAIL; jeder andere Wert gilt als
// model.FAIL (keine versehentlich zu strenge oder zu lasche Schwelle).
func ExitCodeAt(findings []model.Finding, err error, failOn model.Severity) int {
	if err != nil {
		return 2
	}
	if failOn != model.WARN {
		failOn = model.FAIL
	}
	if model.MaxSeverity(findings) >= failOn {
		return 1
	}
	return 0
}
