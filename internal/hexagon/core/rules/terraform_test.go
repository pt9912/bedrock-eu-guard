package rules_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
)

func tf(files ...rules.TerraformFileState) rules.ScanContext {
	return rules.ScanContext{Terraform: rules.TerraformScan{Searched: true, Files: files}}
}

func tfFile(path string, f model.TerraformFile) rules.TerraformFileState {
	f.Path = path
	return rules.TerraformFileState{Path: path, F: &f}
}

func TestTerraformFiles(t *testing.T) {
	cfg := model.DefaultConfig()
	if fs := rules.CheckTerraformFiles(rules.ScanContext{}, cfg); len(fs) != 0 {
		t.Errorf("nicht aktiv: %+v", fs)
	}
	f := sole(t, rules.CheckTerraformFiles(tf(), cfg))
	if f.Severity != model.WARN || f.Reason != model.ReasonNoTerraform {
		t.Errorf("keine Dateien: %+v", f)
	}
	f = sole(t, rules.CheckTerraformFiles(tf(tfFile("main.tf", model.TerraformFile{})), cfg))
	if f.Severity != model.PASS || f.Reason != model.ReasonTerraformFound {
		t.Errorf("gefunden: %+v", f)
	}
	bad := rules.TerraformFileState{Path: "x.tf", Err: &model.FileError{Path: "x.tf", Reason: model.ReasonFileInvalid, Line: 4}}
	fs := rules.CheckTerraformFiles(tf(tfFile("main.tf", model.TerraformFile{}), bad), cfg)
	if len(fs) != 2 || fs[0].Severity != model.PASS || fs[1].Severity != model.FAIL || fs[1].Line != 4 || fs[1].File != "x.tf" {
		t.Errorf("kaputt: %+v", fs)
	}
	only := tf(bad)
	if fs := rules.CheckTerraformFiles(only, cfg); len(fs) != 1 || fs[0].Severity != model.FAIL {
		t.Errorf("nur kaputt: %+v", fs)
	}
	findErr := rules.ScanContext{Terraform: rules.TerraformScan{Searched: true, FindErr: &model.FileError{Reason: model.ReasonFileUnreadable}}}
	if f := sole(t, rules.CheckTerraformFiles(findErr, cfg)); f.Severity != model.FAIL || f.Reason != model.ReasonFileUnreadable {
		t.Errorf("Suchfehler: %+v", f)
	}
}

func TestTerraformRegion(t *testing.T) {
	cfg := model.DefaultConfig()
	p := func(region string, dyn bool, line int) model.AWSProvider {
		return model.AWSProvider{Region: region, RegionDynamic: dyn, Line: line}
	}
	cases := []struct {
		name string
		ps   []model.AWSProvider
		none bool
		sev  model.Severity
		rsn  string
		line int
	}{
		{"erlaubt", []model.AWSProvider{p("eu-central-1", false, 2)}, false, model.PASS, model.ReasonOK, 0},
		{"nicht erlaubt", []model.AWSProvider{p("eu-central-1", false, 2), p("us-east-1", false, 9)}, false, model.FAIL, model.ReasonRegionNotAllowed, 9},
		{"dynamisch", []model.AWSProvider{p("", true, 3)}, false, model.WARN, model.ReasonRegionDynamic, 3},
		{"fehlt", []model.AWSProvider{p("", false, 4)}, false, model.WARN, model.ReasonRegionMissing, 4},
		{"FAIL schlaegt WARN", []model.AWSProvider{p("", true, 3), p("ap-southeast-1", false, 8)}, false, model.FAIL, model.ReasonRegionNotAllowed, 8},
		{"kein Provider", nil, true, 0, "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckTerraformRegion(tf(tfFile("main.tf", model.TerraformFile{Providers: c.ps})), cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("got %+v", fs)
				}
				return
			}
			f := sole(t, fs)
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDTerraformRegion || f.File != "main.tf" {
				t.Fatalf("got %+v", f)
			}
			if c.line != 0 && f.Line != c.line {
				t.Fatalf("line %d", f.Line)
			}
		})
	}
}

func TestTerraformIAM(t *testing.T) {
	cfg := model.DefaultConfig()
	allow := func(actions, res []string, cond, dyn bool, line int) model.IAMStatement {
		return model.IAMStatement{Effect: "Allow", Actions: actions, Resources: res, HasRegionCondition: cond, Dynamic: dyn, Line: line}
	}
	star := []string{"*"}
	cases := []struct {
		name string
		st   []model.IAMStatement
		none bool
		sev  model.Severity
		rsn  string
	}{
		{"bedrock:* auf * ohne Condition", []model.IAMStatement{allow([]string{"bedrock:*"}, star, false, false, 5)}, false, model.FAIL, model.ReasonIAMWildcard},
		{"InvokeModel auf *", []model.IAMStatement{allow([]string{"bedrock:InvokeModel"}, star, false, false, 5)}, false, model.FAIL, model.ReasonIAMWildcard},
		{"InvokeModelWithResponseStream auf *", []model.IAMStatement{allow([]string{"bedrock:InvokeModelWithResponseStream"}, star, false, false, 5)}, false, model.FAIL, model.ReasonIAMWildcard},
		{"Action * auf *", []model.IAMStatement{allow(star, star, false, false, 5)}, false, model.FAIL, model.ReasonIAMWildcard},
		{"mit Region-Condition", []model.IAMStatement{allow([]string{"bedrock:*"}, star, true, false, 5)}, false, model.WARN, model.ReasonIAMWildcardCond},
		{"eng", []model.IAMStatement{allow([]string{"bedrock:InvokeModel"}, []string{"arn:aws:bedrock:eu-central-1::foundation-model/x"}, false, false, 5)}, false, model.PASS, model.ReasonOK},
		{"List auf *: nicht riskant", []model.IAMStatement{allow([]string{"bedrock:ListFoundationModels"}, star, false, false, 5)}, false, model.PASS, model.ReasonOK},
		{"Deny wird ignoriert", []model.IAMStatement{{Effect: "Deny", Actions: []string{"bedrock:*"}, Resources: star, Line: 5}}, true, 0, ""},
		{"kein Bedrock", []model.IAMStatement{allow([]string{"s3:*"}, star, false, false, 5)}, true, 0, ""},
		{"dynamisch", []model.IAMStatement{allow([]string{"bedrock:InvokeModel"}, nil, false, true, 6)}, false, model.WARN, model.ReasonIAMDynamic},
		{"NotAction ohne Aktionsliste: nicht auswertbar", []model.IAMStatement{allow(nil, star, false, true, 6)}, false, model.WARN, model.ReasonIAMDynamic},
		{"riskant + eng: nur das riskante", []model.IAMStatement{
			allow([]string{"bedrock:*"}, star, false, false, 5),
			allow([]string{"bedrock:InvokeModel"}, []string{"arn:x"}, false, false, 9)}, false, model.FAIL, model.ReasonIAMWildcard},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckTerraformIAM(tf(tfFile("iam.tf", model.TerraformFile{Statements: c.st})), cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("got %+v", fs)
				}
				return
			}
			f := sole(t, fs)
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDTerraformIAM || f.File != "iam.tf" {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestTerraformModels(t *testing.T) {
	cfg := model.DefaultConfig()
	refs := func(vs ...string) []model.ModelRef {
		var out []model.ModelRef
		for i, v := range vs {
			out = append(out, model.ModelRef{Value: v, Line: i + 1})
		}
		return out
	}
	cases := []struct {
		name string
		r    []model.ModelRef
		none bool
		sev  model.Severity
		line int
	}{
		{"erlaubt", refs("eu.anthropic.claude-sonnet-4-6"), false, model.PASS, 0},
		{"us. blockiert", refs("eu.anthropic.claude-sonnet-4-6", "us.anthropic.claude-sonnet-4-6"), false, model.FAIL, 2},
		{"global. blockiert", refs("global.anthropic.claude-opus-4"), false, model.FAIL, 1},
		{"ARN mit fremder Region", refs("arn:aws:bedrock:ap-southeast-1::foundation-model/anthropic.claude-sonnet-4-6"), false, model.FAIL, 1},
		{"unklassifiziert", refs("anthropic.claude-opus-4"), false, model.WARN, 1},
		{"FAIL schlaegt WARN", refs("anthropic.claude-opus-4", "us.x"), false, model.FAIL, 2},
		{"keine Funde", nil, true, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckTerraformModels(tf(tfFile("m.tf", model.TerraformFile{ModelRefs: c.r})), cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("got %+v", fs)
				}
				return
			}
			f := sole(t, fs)
			if f.Severity != c.sev || f.RuleID != rules.IDTerraformModels || f.File != "m.tf" {
				t.Fatalf("got %+v", f)
			}
			if c.line != 0 && f.Line != c.line {
				t.Fatalf("line %d", f.Line)
			}
		})
	}
}
