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
