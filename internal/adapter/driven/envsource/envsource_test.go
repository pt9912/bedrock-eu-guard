package envsource

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.EnvSourceReader = Reader{}

const secretValue = "sk-ant-GEHEIM-123"

func entries(f *model.EnvFile) map[string]model.EnvEntry {
	m := map[string]model.EnvEntry{}
	for _, e := range f.Env {
		m[e.Key+"@"+e.Scope] = e
	}
	return m
}

func noSecretLeak(t *testing.T, f *model.EnvFile) {
	t.Helper()
	for _, e := range f.Env {
		if strings.Contains(e.Value, "GEHEIM") || strings.Contains(e.Key, "GEHEIM") {
			t.Errorf("Secret-Wert im Ergebnis: %+v", e)
		}
	}
}

func TestDevcontainerJSONC(t *testing.T) {
	in := `{
  // Kommentar mit "Anführungszeichen" und , Komma
  "name": "dev", /* Block
  über zwei Zeilen */
  "containerEnv": {
    "AWS_REGION": "eu-central-1", // Zeilenkommentar
    "CLAUDE_CODE_USE_BEDROCK": "1",
    "ANTHROPIC_API_KEY": "` + secretValue + `",
  },
  "remoteEnv": { "URL": "https://x.example/a//b" },
}`
	f, ferr := parseDevcontainer(".devcontainer/devcontainer.json", []byte(in))
	if ferr != nil {
		t.Fatal(ferr)
	}
	m := entries(f)
	check := func(k, val string, secret bool, line int) {
		t.Helper()
		e, ok := m[k+"@"]
		if !ok || e.Value != val || e.Secret != secret || e.Line != line {
			t.Errorf("%s: got %+v (ok=%v), want val=%q secret=%v line=%d", k, e, ok, val, secret, line)
		}
	}
	check("AWS_REGION", "eu-central-1", false, 6)
	check("CLAUDE_CODE_USE_BEDROCK", "1", false, 7)
	check("ANTHROPIC_API_KEY", "", true, 8)
	check("URL", "https://x.example/a//b", false, 10) // "//" in String ist kein Kommentar
	noSecretLeak(t, f)
	if f.Kind != model.EnvDevcontainer {
		t.Errorf("kind = %v", f.Kind)
	}
}

func TestDevcontainerInvalid(t *testing.T) {
	cases := map[string]struct {
		in   string
		line int
	}{
		"leer":          {"", 1},
		"Array":         {"[1]", 1},
		"abgeschnitten": {"{\n  \"containerEnv\": {\n    \"A\": \"x\"", 3},
		"Syntaxfehler":  {"{\n  \"containerEnv\": {\n    \"A\": ,\n  }\n}", 3},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f, ferr := parseDevcontainer("d.json", []byte(c.in))
			if f != nil || ferr == nil || ferr.Reason != model.ReasonFileInvalid {
				t.Fatalf("want FILE_INVALID, got %+v, %v", f, ferr)
			}
			if ferr.Line != c.line {
				t.Errorf("Zeile: got %d, want %d", ferr.Line, c.line)
			}
		})
	}
}

func TestDevcontainerNoEnvBlocks(t *testing.T) {
	f, ferr := parseDevcontainer("d.json", []byte(`{"name": "x", "containerEnv": null, "remoteEnv": []}`))
	if ferr != nil || len(f.Env) != 0 {
		t.Errorf("got %+v, %v", f, ferr)
	}
}

func TestComposeMapAndList(t *testing.T) {
	in := `services:
  app:
    image: x
    environment:
      AWS_REGION: eu-central-1
      ANTHROPIC_API_KEY: ` + secretValue + `
      CLAUDE_CODE_USE_BEDROCK: 1
  worker:
    environment:
      - AWS_REGION=us-east-1
      - MY_TOKEN=` + secretValue + `
      - NUR_NAME
    env_file: .env
`
	f, ferr := parseCompose("docker-compose.yml", []byte(in))
	if ferr != nil {
		t.Fatal(ferr)
	}
	m := entries(f)
	want := map[string]model.EnvEntry{
		"AWS_REGION@app":              {Key: "AWS_REGION", Value: "eu-central-1", Scope: "app", Line: 5},
		"ANTHROPIC_API_KEY@app":       {Key: "ANTHROPIC_API_KEY", Secret: true, Scope: "app", Line: 6},
		"CLAUDE_CODE_USE_BEDROCK@app": {Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1", Scope: "app", Line: 7},
		"AWS_REGION@worker":           {Key: "AWS_REGION", Value: "us-east-1", Scope: "worker", Line: 10},
		"MY_TOKEN@worker":             {Key: "MY_TOKEN", Secret: true, Scope: "worker", Line: 11},
		"NUR_NAME@worker":             {Key: "NUR_NAME", Scope: "worker", Line: 12},
	}
	if len(m) != len(want) {
		t.Fatalf("Einträge: %+v", f.Env)
	}
	for k, w := range want {
		if m[k] != w {
			t.Errorf("%s: got %+v, want %+v", k, m[k], w)
		}
	}
	noSecretLeak(t, f)
}

func TestComposeAnchorsAndMerge(t *testing.T) {
	in := `x-env: &common
  AWS_REGION: eu-central-1
services:
  a:
    environment: *common
  b:
    environment:
      <<: *common
      EXTRA: "1"
`
	f, ferr := parseCompose("c.yml", []byte(in))
	if ferr != nil {
		t.Fatal(ferr)
	}
	m := entries(f)
	if m["AWS_REGION@a"].Value != "eu-central-1" || m["AWS_REGION@b"].Value != "eu-central-1" || m["EXTRA@b"].Value != "1" {
		t.Errorf("got %+v", f.Env)
	}
}

func TestComposeInvalidAndEmpty(t *testing.T) {
	const secret = "GEHEIM"
	f, ferr := parseCompose("c.yml", []byte("services:\n  a: [unclosed "+secret+"\n"))
	if f != nil || ferr == nil || ferr.Reason != model.ReasonFileInvalid {
		t.Fatalf("want FILE_INVALID, got %+v, %v", f, ferr)
	}
	if strings.Contains(ferr.Error(), secret) {
		t.Errorf("Fehlertext zitiert Wert: %q", ferr.Error())
	}
	for _, in := range []string{"", "# nichts\n", "version: '3'\n", "services: []\n", "- a\n"} {
		f, ferr := parseCompose("c.yml", []byte(in))
		if ferr != nil || len(f.Env) != 0 {
			t.Errorf("%q: got %+v, %v", in, f, ferr)
		}
	}
}

func TestDockerfileEnvForms(t *testing.T) {
	in := `# Kommentar
FROM base AS build
ENV AWS_REGION=eu-central-1 CLAUDE_CODE_USE_BEDROCK=1
ENV OLD_STYLE alter wert mit leerzeichen
ENV QUOTED="a b" \
    ANTHROPIC_API_KEY=` + secretValue + ` \
    # Kommentar in der Fortsetzung
    LAST=x
FROM scratch
ENV AWS_DEFAULT_REGION us-east-1
`
	f, ferr := parseDockerfile("Dockerfile", []byte(in))
	if ferr != nil {
		t.Fatal(ferr)
	}
	m := entries(f)
	want := map[string]model.EnvEntry{
		"AWS_REGION@build":              {Key: "AWS_REGION", Value: "eu-central-1", Scope: "build", Line: 3},
		"CLAUDE_CODE_USE_BEDROCK@build": {Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1", Scope: "build", Line: 3},
		"OLD_STYLE@build":               {Key: "OLD_STYLE", Value: "alter wert mit leerzeichen", Scope: "build", Line: 4},
		"QUOTED@build":                  {Key: "QUOTED", Value: "a b", Scope: "build", Line: 5},
		"ANTHROPIC_API_KEY@build":       {Key: "ANTHROPIC_API_KEY", Secret: true, Scope: "build", Line: 5},
		"LAST@build":                    {Key: "LAST", Value: "x", Scope: "build", Line: 5},
		"AWS_DEFAULT_REGION@1":          {Key: "AWS_DEFAULT_REGION", Value: "us-east-1", Scope: "1", Line: 10},
	}
	if len(m) != len(want) {
		t.Fatalf("Einträge (%d): %+v", len(f.Env), f.Env)
	}
	for k, w := range want {
		if m[k] != w {
			t.Errorf("%s: got %+v, want %+v", k, m[k], w)
		}
	}
	noSecretLeak(t, f)
}

func TestDockerfileCRLFAndNoStage(t *testing.T) {
	f, ferr := parseDockerfile("Dockerfile", []byte("ENV AWS_REGION=eu-central-1\r\nFROM x\r\n"))
	if ferr != nil || len(f.Env) != 1 || f.Env[0].Line != 1 || f.Env[0].Scope != "" {
		t.Errorf("got %+v, %v", f, ferr)
	}
}

// Variablenreferenzen sind keine Region-Werte (kein Fehlalarm).
func TestInterpolatedRegionsDropped(t *testing.T) {
	d, _ := parseDockerfile("Dockerfile", []byte("ENV AWS_REGION=${REGION}\nENV AWS_DEFAULT_REGION=$R\nENV OTHER=${X}\n"))
	if len(d.Env) != 1 || d.Env[0].Key != "OTHER" {
		t.Errorf("Dockerfile: %+v", d.Env)
	}
	c, _ := parseCompose("c.yml", []byte("services:\n  a:\n    environment:\n      AWS_REGION: ${AWS_REGION:-us-east-1}\n"))
	if len(c.Env) != 0 {
		t.Errorf("Compose: %+v", c.Env)
	}
	j, _ := parseDevcontainer("d.json", []byte(`{"containerEnv": {"AWS_REGION": "${localEnv:AWS_REGION}"}}`))
	if len(j.Env) != 0 {
		t.Errorf("Devcontainer: %+v", j.Env)
	}
}

func TestStripJSONCPreservesLayout(t *testing.T) {
	in := []byte("{\n  // c\n  \"a\": \"//x\", /* b\n */\n  \"b\": [1, 2,],\n}")
	out := stripJSONC(in)
	if len(out) != len(in) || strings.Count(string(out), "\n") != strings.Count(string(in), "\n") {
		t.Fatalf("Länge/Zeilen verändert:\n%q\n%q", in, out)
	}
	if !strings.Contains(string(out), `"//x"`) || strings.Contains(string(out), "// c") {
		t.Errorf("Kommentare falsch behandelt: %q", out)
	}
}

func TestReaderFilesAndPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("ENV AWS_REGION=eu-central-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "c.yml"), []byte("services: [unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New(root)

	if f, err := r.ReadEnvSource(model.EnvDockerfile, "fehlt"); f != nil || err != nil {
		t.Errorf("fehlend: %+v, %v", f, err)
	}
	f, err := r.ReadEnvSource(model.EnvDockerfile, "Dockerfile")
	if err != nil || f == nil || f.Path != "Dockerfile" || len(f.Env) != 1 {
		t.Errorf("Dockerfile: %+v, %v", f, err)
	}
	_, err = r.ReadEnvSource(model.EnvCompose, "c.yml")
	var fe *model.FileError
	if !errors.As(err, &fe) || fe.Reason != model.ReasonFileInvalid {
		t.Errorf("kaputt: %v", err)
	}
	for _, p := range []string{"../x", "/etc/passwd"} {
		if _, err := r.ReadEnvSource(model.EnvDockerfile, p); !errors.As(err, &fe) {
			t.Errorf("%s muss abgelehnt werden: %v", p, err)
		}
	}
	// Verzeichnis statt Datei → nicht lesbar
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ReadEnvSource(model.EnvDockerfile, "dir"); !errors.As(err, &fe) || fe.Reason != model.ReasonFileUnreadable {
		t.Errorf("Verzeichnis: %v", err)
	}
	if _, err := r.ReadEnvSource(model.EnvSourceKind("x"), "Dockerfile"); err == nil {
		t.Error("unbekannte Art muss ein Fehler sein")
	}
}
