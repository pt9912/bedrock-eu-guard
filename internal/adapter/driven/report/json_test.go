package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/report"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.Reporter = (*report.JSON)(nil)

func renderJSON(t *testing.T, fs []model.Finding) (string, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if err := report.NewJSON(&buf).Report(fs); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("keine gültige JSON-Ausgabe: %v\n%s", err, buf.String())
	}
	return buf.String(), doc
}

// Die Schema-Version ist fixiert; eine Änderung muss bewusst erfolgen.
func TestJSONVersionIsPinned(t *testing.T) {
	if report.JSONVersion != 1 {
		t.Fatalf("JSONVersion = %d; Schemaänderung bewusst anpassen (SPEC-021)", report.JSONVersion)
	}
	_, doc := renderJSON(t, nil)
	if doc["version"] != float64(1) {
		t.Errorf("version = %v", doc["version"])
	}
}

func TestJSONEmptyIsEmptyArray(t *testing.T) {
	out, doc := renderJSON(t, nil)
	arr, ok := doc["findings"].([]any)
	if !ok || len(arr) != 0 {
		t.Errorf("findings muss ein leeres Array sein: %v", doc["findings"])
	}
	if !strings.Contains(out, `"findings": []`) {
		t.Errorf("Ausgabe:\n%s", out)
	}
}

func TestJSONFindingFields(t *testing.T) {
	_, doc := renderJSON(t, []model.Finding{
		{RuleID: "LH-FA-CLAUDE-003", Severity: model.FAIL, Reason: "DIRECT_ANTHROPIC", Message: "Direkter Zugang <b> & \"x\" äöü",
			Subject: "SUBJECT-NICHT-AUSGEBEN", File: ".claude/settings.json", Line: 5, Recommendation: "Entfernen"},
		{RuleID: "LH-FA-LENV-001", Severity: model.PASS, Reason: "OK", Message: "ok"},
		{RuleID: "LH-FA-LENV-002", Severity: model.WARN, Reason: "REGION_FALLBACK_DEFAULT", Message: "w", File: "a.tf"},
	})
	fs := doc["findings"].([]any)
	if len(fs) != 3 {
		t.Fatalf("findings: %v", fs)
	}
	f0 := fs[0].(map[string]any)
	want0 := map[string]any{
		"status": "FAIL", "id": "LH-FA-CLAUDE-003", "severity": "error", "reason": "DIRECT_ANTHROPIC",
		"message": "Direkter Zugang <b> & \"x\" äöü", "file": ".claude/settings.json", "line": float64(5), "recommendation": "Entfernen",
	}
	for k, v := range want0 {
		if f0[k] != v {
			t.Errorf("f0[%s] = %v, want %v", k, f0[k], v)
		}
	}
	if _, ok := f0["subject"]; ok {
		t.Error("subject darf nicht ausgegeben werden")
	}
	f1 := fs[1].(map[string]any)
	if f1["status"] != "PASS" || f1["severity"] != "info" || f1["file"] != nil || f1["line"] != nil || f1["recommendation"] != nil {
		t.Errorf("f1 (null-Felder): %v", f1)
	}
	f2 := fs[2].(map[string]any)
	if f2["status"] != "WARN" || f2["severity"] != "warning" || f2["file"] != "a.tf" || f2["line"] != nil {
		t.Errorf("f2 (Datei ohne Zeile): %v", f2)
	}
	for _, f := range fs {
		for _, k := range []string{"status", "id", "message", "severity", "reason", "file", "line", "recommendation"} {
			if _, ok := f.(map[string]any)[k]; !ok {
				t.Errorf("Pflichtfeld %s fehlt", k)
			}
		}
	}
}

func TestJSONDeterministicAndReadable(t *testing.T) {
	fs := []model.Finding{{RuleID: "R", Severity: model.WARN, Reason: "X", Message: "a < b & c"}}
	a, _ := renderJSON(t, fs)
	b, _ := renderJSON(t, fs)
	if a != b {
		t.Error("Ausgabe ist nicht deterministisch")
	}
	if !strings.Contains(a, "a < b & c") || !strings.HasSuffix(a, "}\n") {
		t.Errorf("Ausgabe:\n%s", a)
	}
}

func TestJSONWriteError(t *testing.T) {
	if err := report.NewJSON(failWriter{}).Report(nil); err == nil {
		t.Error("expected write error")
	}
}
