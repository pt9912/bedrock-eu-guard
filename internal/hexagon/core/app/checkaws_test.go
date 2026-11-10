package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeaws"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakesettings"
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

func allDecisions(denied ...string) []model.ActionDecision {
	d := map[string]bool{}
	for _, a := range denied {
		d[a] = true
	}
	var out []model.ActionDecision
	for _, a := range []string{"bedrock:InvokeModel", "bedrock:InvokeModelWithResponseStream", "bedrock:Converse", "bedrock:ConverseStream", "bedrock:ListInferenceProfiles", "bedrock:GetInferenceProfile"} {
		out = append(out, model.ActionDecision{Action: a, Allowed: !d[a]})
	}
	return out
}

func TestCheckAWSWithIAM(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "1", ARN: "arn:aws:sts::1:assumed-role/r/s"}}
	iam := fakeaws.IAM{
		Decisions: allDecisions(),
		Statements: []model.IAMStatement{{Effect: "Allow", Actions: []string{"bedrock:InvokeModel"}, Resources: []string{"*"},
			HasRegionCondition: true, ConditionRegions: []string{"eu-central-1"}}},
	}
	fs, err := app.CheckAWS{AWS: r, IAM: iam}.Run(fakeenv.Env{"AWS_PROFILE": "p"}, model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]model.Severity{}
	for _, f := range fs {
		have[f.RuleID] = f.Severity
	}
	if have["LH-FA-IAM-001"] != model.PASS || have["LH-FA-IAM-003"] != model.PASS {
		t.Fatalf("got %+v", fs)
	}
	if _, ok := have["LH-FA-IAM-002"]; !ok {
		t.Fatalf("IAM-002 fehlt: %+v", fs)
	}
}

func TestCheckAWSIAMFailuresAndSkips(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "1", ARN: "arn"}}
	// Reader nil → keine IAM-Regeln
	fs, _ := app.CheckAWS{AWS: r}.Run(fakeenv.Env{}, model.DefaultConfig())
	for _, f := range fs {
		if strings.HasPrefix(f.RuleID, "LH-FA-IAM") {
			t.Errorf("IAM-Regel ohne Reader: %+v", f)
		}
	}
	// Fehler: verweigert → WARN, unbekannter Fehler → FAIL ohne Fremdtext
	iam := fakeaws.IAM{
		DecisionsErr:  &model.AWSError{Op: "iam:SimulatePrincipalPolicy", Reason: model.AWSAccessDenied},
		StatementsErr: errors.New("raw secret output"),
	}
	fs, err := app.CheckAWS{AWS: r, IAM: iam}.Run(fakeenv.Env{}, model.DefaultConfig())
	if err != nil {
		t.Fatalf("IAM-Fehler sind keine Tool-Fehler: %v", err)
	}
	var warn, fail bool
	for _, f := range fs {
		if f.RuleID == "LH-FA-IAM-001" && f.Severity == model.WARN {
			warn = true
		}
		if f.RuleID == "LH-FA-IAM-002" && f.Severity == model.FAIL && f.Reason == model.AWSUnreachable {
			fail = true
		}
		if strings.Contains(f.Message, "secret") {
			t.Errorf("Fremdausgabe: %+v", f)
		}
	}
	if !warn || !fail {
		t.Fatalf("warn=%v fail=%v: %+v", warn, fail, fs)
	}
}

func TestCheckAWSIAMNoIdentity(t *testing.T) {
	r := fakeaws.Reader{Err: &model.AWSError{Op: "sts:GetCallerIdentity", Reason: model.AWSNoCredentials}}
	fs, err := app.CheckAWS{AWS: r, IAM: fakeaws.IAM{Decisions: allDecisions()}}.Run(fakeenv.Env{}, model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, f := range fs {
		if strings.HasPrefix(f.RuleID, "LH-FA-IAM") && f.Severity == model.FAIL && f.Reason == model.AWSNoCredentials {
			n++
		}
	}
	if n != 3 {
		t.Fatalf("3 FAIL (IAM-001..003) erwartet, got %d: %+v", n, fs)
	}
}

func TestCheckAWSUsesSettingsModels(t *testing.T) {
	r := fakeaws.Reader{Ident: &model.AWSIdentity{Account: "1", ARN: "arn"}}
	st := fakesettings.Reader{
		User: &model.ClaudeSettings{Model: "eu.anthropic.claude-sonnet-4-6",
			Env: []model.EnvEntry{{Key: "ANTHROPIC_API_KEY", Secret: true}, {Key: "ANTHROPIC_DEFAULT_HAIKU_MODEL", Value: "us.anthropic.claude-haiku-4-5"}}},
		Project:    map[string]*model.ClaudeSettings{".claude/settings.json": {Model: "global.anthropic.claude-opus-4"}},
		ProjectErr: map[string]error{".claude/settings.local.json": errors.New("kaputt")},
	}
	run := func(cfg model.Config) []model.Finding {
		fs, err := app.CheckAWS{AWS: r, Bedrock: bedrockOK(), Settings: st}.Run(fakeenv.Env{}, cfg)
		if err != nil {
			t.Fatal(err)
		}
		return fs
	}
	blocked := map[string]bool{}
	for _, f := range run(model.DefaultConfig()) {
		if f.RuleID == "LH-FA-BEDROCK-002" && f.Severity == model.FAIL {
			blocked[f.Subject+"|"+f.File] = true
		}
	}
	if !blocked["ANTHROPIC_DEFAULT_HAIKU_MODEL|~/.claude/settings.json"] || !blocked["model|.claude/settings.json"] || len(blocked) != 2 {
		t.Fatalf("got %v", blocked)
	}
	// ausgeschlossene Projekt-Settings zählen nicht
	cfg := model.DefaultConfig()
	cfg.Exclude = []string{".claude"}
	for _, f := range run(cfg) {
		if f.File == ".claude/settings.json" {
			t.Errorf("ausgeschlossen, aber gemeldet: %+v", f)
		}
	}
	// ohne Settings-Reader: Verhalten wie bisher (nur Umgebung)
	fs, _ := app.CheckAWS{AWS: r, Bedrock: bedrockOK()}.Run(fakeenv.Env{}, model.DefaultConfig())
	for _, f := range fs {
		if f.RuleID == "LH-FA-BEDROCK-002" && f.Severity == model.FAIL {
			t.Errorf("ohne Reader kein FAIL: %+v", f)
		}
	}
}
