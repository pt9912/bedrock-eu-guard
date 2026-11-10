package rules_test

import (
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

func TestClassify(t *testing.T) {
	cfg := model.DefaultConfig()
	for id, want := range map[string]model.ModelClass{
		"eu.anthropic.claude-sonnet-4-6": model.ModelAllowed,
		"anthropic.claude-sonnet-4-6":    model.ModelAllowed,
		"anthropic.claude-haiku-4-5":     model.ModelAllowed,
		"us.anthropic.claude-sonnet-4-6": model.ModelBlocked,
		"global.anthropic.claude-opus-4": model.ModelBlocked,
		"eu.anthropic.claude-us-east-1":  model.ModelBlocked, // Block vor Allow
		"arn:aws:bedrock:us-west-2:1:x":  model.ModelBlocked,
		"anthropic.claude-opus-4":        model.ModelUnclassified,
		"":                               model.ModelUnclassified,
	} {
		if got := rules.Classify(id, cfg); got != want {
			t.Errorf("Classify(%q)=%v want %v", id, got, want)
		}
	}
}

func TestClassifyEmptyAndInvalidPatterns(t *testing.T) {
	if got := rules.Classify("eu.x", model.Config{}); got != model.ModelUnclassified {
		t.Errorf("empty config: %v", got)
	}
	cfg := model.Config{AllowedModelPatterns: []string{"([", `^eu\.`}}
	if got := rules.Classify("eu.x", cfg); got != model.ModelAllowed {
		t.Errorf("invalid pattern must be skipped: %v", got)
	}
}

func TestClassifyCustomConfig(t *testing.T) {
	cfg := model.Config{AllowedModelPatterns: []string{`^custom\.`}, BlockedModelPatterns: []string{`bad`}}
	if rules.Classify("custom.bad", cfg) != model.ModelBlocked || rules.Classify("custom.ok", cfg) != model.ModelAllowed {
		t.Error("custom config not honoured")
	}
}

func TestDefaultModels(t *testing.T) {
	cfg := model.DefaultConfig()
	if fs := rules.CheckDefaultModels(fakeenv.Env{}, cfg); len(fs) != 0 {
		t.Fatalf("unset must yield none: %+v", fs)
	}
	env := fakeenv.Env{
		"ANTHROPIC_DEFAULT_SONNET_MODEL": "eu.anthropic.claude-sonnet-4-6",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "us.anthropic.claude-haiku-4-5",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   "anthropic.claude-opus-4",
	}
	fs := rules.CheckDefaultModels(env, cfg)
	if len(fs) != 3 {
		t.Fatalf("got %+v", fs)
	}
	want := []model.Severity{model.PASS, model.FAIL, model.WARN}
	for i, f := range fs {
		if f.Severity != want[i] || f.RuleID != rules.IDDefaultModels {
			t.Errorf("%d: %+v", i, f)
		}
	}
}

func TestClassifyRegionInID(t *testing.T) {
	cfg := model.DefaultConfig()
	for name, c := range map[string]struct {
		id   string
		want model.ModelClass
	}{
		"andere Region":                  {"anthropic.claude-opus-4-ap-southeast-1", model.ModelBlocked},
		"EU-Region nicht erlaubt":        {"eu.anthropic.claude-sonnet-4-6-eu-west-1", model.ModelBlocked},
		"erlaubte Region, Allow-Muster":  {"eu.anthropic.claude-eu-central-1", model.ModelAllowed},
		"erlaubte Region, unclassified":  {"anthropic.claude-opus-4-eu-central-1", model.ModelUnclassified},
		"ARN mit nicht erlaubter Region": {"arn:aws:bedrock:ap-northeast-1:123:inference-profile/eu.x", model.ModelBlocked},
		"ARN mit erlaubter Region":       {"arn:aws:bedrock:eu-central-1:123:inference-profile/eu.x", model.ModelAllowed}, // Tail eu.x trifft ^eu\.
		"ohne Region":                    {"anthropic.claude-sonnet-4-6", model.ModelAllowed},
		"Modellname ist keine Region":    {"anthropic.claude-opus-4-1", model.ModelUnclassified},
		"mehrere, eine nicht erlaubt":    {"eu.x-eu-central-1-us-gov-west-1", model.ModelBlocked},
		"Grossschreibung nicht erlaubt":  {"eu.x-AP-SOUTHEAST-1", model.ModelBlocked},
		"Grossschreibung erlaubt":        {"eu.x-EU-CENTRAL-1", model.ModelAllowed},
	} {
		t.Run(name, func(t *testing.T) {
			if got := rules.Classify(c.id, cfg); got != c.want {
				t.Errorf("Classify(%q)=%v want %v", c.id, got, c.want)
			}
		})
	}
}

func TestClassifyRegionHonoursAllowedRegions(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.AllowedRegions = []string{"eu-central-1", "eu-west-1"}
	if got := rules.Classify("eu.x-eu-west-1", cfg); got != model.ModelAllowed {
		t.Errorf("got %v", got)
	}
}

func TestDefaultModelsRegionMessage(t *testing.T) {
	env := fakeenv.Env{"ANTHROPIC_DEFAULT_OPUS_MODEL": "anthropic.claude-opus-4-ap-southeast-1"}
	fs := rules.CheckDefaultModels(env, model.DefaultConfig())
	if len(fs) != 1 || fs[0].Severity != model.FAIL || !strings.Contains(fs[0].Message, "ap-southeast-1") {
		t.Fatalf("got %+v", fs)
	}
}

// Reale Bedrock-IDs ohne Region dürfen nicht als Region erkannt werden.
func TestRegionDetectionNoFalsePositivesOnRealIDs(t *testing.T) {
	for _, id := range []string{
		"amazon.nova-pro-v1:0",
		"anthropic.claude-3-5-sonnet-20241022-v2:0",
		"meta.llama3-1-70b-instruct-v1:0",
		"anthropic.claude-sonnet-4-6",
		"eu.anthropic.claude-haiku-4-5",
	} {
		if r, bad := rules.DisallowedRegionInID(id, model.DefaultConfig()); bad {
			t.Errorf("%q: false positive region %q", id, r)
		}
	}
}

func TestClassifyARNTail(t *testing.T) {
	cfg := model.DefaultConfig()
	for id, want := range map[string]model.ModelClass{
		"arn:aws:bedrock:eu-central-1::foundation-model/anthropic.claude-sonnet-4-6":      model.ModelAllowed,
		"arn:aws:bedrock:eu-central-1:123456789012:inference-profile/eu.anthropic.claude": model.ModelAllowed,
		"arn:aws:bedrock:eu-central-1:123456789012:inference-profile/us.anthropic.claude": model.ModelBlocked,
		"arn:aws:bedrock:eu-central-1::foundation-model/global.anthropic.claude":          model.ModelBlocked,
		"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-sonnet-4-6":         model.ModelBlocked,
		"arn:aws:bedrock:eu-central-1::foundation-model/anthropic.claude-opus-4":          model.ModelUnclassified,
		"arn:aws:bedrock:eu-central-1::foundation-model/":                                 model.ModelUnclassified,
		"arn:aws:bedrock:eu-central-1::foundation-model":                                  model.ModelUnclassified,
		"foo/eu.anthropic.x": model.ModelUnclassified, // kein ARN: Tail-Regel gilt nicht
	} {
		if got := rules.Classify(id, cfg); got != want {
			t.Errorf("Classify(%q)=%v want %v", id, got, want)
		}
	}
}
