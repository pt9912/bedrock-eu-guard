package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// slice-026: --fail-on. Ein leeres Root ohne Benutzer-Settings liefert nur WARN
// (CLAUDE-001 FILE_MISSING), keinen FAIL.
func failOnRun(t *testing.T, args ...string) (code int, out, errs string) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, euEnv, "", &o, &e)
	return code, o.String(), e.String()
}

func TestFailOnWarnTurnsWarningsIntoExit1(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"Default: WARN bleibt Exit 0", []string{"local", "--root", root}, 0},
		{"--fail-on fail wie Default", []string{"local", "--root", root, "--fail-on", "fail"}, 0},
		{"--fail-on warn: WARN = Exit 1", []string{"local", "--root", root, "--fail-on", "warn"}, 1},
		{"Option vor anderen Optionen", []string{"local", "--fail-on", "warn", "--root", root}, 1},
	}
	for _, c := range cases {
		code, out, _ := failOnRun(t, c.args...)
		if code != c.want {
			t.Errorf("%s: exit=%d want %d\n%s", c.name, code, c.want, out)
		}
		if !strings.Contains(out, "[WARN]") || strings.Contains(out, "[FAIL]") {
			t.Errorf("%s: Datensatz soll nur WARN enthalten:\n%s", c.name, out)
		}
	}
}

// Die Schwelle ändert weder Befunde noch Ausgabe, nur den Exit-Code.
func TestFailOnDoesNotChangeOutput(t *testing.T) {
	root := t.TempDir()
	for _, format := range []string{"console", "json", "sarif"} {
		_, a, _ := failOnRun(t, "local", "--root", root, "--format", format)
		_, b, _ := failOnRun(t, "local", "--root", root, "--format", format, "--fail-on", "warn")
		if a != b {
			t.Errorf("%s: --fail-on darf die Ausgabe nicht ändern", format)
		}
	}
	_, out, _ := failOnRun(t, "local", "--root", root, "--format", "json", "--fail-on", "warn")
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Errorf("JSON nicht parsebar: %v", err)
	}
}

func TestFailOnWithFailAndCleanData(t *testing.T) {
	if code, _, _ := ciRun(t, "local", "bad", euEnv, "--fail-on", "warn"); code != 1 {
		t.Errorf("bad mit --fail-on warn: exit=%d, want 1", code)
	}
	if code, _, _ := ciRun(t, "terraform", "good", mapEnv{}, "--fail-on", "warn"); code != 0 {
		t.Errorf("good (nur PASS) mit --fail-on warn: exit=%d, want 0", code)
	}
}

func TestFailOnCISummaryShowsActualExit(t *testing.T) {
	code, _, errs := failOnRun(t, "local", "--root", t.TempDir(), "--fail-on", "warn", "--ci")
	if code != 1 || !strings.Contains(errs, "(Exit 1)") {
		t.Errorf("exit=%d stderr=%q", code, errs)
	}
}

// Tool-Fehler hat Vorrang vor der Schwelle (all + fehlendes aws-Binary).
func TestFailOnToolErrorStillExit2(t *testing.T) {
	old := awsBinary
	awsBinary = filepath.Join(t.TempDir(), "gibt-es-nicht")
	t.Cleanup(func() { awsBinary = old })
	if code, _, errs := ciRun(t, "all", "good", euEnv, "--fail-on", "warn"); code != 2 || !strings.Contains(errs, "AWS_CLI_MISSING") {
		t.Errorf("exit=%d stderr=%q", code, errs)
	}
}

// Ungültige Werte und die Option bei "version" sind Aufruffehler (Exit 2, usage).
func TestFailOnInvalidIsUsageError(t *testing.T) {
	for _, args := range [][]string{
		{"local", "--fail-on", "pass"},
		{"local", "--fail-on", "WARN"},
		{"local", "--fail-on", ""},
		{"local", "--fail-on"},
		{"version", "--fail-on", "warn"},
	} {
		code, out, errs := failOnRun(t, args...)
		if code != 2 || !strings.Contains(errs, "usage:") || !strings.Contains(errs, "--fail-on fail|warn") || out != "" {
			t.Errorf("%v: exit=%d stdout=%q stderr=%q", args, code, out, errs)
		}
	}
}
