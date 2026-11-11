package envsource

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

func TestDotenvSyntax(t *testing.T) {
	in := "\ufeff# Kommentar\r\n" +
		"AWS_REGION=eu-central-1\r\n" +
		"export CLAUDE_CODE_USE_BEDROCK=1   # Inline-Kommentar\n" +
		"  SPACED = mit leerzeichen  \n" +
		"QUOTED=\"a # kein Kommentar\" # aber hier\n" +
		"SINGLE='roh $HOME \\n bleibt'\n" +
		"ESC=\"zeile \\\"zwei\\\" \\\\ ende\"\n" +
		"MULTI=\"erste\n" +
		"zweite\" # Rest\n" +
		"NUR_NAME\n" +
		"EMPTY=\n" +
		"ANTHROPIC_API_KEY=" + secretValue + "\n" +
		"MY_TOKEN=\"" + secretValue + "\"\n"
	f, ferr := parseDotenv(".env", []byte(in))
	if ferr != nil {
		t.Fatal(ferr)
	}
	want := map[string]model.EnvEntry{
		"AWS_REGION":              {Key: "AWS_REGION", Value: "eu-central-1", Line: 2},
		"CLAUDE_CODE_USE_BEDROCK": {Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1", Line: 3},
		"SPACED":                  {Key: "SPACED", Value: "mit leerzeichen", Line: 4},
		"QUOTED":                  {Key: "QUOTED", Value: "a # kein Kommentar", Line: 5},
		"SINGLE":                  {Key: "SINGLE", Value: `roh $HOME \n bleibt`, Line: 6},
		"ESC":                     {Key: "ESC", Value: `zeile "zwei" \ ende`, Line: 7},
		"MULTI":                   {Key: "MULTI", Value: "erste\nzweite", Line: 8},
		"EMPTY":                   {Key: "EMPTY", Value: "", Line: 11},
		"ANTHROPIC_API_KEY":       {Key: "ANTHROPIC_API_KEY", Secret: true, Line: 12},
		"MY_TOKEN":                {Key: "MY_TOKEN", Secret: true, Line: 13},
	}
	got := map[string]model.EnvEntry{}
	for _, e := range f.Env {
		got[e.Key] = e
	}
	if len(got) != len(want) {
		t.Fatalf("Einträge (%d): %+v", len(f.Env), f.Env)
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %+v, want %+v", k, got[k], w)
		}
	}
	if f.Kind != dotenvKind {
		t.Errorf("kind = %v", f.Kind)
	}
	noSecretLeak(t, f)
}

func TestDotenvInvalid(t *testing.T) {
	cases := map[string]struct {
		in   string
		line int
	}{
		"Zeile ohne Gleichheitszeichen":   {"AWS_REGION=eu-central-1\nkein gueltiger schluessel mit leerzeichen\n", 2},
		"ungueltiger Schluessel":          {"1ABC=x\n", 1},
		"Schluessel mit Sonderzeichen":    {"A$B=x\n", 1},
		"nicht geschlossenes Zitat":       {"OK=1\nK=\"nie geschlossen " + secretValue + "\nweiter\n", 2},
		"nicht geschlossenes Einzelzitat": {"K='offen\n", 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f, ferr := parseDotenv(".env", []byte(c.in))
			if f != nil || ferr == nil || ferr.Reason != model.ReasonFileInvalid || ferr.Path != ".env" {
				t.Fatalf("want FILE_INVALID, got %+v, %+v", f, ferr)
			}
			if ferr.Line != c.line {
				t.Errorf("Zeile: got %d, want %d", ferr.Line, c.line)
			}
			if strings.Contains(ferr.Error(), "GEHEIM") || strings.Contains(ferr.Error(), "nie geschlossen") {
				t.Errorf("Fehlertext zitiert Inhalt: %q", ferr.Error())
			}
		})
	}
}

func TestDotenvEmptyAndOnlyComments(t *testing.T) {
	for _, in := range []string{"", "\n\n", "# nur Kommentar\n", "NUR_NAME\n"} {
		f, ferr := parseDotenv(".env", []byte(in))
		if ferr != nil || len(f.Env) != 0 {
			t.Errorf("%q: %+v, %v", in, f, ferr)
		}
	}
}

// Variablenreferenzen sind keine Region-Werte (kein Fehlalarm).
func TestDotenvInterpolatedRegionDropped(t *testing.T) {
	f, _ := parseDotenv(".env", []byte("AWS_REGION=${REGION}\nAWS_DEFAULT_REGION=$R\nOTHER=${X}\n"))
	if len(f.Env) != 1 || f.Env[0].Key != "OTHER" {
		t.Errorf("got %+v", f.Env)
	}
}

func TestIsDotenvName(t *testing.T) {
	yes := []string{".env", ".env.local", ".env.production", ".env.development.local", ".env.test"}
	no := []string{".env.example", ".env.sample", ".env.template", ".env.local.example", "env", "dotenv", ".envrc", "x.env", ".environment"}
	for _, n := range yes {
		if !isDotenvName(n) {
			t.Errorf("%s soll erkannt werden", n)
		}
	}
	for _, n := range no {
		if isDotenvName(n) {
			t.Errorf("%s soll nicht erkannt werden", n)
		}
	}
}

func TestFindDotenv(t *testing.T) {
	root := t.TempDir()
	write := func(rel string) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("A=1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, rel := range []string{".env", "apps/api/.env", "apps/api/.env.production", "apps/web/.env.example", "docs/env.txt"} {
		write(rel)
	}
	for _, skip := range []string{".git", ".terraform", "vendor", "node_modules", ".tmp"} {
		write(skip + "/.env")
	}
	if err := os.Symlink(filepath.Join(root, "apps"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	r := New(root)
	got, err := r.Find(dotenvKind)
	want := []string{".env", "apps/api/.env", "apps/api/.env.production"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Find: got %v, %v; want %v", got, err, want)
	}
	if other, err := r.Find(model.EnvCompose); err != nil || len(other) != 0 {
		t.Errorf("andere Arten liefert Find nicht: %v, %v", other, err)
	}
}

func TestReaderReadsDotenvAndMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("AWS_REGION=us-east-1\nANTHROPIC_API_KEY="+secretValue+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New(root)
	f, err := r.ReadEnvSource(dotenvKind, ".env")
	if err != nil || f == nil || len(f.Env) != 2 {
		t.Fatalf("%+v, %v", f, err)
	}
	noSecretLeak(t, f)
	if f, err := r.ReadEnvSource(dotenvKind, "nicht/da/.env"); f != nil || err != nil {
		t.Errorf("fehlend: %+v, %v", f, err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.bad"), []byte("kaputt ohne gleich\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = r.ReadEnvSource(dotenvKind, ".env.bad")
	var fe *model.FileError
	if !errors.As(err, &fe) || fe.Reason != model.ReasonFileInvalid || fe.Line != 1 {
		t.Errorf("kaputt: %v", err)
	}
}
