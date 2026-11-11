// Package fakeenvsource ist ein Test-Double für driven.EnvSourceReader.
package fakeenvsource

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Reader liefert vorbereitete Env-Quellen; Schlüssel ist der relative Pfad.
type Reader struct {
	Files map[string]*model.EnvFile
	Errs  map[string]error
}

// ReadEnvSource implementiert driven.EnvSourceReader.
func (r Reader) ReadEnvSource(_ model.EnvSourceKind, rel string) (*model.EnvFile, error) {
	if err := r.Errs[rel]; err != nil {
		return nil, err
	}
	return r.Files[rel], nil
}

// Finder implementiert driven.EnvFileFinder mit vorbereiteten Pfaden je Art.
type Finder struct {
	Paths map[model.EnvSourceKind][]string
	Err   error
}

// Find implementiert driven.EnvFileFinder.
func (f Finder) Find(kind model.EnvSourceKind) ([]string, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return append([]string(nil), f.Paths[kind]...), nil
}
