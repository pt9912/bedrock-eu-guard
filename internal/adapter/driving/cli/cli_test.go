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

// euEnv ist die EU-konforme Minimalumgebung (LH-ABN-002).
var euEnv = mapEnv{"AWS_REGION": "eu-central-1", "CLAUDE_CODE_USE_BEDROCK": "1"}

func scanArgs(dir string) (args []string, home string) {
	return []string{"local", "--root", dir + "/project"}, dir + "/home"
}

// Positivbeispiel: EU-sichere Settings → kein FAIL, Exit 0.
func TestLocalGoodSettings(t *testing.T) {
	args, home := scanArgs("testdata/good")
	var out, errb bytes.Buffer
	if code := run(args, euEnv, home, &out, &errb); code != 0 {
		t.Fatalf("exit = %d; stdout=%q stderr=%q", code, out.String(), errb.String())
	}
	if strings.Contains(out.String(), "[FAIL]") {
		t.Errorf("unerwartetes FAIL:\n%s", out.String())
	}
}

// Negativbeispiel: direkte Anthropic-Konfiguration, US-Region, blockiertes
// Modell → FAIL mit Datei und Zeile, Exit 1, kein Secret-Wert (LH-NF-004).
func TestLocalBadSettings(t *testing.T) {
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

func TestLocalMissingFilesAreFindingsNotErrors(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"local", "--root", t.TempDir()}, euEnv, t.TempDir(), &out, &errb)
	if code != 0 || !strings.Contains(out.String(), "FILE_MISSING") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out.String(), errb.String())
	}
}

func TestLocalBadArgs(t *testing.T) {
	for _, args := range [][]string{{"local", "--root"}, {"local", "x"}, {"scan"}, {"scan", "--root", "x"}} {
		var out, errb bytes.Buffer
		if code := run(args, mapEnv{}, "", &out, &errb); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
	}
}

func localWith(t *testing.T, env mapEnv, extra ...string) (code int, out, errs string) {
	t.Helper()
	args, home := scanArgs("testdata/good")
	var o, e bytes.Buffer
	code = run(append(args, extra...), env, home, &o, &e)
	return code, o.String(), e.String()
}

func mantleEnv() mapEnv {
	return mapEnv{"AWS_REGION": "eu-central-1", "CLAUDE_CODE_USE_BEDROCK": "1", "CLAUDE_CODE_USE_MANTLE": "1"}
}

// SPEC-020 mantle: warn (Default) → WARN; allow → PASS; block → FAIL.
func TestConfigMantleModes(t *testing.T) {
	cases := []struct {
		extra    []string
		wantCode int
		wantSub  string
	}{
		{nil, 0, "[WARN] LH-FA-LENV-005"},
		{[]string{"--config", "testdata/config/mantle-allow.yml"}, 0, "MANTLE_ALLOWED"},
		{[]string{"--config", "testdata/config/mantle-block.yml"}, 1, "MANTLE_BLOCKED"},
	}
	for _, c := range cases {
		code, out, errs := localWith(t, mantleEnv(), c.extra...)
		if code != c.wantCode || !strings.Contains(out, c.wantSub) {
			t.Errorf("%v: exit=%d want %d, stdout=%q stderr=%q", c.extra, code, c.wantCode, out, errs)
		}
	}
}

// Ungültige eigene Konfiguration → Tool-Fehler Exit 2, ohne Feldwert (LH-NF-004).
func TestInvalidConfigIsToolError(t *testing.T) {
	for _, f := range []string{"invalid-regex.yml", "unknown-field.yml", "empty-regions.yml"} {
		code, out, errs := localWith(t, euEnv, "--config", "testdata/config/"+f)
		if code != 2 || out != "" || errs == "" {
			t.Errorf("%s: exit=%d stdout=%q stderr=%q", f, code, out, errs)
		}
		if strings.Contains(errs, "GEHEIM") {
			t.Errorf("%s: Feldwert in der Fehlermeldung: %q", f, errs)
		}
	}
}

func TestExplicitMissingConfigIsToolError(t *testing.T) {
	code, _, errs := localWith(t, euEnv, "--config", "testdata/config/gibt-es-nicht.yml")
	if code != 2 || errs == "" {
		t.Errorf("exit=%d stderr=%q", code, errs)
	}
}

// Fehlt die Default-Datei im Root, gelten die Defaults.
func TestMissingDefaultConfigMeansDefaults(t *testing.T) {
	if code, out, errs := localWith(t, euEnv); code != 0 {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
}

// slice-006: Devcontainer, Compose und Dockerfile — Positivbeispiel.
func TestLocalEnvSourcesGood(t *testing.T) {
	code, out, errs := localWith(t, euEnv)
	if code != 0 || strings.Contains(out, "[FAIL]") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, want := range []string{"LH-FA-DEVCON-001", "LH-FA-DOCKER-001", "LH-FA-DOCKER-002"} {
		if !strings.Contains(out, "[PASS] "+want) {
			t.Errorf("kein PASS für %s:\n%s", want, out)
		}
	}
}

// Negativbeispiel: direkter Anthropic-Zugang und US-/APAC-Regionen in allen
// drei Quellen → FAIL mit Datei, Exit 1, kein Secret-Wert (LH-NF-004).
func TestLocalEnvSourcesBad(t *testing.T) {
	args, home := scanArgs("testdata/bad")
	var o, e bytes.Buffer
	code := run(args, euEnv, home, &o, &e)
	out := o.String()
	if code != 1 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, e.String())
	}
	for _, want := range []string{
		"[FAIL] LH-FA-DEVCON-001", ".devcontainer/devcontainer.json:",
		"[FAIL] LH-FA-DOCKER-001", "docker-compose.yml:",
		"[FAIL] LH-FA-DOCKER-002", "Dockerfile:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Ausgabe enthält %q nicht:\n%s", want, out)
		}
	}
	for _, s := range []string{out, e.String()} {
		if strings.Contains(s, "GEHEIM") || strings.Contains(s, "sk-ant") {
			t.Errorf("Secret-Wert in der Ausgabe:\n%s", s)
		}
	}
}
