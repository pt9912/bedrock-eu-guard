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
