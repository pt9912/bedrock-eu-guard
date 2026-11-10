// Package envsource implementiert den EnvSourceReader-Port (ARC-006): liest
// Umgebungsvariablen aus devcontainer.json (JSONC), docker-compose-Dateien und
// Dockerfiles. Secret-Werte werden hier geschwärzt (LH-NF-004); der Kern sieht
// nie einen Wert.
package envsource

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Reader liest Env-Quellen relativ zum Scan-Root.
type Reader struct{ root string }

// New liefert einen Reader für root.
func New(root string) Reader { return Reader{root: root} }

// ReadEnvSource liest eine Datei der angegebenen Art. Fehlt sie, gilt
// (nil, nil); ist sie kaputt oder nicht lesbar, liefert der Reader ein
// *model.FileError ohne Dateiinhalt.
func (r Reader) ReadEnvSource(kind model.EnvSourceKind, rel string) (*model.EnvFile, error) {
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
	var (
		f    *model.EnvFile
		ferr *model.FileError
	)
	switch kind {
	case model.EnvDevcontainer:
		f, ferr = parseDevcontainer(display, data)
	case model.EnvCompose:
		f, ferr = parseCompose(display, data)
	case model.EnvDockerfile:
		f, ferr = parseDockerfile(display, data)
	default:
		return nil, &model.FileError{Path: display, Reason: model.ReasonFileUnreadable}
	}
	if ferr != nil {
		return nil, ferr
	}
	return f, nil
}

// dropInterpolatedRegions entfernt AWS_REGION/AWS_DEFAULT_REGION-Einträge mit
// Variablenreferenz ($VAR, ${VAR}): Der Wert ist hier nicht auflösbar und darf
// keinen Fehlalarm "Region nicht erlaubt" erzeugen (Plan slice-006 §8).
func dropInterpolatedRegions(in []model.EnvEntry) []model.EnvEntry {
	out := in[:0]
	for _, e := range in {
		if (e.Key == "AWS_REGION" || e.Key == "AWS_DEFAULT_REGION") && strings.Contains(e.Value, "$") {
			continue
		}
		out = append(out, e)
	}
	return out
}
