package app

import (
	"errors"
	"sort"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// CheckAWS ist der Use-Case des Modus `aws`. Keine Credentials oder AWS
// nicht erreichbar sind Findings (FAIL, Exit 1); ein fehlendes `aws`-Binary
// ist ein Tool-Fehler (Exit 2, SPEC-030).
type CheckAWS struct {
	AWS driven.AWSReader
}

// Run ermittelt die AWS-Fakten einmal und wertet alle AWS-Regeln aus; die
// Findings sind stabil nach RuleID sortiert (SPEC-001).
func (u CheckAWS) Run(env driven.Environment, cfg model.Config) ([]model.Finding, error) {
	s := rules.AWSScan{Env: env}
	id, err := u.AWS.Identity()
	if err != nil {
		var ae *model.AWSError
		if !errors.As(err, &ae) {
			ae = &model.AWSError{Op: "sts:GetCallerIdentity", Reason: model.AWSUnreachable}
		}
		if ae.Reason == model.AWSCLIMissing {
			return nil, ae
		}
		s.IdentityErr = ae
	} else {
		s.Identity = id
	}
	var out []model.Finding
	for _, r := range rules.AWSRules() {
		out = append(out, r(s, cfg)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out, nil
}
