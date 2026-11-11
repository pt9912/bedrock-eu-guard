package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// slice-030: eigenständige IAM-Policy-Dateien (LH-FA-IAM-004), Modus terraform.

func TestPolicyFilesGood(t *testing.T) {
	code, out, errs := modeRun(t, "terraform", "good", mapEnv{})
	if code != 0 || strings.Contains(out, "[FAIL]") || errs != "" {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	if !strings.Contains(out, "[PASS] LH-FA-IAM-004 iam-policy.json") {
		t.Errorf("PASS für die enge Policy erwartet:\n%s", out)
	}
	// Fremdformat im breiten Muster ergibt kein Finding.
	if strings.Contains(out, "foreign-policy.json") {
		t.Errorf("Fremdformat darf kein Finding erzeugen:\n%s", out)
	}
}

func TestPolicyFilesBad(t *testing.T) {
	code, out, _ := modeRun(t, "terraform", "bad", mapEnv{})
	if code != 1 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	for _, want := range []string{
		"[FAIL] LH-FA-IAM-004 iam-policy.json",               // nacktes Dokument, bedrock:* auf *
		"[FAIL] LH-FA-IAM-004 policies/role-policy.json",     // Hülle get-role-policy, InvokeModel auf *
		"[WARN] LH-FA-IAM-004 policies/broken-policy.json:3", // breites Muster, kaputt: nur WARN, mit Zeile
		"FILE_INVALID",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("fehlt %q:\n%s", want, out)
		}
	}
	// Dateiinhalt erscheint nie, in keinem Format.
	for _, format := range []string{"console", "json", "sarif"} {
		_, o, e := modeRun(t, "terraform", "bad", mapEnv{}, "--format", format)
		if strings.Contains(o+e, "GEHEIM") || strings.Contains(o+e, "darf-nicht-erscheinen") {
			t.Errorf("%s: Dateiinhalt in der Ausgabe", format)
		}
	}
}

// Das Muster iam-policy*.json ist eindeutig: dort ist eine kaputte Datei ein FAIL; eine Datei, die
// beide Muster trifft (iam-policy.json), zählt als eindeutig.
func TestPolicyFilesErrorSeverityByPattern(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "iam-policy.json"), []byte("{ kaputt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "andere-policy.json"), []byte("{ kaputt"), 0o644); err != nil {
		t.Fatal(err)
	}
	var o, e strings.Builder
	code := run([]string{"terraform", "--root", root}, mapEnv{}, "", &o, &e)
	out := o.String()
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-IAM-004 iam-policy.json:1") || !strings.Contains(out, "[WARN] LH-FA-IAM-004 andere-policy.json:1") {
		t.Errorf("exit=%d\n%s", code, out)
	}
}

func TestPolicyFilesExclude(t *testing.T) {
	_, out, _ := modeRun(t, "terraform", "bad", mapEnv{}, "--config", "testdata/config/exclude-policies.yml")
	if strings.Contains(out, "policies/") {
		t.Errorf("exclude: [policies] muss die Dateien unterhalb ausnehmen:\n%s", out)
	}
	if !strings.Contains(out, "LH-FA-IAM-004 iam-policy.json") {
		t.Errorf("die Policy im Root bleibt:\n%s", out)
	}
}

// Nur die Modi terraform und all prüfen Policy-Dateien; local nicht.
func TestPolicyFilesOnlyInTerraformAndAll(t *testing.T) {
	if _, out, _ := modeRun(t, "local", "bad", euEnv); strings.Contains(out, "LH-FA-IAM-004") {
		t.Errorf("local darf Policy-Dateien nicht prüfen:\n%s", out)
	}
	useFakeAWS(t, awsOK)
	_, out, _ := modeRun(t, "all", "bad", mapEnv{"AWS_PROFILE": "dev", "AWS_REGION": "eu-central-1", "CLAUDE_CODE_USE_BEDROCK": "1"})
	if !strings.Contains(out, "LH-FA-IAM-004 iam-policy.json") {
		t.Errorf("all muss Policy-Dateien prüfen:\n%s", out)
	}
}

func TestPolicyFilesJSONAndSARIF(t *testing.T) {
	_, out, _ := modeRun(t, "terraform", "bad", mapEnv{}, "--format", "json")
	var doc struct {
		Findings []map[string]any `json:"findings"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("kein JSON: %v", err)
	}
	var seen bool
	for _, f := range doc.Findings {
		if f["id"] == "LH-FA-IAM-004" && f["file"] == "iam-policy.json" && f["status"] == "FAIL" {
			seen = true
		}
	}
	if !seen {
		t.Errorf("JSON-Finding für iam-policy.json fehlt:\n%s", out)
	}
	_, _, _, sarif := sarifRun(t, "terraform", "bad", mapEnv{})
	if sarif["version"] != "2.1.0" {
		t.Errorf("SARIF-Version: %v", sarif["version"])
	}
}
