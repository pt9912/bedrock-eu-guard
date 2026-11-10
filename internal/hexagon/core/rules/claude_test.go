package rules_test

import (
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
)

func settings(env ...model.EnvEntry) *model.ClaudeSettings {
	return &model.ClaudeSettings{Env: env}
}

func e(k, v string, line int) model.EnvEntry { return model.EnvEntry{Key: k, Value: v, Line: line} }
func secret(k string, line int) model.EnvEntry {
	return model.EnvEntry{Key: k, Secret: true, Line: line}
}

func user(s *model.ClaudeSettings) rules.ScanContext {
	return rules.ScanContext{User: rules.SettingsFile{Path: rules.UserSettingsPath, S: s}}
}

func sole(t *testing.T, fs []model.Finding) model.Finding {
	t.Helper()
	if len(fs) != 1 {
		t.Fatalf("want 1 finding, got %+v", fs)
	}
	return fs[0]
}

func TestClaudeUser(t *testing.T) {
	cfg := model.DefaultConfig()
	bedrock := e("CLAUDE_CODE_USE_BEDROCK", "1", 2)
	cases := []struct {
		name string
		sc   rules.ScanContext
		sev  model.Severity
		rsn  string
		line int
	}{
		{"fehlt", rules.ScanContext{User: rules.SettingsFile{Path: rules.UserSettingsPath}}, model.WARN, model.ReasonFileMissing, 0},
		{"ungueltig", rules.ScanContext{User: rules.SettingsFile{Path: rules.UserSettingsPath,
			Err: &model.FileError{Path: "x", Reason: model.ReasonFileInvalid, Line: 7}}}, model.FAIL, model.ReasonFileInvalid, 7},
		{"nicht lesbar", rules.ScanContext{User: rules.SettingsFile{Path: rules.UserSettingsPath,
			Err: &model.FileError{Path: "x", Reason: model.ReasonFileUnreadable}}}, model.FAIL, model.ReasonFileUnreadable, 0},
		{"bedrock ok", user(settings(bedrock)), model.PASS, model.ReasonOK, 0},
		{"bedrock aus", user(settings(e("CLAUDE_CODE_USE_BEDROCK", "0", 2))), model.WARN, model.ReasonBedrockNotEnabled, 0},
		{"leer", user(settings()), model.WARN, model.ReasonBedrockNotEnabled, 0},
		{"api key", user(settings(bedrock, secret("ANTHROPIC_API_KEY", 5))), model.FAIL, model.ReasonDirectAnthropic, 5},
		{"auth token", user(settings(bedrock, secret("ANTHROPIC_AUTH_TOKEN", 6))), model.FAIL, model.ReasonDirectAnthropic, 6},
		{"base url anthropic", user(settings(bedrock, e("ANTHROPIC_BASE_URL", "https://API.anthropic.com/v1", 4))), model.FAIL, model.ReasonDirectAnthropic, 4},
		{"host hit", user(&model.ClaudeSettings{Env: []model.EnvEntry{bedrock}, Hits: []model.Hit{{Pattern: model.HitAnthropicHost, Line: 9}}}), model.FAIL, model.ReasonDirectAnthropic, 9},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := sole(t, rules.CheckClaudeUser(c.sc, cfg))
			if f.Severity != c.sev || f.Reason != c.rsn || f.Line != c.line || f.RuleID != rules.IDClaudeUser || f.File != rules.UserSettingsPath {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func proj(rel string, s *model.ClaudeSettings) rules.SettingsFile {
	return rules.SettingsFile{Path: rel, S: s}
}

func TestClaudeProject(t *testing.T) {
	cfg := model.DefaultConfig()
	f1, f2 := ".claude/settings.json", ".claude/settings.local.json"
	t.Run("fehlende Dateien: kein Finding", func(t *testing.T) {
		sc := rules.ScanContext{Projects: []rules.SettingsFile{{Path: f1}, {Path: f2}}}
		if fs := rules.CheckClaudeProject(sc, cfg); len(fs) != 0 {
			t.Fatalf("got %+v", fs)
		}
	})
	t.Run("ungueltig: FAIL, andere Datei wird weiter geprueft", func(t *testing.T) {
		sc := rules.ScanContext{Projects: []rules.SettingsFile{
			{Path: f1, Err: &model.FileError{Path: f1, Reason: model.ReasonFileInvalid, Line: 3}},
			proj(f2, settings(e("AWS_REGION", "eu-central-1", 2)))}}
		fs := rules.CheckClaudeProject(sc, cfg)
		if len(fs) != 2 || fs[0].Reason != model.ReasonFileInvalid || fs[0].Line != 3 || fs[1].Severity != model.PASS || fs[1].File != f2 {
			t.Fatalf("got %+v", fs)
		}
	})
	cases := []struct {
		name string
		s    *model.ClaudeSettings
		sev  model.Severity
		line int
	}{
		{"eu-safe", &model.ClaudeSettings{Model: "eu.anthropic.claude-sonnet-4-6", Env: []model.EnvEntry{e("AWS_REGION", "eu-central-1", 3)}}, model.PASS, 0},
		{"leer", settings(), model.PASS, 0},
		{"blockiertes Modell", &model.ClaudeSettings{Model: "us.anthropic.claude-sonnet-4-6"}, model.FAIL, 0},
		{"Modell per env", settings(e("ANTHROPIC_DEFAULT_OPUS_MODEL", "global.x", 4)), model.FAIL, 0},
		{"unklassifiziert", &model.ClaudeSettings{Model: "anthropic.claude-opus-4"}, model.WARN, 0},
		{"Region nicht erlaubt", settings(e("AWS_REGION", "us-east-1", 6)), model.FAIL, 6},
		{"api key", settings(secret("ANTHROPIC_API_KEY", 8)), model.FAIL, 8},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sc := rules.ScanContext{Projects: []rules.SettingsFile{proj(f1, c.s)}}
			f := sole(t, rules.CheckClaudeProject(sc, cfg))
			if f.Severity != c.sev || f.Line != c.line || f.File != f1 || f.RuleID != rules.IDClaudeProject {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestClaudeEndpoints(t *testing.T) {
	cfg := model.DefaultConfig()
	f1 := ".claude/settings.json"
	cases := []struct {
		name string
		s    *model.ClaudeSettings
		sev  model.Severity
		rsn  string
	}{
		{"sauber", settings(), model.PASS, model.ReasonOK},
		{"api key", settings(secret("ANTHROPIC_API_KEY", 2)), model.FAIL, model.ReasonDirectAnthropic},
		{"host", &model.ClaudeSettings{Hits: []model.Hit{{Pattern: model.HitAnthropicHost, Line: 4}}}, model.FAIL, model.ReasonDirectAnthropic},
		{"fremder Host", settings(e("ANTHROPIC_BASE_URL", "https://proxy.example.eu", 3)), model.WARN, model.ReasonEndpointForeign},
		{"nur Text API_KEY", &model.ClaudeSettings{Hits: []model.Hit{{Pattern: model.HitAPIKey, Line: 5}}}, model.WARN, model.ReasonEndpointTextOnly},
		{"nur Text BASE_URL", &model.ClaudeSettings{Hits: []model.Hit{{Pattern: model.HitBaseURL, Line: 5}}}, model.WARN, model.ReasonEndpointTextOnly},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sc := rules.ScanContext{Projects: []rules.SettingsFile{proj(f1, c.s)}}
			f := sole(t, rules.CheckClaudeEndpoints(sc, cfg))
			if f.Severity != c.sev || f.Reason != c.rsn || f.File != f1 {
				t.Fatalf("got %+v", f)
			}
		})
	}
	t.Run("fehlend/kaputt: kein Finding, User wird mitgeprueft", func(t *testing.T) {
		sc := rules.ScanContext{
			User:     rules.SettingsFile{Path: rules.UserSettingsPath, S: settings(secret("ANTHROPIC_AUTH_TOKEN", 1))},
			Projects: []rules.SettingsFile{{Path: f1}, {Path: ".claude/settings.local.json", Err: &model.FileError{Reason: model.ReasonFileInvalid}}}}
		f := sole(t, rules.CheckClaudeEndpoints(sc, cfg))
		if f.File != rules.UserSettingsPath || f.Severity != model.FAIL {
			t.Fatalf("got %+v", f)
		}
	})
}

func TestClaudeWizard(t *testing.T) {
	cfg := model.DefaultConfig()
	bedrock := e("CLAUDE_CODE_USE_BEDROCK", "1", 2)
	cases := []struct {
		name string
		s    *model.ClaudeSettings
		none bool
		sev  model.Severity
		rsn  string
	}{
		{"ok", settings(bedrock, e("AWS_REGION", "eu-central-1", 3)), false, model.PASS, model.ReasonOK},
		{"Region fehlt", settings(bedrock), false, model.FAIL, model.ReasonRegionMissing},
		{"Region leer", settings(bedrock, e("AWS_REGION", "", 3)), false, model.FAIL, model.ReasonRegionMissing},
		{"us-east-1", settings(bedrock, e("AWS_REGION", "us-east-1", 3)), false, model.FAIL, model.ReasonRegionNotAllowed},
		{"parallel Anthropic", settings(bedrock, e("AWS_REGION", "eu-central-1", 3), secret("ANTHROPIC_API_KEY", 4)), false, model.FAIL, model.ReasonDirectAnthropic},
		{"kein Bedrock: nicht zustaendig", settings(e("AWS_REGION", "us-east-1", 3)), true, 0, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckClaudeWizard(user(c.s), cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("got %+v", fs)
				}
				return
			}
			if f := sole(t, fs); f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDClaudeWizard {
				t.Fatalf("got %+v", f)
			}
		})
	}
	if fs := rules.CheckClaudeWizard(rules.ScanContext{User: rules.SettingsFile{Path: "x"}}, cfg); len(fs) != 0 {
		t.Fatalf("fehlende Datei: %+v", fs)
	}
}

func TestClaudeNoSecretOrContentInFindings(t *testing.T) {
	cfg := model.DefaultConfig()
	s := &model.ClaudeSettings{Env: []model.EnvEntry{
		{Key: "ANTHROPIC_API_KEY", Secret: true, Value: "", Line: 1},
		e("CLAUDE_CODE_USE_BEDROCK", "1", 2)}}
	sc := user(s)
	sc.Projects = []rules.SettingsFile{proj(".claude/settings.json", s)}
	var all []model.Finding
	for _, r := range rules.ScanRules() {
		all = append(all, r(sc, cfg)...)
	}
	for _, f := range all {
		if strings.Contains(f.Message+f.Recommendation, "sk-") {
			t.Fatalf("leak: %+v", f)
		}
	}
}
