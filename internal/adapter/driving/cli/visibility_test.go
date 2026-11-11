package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// slice-028: visibilityRegions mit Fake-aws. us-east-1 zeigt ein blockiertes Profil,
// us-west-2 ist nicht lesbar; stderr des Fake enthält Text, der nie erscheinen darf.
const awsVisibility = `case "$*" in
  *"bedrock list-inference-profiles"*"us-east-1"*) echo '{"inferenceProfileSummaries":[{"inferenceProfileId":"us.anthropic.claude-opus-4","inferenceProfileName":"US Opus","type":"SYSTEM_DEFINED"}]}' ;;
  *"bedrock list-inference-profiles"*"us-west-2"*) echo 'An error occurred (AccessDeniedException) LEAK-STDERR-TEXT' >&2; exit 254 ;;
  *) ` + awsOK + ` ;;
esac`

func visibilityConfig(t *testing.T, regions string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "bedrock-eu-check.yml")
	if err := os.WriteFile(p, []byte("visibilityRegions: "+regions+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestVisibilityRegionsReportWarnings(t *testing.T) {
	useFakeAWS(t, awsVisibility)
	env := mapEnv{"AWS_PROFILE": "dev"}
	code, out, errs := awsRun(t, env, "--config", visibilityConfig(t, "[us-east-1, us-west-2]"))
	if code != 0 {
		t.Errorf("Zusatzregionen liefern nur WARN, nie FAIL: exit=%d\n%s\n%s", code, out, errs)
	}
	if !strings.Contains(out, "[WARN] LH-FA-BEDROCK-002") || !strings.Contains(out, "PROFILES_VISIBLE_NOT_ALLOWED") || !strings.Contains(out, "us-east-1") {
		t.Errorf("WARN für sichtbares blockiertes Profil in us-east-1 erwartet:\n%s", out)
	}
	if !strings.Contains(out, "VISIBILITY_UNKNOWN") || !strings.Contains(out, "us-west-2") {
		t.Errorf("WARN VISIBILITY_UNKNOWN für us-west-2 erwartet:\n%s", out)
	}
	if strings.Contains(out+errs, "LEAK-STDERR-TEXT") {
		t.Errorf("stderr des aws-Aufrufs darf nie erscheinen:\n%s\n%s", out, errs)
	}
	if strings.Contains(out, "[FAIL]") {
		t.Errorf("kein FAIL erwartet:\n%s", out)
	}
}

// Ohne das Feld bleibt das Verhalten unverändert: keine Meldung zu Zusatzregionen.
func TestVisibilityRegionsDefaultUnchanged(t *testing.T) {
	useFakeAWS(t, awsVisibility)
	_, out, _ := awsRun(t, mapEnv{"AWS_PROFILE": "dev"})
	if strings.Contains(out, "VISIBILITY_UNKNOWN") || strings.Contains(out, "us-east-1") || strings.Contains(out, "us-west-2") {
		t.Errorf("ohne visibilityRegions keine Zusatzregionen:\n%s", out)
	}
}

// Eine Region in allowedRegions und visibilityRegions ist ein Konfigurationsfehler (Exit 2).
func TestVisibilityRegionsOverlapIsConfigError(t *testing.T) {
	useFakeAWS(t, awsVisibility)
	code, _, errs := awsRun(t, mapEnv{"AWS_PROFILE": "dev"}, "--config", visibilityConfig(t, "[eu-central-1]"))
	if code != 2 || !strings.Contains(errs, "visibilityRegions") || strings.Contains(errs, "eu-central-1") {
		t.Errorf("exit=%d stderr=%q", code, errs)
	}
}
