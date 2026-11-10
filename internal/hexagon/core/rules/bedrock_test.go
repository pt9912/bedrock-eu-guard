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
