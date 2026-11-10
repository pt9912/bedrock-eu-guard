package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeaws"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

func TestCheckAWSHappyPath(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "123456789012", ARN: "arn:aws:iam::123456789012:user/a"}}
	cfg := model.DefaultConfig()
	cfg.AllowedAccounts = []string{"123456789012"}
	fs, err := app.CheckAWS{AWS: r}.Run(fakeenv.Env{"AWS_PROFILE": "dev"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if model.MaxSeverity(fs) != model.PASS || len(fs) != 4 {
		t.Fatalf("got %+v", fs)
	}
	if app.ExitCode(fs, nil) != 0 {
		t.Fatal("exit != 0")
	}
	for i := 1; i < len(fs); i++ {
		if fs[i-1].RuleID > fs[i].RuleID {
			t.Fatalf("nicht sortiert: %+v", fs)
		}
	}
}

func TestCheckAWSNoCredentialsIsFindingNotToolError(t *testing.T) {
	r := fakeaws.Reader{Err: &model.AWSError{Op: "sts:GetCallerIdentity", Reason: model.AWSNoCredentials}}
	fs, err := app.CheckAWS{AWS: r}.Run(fakeenv.Env{}, model.DefaultConfig())
	if err != nil {
		t.Fatalf("kein Tool-Fehler erwartet: %v", err)
	}
	var fail bool
	for _, f := range fs {
		if f.RuleID == "LH-FA-AWS-001" && f.Severity == model.FAIL && f.Reason == model.AWSNoCredentials {
			fail = true
		}
	}
	if !fail || app.ExitCode(fs, nil) != 1 {
		t.Fatalf("got %+v", fs)
	}
}

func TestCheckAWSUnknownErrorBecomesUnreachable(t *testing.T) {
	fs, err := app.CheckAWS{AWS: fakeaws.Reader{Err: errors.New("boom: secret-output")}}.Run(fakeenv.Env{}, model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fs {
		if f.RuleID == "LH-FA-AWS-001" && (f.Reason != model.AWSUnreachable || f.Severity != model.FAIL) {
			t.Fatalf("got %+v", f)
		}
		if f.Message == "boom: secret-output" {
			t.Fatal("Fremdausgabe in Meldung")
		}
	}
}

func TestCheckAWSCLIMissingIsToolError(t *testing.T) {
	r := fakeaws.Reader{Err: &model.AWSError{Op: "sts:GetCallerIdentity", Reason: model.AWSCLIMissing}}
	fs, err := app.CheckAWS{AWS: r}.Run(fakeenv.Env{}, model.DefaultConfig())
	if err == nil || fs != nil {
		t.Fatalf("Tool-Fehler erwartet: %+v, %v", fs, err)
	}
	if app.ExitCode(nil, err) != 2 {
		t.Fatal("exit != 2")
	}
}

func TestCheckAWSAccountMismatch(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "111111111111", ARN: "arn"}}
	cfg := model.DefaultConfig()
	cfg.AllowedAccounts = []string{"123456789012"}
	fs, _ := app.CheckAWS{AWS: r}.Run(fakeenv.Env{"AWS_PROFILE": "p"}, cfg)
	if app.ExitCode(fs, nil) != 1 {
		t.Fatalf("got %+v", fs)
	}
}

func bedrockOK() fakeaws.Bedrock {
	return fakeaws.Bedrock{
		Profiles: map[string][]model.InferenceProfile{"eu-central-1": {{ID: "eu.anthropic.claude-sonnet-4-6"}}},
		Models:   map[string][]model.FoundationModel{"eu-central-1": {{ID: "anthropic.claude-sonnet-4-6"}}},
	}
}

func TestCheckAWSWithBedrock(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "123456789012", ARN: "arn"}}
	env := fakeenv.Env{"AWS_PROFILE": "p", "ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6"}
	fs, err := app.CheckAWS{AWS: r, Bedrock: bedrockOK()}.Run(env, model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, f := range fs {
		have[f.RuleID] = true
		if f.Severity == model.FAIL {
			t.Errorf("unerwartetes FAIL: %+v", f)
		}
	}
	for _, id := range []string{"LH-FA-BEDROCK-001", "LH-FA-BEDROCK-002", "LH-FA-BEDROCK-003", "LH-FA-BEDROCK-004"} {
		if !have[id] {
			t.Errorf("fehlend %s: %+v", id, fs)
		}
	}
}

func TestCheckAWSWithoutBedrockReaderSkipsBedrockRules(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "1", ARN: "a"}}
	fs, _ := app.CheckAWS{AWS: r}.Run(fakeenv.Env{"ANTHROPIC_MODEL": "us.x"}, model.DefaultConfig())
	for _, f := range fs {
		if len(f.RuleID) > 15 && f.RuleID[:15] == "LH-FA-BEDROCK-0" {
			t.Errorf("Bedrock-Regel ohne Reader: %+v", f)
		}
	}
}

func TestCheckAWSBedrockNoCredentialsReportsFailPerRegion(t *testing.T) {
	r := fakeaws.Reader{Err: &model.AWSError{Op: "sts:GetCallerIdentity", Reason: model.AWSNoCredentials}}
	cfg := model.DefaultConfig()
	cfg.AllowedRegions = []string{"eu-central-1", "eu-west-1"}
	fs, err := app.CheckAWS{AWS: r, Bedrock: bedrockOK()}.Run(fakeenv.Env{}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, f := range fs {
		if f.RuleID == "LH-FA-BEDROCK-001" && f.Severity == model.FAIL && f.Reason == model.AWSNoCredentials {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("FAIL je Region erwartet, got %d: %+v", n, fs)
	}
}

func TestCheckAWSBedrockErrorsAreFindings(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "1", ARN: "a"}}
	b := fakeaws.Bedrock{
		ProfilesErr: map[string]error{"eu-central-1": errors.New("raw output secret")},
		ModelsErr:   map[string]error{"eu-central-1": &model.AWSError{Op: "bedrock:ListFoundationModels", Reason: model.AWSAccessDenied}},
	}
	fs, err := app.CheckAWS{AWS: r, Bedrock: b}.Run(fakeenv.Env{"ANTHROPIC_MODEL": "anthropic.claude-sonnet-4-6"}, model.DefaultConfig())
	if err != nil {
		t.Fatalf("Bedrock-Fehler sind keine Tool-Fehler: %v", err)
	}
	var prof, avail bool
	for _, f := range fs {
		if f.RuleID == "LH-FA-BEDROCK-001" && f.Severity == model.FAIL && f.Reason == model.AWSUnreachable {
			prof = true
		}
		if f.RuleID == "LH-FA-BEDROCK-004" && f.Severity == model.WARN {
			avail = true
		}
		if strings.Contains(f.Message, "secret") {
			t.Errorf("Fremdausgabe: %+v", f)
		}
	}
	if !prof || !avail {
		t.Fatalf("prof=%v avail=%v: %+v", prof, avail, fs)
	}
}
