package rules_test

import (
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

func prof(ids ...string) []model.InferenceProfile {
	var out []model.InferenceProfile
	for _, id := range ids {
		out = append(out, model.InferenceProfile{ID: id})
	}
	return out
}

func mods(ids ...string) []model.FoundationModel {
	var out []model.FoundationModel
	for _, id := range ids {
		out = append(out, model.FoundationModel{ID: id})
	}
	return out
}

func bedrock(env fakeenv.Env, regions ...rules.RegionData) rules.AWSScan {
	return rules.AWSScan{Env: env, Bedrock: &rules.BedrockScan{Primary: "eu-central-1", Regions: regions}}
}

func TestQueryAndPrimaryRegion(t *testing.T) {
	cfg := model.Config{AllowedRegions: []string{"eu-central-1", "eu-west-1", "eu-central-1", ""}}
	if got := rules.QueryRegions(cfg); len(got) != 2 || got[0] != "eu-central-1" || got[1] != "eu-west-1" {
		t.Errorf("QueryRegions: %v", got)
	}
	if got := rules.PrimaryRegion(fakeenv.Env{"AWS_REGION": "eu-west-1"}, cfg); got != "eu-west-1" {
		t.Errorf("Primary mit AWS_REGION: %q", got)
	}
	if got := rules.PrimaryRegion(fakeenv.Env{"AWS_REGION": "us-east-1"}, cfg); got != "eu-central-1" {
		t.Errorf("Primary nicht erlaubt: %q", got)
	}
	if got := rules.PrimaryRegion(fakeenv.Env{}, model.Config{}); got != "" {
		t.Errorf("Primary leer: %q", got)
	}
}

func TestBedrockProfiles(t *testing.T) {
	cfg := model.DefaultConfig()
	if fs := rules.CheckBedrockProfiles(rules.AWSScan{Env: fakeenv.Env{}}, cfg); len(fs) != 0 {
		t.Errorf("ohne Reader: %+v", fs)
	}
	s := bedrock(fakeenv.Env{},
		rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.a", "eu.b")},
		rules.RegionData{Region: "eu-west-1"},
		rules.RegionData{Region: "eu-north-1", ProfilesErr: &model.AWSError{Op: "bedrock:ListInferenceProfiles", Reason: model.AWSAccessDenied}})
	fs := rules.CheckBedrockProfiles(s, cfg)
	if len(fs) != 3 || fs[0].Severity != model.PASS || fs[1].Severity != model.WARN || fs[1].Reason != model.ReasonProfilesNone ||
		fs[2].Severity != model.FAIL || fs[2].Reason != model.AWSAccessDenied || !strings.Contains(fs[2].Message, "eu-north-1") {
		t.Errorf("got %+v", fs)
	}
}

func TestBedrockBlocked(t *testing.T) {
	cfg := model.DefaultConfig()
	cases := []struct {
		name string
		env  fakeenv.Env
		rd   rules.RegionData
		want []model.Severity
	}{
		{"alles sauber", fakeenv.Env{"ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6"}, rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.anthropic.claude-sonnet-4-6")}, []model.Severity{model.PASS}},
		{"verwendet blockiert", fakeenv.Env{"ANTHROPIC_MODEL": "us.anthropic.claude-sonnet-4-6"}, rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.x")}, []model.Severity{model.FAIL}},
		{"nur sichtbar", fakeenv.Env{}, rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.x", "global.anthropic.claude-opus-4")}, []model.Severity{model.WARN}},
		{"verwendet und sichtbar: nur FAIL", fakeenv.Env{"ANTHROPIC_MODEL": "global.anthropic.claude-opus-4"}, rules.RegionData{Region: "eu-central-1", Profiles: prof("global.anthropic.claude-opus-4")}, []model.Severity{model.FAIL}},
		{"FAIL und WARN", fakeenv.Env{"ANTHROPIC_MODEL": "us.x"}, rules.RegionData{Region: "eu-central-1", Profiles: prof("global.y")}, []model.Severity{model.FAIL, model.WARN}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckBedrockBlocked(bedrock(c.env, c.rd), cfg)
			if len(fs) != len(c.want) {
				t.Fatalf("got %+v", fs)
			}
			for i, f := range fs {
				if f.Severity != c.want[i] || f.RuleID != rules.IDBedrockBlocked {
					t.Fatalf("got %+v", fs)
				}
			}
		})
	}
}

func TestBedrockAllowlist(t *testing.T) {
	cfg := model.DefaultConfig()
	if fs := rules.CheckBedrockAllowlist(bedrock(fakeenv.Env{}), cfg); len(fs) != 0 {
		t.Errorf("nichts verwendet: %+v", fs)
	}
	ok := rules.CheckBedrockAllowlist(bedrock(fakeenv.Env{"ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6", "ANTHROPIC_DEFAULT_HAIKU_MODEL": "anthropic.claude-haiku-4-5"}), cfg)
	if len(ok) != 1 || ok[0].Severity != model.PASS {
		t.Errorf("abgedeckt (Profil-Muster bzw. SPEC-010 allowed): %+v", ok)
	}
	bad := rules.CheckBedrockAllowlist(bedrock(fakeenv.Env{"ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6", "ANTHROPIC_DEFAULT_OPUS_MODEL": "anthropic.claude-opus-4"}), cfg)
	if len(bad) != 1 || bad[0].Severity != model.FAIL || bad[0].Subject != "ANTHROPIC_DEFAULT_OPUS_MODEL" {
		t.Errorf("nicht abgedeckt: %+v", bad)
	}
	// Muster ist anchored: "xeu.foo" darf nicht auf "eu.*" passen.
	if fs := rules.CheckBedrockAllowlist(bedrock(fakeenv.Env{"ANTHROPIC_MODEL": "xeu.foo"}), cfg); len(fs) != 1 || fs[0].Severity != model.FAIL {
		t.Errorf("anchored: %+v", fs)
	}
	custom := cfg
	custom.AllowedProfiles = []string{"my-profile-.*"}
	if fs := rules.CheckBedrockAllowlist(bedrock(fakeenv.Env{"ANTHROPIC_MODEL": "my-profile-1"}), custom); len(fs) != 1 || fs[0].Severity != model.PASS {
		t.Errorf("eigene Allowlist: %+v", fs)
	}
}

func TestBedrockAvailability(t *testing.T) {
	cfg := model.DefaultConfig()
	region := func(models, profiles []string) rules.RegionData {
		return rules.RegionData{Region: "eu-central-1", Models: mods(models...), Profiles: prof(profiles...)}
	}
	cases := []struct {
		name string
		env  fakeenv.Env
		rd   rules.RegionData
		none bool
		sev  model.Severity
		rsn  string
	}{
		{"Basismodell verfügbar", fakeenv.Env{"ANTHROPIC_MODEL": "anthropic.claude-sonnet-4-6"}, region([]string{"anthropic.claude-sonnet-4-6"}, nil), false, model.PASS, model.ReasonModelAvailable},
		{"Profil mit Basismodell in Region", fakeenv.Env{"ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6"}, region([]string{"anthropic.claude-sonnet-4-6"}, nil), false, model.PASS, model.ReasonModelAvailable},
		{"Profil direkt gelistet", fakeenv.Env{"ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6"}, region(nil, []string{"eu.anthropic.claude-sonnet-4-6"}), false, model.PASS, model.ReasonModelAvailable},
		{"nicht verfügbar", fakeenv.Env{"ANTHROPIC_MODEL": "anthropic.claude-opus-4"}, region([]string{"anthropic.claude-haiku-4-5"}, []string{"eu.x"}), false, model.FAIL, model.ReasonModelUnavailable},
		{"ARN: unklar", fakeenv.Env{"ANTHROPIC_MODEL": "arn:aws:bedrock:eu-central-1::foundation-model/x"}, region([]string{"x"}, nil), false, model.WARN, model.ReasonModelUnknown},
		{"Listenfehler: unklar", fakeenv.Env{"ANTHROPIC_MODEL": "anthropic.claude-opus-4"}, rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.x"), ModelsErr: &model.AWSError{Reason: model.AWSAccessDenied}}, false, model.WARN, model.ReasonModelUnknown},
		{"beide Listen fehlerhaft: unklar", fakeenv.Env{"ANTHROPIC_MODEL": "anthropic.claude-opus-4"}, rules.RegionData{Region: "eu-central-1", ModelsErr: &model.AWSError{Reason: model.AWSUnreachable}, ProfilesErr: &model.AWSError{Reason: model.AWSUnreachable}}, false, model.WARN, model.ReasonModelUnknown},
		{"nichts konfiguriert", fakeenv.Env{}, region(nil, nil), true, 0, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckBedrockAvailability(bedrock(c.env, c.rd), cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("got %+v", fs)
				}
				return
			}
			f := sole(t, fs)
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDBedrockAvailability {
				t.Fatalf("got %+v", f)
			}
		})
	}
	s := bedrock(fakeenv.Env{"ANTHROPIC_MODEL": "x"})
	s.Bedrock.Primary = "eu-west-1" // keine Daten zur Primärregion
	if f := sole(t, rules.CheckBedrockAvailability(s, cfg)); f.Severity != model.WARN {
		t.Errorf("Primärregion ohne Daten: %+v", f)
	}
}

func withSettings(s rules.AWSScan, ms ...rules.SettingsModel) rules.AWSScan {
	s.Settings = ms
	return s
}

func TestBedrockUsedFromSettings(t *testing.T) {
	cfg := model.DefaultConfig()
	rd := rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.anthropic.claude-sonnet-4-6"), Models: mods("anthropic.claude-sonnet-4-6")}
	blocked := rules.SettingsModel{Source: ".claude/settings.json", Var: "model", ID: "us.anthropic.claude-opus-4"}

	// BEDROCK-002: blockiertes Modell aus den Settings → FAIL mit Herkunft
	fs := rules.CheckBedrockBlocked(withSettings(bedrock(fakeenv.Env{}, rd), blocked), cfg)
	if len(fs) != 1 || fs[0].Severity != model.FAIL || fs[0].File != ".claude/settings.json" || !strings.Contains(fs[0].Message, "in .claude/settings.json") {
		t.Errorf("002: %+v", fs)
	}
	// BEDROCK-003: nicht abgedeckt → FAIL; abgedeckt → PASS
	fs = rules.CheckBedrockAllowlist(withSettings(bedrock(fakeenv.Env{}, rd), rules.SettingsModel{Source: "~/.claude/settings.json", Var: "model", ID: "anthropic.claude-opus-4"}), cfg)
	if len(fs) != 1 || fs[0].Severity != model.FAIL || fs[0].File != "~/.claude/settings.json" {
		t.Errorf("003: %+v", fs)
	}
	fs = rules.CheckBedrockAllowlist(withSettings(bedrock(fakeenv.Env{}, rd), rules.SettingsModel{Source: "s", Var: "model", ID: "eu.anthropic.claude-sonnet-4-6"}), cfg)
	if len(fs) != 1 || fs[0].Severity != model.PASS {
		t.Errorf("003 pass: %+v", fs)
	}
	// BEDROCK-004: Verfügbarkeit für Settings-Modell
	fs = rules.CheckBedrockAvailability(withSettings(bedrock(fakeenv.Env{}, rd), rules.SettingsModel{Source: "s", Var: "model", ID: "eu.anthropic.claude-sonnet-4-6"}), cfg)
	if len(fs) != 1 || fs[0].Severity != model.PASS || fs[0].File != "s" {
		t.Errorf("004: %+v", fs)
	}
	// Env und Settings: gleiche Variable und ID nur einmal; andere ID zusätzlich
	env := fakeenv.Env{"ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6"}
	s := withSettings(bedrock(env, rd),
		rules.SettingsModel{Source: "s", Var: "ANTHROPIC_MODEL", ID: "eu.anthropic.claude-sonnet-4-6"},
		rules.SettingsModel{Source: "s", Var: "ANTHROPIC_MODEL", ID: "anthropic.claude-opus-4"})
	fs = rules.CheckBedrockAllowlist(s, cfg)
	if len(fs) != 1 || fs[0].Severity != model.FAIL || !strings.Contains(fs[0].Message, "anthropic.claude-opus-4") {
		t.Errorf("Dedup/Zusatz: %+v", fs)
	}
}

func TestVisibilityRegions(t *testing.T) {
	cfg := model.Config{AllowedRegions: []string{"eu-central-1"}, VisibilityRegions: []string{"us-east-1", "us-east-1", "", "eu-central-1", "us-west-2"}}
	got := rules.VisibilityRegions(cfg)
	if len(got) != 2 || got[0] != "us-east-1" || got[1] != "us-west-2" {
		t.Errorf("got %v", got)
	}
	if rules.VisibilityRegions(model.DefaultConfig()) != nil {
		t.Error("Default: keine Zusatzregionen")
	}
}

func TestBedrockVisibilityRegions(t *testing.T) {
	cfg := model.DefaultConfig()
	allowed := rules.RegionData{Region: "eu-central-1", Profiles: prof("eu.anthropic.claude-sonnet-4-6")}
	vis := func(profiles []model.InferenceProfile, err *model.AWSError) rules.RegionData {
		return rules.RegionData{Region: "us-east-1", VisibilityOnly: true, Profiles: profiles, ProfilesErr: err}
	}
	// BEDROCK-001 betrifft Zusatzregionen nicht
	s := bedrock(fakeenv.Env{}, allowed, vis(prof("us.anthropic.claude-opus-4"), nil))
	if fs := rules.CheckBedrockProfiles(s, cfg); len(fs) != 1 || fs[0].Subject != "eu-central-1" {
		t.Errorf("001 darf Zusatzregionen nicht melden: %+v", fs)
	}
	// BEDROCK-002: sichtbares blockiertes Profil → WARN mit Region, nie FAIL
	fs := rules.CheckBedrockBlocked(s, cfg)
	if len(fs) != 1 || fs[0].Severity != model.WARN || fs[0].Reason != model.ReasonProfilesVisible || !strings.Contains(fs[0].Message, "us-east-1") {
		t.Errorf("002 sichtbar: %+v", fs)
	}
	// nur erlaubte Profile in der Zusatzregion → PASS
	fs = rules.CheckBedrockBlocked(bedrock(fakeenv.Env{}, allowed, vis(prof("eu.x"), nil)), cfg)
	if len(fs) != 1 || fs[0].Severity != model.PASS {
		t.Errorf("002 nur erlaubte: %+v", fs)
	}
	// nicht lesbare Zusatzregion → WARN VISIBILITY_UNKNOWN (kein FAIL), mit Region
	fs = rules.CheckBedrockBlocked(bedrock(fakeenv.Env{}, allowed, vis(nil, &model.AWSError{Op: "bedrock:ListInferenceProfiles", Reason: model.AWSAccessDenied})), cfg)
	if len(fs) != 1 || fs[0].Severity != model.WARN || fs[0].Reason != model.ReasonVisibilityUnknown || fs[0].Subject != "us-east-1" {
		t.Errorf("002 unbekannt: %+v", fs)
	}
	// verwendet (FAIL) + sichtbar in Zusatzregion (WARN): beides
	env := fakeenv.Env{"ANTHROPIC_MODEL": "us.anthropic.claude-opus-4"}
	fs = rules.CheckBedrockBlocked(bedrock(env, allowed, vis(prof("global.anthropic.claude-sonnet-4-6"), nil)), cfg)
	if len(fs) != 2 || fs[0].Severity != model.FAIL || fs[1].Severity != model.WARN {
		t.Errorf("002 verwendet+sichtbar: %+v", fs)
	}
	// Verfügbarkeit (004) nutzt weiter die Primärregion, nicht die Zusatzregion
	allowedWithModel := rules.RegionData{Region: "eu-central-1", Models: mods("anthropic.claude-sonnet-4-6")}
	s2 := bedrock(fakeenv.Env{"ANTHROPIC_MODEL": "anthropic.claude-sonnet-4-6"}, vis(prof("x"), nil), allowedWithModel)
	if fs := rules.CheckBedrockAvailability(s2, cfg); len(fs) != 1 || fs[0].Severity != model.PASS {
		t.Errorf("004: %+v", fs)
	}
}
