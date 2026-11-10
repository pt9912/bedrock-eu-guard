package fakeaws

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// IAM implementiert driven.IAMReader.
type IAM struct {
	Decisions     []model.ActionDecision
	DecisionsErr  error
	Statements    []model.IAMStatement
	StatementsErr error
}

// SimulateBedrock implementiert driven.IAMReader.
func (i IAM) SimulateBedrock(_ string, _ []string) ([]model.ActionDecision, error) {
	return i.Decisions, i.DecisionsErr
}

// PrincipalPolicies implementiert driven.IAMReader.
func (i IAM) PrincipalPolicies(_ string) ([]model.IAMStatement, error) {
	return i.Statements, i.StatementsErr
}
