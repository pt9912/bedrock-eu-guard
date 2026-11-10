package rules_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

func TestCheckMantle(t *testing.T) {
	cases := []struct {
		name       string
		env        fakeenv.Env
		wantNone   bool
		wantSev    model.Severity
		wantReason string
	}{
		{name: "nicht gesetzt", env: fakeenv.Env{}, wantNone: true},
		{name: "Wert 0", env: fakeenv.Env{"CLAUDE_CODE_USE_MANTLE": "0"}, wantNone: true},
		{name: "aktiv ohne Modellvariablen", env: fakeenv.Env{"CLAUDE_CODE_USE_MANTLE": "1"},
			wantSev: model.WARN, wantReason: model.ReasonMantleUnverified},
		{name: "aktiv, leere Modellvariable", env: fakeenv.Env{"CLAUDE_CODE_USE_MANTLE": "1", "ANTHROPIC_DEFAULT_SONNET_MODEL": ""},
			wantSev: model.WARN, wantReason: model.ReasonMantleUnverified},
		{name: "aktiv, alle erlaubt", env: fakeenv.Env{"CLAUDE_CODE_USE_MANTLE": "1",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "eu.anthropic.claude-sonnet-4-6",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "anthropic.claude-haiku-4-5"},
			wantSev: model.PASS, wantReason: model.ReasonMantleVerified},
		{name: "aktiv, eine blockiert", env: fakeenv.Env{"CLAUDE_CODE_USE_MANTLE": "1",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "eu.anthropic.claude-sonnet-4-6",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   "us.anthropic.claude-opus-4"},
			wantSev: model.WARN, wantReason: model.ReasonMantleUnverified},
		{name: "aktiv, eine unklassifiziert", env: fakeenv.Env{"CLAUDE_CODE_USE_MANTLE": "1",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "irgendwas"},
			wantSev: model.WARN, wantReason: model.ReasonMantleUnverified},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := rules.CheckMantle(c.env, model.DefaultConfig())
			if c.wantNone {
				if len(got) != 0 {
					t.Fatalf("want no finding, got %+v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("want 1 finding, got %+v", got)
			}
			if got[0].RuleID != rules.IDMantle || got[0].Severity != c.wantSev || got[0].Reason != c.wantReason {
				t.Errorf("got %+v, want sev=%v reason=%s", got[0], c.wantSev, c.wantReason)
			}
		})
	}
}
