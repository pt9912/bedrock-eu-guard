package rules_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
)

func policyScan(files ...rules.IAMPolicyFileState) rules.ScanContext {
	return rules.ScanContext{IAMPolicies: rules.IAMPolicyScan{Searched: true, Files: files}}
}

func policy(path string, st ...model.IAMStatement) rules.IAMPolicyFileState {
	return rules.IAMPolicyFileState{Path: path, F: &model.IAMPolicyFile{Statements: st}}
}

func TestIAMPolicyFilesAssessment(t *testing.T) {
	cfg := model.DefaultConfig()
	star := []string{"*"}
	allow := func(actions, res []string, cond bool) model.IAMStatement {
		return model.IAMStatement{Effect: "Allow", Actions: actions, Resources: res, HasRegionCondition: cond}
	}
	cases := []struct {
		name string
		st   []model.IAMStatement
		none bool
		sev  model.Severity
		rsn  string
	}{
		{"bedrock:* auf * ohne Condition", []model.IAMStatement{allow([]string{"bedrock:*"}, star, false)}, false, model.FAIL, model.ReasonIAMWildcard},
		{"InvokeModel auf *", []model.IAMStatement{allow([]string{"bedrock:InvokeModel"}, star, false)}, false, model.FAIL, model.ReasonIAMWildcard},
		{"mit Region-Condition", []model.IAMStatement{allow([]string{"bedrock:*"}, star, true)}, false, model.WARN, model.ReasonIAMWildcardCond},
		{"eng", []model.IAMStatement{allow([]string{"bedrock:InvokeModel"}, []string{"arn:aws:bedrock:eu-central-1::foundation-model/x"}, false)}, false, model.PASS, model.ReasonOK},
		{"nicht auswertbar", []model.IAMStatement{{Effect: "Allow", Actions: []string{"bedrock:InvokeModel"}, Dynamic: true}}, false, model.WARN, model.ReasonIAMDynamic},
		{"NotAction ohne Aktionsliste: nicht auswertbar", []model.IAMStatement{{Effect: "Allow", Resources: star, Dynamic: true}}, false, model.WARN, model.ReasonIAMDynamic},
		{"ohne Bedrock-Aktion: kein Finding", []model.IAMStatement{allow([]string{"s3:*"}, star, false)}, true, 0, ""},
		{"Fremdformat (leere Liste): kein Finding", nil, true, 0, ""},
	}
	for _, c := range cases {
		for _, path := range []string{"iam-policy.json", "deploy/bedrock-policy.json"} {
			t.Run(c.name+"|"+path, func(t *testing.T) {
				fs := rules.CheckIAMPolicyFiles(policyScan(policy(path, c.st...)), cfg)
				if c.none {
					if len(fs) != 0 {
						t.Fatalf("got %+v", fs)
					}
					return
				}
				f := sole(t, fs)
				if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDIAMPolicyFile || f.File != path || f.Subject != path {
					t.Fatalf("got %+v", f)
				}
			})
		}
	}
}

// Fehlerregel je Muster: eindeutiger Name FAIL, breites Muster WARN (LH-FA-IAM-004).
func TestIAMPolicyFilesFileErrors(t *testing.T) {
	cfg := model.DefaultConfig()
	cases := []struct {
		path   string
		reason string
		sev    model.Severity
	}{
		{"iam-policy.json", model.ReasonFileInvalid, model.FAIL},
		{"sub/iam-policy-prod.json", model.ReasonFileUnreadable, model.FAIL},
		{"sub/app-policy.json", model.ReasonFileInvalid, model.WARN},
		{"sub/app-policy.json", model.ReasonFileUnreadable, model.WARN},
	}
	for _, c := range cases {
		st := rules.IAMPolicyFileState{Path: c.path, Err: &model.FileError{Path: c.path, Reason: c.reason}}
		f := sole(t, rules.CheckIAMPolicyFiles(policyScan(st), cfg))
		if f.Severity != c.sev || f.Reason != c.reason || f.RuleID != rules.IDIAMPolicyFile || f.File != c.path {
			t.Errorf("%s/%s: got %+v", c.path, c.reason, f)
		}
		if f.Message != "Datei ungültig" && f.Message != "Datei nicht lesbar" {
			t.Errorf("Meldung nennt mehr als die Fehlerklasse: %q", f.Message)
		}
	}
}

func TestIAMPolicyFilesNotSearchedAndFindError(t *testing.T) {
	cfg := model.DefaultConfig()
	if fs := rules.CheckIAMPolicyFiles(rules.ScanContext{}, cfg); len(fs) != 0 {
		t.Fatalf("nicht aktiv: %+v", fs)
	}
	sc := rules.ScanContext{IAMPolicies: rules.IAMPolicyScan{Searched: true, FindErr: &model.FileError{Path: ".", Reason: model.ReasonFileUnreadable}}}
	f := sole(t, rules.CheckIAMPolicyFiles(sc, cfg))
	if f.Severity != model.FAIL || f.Reason != model.ReasonFileUnreadable || f.RuleID != rules.IDIAMPolicyFile {
		t.Fatalf("got %+v", f)
	}
	// Gefunden, aber ohne Dateien: kein Finding (kein Zwang, dass es Policy-Dateien gibt).
	if fs := rules.CheckIAMPolicyFiles(policyScan(), cfg); len(fs) != 0 {
		t.Fatalf("leer: %+v", fs)
	}
}
