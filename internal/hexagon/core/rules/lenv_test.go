package rules_test

import (
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

func one(t *testing.T, fs []model.Finding) model.Finding {
	t.Helper()
	if len(fs) != 1 {
		t.Fatalf("want 1 finding, got %d: %+v", len(fs), fs)
	}
	return fs[0]
}

func TestAWSRegion(t *testing.T) {
	cfg := model.DefaultConfig()
	cases := []struct {
		name string
		env  fakeenv.Env
		sev  model.Severity
		rsn  string
		none bool
	}{
		{"allowed", fakeenv.Env{"AWS_REGION": "eu-central-1"}, model.PASS, model.ReasonOK, false},
		{"missing", fakeenv.Env{}, model.FAIL, model.ReasonRegionMissing, false},
		{"empty", fakeenv.Env{"AWS_REGION": ""}, model.FAIL, model.ReasonRegionMissing, false},
		{"not allowed", fakeenv.Env{"AWS_REGION": "us-east-1"}, model.FAIL, model.ReasonRegionNotAllowed, false},
		{"defers to 002", fakeenv.Env{"AWS_DEFAULT_REGION": "eu-central-1"}, 0, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckAWSRegion(c.env, cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("want none, got %+v", fs)
				}
				return
			}
			f := one(t, fs)
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDRegion {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestAWSRegionCustomConfig(t *testing.T) {
	cfg := model.Config{AllowedRegions: []string{"eu-west-1"}}
	f := one(t, rules.CheckAWSRegion(fakeenv.Env{"AWS_REGION": "eu-west-1"}, cfg))
	if f.Severity != model.PASS {
		t.Fatalf("got %+v", f)
	}
}

func TestAWSDefaultRegion(t *testing.T) {
	cfg := model.DefaultConfig()
	cases := []struct {
		name string
		env  fakeenv.Env
		sev  model.Severity
		rsn  string
		none bool
	}{
		{"fallback allowed", fakeenv.Env{"AWS_DEFAULT_REGION": "eu-central-1"}, model.WARN, model.ReasonRegionFallback, false},
		{"fallback not allowed", fakeenv.Env{"AWS_DEFAULT_REGION": "us-west-2"}, model.FAIL, model.ReasonRegionNotAllowed, false},
		{"both missing", fakeenv.Env{}, 0, "", true},
		{"region set", fakeenv.Env{"AWS_REGION": "eu-central-1", "AWS_DEFAULT_REGION": "us-west-2"}, 0, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := rules.CheckAWSDefaultRegion(c.env, cfg)
			if c.none {
				if len(fs) != 0 {
					t.Fatalf("want none, got %+v", fs)
				}
				return
			}
			f := one(t, fs)
			if f.Severity != c.sev || f.Reason != c.rsn || f.RuleID != rules.IDDefaultRegion {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestUseBedrock(t *testing.T) {
	cfg := model.DefaultConfig()
	for name, c := range map[string]struct {
		env fakeenv.Env
		sev model.Severity
	}{
		"one":     {fakeenv.Env{"CLAUDE_CODE_USE_BEDROCK": "1"}, model.PASS},
		"missing": {fakeenv.Env{}, model.FAIL},
		"zero":    {fakeenv.Env{"CLAUDE_CODE_USE_BEDROCK": "0"}, model.FAIL},
		"true":    {fakeenv.Env{"CLAUDE_CODE_USE_BEDROCK": "true"}, model.FAIL},
	} {
		t.Run(name, func(t *testing.T) {
			if f := one(t, rules.CheckUseBedrock(c.env, cfg)); f.Severity != c.sev {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestAnthropicKeyNeverLeaksValue(t *testing.T) {
	cfg := model.DefaultConfig()
	f := one(t, rules.CheckAnthropicKey(fakeenv.Env{"ANTHROPIC_API_KEY": "sk-secret-123"}, cfg))
	if f.Severity != model.FAIL || f.Reason != model.ReasonAnthropicKeySet {
		t.Fatalf("got %+v", f)
	}
	if strings.Contains(f.Message, "sk-secret-123") || strings.Contains(f.Subject, "sk-secret-123") {
		t.Fatalf("secret leaked: %+v", f)
	}
	f = one(t, rules.CheckAnthropicKey(fakeenv.Env{}, cfg))
	if f.Severity != model.PASS || f.Reason != model.ReasonAnthropicKeyNotSet {
		t.Fatalf("got %+v", f)
	}
}
