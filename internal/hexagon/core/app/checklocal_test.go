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

func TestLocalIncludesDefaultModels(t *testing.T) {
	env := fakeenv.Env{"ANTHROPIC_DEFAULT_SONNET_MODEL": "us.anthropic.claude-sonnet-4-6"}
	fs, _ := app.CheckLocal{}.Run(env, model.DefaultConfig())
	for _, f := range fs {
		if f.RuleID == "LH-FA-LENV-006" && f.Severity == model.FAIL {
			return
		}
	}
	t.Fatalf("no LENV-006 FAIL: %+v", fs)
}

func TestExitCodeAt(t *testing.T) {
	f := func(sevs ...model.Severity) []model.Finding {
		var out []model.Finding
		for _, s := range sevs {
			out = append(out, model.Finding{Severity: s})
		}
		return out
	}
	boom := errors.New("x")
	cases := []struct {
		name   string
		fs     []model.Finding
		err    error
		failOn model.Severity
		want   int
	}{
		{"leer, Standard", nil, nil, model.FAIL, 0},
		{"nur PASS, Standard", f(model.PASS), nil, model.FAIL, 0},
		{"WARN, Standard", f(model.PASS, model.WARN), nil, model.FAIL, 0},
		{"FAIL, Standard", f(model.WARN, model.FAIL), nil, model.FAIL, 1},
		{"nur PASS, warn", f(model.PASS), nil, model.WARN, 0},
		{"leer, warn", nil, nil, model.WARN, 0},
		{"WARN, warn", f(model.PASS, model.WARN), nil, model.WARN, 1},
		{"FAIL, warn", f(model.FAIL), nil, model.WARN, 1},
		{"Fehler schlägt Schwelle (warn)", f(model.PASS), boom, model.WARN, 2},
		{"Fehler schlägt Befunde", f(model.FAIL), boom, model.FAIL, 2},
		{"PASS als Schwelle gilt als FAIL", f(model.WARN), nil, model.PASS, 0},
		{"unbekannte Schwelle gilt als FAIL", f(model.WARN), nil, model.Severity(99), 0},
		{"unbekannte Schwelle, FAIL", f(model.FAIL), nil, model.Severity(99), 1},
	}
	for _, c := range cases {
		if got := app.ExitCodeAt(c.fs, c.err, c.failOn); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
	// ExitCode ist ExitCodeAt mit Schwelle FAIL (Standardverhalten unverändert)
	for _, fs := range [][]model.Finding{nil, f(model.WARN), f(model.FAIL)} {
		if app.ExitCode(fs, nil) != app.ExitCodeAt(fs, nil, model.FAIL) {
			t.Errorf("ExitCode weicht ab: %+v", fs)
		}
	}
}
