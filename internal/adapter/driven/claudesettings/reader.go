// Package claudesettings implementiert den ClaudeSettingsReader-Port
// (ARC-006): liest settings.json, ermittelt Zeilen und schwärzt Secrets
// (LH-NF-004). Der Kern sieht nie einen Secret-Wert.
package claudesettings

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/secretkey"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

const userDisplayPath = "~/.claude/settings.json"

// Reader liest Claude-Code-Settings aus Home und Scan-Root.
type Reader struct{ root, home string }

// New liefert einen Reader. home wird injiziert (Tests ohne echtes ~);
// ist home leer, gilt die User-Datei als nicht vorhanden.
func New(root, home string) Reader { return Reader{root: root, home: home} }

// ReadUser liest ~/.claude/settings.json.
func (r Reader) ReadUser() (*model.ClaudeSettings, error) {
	if r.home == "" {
		return nil, nil
	}
	return readFile(filepath.Join(r.home, ".claude", "settings.json"), userDisplayPath)
}

// ReadProject liest eine Projektdatei relativ zum Scan-Root.
func (r Reader) ReadProject(rel string) (*model.ClaudeSettings, error) {
	clean := filepath.Clean(rel)
	if filepath.IsAbs(rel) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, &model.FileError{Path: rel, Reason: model.ReasonFileUnreadable}
	}
	return readFile(filepath.Join(r.root, clean), filepath.ToSlash(clean))
}

func readFile(full, display string) (*model.ClaudeSettings, error) {
	data, err := os.ReadFile(full)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, &model.FileError{Path: display, Reason: model.ReasonFileUnreadable}
	}
	s, ferr := Parse(display, data)
	if ferr != nil {
		return nil, ferr
	}
	return s, nil
}

// Parse parst den Inhalt einer settings.json. Fehler tragen Pfad, Grund
// und Zeile, nie Dateiinhalt.
func Parse(display string, data []byte) (*model.ClaudeSettings, *model.FileError) {
	s := &model.ClaudeSettings{Path: display, Hits: findHits(data)}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	invalid := func(err error) *model.FileError {
		off := dec.InputOffset()
		var se *json.SyntaxError
		// SyntaxError.Offset ist beim Stream-Decoder teils relativ zum
		// aktuellen Wert; der größere Wert trifft die Fehlerzeile.
		if errors.As(err, &se) && se.Offset > off {
			off = se.Offset
		}
		return &model.FileError{Path: display, Reason: model.ReasonFileInvalid, Line: lineAt(data, off)}
	}

	if d, err := dec.Token(); err != nil || d != json.Delim('{') {
		return nil, invalid(err)
	}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, invalid(err)
		}
		key, _ := kt.(string)
		switch key {
		case "model":
			var v any
			if err := dec.Decode(&v); err != nil {
				return nil, invalid(err)
			}
			if str, ok := v.(string); ok {
				s.Model = str
			}
		case "apiKeyHelper":
			// Nur die Zeile des Schlüssels; der Wert (ein Befehl, der Secrets
			// enthalten kann) wird gelesen und sofort verworfen (LH-NF-004).
			s.APIKeyHelperLine = lineAt(data, dec.InputOffset())
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return nil, invalid(err)
			}
		case "env":
			if err := parseEnv(dec, data, s); err != nil {
				return nil, invalid(err)
			}
		default:
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return nil, invalid(err)
			}
		}
	}
	if _, err := dec.Token(); err != nil { // schließende Klammer
		return nil, invalid(err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, &model.FileError{Path: display, Reason: model.ReasonFileInvalid, Line: lineAt(data, dec.InputOffset())}
	}
	return s, nil
}

func parseEnv(dec *json.Decoder, data []byte, s *model.ClaudeSettings) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		// kein Objekt (z. B. null oder Array): Rest überspringen
		if ok && d == '[' {
			for dec.More() {
				var skip json.RawMessage
				if err := dec.Decode(&skip); err != nil {
					return err
				}
			}
			_, err := dec.Token()
			return err
		}
		return nil
	}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return err
		}
		key, _ := kt.(string)
		line := lineAt(data, dec.InputOffset())
		var v any
		if err := dec.Decode(&v); err != nil {
			return err
		}
		e := model.EnvEntry{Key: key, Line: line}
		if secretkey.IsSecret(key) {
			e.Secret = true
		} else {
			e.Value = stringify(v)
		}
		s.Env = append(s.Env, e)
	}
	_, err = dec.Token()
	return err
}

// stringify bildet skalare JSON-Werte auf Strings ab (Zahl 1 → "1").
func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		if x {
			return "true"
		}
		return "false"
	}
	return ""
}

func lineAt(data []byte, off int64) int {
	if off > int64(len(data)) {
		off = int64(len(data))
	}
	if off < 0 {
		off = 0
	}
	return 1 + bytes.Count(data[:off], []byte("\n"))
}

// findHits sucht die Hit*-Muster im Rohinhalt und liefert nur Muster und
// Zeile (nie Textausschnitte), je Muster höchstens einen Treffer pro Zeile.
func findHits(data []byte) []model.Hit {
	var hits []model.Hit
	for i, ln := range strings.Split(string(data), "\n") {
		lower := strings.ToLower(ln)
		if strings.Contains(lower, model.HitAnthropicHost) {
			hits = append(hits, model.Hit{Pattern: model.HitAnthropicHost, Line: i + 1})
		}
		if strings.Contains(ln, model.HitBaseURL) {
			hits = append(hits, model.Hit{Pattern: model.HitBaseURL, Line: i + 1})
		}
		if strings.Contains(ln, model.HitAPIKey) {
			hits = append(hits, model.Hit{Pattern: model.HitAPIKey, Line: i + 1})
		}
	}
	return hits
}
