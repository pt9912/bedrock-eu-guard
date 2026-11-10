package rules_test

import (
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
