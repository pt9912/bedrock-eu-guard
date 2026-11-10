package claudesettings_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/claudesettings"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.ClaudeSettingsReader = claudesettings.Reader{}

const secretValue = "sk-ant-GEHEIM-123"

const sample = `{
  "model": "eu.anthropic.claude-sonnet-4-6",
  "env": {
    "AWS_REGION": "eu-central-1",
    "CLAUDE_CODE_USE_BEDROCK": 1,
    "ANTHROPIC_API_KEY": "` + secretValue + `",
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
    "MY_FLAG": true
  }
}`

func TestParseEnvLinesAndStringify(t *testing.T) {
	s, ferr := claudesettings.Parse("p", []byte(sample))
	if ferr != nil {
		t.Fatal(ferr)
	}
	if s.Model != "eu.anthropic.claude-sonnet-4-6" || s.Path != "p" {
		t.Errorf("model/path: %+v", s)
	}
	want := map[string]struct {
		val    string
		secret bool
		line   int
	}{
		"AWS_REGION":              {"eu-central-1", false, 4},
		"CLAUDE_CODE_USE_BEDROCK": {"1", false, 5}, // Zahl → "1"
		"ANTHROPIC_API_KEY":       {"", true, 6},
		"ANTHROPIC_BASE_URL":      {"https://api.anthropic.com", false, 7},
		"MY_FLAG":                 {"true", false, 8},
	}
	if len(s.Env) != len(want) {
		t.Fatalf("env: %+v", s.Env)
	}
	for _, e := range s.Env {
		w := want[e.Key]
		if e.Value != w.val || e.Secret != w.secret || e.Line != w.line {
			t.Errorf("%s: got %+v, want %+v", e.Key, e, w)
		}
	}
}

func TestParseHitsOnlyPatternAndLine(t *testing.T) {
	s, _ := claudesettings.Parse("p", []byte(sample))
	want := []model.Hit{
		{Pattern: model.HitAPIKey, Line: 6},
		{Pattern: model.HitAnthropicHost, Line: 7},
		{Pattern: model.HitBaseURL, Line: 7},
	}
	if !reflect.DeepEqual(s.Hits, want) {
		t.Errorf("hits: got %+v, want %+v", s.Hits, want)
	}
}

func TestNoSecretValueAnywhere(t *testing.T) {
	s, _ := claudesettings.Parse("p", []byte(sample))
	if strings.Contains(strings.Repeat("|", 1)+dump(s), secretValue) {
		t.Fatal("Secret-Wert im geparsten Ergebnis")
	}
}

func dump(s *model.ClaudeSettings) string {
	var b strings.Builder
	b.WriteString(s.Path + s.Model)
	for _, e := range s.Env {
		b.WriteString(e.Key + e.Value)
	}
	for _, h := range s.Hits {
		b.WriteString(h.Pattern)
	}
	return b.String()
}

func TestSecretKeys(t *testing.T) {
	secrets := []string{
		"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN", "AWS_BEARER_TOKEN_BEDROCK", "DB_PASSWORD", "GCP_CREDENTIALS",
		"some_private_key", "MY_KEY", "apiKeyHelper",
	}
	plain := []string{"AWS_REGION", "CLAUDE_CODE_USE_BEDROCK", "ANTHROPIC_BASE_URL", "ANTHROPIC_DEFAULT_SONNET_MODEL", "HOME"}
	var parts []string
	for _, k := range append(append([]string{}, secrets...), plain...) {
		parts = append(parts, `"`+k+`": "`+secretValue+`"`)
	}
	s, ferr := claudesettings.Parse("p", []byte(`{"env": {`+strings.Join(parts, ",")+`}}`))
	if ferr != nil {
		t.Fatal(ferr)
	}
	got := map[string]model.EnvEntry{}
	for _, e := range s.Env {
		got[e.Key] = e
	}
	for _, k := range secrets {
		if !got[k].Secret || got[k].Value != "" {
			t.Errorf("%s soll Secret mit leerem Wert sein: %+v", k, got[k])
		}
	}
	for _, k := range plain {
		if got[k].Secret || got[k].Value != secretValue {
			t.Errorf("%s soll Klartext sein: %+v", k, got[k])
		}
	}
}

func TestParseInvalid(t *testing.T) {
	cases := map[string]struct {
		in   string
		line int
	}{
		"leer":              {"", 1},
		"kein Objekt":       {"[1,2]", 1},
		"Syntaxfehler":      {"{\n  \"env\": {\n    \"A\": ,\n  }\n}", 3},
		"abgeschnitten":     {"{\"env\": {\"A\": \"x\"", 1},
		"Text nach Objekt":  {"{}\nmüll", 2},
		"Geheimnis im Text": {"{ \"x\": " + secretValue + " }", 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s, ferr := claudesettings.Parse("p", []byte(c.in))
			if s != nil || ferr == nil {
				t.Fatalf("want FileError, got %+v, %v", s, ferr)
			}
			if ferr.Reason != model.ReasonFileInvalid || ferr.Path != "p" {
				t.Errorf("got %+v", ferr)
			}
			if ferr.Line != c.line {
				t.Errorf("Zeile: got %d, want %d", ferr.Line, c.line)
			}
			if strings.Contains(ferr.Error(), secretValue) || strings.Contains(ferr.Error(), "müll") {
				t.Errorf("Fehlertext zitiert Inhalt: %q", ferr.Error())
			}
		})
	}
}

func TestParseCRLFLines(t *testing.T) {
	in := "{\r\n  \"env\": {\r\n    \"AWS_REGION\": \"eu-central-1\"\r\n  }\r\n}"
	s, ferr := claudesettings.Parse("p", []byte(in))
	if ferr != nil || s.Env[0].Line != 3 {
		t.Errorf("got %+v, %v", s, ferr)
	}
}

func TestParseEnvNotObject(t *testing.T) {
	for _, in := range []string{`{"env": null}`, `{"env": [1, 2]}`, `{"env": "x"}`, `{"other": {"a": 1}}`} {
		s, ferr := claudesettings.Parse("p", []byte(in))
		if ferr != nil || len(s.Env) != 0 {
			t.Errorf("%s: got %+v, %v", in, s, ferr)
		}
	}
}

func TestReaderFiles(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	write := func(base, rel, content string) {
		full := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := claudesettings.New(root, home)

	if s, err := r.ReadUser(); s != nil || err != nil {
		t.Errorf("fehlende User-Datei: got %+v, %v", s, err)
	}
	write(home, ".claude/settings.json", `{"model": "m"}`)
	s, err := r.ReadUser()
	if err != nil || s == nil || s.Path != "~/.claude/settings.json" || s.Model != "m" {
		t.Errorf("User: got %+v, %v", s, err)
	}

	if s, err := r.ReadProject(".claude/settings.json"); s != nil || err != nil {
		t.Errorf("fehlende Projektdatei: got %+v, %v", s, err)
	}
	write(root, ".claude/settings.local.json", `{ kaputt`)
	_, err = r.ReadProject(".claude/settings.local.json")
	fe, ok := err.(*model.FileError)
	if !ok || fe.Reason != model.ReasonFileInvalid || fe.Path != ".claude/settings.local.json" {
		t.Errorf("kaputt: got %v", err)
	}

	if _, err := r.ReadProject("../x"); err == nil {
		t.Error("Pfad außerhalb des Roots muss abgelehnt werden")
	}
	if s, err := claudesettings.New(root, "").ReadUser(); s != nil || err != nil {
		t.Errorf("ohne Home: got %+v, %v", s, err)
	}
}

const helperSecret = "GEHEIM-HELPER-BEFEHL"

// apiKeyHelper: nur die Zeile des Top-Level-Schlüssels; der Befehl erscheint nirgends.
func TestAPIKeyHelperLineOnly(t *testing.T) {
	in := `{
  "model": "m",
  "apiKeyHelper": "/usr/local/bin/get-key --token=` + helperSecret + `",
  "env": {
    "AWS_REGION": "eu-central-1"
  }
}`
	s, ferr := claudesettings.Parse("p", []byte(in))
	if ferr != nil {
		t.Fatal(ferr)
	}
	if s.APIKeyHelperLine != 3 {
		t.Errorf("APIKeyHelperLine = %d, want 3", s.APIKeyHelperLine)
	}
	if strings.Contains(dump(s), helperSecret) {
		t.Errorf("Befehl im Ergebnis: %s", dump(s))
	}
	for _, e := range s.Env {
		if strings.Contains(e.Value, helperSecret) || strings.Contains(e.Key, "apiKeyHelper") {
			t.Errorf("apiKeyHelper darf nicht in Env stehen: %+v", e)
		}
	}
}

func TestAPIKeyHelperAbsentOrNotTopLevel(t *testing.T) {
	cases := map[string]string{
		"fehlt":            `{"model": "m"}`,
		"nur im env-Block": `{"env": {"apiKeyHelper": "` + helperSecret + `"}}`,
		"verschachtelt":    `{"other": {"apiKeyHelper": "x"}}`,
	}
	for name, in := range cases {
		s, ferr := claudesettings.Parse("p", []byte(in))
		if ferr != nil || s.APIKeyHelperLine != 0 {
			t.Errorf("%s: line=%d err=%v", name, s.APIKeyHelperLine, ferr)
		}
	}
	// Im env-Block bleibt der Schlüssel ein Secret: gesetzt, aber ohne Wert.
	s, _ := claudesettings.Parse("p", []byte(cases["nur im env-Block"]))
	if len(s.Env) != 1 || !s.Env[0].Secret || s.Env[0].Value != "" {
		t.Errorf("env: %+v", s.Env)
	}
}

func TestAPIKeyHelperNonStringValueAndCRLF(t *testing.T) {
	s, ferr := claudesettings.Parse("p", []byte("{\r\n  \"apiKeyHelper\": [\"a\", {\"b\": 1}]\r\n}"))
	if ferr != nil || s.APIKeyHelperLine != 2 {
		t.Errorf("line=%d err=%v", s.APIKeyHelperLine, ferr)
	}
}
