package configyaml_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/configyaml"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.ConfigLoader = configyaml.Loader{}

func TestParseFullDocument(t *testing.T) {
	in := `
allowedRegions: [eu-central-1, eu-west-1]
allowedAccounts: ["123456789012"]
allowedModelPatterns: ['^eu\.']
blockedModelPatterns: ['^us\.', 'us-east-1']
allowedProfiles: ['eu.*']
mantle: block
`
	p, err := configyaml.Parse("c.yml", []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	want := &model.PartialConfig{
		AllowedRegions:       []string{"eu-central-1", "eu-west-1"},
		AllowedAccounts:      []string{"123456789012"},
		AllowedModelPatterns: []string{`^eu\.`},
		BlockedModelPatterns: []string{`^us\.`, "us-east-1"},
		AllowedProfiles:      []string{"eu.*"},
		Mantle:               "block",
	}
	if !reflect.DeepEqual(p, want) {
		t.Errorf("got %+v, want %+v", p, want)
	}
}

// Abwesend (nil) und explizit leer ([]) müssen unterscheidbar bleiben,
// sonst greifen die Defaults falsch (LH-NF-007).
func TestAbsentVersusExplicitEmpty(t *testing.T) {
	p, err := configyaml.Parse("c.yml", []byte("allowedAccounts: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	if p.AllowedAccounts == nil || len(p.AllowedAccounts) != 0 {
		t.Errorf("allowedAccounts: [] muss nicht-nil leer sein, got %#v", p.AllowedAccounts)
	}
	if p.AllowedRegions != nil || p.Mantle != "" {
		t.Errorf("abwesende Felder müssen nil/leer bleiben: %+v", p)
	}
}

func TestEmptyDocumentMeansDefaults(t *testing.T) {
	for _, in := range []string{"", "# nur Kommentar\n", "{}\n"} {
		p, err := configyaml.Parse("c.yml", []byte(in))
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if !reflect.DeepEqual(p, &model.PartialConfig{}) {
			t.Errorf("%q: got %+v", in, p)
		}
	}
}

func TestInvalidDocuments(t *testing.T) {
	const secret = "GEHEIMER-WERT"
	cases := map[string]struct {
		in   string
		line int
	}{
		"unbekanntes Feld":   {"allowedRegions: [eu-central-1]\nfoo: " + secret + "\n", 2},
		"Syntaxfehler":       {"allowedRegions: [eu-central-1\nmantle: warn\n", 0},
		"falscher Typ":       {"allowedRegions: " + secret + "\n", 1},
		"zweites Dokument":   {"mantle: warn\n---\nmantle: allow\n", 0},
		"mantle kein String": {"mantle: [a, b]\n", 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p, err := configyaml.Parse("c.yml", []byte(c.in))
			var fe *model.FileError
			if p != nil || !errors.As(err, &fe) {
				t.Fatalf("want FileError, got %+v, %v", p, err)
			}
			if fe.Reason != model.ReasonFileInvalid || fe.Path != "c.yml" {
				t.Errorf("got %+v", fe)
			}
			if c.line > 0 && fe.Line != c.line {
				t.Errorf("Zeile: got %d, want %d", fe.Line, c.line)
			}
			if strings.Contains(err.Error(), secret) {
				t.Errorf("Fehlertext zitiert Wert: %q", err.Error())
			}
		})
	}
}

func TestLoaderFiles(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "fehlt.yml")
	if p, err := configyaml.New(missing).Load(); p != nil || err != nil {
		t.Errorf("fehlende Datei: got %+v, %v", p, err)
	}

	good := filepath.Join(dir, "gut.yml")
	if err := os.WriteFile(good, []byte("mantle: allow\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := configyaml.New(good).Load()
	if err != nil || p == nil || p.Mantle != "allow" {
		t.Errorf("gut: got %+v, %v", p, err)
	}

	bad := filepath.Join(dir, "kaputt.yml")
	if err := os.WriteFile(bad, []byte("foo: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = configyaml.New(bad).Load()
	var fe *model.FileError
	if !errors.As(err, &fe) || fe.Reason != model.ReasonFileInvalid {
		t.Errorf("kaputt: got %v", err)
	}

	// Verzeichnis statt Datei: vorhanden, aber nicht lesbar.
	_, err = configyaml.New(dir).Load()
	if !errors.As(err, &fe) || fe.Reason != model.ReasonFileUnreadable {
		t.Errorf("Verzeichnis: got %v", err)
	}
}

// "allowedAccounts:" ohne Wert (YAML null) wird wie ein fehlendes Feld behandelt.
func TestNullFieldIsAbsent(t *testing.T) {
	p, err := configyaml.Parse("c.yml", []byte("allowedAccounts:\nmantle: warn\n"))
	if err != nil {
		t.Fatal(err)
	}
	if p.AllowedAccounts != nil {
		t.Errorf("null muss nil bleiben, got %#v", p.AllowedAccounts)
	}
}
