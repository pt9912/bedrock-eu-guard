package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

// LH-ABN-001: ANTHROPIC_API_KEY gesetzt, AWS_REGION fehlt.
func TestAcceptanceABN001(t *testing.T) {
	env := fakeenv.Env{"ANTHROPIC_API_KEY": "secret"}
	fs, err := app.CheckLocal{}.Run(env, model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	fails := 0
	for _, f := range fs {
		if f.Severity == model.FAIL {
			fails++
		}
		if strings.Contains(f.Message, "secret") {
			t.Fatalf("secret leaked: %+v", f)
		}
	}
	if fails < 2 {
		t.Fatalf("want >=2 FAIL, got %d: %+v", fails, fs)
	}
	if got := app.ExitCode(fs, nil); got != 1 {
		t.Fatalf("exit=%d", got)
	}
}

// LH-ABN-002: EU-konforme Minimalumgebung.
func TestAcceptanceABN002(t *testing.T) {
	env := fakeenv.Env{"AWS_REGION": "eu-central-1", "CLAUDE_CODE_USE_BEDROCK": "1"}
	fs, err := app.CheckLocal{}.Run(env, model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if m := model.MaxSeverity(fs); m != model.PASS {
		t.Fatalf("max=%v: %+v", m, fs)
	}
	if got := app.ExitCode(fs, nil); got != 0 {
		t.Fatalf("exit=%d", got)
	}
}

func TestStableOrderByRuleID(t *testing.T) {
	fs, _ := app.CheckLocal{}.Run(fakeenv.Env{}, model.DefaultConfig())
	for i := 1; i < len(fs); i++ {
		if fs[i-1].RuleID > fs[i].RuleID {
			t.Fatalf("not sorted: %+v", fs)
		}
	}
}

func TestExitCode(t *testing.T) {
	w := []model.Finding{{Severity: model.WARN}}
	f := []model.Finding{{Severity: model.WARN}, {Severity: model.FAIL}}
	for name, c := range map[string]struct {
		fs   []model.Finding
		err  error
		want int
	}{
		"empty": {nil, nil, 0},
		"warn":  {w, nil, 0},
		"fail":  {f, nil, 1},
		"error": {f, errors.New("x"), 2},
	} {
		if got := app.ExitCode(c.fs, c.err); got != c.want {
			t.Errorf("%s: got %d want %d", name, got, c.want)
		}
	}
}
