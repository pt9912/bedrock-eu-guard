package app_test

import (
	"errors"
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
