// Package configyaml implementiert den ConfigLoader-Port (ARC-006): liest
// bedrock-eu-check.yml (SPEC-020) und bildet es 1:1 auf model.PartialConfig
// ab. Defaults und Validierung liegen im Kern (model.Resolve).
package configyaml

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Loader liest eine Konfigurationsdatei.
type Loader struct{ path string }

// New liefert einen Loader für die Datei path.
func New(path string) Loader { return Loader{path: path} }

// Load liest die Datei. Fehlt sie, gilt (nil, nil) = Defaults. Ist sie nicht
// lesbar oder ungültig, liefert Load ein *model.FileError (ohne Dateiinhalt).
func (l Loader) Load() (*model.PartialConfig, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, &model.FileError{Path: l.path, Reason: model.ReasonFileUnreadable}
	}
	return Parse(l.path, data)
}

// document spiegelt das Schema aus SPEC-020. Zeiger-Slices unterscheiden
// "Feld fehlt" (nil) von "explizit leer" ([]).
type document struct {
	AllowedRegions       *[]string `yaml:"allowedRegions"`
	AllowedAccounts      *[]string `yaml:"allowedAccounts"`
	AllowedModelPatterns *[]string `yaml:"allowedModelPatterns"`
	BlockedModelPatterns *[]string `yaml:"blockedModelPatterns"`
	AllowedProfiles      *[]string `yaml:"allowedProfiles"`
	Exclude              *[]string `yaml:"exclude"`
	Mantle               *string   `yaml:"mantle"`
}

var lineRE = regexp.MustCompile(`line (\d+)`)

// Parse parst den Inhalt einer Konfigurationsdatei. Unbekannte Felder und
// Syntaxfehler ergeben ein *model.FileError (FILE_INVALID, Zeile), nie mit
// Dateiinhalt.
func Parse(path string, data []byte) (*model.PartialConfig, error) {
	var doc document
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&doc); err != nil && !errors.Is(err, io.EOF) {
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid, Line: errorLine(err)}
	}
	// Mehr als ein YAML-Dokument ist keine gültige Konfiguration.
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid}
	}
	return doc.partial(), nil
}

func (d document) partial() *model.PartialConfig {
	p := &model.PartialConfig{
		AllowedRegions:       deref(d.AllowedRegions),
		AllowedAccounts:      deref(d.AllowedAccounts),
		AllowedModelPatterns: deref(d.AllowedModelPatterns),
		BlockedModelPatterns: deref(d.BlockedModelPatterns),
		AllowedProfiles:      deref(d.AllowedProfiles),
		Exclude:              deref(d.Exclude),
	}
	if d.Mantle != nil {
		p.Mantle = *d.Mantle
	}
	return p
}

// deref liefert nil für ein fehlendes Feld und einen nicht-nil Slice (ggf.
// leer) für ein vorhandenes. "allowedX:" ohne Wert (YAML null) liefert yaml.v3
// als nil und gilt damit wie ein fehlendes Feld (Default greift).
func deref(s *[]string) []string {
	if s == nil {
		return nil
	}
	if *s == nil {
		return []string{}
	}
	return *s
}

// errorLine liest die Zeilennummer aus einer yaml.v3-Fehlermeldung, ohne die
// Meldung (die Werte zitieren kann) weiterzugeben.
func errorLine(err error) int {
	if m := lineRE.FindStringSubmatch(err.Error()); m != nil {
		if n, convErr := strconv.Atoi(m[1]); convErr == nil {
			return n
		}
	}
	return 0
}
