package cli

import (
	"bytes"
	"strings"
	"testing"
)

type mapEnv map[string]string

func (m mapEnv) Lookup(k string) (string, bool) { v, ok := m[k]; return v, ok }

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"bogus"}, {"local", "extra"}} {
		var out, errb bytes.Buffer
		if code := run(args, mapEnv{}, "", &out, &errb); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
		if !strings.Contains(errb.String(), "usage:") || out.Len() != 0 {
			t.Errorf("run(%v): stderr=%q stdout=%q", args, errb.String(), out.String())
		}
	}
}

func TestLocalPrintsSummary(t *testing.T) {
	var out, errb bytes.Buffer
	run([]string{"local"}, mapEnv{}, "", &out, &errb)
	if !strings.Contains(out.String(), "Summary:") {
		t.Errorf("stdout=%q stderr=%q", out.String(), errb.String())
	}
}

// LH-ABN-001: kritische lokale Fehlkonfiguration → ≥ 2 FAIL, Exit 1, kein Secret-Wert.
func TestAcceptanceCriticalMisconfiguration(t *testing.T) {
	const secret = "sk-ant-secret-123"
	env := mapEnv{
		"AWS_REGION":        "us-east-1",
		"ANTHROPIC_API_KEY": secret,
	}
	var out, errb bytes.Buffer
	if code := run([]string{"local"}, env, "", &out, &errb); code != 1 {
		t.Errorf("exit = %d, want 1; stdout=%q stderr=%q", code, out.String(), errb.String())
	}
	if n := strings.Count(out.String(), "[FAIL]"); n < 2 {
		t.Errorf("FAIL-Zeilen = %d, want >= 2:\n%s", n, out.String())
	}
	if strings.Contains(out.String(), secret) || strings.Contains(errb.String(), secret) {
		t.Error("Secret-Wert in der Ausgabe (LH-NF-004)")
	}
}

// LH-ABN-002: EU-konforme Minimalumgebung → kein FAIL, Exit 0.
func TestAcceptanceEUMinimalEnvironment(t *testing.T) {
	env := mapEnv{
		"AWS_REGION":              "eu-central-1",
		"CLAUDE_CODE_USE_BEDROCK": "1",
	}
	var out, errb bytes.Buffer
	if code := run([]string{"local"}, env, "", &out, &errb); code != 0 {
		t.Errorf("exit = %d, want 0; stdout=%q stderr=%q", code, out.String(), errb.String())
	}
	if strings.Contains(out.String(), "[FAIL]") {
		t.Errorf("unerwartetes FAIL:\n%s", out.String())
	}
}

func scanArgs(dir string) (args []string, home string) {
	return []string{"scan", "--root", dir + "/project"}, dir + "/home"
}

// Positivbeispiel: EU-sichere Settings → kein FAIL, Exit 0.
func TestScanGoodSettings(t *testing.T) {
	args, home := scanArgs("testdata/good")
	var out, errb bytes.Buffer
	if code := run(args, mapEnv{}, home, &out, &errb); code != 0 {
		t.Fatalf("exit = %d; stdout=%q stderr=%q", code, out.String(), errb.String())
	}
	if strings.Contains(out.String(), "[FAIL]") {
		t.Errorf("unerwartetes FAIL:\n%s", out.String())
	}
}

// Negativbeispiel: direkte Anthropic-Konfiguration, US-Region, blockiertes
// Modell → FAIL mit Datei und Zeile, Exit 1, kein Secret-Wert (LH-NF-004).
func TestScanBadSettings(t *testing.T) {
	args, home := scanArgs("testdata/bad")
	var out, errb bytes.Buffer
	if code := run(args, mapEnv{}, home, &out, &errb); code != 1 {
		t.Fatalf("exit = %d; stdout=%q stderr=%q", code, out.String(), errb.String())
	}
	for _, want := range []string{"[FAIL]", "~/.claude/settings.json:", ".claude/settings.json:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Ausgabe enthält %q nicht:\n%s", want, out.String())
		}
	}
	for _, s := range []string{out.String(), errb.String()} {
		if strings.Contains(s, "GEHEIM") || strings.Contains(s, "sk-ant") {
			t.Errorf("Secret-Wert in der Ausgabe:\n%s", s)
		}
	}
}

func TestScanMissingFilesAreFindingsNotErrors(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"scan", "--root", t.TempDir()}, mapEnv{}, t.TempDir(), &out, &errb)
	if code != 0 || !strings.Contains(out.String(), "FILE_MISSING") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out.String(), errb.String())
	}
}

func TestScanBadArgs(t *testing.T) {
	for _, args := range [][]string{{"scan", "--root"}, {"scan", "x"}, {"scan", "--root", ""}} {
		var out, errb bytes.Buffer
		if code := run(args, mapEnv{}, "", &out, &errb); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
	}
}
