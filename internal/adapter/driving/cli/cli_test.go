package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	run([]string{"local", "--root", t.TempDir()}, mapEnv{}, "", &out, &errb)
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
	if code := run([]string{"local", "--root", t.TempDir()}, env, "", &out, &errb); code != 1 {
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
	if code := run([]string{"local", "--root", t.TempDir()}, env, "", &out, &errb); code != 0 {
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

func modeRun(t *testing.T, mode, dir string, env mapEnv, extra ...string) (code int, out, errs string) {
	t.Helper()
	args, home := scanArgs("testdata/" + dir)
	args[0] = mode
	var o, e bytes.Buffer
	code = run(append(args, extra...), env, home, &o, &e)
	return code, o.String(), e.String()
}

// Modus terraform (LH-MOD-003): nur Terraform-Regeln — Positivbeispiel.
func TestTerraformModeGood(t *testing.T) {
	code, out, errs := modeRun(t, "terraform", "good", mapEnv{})
	if code != 0 || strings.Contains(out, "[FAIL]") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, want := range []string{"LH-FA-TF-001", "LH-FA-TF-002", "LH-FA-TF-003", "LH-FA-TF-004"} {
		if !strings.Contains(out, "[PASS] "+want) {
			t.Errorf("kein PASS für %s:\n%s", want, out)
		}
	}
	for _, other := range []string{"LH-FA-LENV", "LH-FA-CLAUDE", "LH-FA-DEVCON", "LH-FA-DOCKER"} {
		if strings.Contains(out, other) {
			t.Errorf("Modus terraform darf %s nicht ausführen:\n%s", other, out)
		}
	}
}

// Negativbeispiel: US-Region, Bedrock-Wildcard auf "*", blockierte Modell-ID → FAIL
// mit Datei und Zeile, Exit 1, kein Wert aus variable/tfvars (LH-NF-004).
func TestTerraformModeBad(t *testing.T) {
	code, out, errs := modeRun(t, "terraform", "bad", mapEnv{})
	if code != 1 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, want := range []string{"[FAIL] LH-FA-TF-002", "[FAIL] LH-FA-TF-003", "[FAIL] LH-FA-TF-004", "main.tf:"} {
		if !strings.Contains(out, want) {
			t.Errorf("Ausgabe enthält %q nicht:\n%s", want, out)
		}
	}
	for _, s := range []string{out, errs} {
		if strings.Contains(s, "GEHEIM") || strings.Contains(s, "sk-ant") {
			t.Errorf("Wert aus variable/Secret in der Ausgabe:\n%s", s)
		}
	}
}

// Modus local führt Terraform NICHT aus (LH-MOD-001 vs. LH-MOD-003).
func TestLocalModeExcludesTerraform(t *testing.T) {
	_, out, _ := modeRun(t, "local", "good", euEnv)
	if strings.Contains(out, "LH-FA-TF-") {
		t.Errorf("local darf keine Terraform-Regeln ausführen:\n%s", out)
	}
}

// Modus all = local + terraform.
func TestAllModeRunsBoth(t *testing.T) {
	useFakeAWS(t, awsOK)
	code, out, errs := modeRun(t, "all", "good", euEnv)
	if code != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, want := range []string{"LH-FA-LENV-001", "LH-FA-CLAUDE-001", "LH-FA-DEVCON-001", "LH-FA-TF-001", "[PASS] LH-FA-AWS-001"} {
		if !strings.Contains(out, want) {
			t.Errorf("all enthält %s nicht:\n%s", want, out)
		}
	}
	if code, _, _ := modeRun(t, "all", "bad", euEnv); code != 1 {
		t.Errorf("all bad: exit=%d", code)
	}
}

func TestUnknownModeIsUsageError(t *testing.T) {
	var o, e bytes.Buffer
	if code := run([]string{"bogus"}, euEnv, "", &o, &e); code != 2 || !strings.Contains(e.String(), "usage:") {
		t.Errorf("exit=%d stderr=%q", code, e.String())
	}
}

const (
	awsOK = `case "$1 $2" in
  "sts get-caller-identity") echo '{"UserId":"AIDAEXAMPLE","Account":"123456789012","Arn":"arn:aws:iam::123456789012:user/tester"}' ;;
  "bedrock list-inference-profiles") echo '{"inferenceProfileSummaries":[{"inferenceProfileId":"eu.anthropic.claude-sonnet-4-6","inferenceProfileName":"EU Sonnet","type":"SYSTEM_DEFINED"},{"inferenceProfileId":"us.anthropic.claude-sonnet-4-6","inferenceProfileName":"US Sonnet","type":"SYSTEM_DEFINED"}]}' ;;
  "bedrock list-foundation-models") echo '{"modelSummaries":[{"modelId":"anthropic.claude-sonnet-4-6"}]}' ;;
  "iam simulate-principal-policy") echo '{"EvaluationResults":[{"EvalActionName":"bedrock:InvokeModel","EvalDecision":"allowed"},{"EvalActionName":"bedrock:InvokeModelWithResponseStream","EvalDecision":"allowed"},{"EvalActionName":"bedrock:Converse","EvalDecision":"allowed"},{"EvalActionName":"bedrock:ConverseStream","EvalDecision":"allowed"},{"EvalActionName":"bedrock:ListInferenceProfiles","EvalDecision":"allowed"},{"EvalActionName":"bedrock:GetInferenceProfile","EvalDecision":"allowed"}]}' ;;
  "iam list-attached-user-policies") echo '{"AttachedPolicies":[]}' ;;
  "iam list-user-policies") echo '{"PolicyNames":["p"]}' ;;
  "iam get-user-policy") echo '{"PolicyDocument":{"Statement":[{"Effect":"Allow","Action":["bedrock:InvokeModel","bedrock:Converse"],"Resource":"arn:aws:bedrock:eu-central-1::foundation-model/*"}]}}' ;;
  "iam list-groups-for-user") echo '{"Groups":[]}' ;;
  *) echo "unerwarteter Aufruf" >&2; exit 2 ;;
esac`
	// stderr enthält absichtlich Text, der nie in der Ausgabe erscheinen darf.
	awsNoCreds = `echo 'Unable to locate credentials LEAK-STDERR-TEXT' >&2; exit 255`
)

// useFakeAWS ersetzt das aws-Binary durch ein Skript (nie echtes AWS).
func useFakeAWS(t *testing.T, body string) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "aws")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	old := awsBinary
	awsBinary = bin
	t.Cleanup(func() { awsBinary = old })
}

func awsRun(t *testing.T, env mapEnv, extra ...string) (code int, out, errs string) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(append([]string{"aws", "--root", t.TempDir()}, extra...), env, "", &o, &e)
	return code, o.String(), e.String()
}

func TestAWSModeIdentityAndAccounts(t *testing.T) {
	useFakeAWS(t, awsOK)
	env := mapEnv{"AWS_PROFILE": "dev"}

	code, out, errs := awsRun(t, env)
	if code != 0 || !strings.Contains(out, "[PASS] LH-FA-AWS-001") || !strings.Contains(out, "123456789012") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	if strings.Contains(out, "LH-FA-LENV") || strings.Contains(out, "LH-FA-TF-") {
		t.Errorf("Modus aws darf keine anderen Regeln ausführen:\n%s", out)
	}

	code, out, _ = awsRun(t, env, "--allowed-account", "123456789012")
	if code != 0 || !strings.Contains(out, "[PASS] LH-FA-AWS-003") {
		t.Errorf("erlaubtes Konto: exit=%d stdout=%q", code, out)
	}
	code, out, _ = awsRun(t, env, "--allowed-account", "999999999999")
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-AWS-003") {
		t.Errorf("nicht erlaubtes Konto: exit=%d stdout=%q", code, out)
	}
	code, _, errs = awsRun(t, env, "--allowed-account", "kein-konto")
	if code != 2 || errs == "" {
		t.Errorf("ungültige Konto-ID muss Exit 2 sein: exit=%d stderr=%q", code, errs)
	}
}

// Keine Credentials → FAIL (Exit 1), kein stderr-Text und kein Credential-Wert in der Ausgabe.
func TestAWSModeNoCredentials(t *testing.T) {
	useFakeAWS(t, awsNoCreds)
	code, out, errs := awsRun(t, mapEnv{"ANTHROPIC_API_KEY": "sk-ant-GEHEIM-AWS"})
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-AWS-001") || !strings.Contains(out, "[FAIL] LH-FA-AUTH-001") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, s := range []string{out, errs} {
		if strings.Contains(s, "LEAK-STDERR-TEXT") || strings.Contains(s, "GEHEIM") || strings.Contains(s, "sk-ant") {
			t.Errorf("Fremdausgabe oder Secret in der Ausgabe:\n%s", s)
		}
	}
}

// aws-Binary fehlt → Tool-Fehler (Exit 2); bei "all" erscheinen die lokalen Befunde trotzdem.
func TestAWSBinaryMissing(t *testing.T) {
	old := awsBinary
	awsBinary = filepath.Join(t.TempDir(), "gibt-es-nicht")
	t.Cleanup(func() { awsBinary = old })

	if code, _, errs := awsRun(t, mapEnv{}); code != 2 || !strings.Contains(errs, "AWS_CLI_MISSING") {
		t.Errorf("aws: exit=%d stderr=%q", code, errs)
	}
	code, out, errs := modeRun(t, "all", "good", euEnv)
	if code != 2 || !strings.Contains(errs, "AWS_CLI_MISSING") || !strings.Contains(out, "LH-FA-LENV-001") {
		t.Errorf("all: exit=%d stdout=%q stderr=%q", code, out, errs)
	}
}

// slice-009: verwendete Modelle (Umgebung) gegen Allowlist und Verfügbarkeit.
func TestAWSModeBedrockProfiles(t *testing.T) {
	useFakeAWS(t, awsOK)

	// EU-Profil, verfügbar, nur das US-Profil ist sichtbar → WARN, Exit 0.
	code, out, errs := awsRun(t, mapEnv{"AWS_PROFILE": "dev", "ANTHROPIC_MODEL": "eu.anthropic.claude-sonnet-4-6"})
	if code != 0 || !strings.Contains(out, "LH-FA-BEDROCK-001") || !strings.Contains(out, "[WARN] LH-FA-BEDROCK-002") ||
		!strings.Contains(out, "[PASS] LH-FA-BEDROCK-003") || !strings.Contains(out, "[PASS] LH-FA-BEDROCK-004") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}

	// Verwendetes US-Profil → FAIL, Exit 1.
	code, out, _ = awsRun(t, mapEnv{"AWS_PROFILE": "dev", "ANTHROPIC_MODEL": "us.anthropic.claude-sonnet-4-6"})
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-BEDROCK-002") {
		t.Errorf("US-Profil: exit=%d stdout=%q", code, out)
	}
}

func TestAWSModeBedrockErrorsAreFindings(t *testing.T) {
	useFakeAWS(t, `case "$1" in sts) echo '{"Account":"123456789012","Arn":"a","UserId":"u"}' ;; *) echo 'AccessDeniedException LEAK-STDERR-TEXT' >&2; exit 255 ;; esac`)
	code, out, errs := awsRun(t, mapEnv{"AWS_PROFILE": "dev"})
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-BEDROCK-001") || !strings.Contains(out, "AWS_ACCESS_DENIED") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, s := range []string{out, errs} {
		if strings.Contains(s, "LEAK-STDERR-TEXT") {
			t.Errorf("stderr-Text in der Ausgabe:\n%s", s)
		}
	}
}

// slice-010: IAM-Rechte und -Policies des eigenen Principals.
func TestAWSModeIAM(t *testing.T) {
	useFakeAWS(t, awsOK)
	code, out, errs := awsRun(t, mapEnv{"AWS_PROFILE": "dev"})
	if code != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, want := range []string{"[PASS] LH-FA-IAM-001", "[PASS] LH-FA-IAM-002", "[PASS] LH-FA-IAM-003"} {
		if !strings.Contains(out, want) {
			t.Errorf("%s fehlt:\n%s", want, out)
		}
	}
}

// Zu breite Bedrock-Policy ohne Region-Condition → FAIL (IAM-002, IAM-003).
func TestAWSModeBroadPolicy(t *testing.T) {
	broad := strings.Replace(awsOK, `"Resource":"arn:aws:bedrock:eu-central-1::foundation-model/*"`, `"Resource":"*"`, 1)
	useFakeAWS(t, broad)
	code, out, errs := awsRun(t, mapEnv{"AWS_PROFILE": "dev"})
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-IAM-002") || !strings.Contains(out, "[FAIL] LH-FA-IAM-003") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
}

// Rechte-Simulation nicht erlaubt → WARN statt stillem Erfolg, kein stderr-Text.
func TestAWSModeIAMSimulationDenied(t *testing.T) {
	denied := strings.Replace(awsOK, `"iam simulate-principal-policy") echo`, `"iam simulate-principal-policy") echo 'AccessDenied LEAK-STDERR-TEXT' >&2; exit 255; echo`, 1)
	useFakeAWS(t, denied)
	code, out, errs := awsRun(t, mapEnv{"AWS_PROFILE": "dev"})
	if code != 0 || !strings.Contains(out, "[WARN] LH-FA-IAM-001") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	for _, s := range []string{out, errs} {
		if strings.Contains(s, "LEAK-STDERR-TEXT") {
			t.Errorf("stderr-Text in der Ausgabe:\n%s", s)
		}
	}
}

// slice-011: --format json
func jsonRun(t *testing.T, mode, dir string, env mapEnv, extra ...string) (code int, out, errs string, doc map[string]any) {
	t.Helper()
	args, home := scanArgs("testdata/" + dir)
	args[0] = mode
	var o, e bytes.Buffer
	code = run(append(args, extra...), env, home, &o, &e)
	if o.Len() > 0 {
		if err := json.Unmarshal(o.Bytes(), &doc); err != nil {
			t.Fatalf("stdout ist kein JSON: %v\n%s", err, o.String())
		}
	}
	return code, o.String(), e.String(), doc
}

func TestFormatJSONGoodAndBad(t *testing.T) {
	code, out, errs, doc := jsonRun(t, "local", "good", euEnv, "--format", "json")
	if code != 0 || errs != "" || doc["version"] != float64(1) {
		t.Fatalf("good: exit=%d stderr=%q doc=%v", code, errs, doc["version"])
	}
	if fs, _ := doc["findings"].([]any); len(fs) == 0 {
		t.Errorf("good: keine Findings:\n%s", out)
	}

	code, out, _, doc = jsonRun(t, "local", "bad", euEnv, "--format", "json")
	if code != 1 {
		t.Fatalf("bad: exit=%d", code)
	}
	var fails int
	for _, f := range doc["findings"].([]any) {
		m := f.(map[string]any)
		if m["status"] == "FAIL" {
			fails++
			if m["severity"] != "error" {
				t.Errorf("FAIL muss severity error haben: %v", m)
			}
		}
	}
	if fails == 0 {
		t.Errorf("bad: kein FAIL im JSON:\n%s", out)
	}
	if strings.Contains(out, "GEHEIM") || strings.Contains(out, "sk-ant") {
		t.Errorf("Secret im JSON:\n%s", out)
	}
	// Datei und Zeile erscheinen als eigene Felder.
	if !strings.Contains(out, `"file": ".claude/settings.json"`) && !strings.Contains(out, `"file": "main.tf"`) {
		t.Errorf("file-Feld fehlt:\n%s", out)
	}
}

func TestFormatJSONToolErrorKeepsStdoutParsable(t *testing.T) {
	old := awsBinary
	awsBinary = filepath.Join(t.TempDir(), "gibt-es-nicht")
	t.Cleanup(func() { awsBinary = old })
	code, out, errs, doc := jsonRun(t, "all", "good", euEnv, "--format", "json")
	if code != 2 || !strings.Contains(errs, "AWS_CLI_MISSING") || doc == nil {
		t.Fatalf("exit=%d stderr=%q stdout=%q", code, errs, out)
	}
	if strings.Contains(out, "AWS_CLI_MISSING") {
		t.Error("Fehlermeldung gehört auf stderr, nicht in die JSON-Ausgabe")
	}
}

func TestFormatInvalidIsUsageError(t *testing.T) {
	for _, args := range [][]string{
		{"local", "--format", "xml"}, {"local", "--format"}, {"local", "--format", ""}, {"local", "--format", "SARIF"},
	} {
		var o, e bytes.Buffer
		if code := run(args, euEnv, "", &o, &e); code != 2 || o.Len() != 0 || !strings.Contains(e.String(), "usage:") {
			t.Errorf("%v: exit=%d stdout=%q stderr=%q", args, code, o.String(), e.String())
		}
	}
}

// slice-012: --format sarif
func sarifRun(t *testing.T, mode, dir string, env mapEnv, extra ...string) (code int, out, errs string, doc map[string]any) {
	t.Helper()
	args, home := scanArgs("testdata/" + dir)
	args[0] = mode
	var o, e bytes.Buffer
	code = run(append(args, append([]string{"--format", "sarif"}, extra...)...), env, home, &o, &e)
	if o.Len() > 0 {
		if err := json.Unmarshal(o.Bytes(), &doc); err != nil {
			t.Fatalf("stdout ist kein SARIF-JSON: %v\n%s", err, o.String())
		}
	}
	return code, o.String(), e.String(), doc
}

func TestFormatSARIFTerraformBad(t *testing.T) {
	code, out, errs, doc := sarifRun(t, "terraform", "bad", mapEnv{})
	if code != 1 || errs != "" || doc["version"] != "2.1.0" {
		t.Fatalf("exit=%d stderr=%q version=%v", code, errs, doc["version"])
	}
	run0 := doc["runs"].([]any)[0].(map[string]any)
	results := run0["results"].([]any)
	var errors int
	for _, r := range results {
		m := r.(map[string]any)
		if m["level"] == "error" {
			errors++
		}
		loc := m["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)
		// Terraform-Regeln verweisen auf main.tf; die Regel für eigenständige Policy-Dateien (IAM-004) auf deren Pfad.
		if rid, _ := m["ruleId"].(string); strings.HasPrefix(rid, "LH-FA-TF-") && loc["artifactLocation"].(map[string]any)["uri"] != "main.tf" {
			t.Errorf("Terraform-Findings verweisen auf main.tf: %v", loc)
		}
	}
	if errors == 0 {
		t.Errorf("keine error-Results:\n%s", out)
	}
	drv := run0["tool"].(map[string]any)["driver"].(map[string]any)
	if drv["name"] != "bedrock-eu-check" || drv["version"] != "dev" {
		t.Errorf("driver: %v", drv)
	}
	if strings.Contains(out, "GEHEIM") || strings.Contains(out, "sk-ant") {
		t.Errorf("Secret in SARIF:\n%s", out)
	}
}

// Findings ohne Fundstelle bekommen eine als Ersatz markierte Position.
func TestFormatSARIFFallbackLocation(t *testing.T) {
	_, out, _, doc := sarifRun(t, "local", "bad", euEnv)
	results := doc["runs"].([]any)[0].(map[string]any)["results"].([]any)
	var synthetic int
	for _, r := range results {
		m := r.(map[string]any)
		if p, _ := m["properties"].(map[string]any); p != nil && p["synthetic"] == true {
			synthetic++
		}
	}
	if synthetic == 0 || !strings.Contains(out, `"uri": "README.md"`) {
		t.Errorf("Ersatzposition erwartet (README.md, da keine Konfigurationsdatei):\n%s", out)
	}
}

func TestFormatSARIFGoodHasNoResults(t *testing.T) {
	code, _, _, doc := sarifRun(t, "terraform", "good", mapEnv{})
	results := doc["runs"].([]any)[0].(map[string]any)["results"].([]any)
	if code != 0 || len(results) != 0 {
		t.Errorf("exit=%d results=%v", code, results)
	}
}

// slice-013: --ci
func ciRun(t *testing.T, mode, dir string, env mapEnv, extra ...string) (code int, out, errs string) {
	t.Helper()
	args, home := scanArgs("testdata/" + dir)
	args[0] = mode
	var o, e bytes.Buffer
	code = run(append(args, extra...), env, home, &o, &e)
	return code, o.String(), e.String()
}

// --ci ändert weder Exit-Codes noch stdout, ergänzt aber eine Zusammenfassung auf stderr.
func TestCIModeKeepsStdoutAndExitCodes(t *testing.T) {
	for _, c := range []struct {
		dir      string
		wantCode int
	}{{"good", 0}, {"bad", 1}} {
		codeA, outA, errA := ciRun(t, "local", c.dir, euEnv)
		codeB, outB, errB := ciRun(t, "local", c.dir, euEnv, "--ci")
		if codeA != c.wantCode || codeB != c.wantCode {
			t.Errorf("%s: exit ohne --ci=%d, mit --ci=%d, want %d", c.dir, codeA, codeB, c.wantCode)
		}
		if outA != outB {
			t.Errorf("%s: --ci darf stdout nicht ändern", c.dir)
		}
		if errA != "" || !strings.HasPrefix(errB, "bedrock-eu-check: ") || !strings.Contains(errB, "(Exit "+string(rune('0'+c.wantCode))+")") {
			t.Errorf("%s: stderr ohne --ci=%q, mit --ci=%q", c.dir, errA, errB)
		}
	}
}

func TestCIModeWithJSONAndSARIFKeepsStdoutParsable(t *testing.T) {
	for _, format := range []string{"json", "sarif"} {
		code, out, errs := ciRun(t, "terraform", "bad", mapEnv{}, "--ci", "--format", format)
		var doc map[string]any
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			t.Fatalf("%s: stdout ist kein JSON: %v", format, err)
		}
		if code != 1 || strings.Count(errs, "\n") != 1 || !strings.Contains(errs, "FAIL (Exit 1)") {
			t.Errorf("%s: exit=%d stderr=%q", format, code, errs)
		}
		if strings.Contains(out+errs, "GEHEIM") || strings.Contains(out+errs, "sk-ant") {
			t.Errorf("%s: Secret in der Ausgabe", format)
		}
	}
}

func TestCIModeToolErrorStillExit2(t *testing.T) {
	old := awsBinary
	awsBinary = filepath.Join(t.TempDir(), "gibt-es-nicht")
	t.Cleanup(func() { awsBinary = old })
	code, out, errs := ciRun(t, "all", "good", euEnv, "--ci", "--format", "json")
	if code != 2 || !strings.Contains(errs, "AWS_CLI_MISSING") || !strings.Contains(errs, "(Exit 2)") {
		t.Errorf("exit=%d stderr=%q", code, errs)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Errorf("stdout muss auch bei Tool-Fehler parsebar bleiben: %v", err)
	}
}

func TestCIFlagTakesNoValue(t *testing.T) {
	var o, e bytes.Buffer
	if code := run([]string{"local", "--ci", "--root", t.TempDir()}, euEnv, "", &o, &e); code == 2 && strings.Contains(e.String(), "usage:") {
		t.Errorf("--ci vor weiteren Optionen muss gültig sein: %q", e.String())
	}
}

// slice-019: exclude — ausgeschlossene Pfade werden nicht gemeldet.
func TestExcludeSkipsPaths(t *testing.T) {
	// Ohne exclude meldet der bad-Datensatz Terraform-FAILs.
	if code, out, _ := modeRun(t, "terraform", "bad", mapEnv{}); code != 1 || !strings.Contains(out, "main.tf") {
		t.Fatalf("Ausgangslage: exit=%d stdout=%q", code, out)
	}
	cfg := []string{"--config", "testdata/config/exclude-bad.yml"}
	code, out, errs := modeRun(t, "terraform", "bad", mapEnv{}, cfg...)
	if code != 0 || strings.Contains(out, "[FAIL]") || strings.Contains(out, "main.tf") || errs != "" {
		t.Errorf("terraform mit exclude: exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	// local: Projektdateien (Claude-Projekt-Settings, Devcontainer, Compose, Dockerfile) entfallen;
	// die Benutzer-Settings im Home-Verzeichnis liegen außerhalb des Scan-Roots und bleiben.
	_, out, _ = modeRun(t, "local", "bad", euEnv, cfg...)
	for _, rule := range []string{"LH-FA-CLAUDE-002", "LH-FA-DEVCON-001", "LH-FA-DOCKER-001", "LH-FA-DOCKER-002"} {
		if strings.Contains(out, rule) {
			t.Errorf("%s betrifft nur ausgeschlossene Pfade, erscheint aber:\n%s", rule, out)
		}
	}
	if !strings.Contains(out, "~/.claude/settings.json") {
		t.Errorf("Benutzer-Settings außerhalb des Scan-Roots müssen bleiben:\n%s", out)
	}
}

func TestExcludeInvalidPatternIsToolError(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "c.yml")
	if err := os.WriteFile(cfgFile, []byte("exclude: [\"[abc\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errs := modeRun(t, "terraform", "good", mapEnv{}, "--config", cfgFile)
	if code != 2 || out != "" || errs == "" || strings.Contains(errs, "[abc") {
		t.Errorf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
}

// slice-014: version
func withBuildInfo(t *testing.T, v, c string) {
	t.Helper()
	oldV, oldC := version, commit
	version, commit = v, c
	t.Cleanup(func() { version, commit = oldV, oldC })
}

func TestVersionCommand(t *testing.T) {
	var o, e bytes.Buffer
	if code := run([]string{"version"}, euEnv, "", &o, &e); code != 0 || o.String() != "bedrock-eu-check dev (unknown)\n" || e.Len() != 0 {
		t.Errorf("Defaults: exit=%d stdout=%q stderr=%q", code, o.String(), e.String())
	}
	withBuildInfo(t, "1.2.3", "abc123def456")
	o.Reset()
	if code := run([]string{"version"}, euEnv, "", &o, &e); code != 0 || o.String() != "bedrock-eu-check 1.2.3 (abc123def456)\n" {
		t.Errorf("gesetzt: exit=%d stdout=%q", code, o.String())
	}
}

func TestVersionRejectsExtraArguments(t *testing.T) {
	for _, args := range [][]string{{"version", "--ci"}, {"version", "x"}} {
		var o, e bytes.Buffer
		if code := run(args, euEnv, "", &o, &e); code != 2 || o.Len() != 0 || !strings.Contains(e.String(), "usage:") {
			t.Errorf("%v: exit=%d stdout=%q stderr=%q", args, code, o.String(), e.String())
		}
	}
}

func TestSARIFUsesBuildVersion(t *testing.T) {
	withBuildInfo(t, "9.9.9", "deadbeef0000")
	_, _, _, doc := sarifRun(t, "terraform", "good", mapEnv{})
	drv := doc["runs"].([]any)[0].(map[string]any)["tool"].(map[string]any)["driver"].(map[string]any)
	if drv["version"] != "9.9.9" {
		t.Errorf("driver.version = %v (Version, nicht Commit)", drv["version"])
	}
}

// slice-021: ein in den Claude-Settings konfiguriertes Modell zählt als „verwendet“ (aws-Modus).
func TestAWSModeUsesModelFromSettings(t *testing.T) {
	useFakeAWS(t, awsOK)
	var o, e bytes.Buffer
	args := []string{"aws", "--root", "testdata/usedmodel/project"}
	code := run(args, mapEnv{"AWS_PROFILE": "dev"}, "", &o, &e)
	out := o.String()
	if code != 1 || !strings.Contains(out, "[FAIL] LH-FA-BEDROCK-002") || !strings.Contains(out, ".claude/settings.json") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, e.String())
	}
	// Ohne Modell in den Settings (leeres Root) bleibt es beim bisherigen Verhalten: kein FAIL.
	o.Reset()
	if code := run([]string{"aws", "--root", t.TempDir()}, mapEnv{"AWS_PROFILE": "dev"}, "", &o, &e); code != 0 || strings.Contains(o.String(), "[FAIL] LH-FA-BEDROCK-002") {
		t.Errorf("ohne Settings-Modell: exit=%d stdout=%q", code, o.String())
	}
}

// slice-020: apiKeyHelper → WARN (kein stilles PASS), Befehl erscheint nirgends.
func TestAPIKeyHelperIsWarningWithoutCommand(t *testing.T) {
	for _, format := range []string{"console", "json", "sarif"} {
		var o, e bytes.Buffer
		args := []string{"local", "--root", t.TempDir(), "--format", format}
		code := run(args, euEnv, "testdata/helper/home", &o, &e)
		out := o.String() + e.String()
		if code != 0 {
			t.Errorf("%s: exit=%d (WARN lässt den Lauf nicht scheitern): %s", format, code, out)
		}
		if !strings.Contains(out, "APIKEY_HELPER") {
			t.Errorf("%s: WARN APIKEY_HELPER erwartet:\n%s", format, out)
		}
		if strings.Contains(out, "GEHEIM") || strings.Contains(out, "get-key") {
			t.Errorf("%s: Befehl des Helpers in der Ausgabe:\n%s", format, out)
		}
	}
}

// slice-022: .env-Dateien (LH-FA-LENV-007) — rekursiv, Beispieldateien ausgenommen.
func TestDotenvGood(t *testing.T) {
	code, out, errs := modeRun(t, "local", "good", euEnv)
	if code != 0 || errs != "" {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, errs)
	}
	if !strings.Contains(out, "[PASS] LH-FA-LENV-007 .env") {
		t.Errorf("PASS für die .env erwartet:\n%s", out)
	}
}

func TestDotenvBadRecursiveWithoutSecrets(t *testing.T) {
	for _, format := range []string{"console", "json", "sarif"} {
		code, out, errs := modeRun(t, "local", "bad", euEnv, "--format", format)
		if code != 1 {
			t.Fatalf("%s: exit=%d stderr=%q", format, code, errs)
		}
		for _, want := range []string{"LH-FA-LENV-007", "apps/api/.env.production"} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: %q fehlt:\n%s", format, want, out)
			}
		}
		// Beispieldateien (.env.example) werden nicht geprüft, und kein Secret-Wert erscheint.
		for _, s := range []string{out, errs} {
			if strings.Contains(s, ".env.example") || strings.Contains(s, "GEHEIM") || strings.Contains(s, "sk-ant") {
				t.Errorf("%s: Beispieldatei geprüft oder Secret in der Ausgabe:\n%s", format, s)
			}
		}
	}
	_, out, _ := modeRun(t, "local", "bad", euEnv)
	if !strings.Contains(out, "[FAIL] LH-FA-LENV-007 .env") {
		t.Errorf("FAIL für die .env im Root erwartet:\n%s", out)
	}
}

func TestDotenvExcludeSkipsSubdirectories(t *testing.T) {
	_, out, _ := modeRun(t, "local", "bad", euEnv, "--config", "testdata/config/exclude-apps.yml")
	if strings.Contains(out, "apps/api") {
		t.Errorf("exclude: [apps] muss die verschachtelte .env ausnehmen:\n%s", out)
	}
	if !strings.Contains(out, "LH-FA-LENV-007 .env") {
		t.Errorf("die .env im Root bleibt:\n%s", out)
	}
}

func TestDotenvOnlyInLocalMode(t *testing.T) {
	_, out, _ := modeRun(t, "terraform", "bad", mapEnv{})
	if strings.Contains(out, "LH-FA-LENV-007") {
		t.Errorf("terraform darf keine .env-Regel ausführen:\n%s", out)
	}
}
