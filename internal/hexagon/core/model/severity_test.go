package model_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

func TestSeverityLevels(t *testing.T) {
	cases := []struct {
		s     model.Severity
		str   string
		level string
		sarif string
	}{
		{model.PASS, "PASS", "info", "none"},
		{model.WARN, "WARN", "warning", "warning"},
		{model.FAIL, "FAIL", "error", "error"},
		{model.Severity(99), "UNKNOWN", "info", "none"},
	}
	for _, c := range cases {
		if c.s.String() != c.str || c.s.Level() != c.level || c.s.SARIFLevel() != c.sarif {
			t.Errorf("%v: String=%q Level=%q SARIFLevel=%q", int(c.s), c.s.String(), c.s.Level(), c.s.SARIFLevel())
		}
	}
}

func TestParseFailOn(t *testing.T) {
	for in, want := range map[string]struct {
		s  model.Severity
		ok bool
	}{
		"fail": {model.FAIL, true},
		"warn": {model.WARN, true},
		"":     {model.PASS, false},
		"pass": {model.PASS, false},
		"WARN": {model.PASS, false},
		"info": {model.PASS, false},
	} {
		s, ok := model.ParseFailOn(in)
		if s != want.s || ok != want.ok {
			t.Errorf("ParseFailOn(%q) = %v,%v want %v,%v", in, s, ok, want.s, want.ok)
		}
	}
}

func TestStrictIAMPolicyName(t *testing.T) {
	yes := []string{"iam-policy.json", "sub/iam-policy.json", "iam-policy-prod.json", "a/b/iam-policy.v2.json"}
	no := []string{"app-policy.json", "sub/deploy-policy.json", "iam-policy", "iam-policy.yaml", "my-iam-policy.json", "iam-policy.json.bak"}
	for _, n := range yes {
		if !model.StrictIAMPolicyName(n) {
			t.Errorf("%q sollte eindeutig sein", n)
		}
	}
	for _, n := range no {
		if model.StrictIAMPolicyName(n) {
			t.Errorf("%q sollte nicht eindeutig sein", n)
		}
	}
}
