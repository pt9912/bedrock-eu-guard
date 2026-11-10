package awscli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.AWSReader = Reader{}

const leak = "SECRET-LEAK-TEXT-12345"

// fakeAWS schreibt ein Shell-Skript als Ersatz für das aws-Binary. body ist
// der Skriptinhalt nach dem Shebang; $LOG zeigt auf eine Protokolldatei.
func fakeAWS(t *testing.T, body string) (bin, log string) {
	t.Helper()
	dir := t.TempDir()
	log = filepath.Join(dir, "log")
	bin = filepath.Join(dir, "aws")
	script := "#!/bin/sh\nLOG='" + log + "'\n" + body + "\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin, log
}

func asAWSError(t *testing.T, err error) *model.AWSError {
	t.Helper()
	var ae *model.AWSError
	if !errors.As(err, &ae) {
		t.Fatalf("want *model.AWSError, got %v", err)
	}
	if ae.Op != "sts:GetCallerIdentity" {
		t.Errorf("Op = %q", ae.Op)
	}
	if strings.Contains(err.Error(), leak) {
		t.Errorf("Fehlertext enthält CLI-Ausgabe: %q", err.Error())
	}
	return ae
}

func TestIdentitySuccessArgsAndEnv(t *testing.T) {
	bin, log := fakeAWS(t, `echo "$@" > "$LOG"; env >> "$LOG"
echo '{"UserId":"AIDAEXAMPLE","Account":"123456789012","Arn":"arn:aws:iam::123456789012:user/tester"}'`)
	t.Setenv("AWS_PROFILE", "dev")
	t.Setenv("ANTHROPIC_API_KEY", leak)
	t.Setenv("AWS_PAGER", "less")

	id, err := New(bin, "eu-central-1").Identity()
	if err != nil {
		t.Fatal(err)
	}
	if id.Account != "123456789012" || id.ARN != "arn:aws:iam::123456789012:user/tester" || id.UserID != "AIDAEXAMPLE" {
		t.Errorf("got %+v", id)
	}
	b, _ := os.ReadFile(log)
	got := string(b)
	if !strings.HasPrefix(got, "sts get-caller-identity --output json --region eu-central-1\n") {
		t.Errorf("Argumente: %q", strings.SplitN(got, "\n", 2)[0])
	}
	if !strings.Contains(got, "AWS_PROFILE=dev") || !strings.Contains(got, "AWS_PAGER=\n") {
		t.Errorf("AWS_* muss durchgereicht, Pager abgeschaltet sein:\n%s", got)
	}
	if strings.Contains(got, "ANTHROPIC_API_KEY") || strings.Contains(got, leak) || strings.Contains(got, "AWS_PAGER=less") {
		t.Errorf("fremde Variablen dürfen die CLI nicht erreichen:\n%s", got)
	}
}

func TestIdentityWithoutRegionOmitsFlag(t *testing.T) {
	bin, log := fakeAWS(t, `echo "$@" > "$LOG"; echo '{"Account":"1","Arn":"a","UserId":"u"}'`)
	if _, err := New(bin, "").Identity(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(log)
	if strings.Contains(string(b), "--region") {
		t.Errorf("Argumente: %q", b)
	}
}

func TestErrorClassification(t *testing.T) {
	cases := map[string]struct {
		stderr string
		want   string
	}{
		"keine Credentials":  {"Unable to locate credentials. You can configure credentials by running \"aws configure\".", model.AWSNoCredentials},
		"abgelaufen":         {"An error occurred (ExpiredToken) when calling the GetCallerIdentity operation: The security token included in the request is expired", model.AWSNoCredentials},
		"ungueltiger Key":    {"An error occurred (InvalidClientTokenId) when calling the GetCallerIdentity operation", model.AWSNoCredentials},
		"Profil fehlt":       {"The config profile (dev) could not be found", model.AWSNoCredentials},
		"SSO abgelaufen":     {"Error loading SSO Token: Token for x does not exist", model.AWSNoCredentials},
		"Zugriff verweigert": {"An error occurred (AccessDenied) when calling the GetCallerIdentity operation", model.AWSAccessDenied},
		"nicht erreichbar":   {"Could not connect to the endpoint URL: \"https://sts.eu-central-1.amazonaws.com/\"", model.AWSUnreachable},
		"unbekannter Fehler": {"irgendetwas anderes", model.AWSUnreachable},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			bin, _ := fakeAWS(t, "echo '"+strings.ReplaceAll(c.stderr, "'", "")+" "+leak+"' >&2; exit 255")
			id, err := New(bin, "eu-central-1").Identity()
			if id != nil {
				t.Fatalf("keine Identität erwartet: %+v", id)
			}
			if ae := asAWSError(t, err); ae.Reason != c.want {
				t.Errorf("Reason = %q, want %q", ae.Reason, c.want)
			}
		})
	}
}

func TestBinaryMissing(t *testing.T) {
	_, err := New(filepath.Join(t.TempDir(), "gibt-es-nicht"), "").Identity()
	if ae := asAWSError(t, err); ae.Reason != model.AWSCLIMissing {
		t.Errorf("Reason = %q", ae.Reason)
	}
	t.Setenv("PATH", t.TempDir())
	_, err = New("", "").Identity()
	if ae := asAWSError(t, err); ae.Reason != model.AWSCLIMissing {
		t.Errorf("Default-Binary im leeren PATH: Reason = %q", ae.Reason)
	}
}

func TestInvalidOutputIsUnreachable(t *testing.T) {
	for _, body := range []string{"echo 'kein json " + leak + "'", "echo '{}'", "echo '{\"Account\":\"\"}'"} {
		bin, _ := fakeAWS(t, body)
		_, err := New(bin, "").Identity()
		if ae := asAWSError(t, err); ae.Reason != model.AWSUnreachable {
			t.Errorf("%s: Reason = %q", body, ae.Reason)
		}
	}
}

func TestTimeout(t *testing.T) {
	old := timeout
	timeout = 200 * time.Millisecond
	t.Cleanup(func() { timeout = old })
	bin, _ := fakeAWS(t, "sleep 5")
	start := time.Now()
	_, err := New(bin, "").Identity()
	if ae := asAWSError(t, err); ae.Reason != model.AWSUnreachable {
		t.Errorf("Reason = %q", ae.Reason)
	}
	if time.Since(start) > 4*time.Second {
		t.Errorf("Timeout griff nicht rechtzeitig: %v", time.Since(start))
	}
}
