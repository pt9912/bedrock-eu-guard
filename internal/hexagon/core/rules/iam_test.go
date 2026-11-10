package rules_test

import (
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

func iamScan(i *rules.IAMScan) rules.AWSScan {
	return rules.AWSScan{Env: fakeenv.Env{}, IAM: i}
}

func decisions(denied ...string) []model.ActionDecision {
	deny := map[string]bool{}
	for _, d := range denied {
		deny[d] = true
	}
	var out []model.ActionDecision
	for _, a := range append(append([]string{}, rules.IAMInvokeActions...), rules.IAMDiscoveryActions...) {
		out = append(out, model.ActionDecision{Action: a, Allowed: !deny[a]})
	}
	return out
}

func TestIAMRights(t *testing.T) {
	cfg := model.DefaultConfig()
	if fs := rules.CheckIAMRights(rules.AWSScan{Env: fakeenv.Env{}}, cfg); len(fs) != 0 {
		t.Errorf("ohne Reader: %+v", fs)
	}
	cases := []struct {
		name string
		i    *rules.IAMScan
		sev  model.Severity
		rsn  string
		sub  string
	}{
		{"alles erlaubt", &rules.IAMScan{Decisions: decisions()}, model.PASS, model.ReasonOK, ""},
		{"Invoke fehlt", &rules.IAMScan{Decisions: decisions("bedrock:InvokeModel")}, model.FAIL, model.ReasonIAMInvokeDenied, "bedrock:InvokeModel"},
		{"Converse fehlt", &rules.IAMScan{Decisions: decisions("bedrock:ConverseStream")}, model.FAIL, model.ReasonIAMInvokeDenied, "ConverseStream"},
		{"nur Discovery fehlt", &rules.IAMScan{Decisions: decisions("bedrock:ListInferenceProfiles")}, model.WARN, model.ReasonIAMDiscoveryDenied, "ListInferenceProfiles"},
		{"Invoke schlaegt Discovery", &rules.IAMScan{Decisions: decisions("bedrock:Converse", "bedrock:GetInferenceProfile")}, model.FAIL, model.ReasonIAMInvokeDenied, "Converse"},
		{"keine Antwort = fehlt", &rules.IAMScan{}, model.FAIL, model.ReasonIAMInvokeDenied, ""},
		{"Simulation verweigert", &rules.IAMScan{DecisionsErr: &model.AWSError{Op: "iam:SimulatePrincipalPolicy", Reason: model.AWSAccessDenied}}, model.WARN, model.ReasonIAMUnknown, "verweigert"},
		{"keine Credentials", &rules.IAMScan{DecisionsErr: &model.AWSError{Op: "iam:SimulatePrincipalPolicy", Reason: model.AWSNoCredentials}}, model.FAIL, model.AWSNoCredentials, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := sole(t, rules.CheckIAMRights(iamScan(c.i), cfg))
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDIAMRights || !strings.Contains(f.Message, c.sub) {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func allowStmt(actions, res []string, cond bool, regions []string, line int) model.IAMStatement {
	return model.IAMStatement{Effect: "Allow", Actions: actions, Resources: res, HasRegionCondition: cond, ConditionRegions: regions, Line: line}
}

func TestIAMBroad(t *testing.T) {
	cfg := model.DefaultConfig()
	star := []string{"*"}
	cases := []struct {
		name string
		i    *rules.IAMScan
		sev  model.Severity
		rsn  string
	}{
		{"bedrock:* auf *", &rules.IAMScan{Statements: []model.IAMStatement{allowStmt([]string{"bedrock:*"}, star, false, nil, 0)}}, model.FAIL, model.ReasonIAMWildcard},
		{"mit Condition", &rules.IAMScan{Statements: []model.IAMStatement{allowStmt([]string{"bedrock:*"}, star, true, nil, 0)}}, model.WARN, model.ReasonIAMWildcardCond},
		{"eng", &rules.IAMScan{Statements: []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, []string{"arn:aws:bedrock:eu-central-1::foundation-model/x"}, false, nil, 0)}}, model.PASS, model.ReasonOK},
		{"Fehler verweigert", &rules.IAMScan{StatementsErr: &model.AWSError{Op: "iam:GetPolicy", Reason: model.AWSAccessDenied}}, model.WARN, model.ReasonIAMUnknown},
		{"Fehler unreachable", &rules.IAMScan{StatementsErr: &model.AWSError{Op: "iam:GetPolicy", Reason: model.AWSUnreachable}}, model.FAIL, model.AWSUnreachable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := sole(t, rules.CheckIAMBroad(iamScan(c.i), cfg))
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDIAMBroad || f.Subject != "iam" {
				t.Fatalf("got %+v", f)
			}
		})
	}
	if fs := rules.CheckIAMBroad(iamScan(&rules.IAMScan{Statements: []model.IAMStatement{allowStmt([]string{"s3:*"}, star, false, nil, 0)}}), cfg); len(fs) != 0 {
		t.Errorf("kein Bedrock: %+v", fs)
	}
}

func TestIAMRegionCondition(t *testing.T) {
	cfg := model.DefaultConfig()
	star := []string{"*"}
	arnEU := []string{"arn:aws:bedrock:eu-central-1::foundation-model/a", "arn:aws:bedrock:eu-central-1::foundation-model/b"}
	cases := []struct {
		name string
		st   []model.IAMStatement
		sev  model.Severity
		rsn  string
		n    int
	}{
		{"Condition mit erlaubter Region", []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, star, true, []string{"eu-central-1"}, 3)}, model.PASS, model.ReasonOK, 1},
		{"Condition mit fremder Region", []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, star, true, []string{"eu-central-1", "us-east-1"}, 3)}, model.FAIL, model.ReasonRegionConditionNotAllowed, 1},
		{"Condition ohne Regionen", []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, star, true, nil, 3)}, model.WARN, model.ReasonRegionConditionUnknown, 1},
		{"ohne Condition", []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, star, false, nil, 3)}, model.FAIL, model.ReasonRegionConditionMissing, 1},
		{"ARN-gebunden", []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, arnEU, false, nil, 3)}, model.PASS, model.ReasonOK, 1},
		{"ARN fremde Region", []model.IAMStatement{allowStmt([]string{"bedrock:InvokeModel"}, []string{"arn:aws:bedrock:us-east-1::foundation-model/a"}, false, nil, 3)}, model.FAIL, model.ReasonRegionConditionMissing, 1},
		{"gemischt: eines ohne", []model.IAMStatement{
			allowStmt([]string{"bedrock:InvokeModel"}, star, true, []string{"eu-central-1"}, 3),
			allowStmt([]string{"bedrock:Converse"}, star, false, nil, 9)}, model.FAIL, model.ReasonRegionConditionMissing, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckIAMRegionCondition(iamScan(&rules.IAMScan{Statements: c.st}), cfg)
			if len(fs) != c.n {
				t.Fatalf("got %+v", fs)
			}
			if f := fs[0]; f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDIAMRegionCnd {
				t.Fatalf("got %+v", f)
			}
		})
	}
	other := []model.IAMStatement{allowStmt([]string{"s3:*"}, star, false, nil, 0), {Effect: "Deny", Actions: []string{"bedrock:*"}, Resources: star}}
	if fs := rules.CheckIAMRegionCondition(iamScan(&rules.IAMScan{Statements: other}), cfg); len(fs) != 0 {
		t.Errorf("kein Bedrock-Allow: %+v", fs)
	}
	if fs := rules.CheckIAMRegionCondition(rules.AWSScan{Env: fakeenv.Env{}}, cfg); len(fs) != 0 {
		t.Errorf("ohne Reader: %+v", fs)
	}
}
