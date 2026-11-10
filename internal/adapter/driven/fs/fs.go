// Package fs implementiert den Filesystem-Port (ARC-006): nur lesender
// Zugriff relativ zu einem Scan-Root. Pfade, die den Root verlassen
// (absolut oder mit ".."), werden abgelehnt.
package fs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrOutsideRoot meldet einen Pfad außerhalb des Scan-Roots.
var ErrOutsideRoot = errors.New("pfad liegt außerhalb des scan-roots")

// FS liest Dateien unterhalb eines Roots.
type FS struct{ root string }

// New liefert einen FS-Adapter für root.
func New(root string) FS { return FS{root: root} }

func (f FS) resolve(p string) (string, error) {
	clean := filepath.Clean(p)
	if filepath.IsAbs(p) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q: %w", p, ErrOutsideRoot)
	}
	return filepath.Join(f.root, clean), nil
}

// ReadFile liest eine Datei; fehlt sie, gilt errors.Is(err, fs.ErrNotExist).
func (f FS) ReadFile(path string) ([]byte, error) {
	full, err := f.resolve(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(full)
}

// Exists meldet, ob der Pfad existiert.
func (f FS) Exists(path string) (bool, error) {
	full, err := f.resolve(path)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(full); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Glob liefert Treffer als Root-relative Pfade mit "/" in stabiler Reihenfolge.
func (f FS) Glob(pattern string) ([]string, error) {
	if _, err := f.resolve(pattern); err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(filepath.Join(f.root, filepath.Clean(pattern)))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		rel, err := filepath.Rel(f.root, m)
		if err != nil {
			return nil, err
		}
		out = append(out, filepath.ToSlash(rel))
	}
	sort.Strings(out)
	return out, nil
}
