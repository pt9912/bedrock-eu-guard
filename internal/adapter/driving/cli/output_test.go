package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// slice-027: --output FILE.
func outRun(t *testing.T, args ...string) (code int, out, errs string) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, euEnv, "", &o, &e)
	return code, o.String(), e.String()
}

// Die Datei enthält genau das, was sonst auf stdout stünde; stdout bleibt leer, Exit-Code gleich.
func TestOutputFileEqualsStdout(t *testing.T) {
	root := t.TempDir()
	for _, format := range []string{"console", "json", "sarif"} {
		wantCode, want, _ := outRun(t, "local", "--root", root, "--format", format)
		file := filepath.Join(t.TempDir(), "ergebnis.out")
		code, out, errs := outRun(t, "local", "--root", root, "--format", format, "--output", file)
		got, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("%s: Datei fehlt: %v", format, err)
		}
		if string(got) != want || code != wantCode || out != "" || errs != "" {
			t.Errorf("%s: exit=%d (want %d) stdout=%q stderr=%q, Datei gleich stdout: %v", format, code, wantCode, out, errs, string(got) == want)
		}
		if format != "console" {
			var doc map[string]any
			if json.Unmarshal(got, &doc) != nil {
				t.Errorf("%s: Datei ist kein JSON", format)
			}
		}
	}
}

// --ci schreibt weiter auf stderr, Exit-Code richtet sich nach --fail-on.
func TestOutputFileWithCIAndFailOn(t *testing.T) {
	file := filepath.Join(t.TempDir(), "r.json")
	code, out, errs := outRun(t, "local", "--root", t.TempDir(), "--format", "json", "--ci", "--fail-on", "warn", "--output", file)
	if code != 1 || out != "" || !strings.Contains(errs, "(Exit 1)") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	if _, err := os.Stat(file); err != nil {
		t.Errorf("Datei fehlt: %v", err)
	}
}

// Eine bestehende Datei wird vollständig ersetzt; es bleiben keine temporären Dateien übrig.
func TestOutputFileReplacesExistingAtomically(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "r.sarif")
	if err := os.WriteFile(file, []byte("ALT-INHALT\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := outRun(t, "local", "--root", t.TempDir(), "--format", "sarif", "--output", file); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "ALT-INHALT") || !strings.Contains(string(got), "\"version\"") {
		t.Errorf("Datei nicht ersetzt: %q", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("temporäre Dateien übrig: %v", entries)
	}
}

// Schreibfehler (Verzeichnis fehlt, Ziel ist ein Verzeichnis) → Exit 2 mit klarer Meldung.
func TestOutputFileWriteErrorIsToolError(t *testing.T) {
	for name, target := range map[string]string{
		"Verzeichnis fehlt": filepath.Join(t.TempDir(), "gibt-es-nicht", "r.json"),
		"Ziel ist Verzeichnis": t.TempDir(),
	} {
		code, out, errs := outRun(t, "local", "--root", t.TempDir(), "--format", "json", "--output", target)
		if code != 2 || out != "" || !strings.Contains(errs, "--output") {
			t.Errorf("%s: exit=%d stdout=%q stderr=%q", name, code, out, errs)
		}
	}
}

// Tool-Fehler vor dem Ergebnis: keine Datei anlegen, bestehende unangetastet lassen.
func TestOutputFileNotWrittenOnEarlyToolError(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.yml")
	if err := os.WriteFile(bad, []byte("unbekanntes_feld: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fresh := filepath.Join(dir, "neu.json")
	if code, _, _ := outRun(t, "local", "--root", t.TempDir(), "--config", bad, "--format", "json", "--output", fresh); code != 2 {
		t.Fatalf("exit=%d", code)
	}
	if _, err := os.Stat(fresh); err == nil {
		t.Error("Datei darf bei Konfigurationsfehler nicht angelegt werden")
	}
	old := filepath.Join(dir, "alt.json")
	_ = os.WriteFile(old, []byte("ALT\n"), 0o644)
	outRun(t, "local", "--root", t.TempDir(), "--config", bad, "--output", old)
	if got, _ := os.ReadFile(old); string(got) != "ALT\n" {
		t.Errorf("bestehende Datei verändert: %q", got)
	}
}

// Bereits ermittelte Befunde bleiben auch bei Tool-Fehler im Ergebnis (wie auf stdout).
func TestOutputFilePartialResultOnToolError(t *testing.T) {
	old := awsBinary
	awsBinary = filepath.Join(t.TempDir(), "gibt-es-nicht")
	t.Cleanup(func() { awsBinary = old })
	file := filepath.Join(t.TempDir(), "r.json")
	args, home := scanArgs("testdata/good")
	args[0] = "all"
	var o, e bytes.Buffer
	code := run(append(args, "--format", "json", "--output", file), euEnv, home, &o, &e)
	var doc map[string]any
	data, err := os.ReadFile(file)
	if code != 2 || err != nil || json.Unmarshal(data, &doc) != nil || o.Len() != 0 {
		t.Errorf("exit=%d err=%v stdout=%q", code, err, o.String())
	}
}

// Ungültig: --output ohne Wert, leerer Wert, bei version.
func TestOutputFileUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"local", "--output"}, {"local", "--output", ""}, {"version", "--output", "x"}} {
		if code, _, errs := outRun(t, args...); code != 2 || !strings.Contains(errs, "usage:") || !strings.Contains(errs, "[--output FILE]") {
			t.Errorf("%v: exit=%d stderr=%q", args, code, errs)
		}
	}
}
