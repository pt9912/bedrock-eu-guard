// Package policyfile implementiert den IAMPolicyFileReader-Port (ARC-006): findet
// eigenständige IAM-Policy-Dateien im Scan-Root und liest sie zu flachen
// model.IAMStatement (LH-FA-IAM-004). Die Auslegung der Statements übernimmt das
// Paket iampolicy, damit Terraform, AWS-Modus und Policy-Dateien dieselbe Sicht
// liefern. Dateiinhalte erscheinen nie in Fehlern.
package policyfile

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/iampolicy"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Reader findet und liest Policy-Dateien relativ zum Scan-Root.
type Reader struct{ root string }

// New liefert einen Reader für root.
func New(root string) Reader { return Reader{root: root} }

// skipDirs sind Verzeichnisse, die nie durchsucht werden (wie bei Terraform und .env).
var skipDirs = map[string]bool{".git": true, ".terraform": true, "vendor": true, "node_modules": true, ".tmp": true}

// isPolicyName entspricht den Mustern iam-policy*.json und *-policy.json (Basisname).
func isPolicyName(name string) bool {
	for _, pat := range []string{"iam-policy*.json", "*-policy.json"} {
		if ok, _ := path.Match(pat, name); ok {
			return true
		}
	}
	return false
}

// Find sucht rekursiv nach Policy-Dateien und liefert sortierte relative Pfade
// mit "/" als Trenner. Symlinks werden nicht verfolgt; "exclude" wendet der Kern an.
func (r Reader) Find() ([]string, error) {
	var out []string
	err := filepath.WalkDir(r.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != r.root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() && isPolicyName(d.Name()) {
			rel, relErr := filepath.Rel(r.root, p)
			if relErr != nil {
				return relErr
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// Read liest eine Datei. Fehlt sie, gilt (nil, nil); nicht lesbar ergibt
// FILE_UNREADABLE, kein gültiges JSON FILE_INVALID (mit Zeile, ohne Inhalt).
// Gültiges JSON ohne erkennbares Policy-Dokument (Fremdformat) liefert eine
// leere Statement-Liste.
func (r Reader) Read(rel string) (*model.IAMPolicyFile, error) {
	clean := filepath.Clean(rel)
	display := filepath.ToSlash(clean)
	if filepath.IsAbs(rel) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, &model.FileError{Path: rel, Reason: model.ReasonFileUnreadable}
	}
	data, err := os.ReadFile(filepath.Join(r.root, clean))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, &model.FileError{Path: display, Reason: model.ReasonFileUnreadable}
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")) // BOM
	if !json.Valid(data) {
		return nil, &model.FileError{Path: display, Reason: model.ReasonFileInvalid, Line: errorLine(data)}
	}
	doc, ok := unwrap(data)
	if !ok {
		return &model.IAMPolicyFile{}, nil
	}
	stmts, ok := iampolicy.Statements(doc, 0)
	if !ok {
		return &model.IAMPolicyFile{}, nil
	}
	return &model.IAMPolicyFile{Statements: stmts}, nil
}

// unwrap liefert das Policy-Dokument (Objekt mit "Statement"): entweder das
// Dokument selbst, "PolicyVersion.Document" (aws iam get-policy-version) oder
// "PolicyDocument" (aws iam get-role-policy). URL-kodierte String-Dokumente und
// andere Formen gelten als Fremdformat.
func unwrap(data []byte) ([]byte, bool) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, false // kein Objekt (zum Beispiel ein Array)
	}
	if _, ok := top["Statement"]; ok {
		return data, true
	}
	if raw, ok := top["PolicyVersion"]; ok {
		var pv struct {
			Document json.RawMessage `json:"Document"`
		}
		if json.Unmarshal(raw, &pv) == nil && isObject(pv.Document) {
			return pv.Document, true
		}
		return nil, false
	}
	if raw, ok := top["PolicyDocument"]; ok && isObject(raw) {
		return raw, true
	}
	return nil, false
}

func isObject(raw json.RawMessage) bool {
	return len(bytes.TrimSpace(raw)) > 0 && bytes.TrimSpace(raw)[0] == '{'
}

// errorLine liefert die Zeile (1-basiert) des ersten Syntaxfehlers, ohne Inhalt
// zu zitieren; 0, wenn sie nicht bestimmbar ist.
func errorLine(data []byte) int {
	var v any
	err := json.Unmarshal(data, &v)
	var se *json.SyntaxError
	if errors.As(err, &se) && se.Offset > 0 && int(se.Offset) <= len(data) {
		return 1 + bytes.Count(data[:se.Offset], []byte("\n"))
	}
	return 0
}
