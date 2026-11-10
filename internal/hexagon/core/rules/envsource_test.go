package rules_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
)

func envFile(kind model.EnvSourceKind, env ...model.EnvEntry) rules.EnvSourceFile {
	return rules.EnvSourceFile{Path: "f", F: &model.EnvFile{Path: "f", Kind: kind, Env: env}}
}

func TestDevcontainer(t *testing.T) {
	cfg := model.DefaultConfig()
	bedrock := e("CLAUDE_CODE_USE_BEDROCK", "1", 3)
	region := e("AWS_REGION", "eu-central-1", 4)
	cases := []struct {
		name string
		f    rules.EnvSourceFile
		sev  model.Severity
		rsn  string
		line int
	}{
		{"pass", envFile(model.EnvDevcontainer, bedrock, region), model.PASS, model.ReasonOK, 0},
		{"api key", envFile(model.EnvDevcontainer, bedrock, region, secret("ANTHROPIC_API_KEY", 9)), model.FAIL, model.ReasonDirectAnthropic, 9},
		{"region falsch", envFile(model.EnvDevcontainer, bedrock, e("AWS_REGION", "us-east-1", 5)), model.FAIL, model.ReasonRegionNotAllowed, 5},
		{"nichts", envFile(model.EnvDevcontainer, e("FOO", "bar", 2)), model.WARN, model.ReasonNoClaudeConfig, 0},
		{"leer", envFile(model.EnvDevcontainer), model.WARN, model.ReasonNoClaudeConfig, 0},
		{"nur Region", envFile(model.EnvDevcontainer, region), model.WARN, model.ReasonEnvIncomplete, 0},
		{"nur Bedrock", envFile(model.EnvDevcontainer, bedrock), model.WARN, model.ReasonEnvIncomplete, 0},
		{"Bedrock=0", envFile(model.EnvDevcontainer, e("CLAUDE_CODE_USE_BEDROCK", "0", 3), region), model.WARN, model.ReasonEnvIncomplete, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := sole(t, rules.CheckDevcontainer(rules.ScanContext{Devcontainers: []rules.EnvSourceFile{c.f}}, cfg))
			if f.Severity != c.sev || f.Reason != c.rsn || f.Line != c.line || f.RuleID != rules.IDDevcontainer {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestCompose(t *testing.T) {
	cfg := model.DefaultConfig()
	svc := func(scope string, en model.EnvEntry) model.EnvEntry { en.Scope = scope; return en }
	cases := []struct {
		name string
		f    rules.EnvSourceFile
		none bool
		sev  model.Severity
	}{
		{"nur erlaubt", envFile(model.EnvCompose, svc("web", e("AWS_REGION", "eu-central-1", 4)), svc("db", e("AWS_DEFAULT_REGION", "eu-central-1", 8))), false, model.PASS},
		{"ein Service falsch", envFile(model.EnvCompose, svc("web", e("AWS_REGION", "eu-central-1", 4)), svc("db", e("AWS_REGION", "us-west-2", 8))), false, model.FAIL},
		{"api key", envFile(model.EnvCompose, svc("web", secret("ANTHROPIC_API_KEY", 6))), false, model.FAIL},
		{"keine relevanten Variablen", envFile(model.EnvCompose, e("TZ", "UTC", 2)), true, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckCompose(rules.ScanContext{Composes: []rules.EnvSourceFile{c.f}}, cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("got %+v", fs)
				}
				return
			}
			if f := sole(t, fs); f.Severity != c.sev || f.RuleID != rules.IDCompose {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestDockerfile(t *testing.T) {
	cfg := model.DefaultConfig()
	cases := []struct {
		name string
		f    rules.EnvSourceFile
		sev  model.Severity
		rsn  string
	}{
		{"region ok", envFile(model.EnvDockerfile, e("AWS_REGION", "eu-central-1", 3)), model.PASS, model.ReasonOK},
		{"keine Region", envFile(model.EnvDockerfile, e("FOO", "x", 2)), model.WARN, model.ReasonRegionNotSetEnv},
		{"region falsch", envFile(model.EnvDockerfile, e("AWS_REGION", "ap-southeast-1", 3)), model.FAIL, model.ReasonRegionNotAllowed},
		{"api key", envFile(model.EnvDockerfile, e("AWS_REGION", "eu-central-1", 3), secret("ANTHROPIC_API_KEY", 4)), model.FAIL, model.ReasonDirectAnthropic},
		{"base url anthropic", envFile(model.EnvDockerfile, e("ANTHROPIC_BASE_URL", "https://api.anthropic.com", 5)), model.FAIL, model.ReasonDirectAnthropic},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := sole(t, rules.CheckDockerfile(rules.ScanContext{Dockerfiles: []rules.EnvSourceFile{c.f}}, cfg))
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDDockerfile {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestEnvSourceMissingAndBroken(t *testing.T) {
	cfg := model.DefaultConfig()
	sc := rules.ScanContext{
		Devcontainers: []rules.EnvSourceFile{{Path: "a"}},
		Composes:      []rules.EnvSourceFile{{Path: "b", Err: &model.FileError{Path: "b", Reason: model.ReasonFileInvalid, Line: 5}}},
	}
	if fs := rules.CheckDevcontainer(sc, cfg); len(fs) != 0 {
		t.Errorf("fehlend: %+v", fs)
	}
	f := sole(t, rules.CheckCompose(sc, cfg))
	if f.Severity != model.FAIL || f.Reason != model.ReasonFileInvalid || f.Line != 5 || f.RuleID != rules.IDCompose {
		t.Errorf("kaputt: %+v", f)
	}
}
