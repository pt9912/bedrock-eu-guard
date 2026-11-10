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
	// Bedrock ist optional (nil = keine Bedrock-Regeln).
	Bedrock driven.BedrockReader
	// IAM ist optional (nil = keine IAM-Regeln).
	IAM driven.IAMReader
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
	if u.Bedrock != nil {
		s.Bedrock = u.readBedrock(env, cfg, s.IdentityErr)
	}
	if u.IAM != nil {
		s.IAM = u.readIAM(s.Identity, s.IdentityErr)
	}
	var out []model.Finding
	for _, r := range rules.AWSRules() {
		out = append(out, r(s, cfg)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out, nil
}

// readBedrock liest die Bedrock-Daten je Region. Ist die Identität bereits
// als unbrauchbar bekannt (keine Credentials / nicht erreichbar), werden die
// Aufrufe nicht wiederholt, sondern derselbe Fehler je Region gemeldet.
func (u CheckAWS) readBedrock(env driven.Environment, cfg model.Config, idErr *model.AWSError) *rules.BedrockScan {
	b := &rules.BedrockScan{Primary: rules.PrimaryRegion(env, cfg)}
	for _, region := range rules.QueryRegions(cfg) {
		rd := rules.RegionData{Region: region}
		if idErr != nil && (idErr.Reason == model.AWSNoCredentials || idErr.Reason == model.AWSUnreachable) {
			e := &model.AWSError{Op: "bedrock:ListInferenceProfiles", Reason: idErr.Reason}
			rd.ProfilesErr = e
			rd.ModelsErr = &model.AWSError{Op: "bedrock:ListFoundationModels", Reason: idErr.Reason}
			b.Regions = append(b.Regions, rd)
			continue
		}
		rd.Profiles, rd.ProfilesErr = asAWSError("bedrock:ListInferenceProfiles", func() ([]model.InferenceProfile, error) {
			return u.Bedrock.ListInferenceProfiles(region)
		})
		rd.Models, rd.ModelsErr = asAWSError("bedrock:ListFoundationModels", func() ([]model.FoundationModel, error) {
			return u.Bedrock.ListFoundationModels(region)
		})
		b.Regions = append(b.Regions, rd)
	}
	return b
}

func asAWSError[T any](op string, call func() ([]T, error)) ([]T, *model.AWSError) {
	v, err := call()
	if err == nil {
		return v, nil
	}
	var ae *model.AWSError
	if !errors.As(err, &ae) {
		ae = &model.AWSError{Op: op, Reason: model.AWSUnreachable}
	}
	return nil, ae
}

// readIAM liest die IAM-Sicht des Principals. Ohne ermittelte Identität wird
// derselbe Fehler gemeldet, statt die Aufrufe zu wiederholen.
func (u CheckAWS) readIAM(id *model.AWSIdentity, idErr *model.AWSError) *rules.IAMScan {
	if id == nil {
		reason := model.AWSUnreachable
		if idErr != nil {
			reason = idErr.Reason
		}
		return &rules.IAMScan{
			DecisionsErr:  &model.AWSError{Op: "iam:SimulatePrincipalPolicy", Reason: reason},
			StatementsErr: &model.AWSError{Op: "iam:GetPolicy", Reason: reason},
		}
	}
	actions := append(append([]string{}, rules.IAMInvokeActions...), rules.IAMDiscoveryActions...)
	d, derr := asAWSError("iam:SimulatePrincipalPolicy", func() ([]model.ActionDecision, error) {
		return u.IAM.SimulateBedrock(id.ARN, actions)
	})
	st, serr := asAWSError("iam:GetPolicy", func() ([]model.IAMStatement, error) {
		return u.IAM.PrincipalPolicies(id.ARN)
	})
	return &rules.IAMScan{Decisions: d, DecisionsErr: derr, Statements: st, StatementsErr: serr}
}
