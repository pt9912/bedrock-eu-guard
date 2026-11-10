// Package memfs ist ein map-basiertes Test-Double für driven.Filesystem.
package memfs

import (
	"io/fs"
	"path"
	"sort"
)

// FS bildet relative Pfade auf Dateiinhalte ab.
type FS map[string]string

// ReadFile liefert den Inhalt oder fs.ErrNotExist.
func (m FS) ReadFile(p string) ([]byte, error) {
	c, ok := m[p]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return []byte(c), nil
}

// Exists meldet, ob die Datei existiert.
func (m FS) Exists(p string) (bool, error) {
	_, ok := m[p]
	return ok, nil
}

// Glob liefert alle passenden Pfade (path.Match), sortiert.
func (m FS) Glob(pattern string) ([]string, error) {
	var out []string
	for p := range m {
		ok, err := path.Match(pattern, p)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out, nil
}
