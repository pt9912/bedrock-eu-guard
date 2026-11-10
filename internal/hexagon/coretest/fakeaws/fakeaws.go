// Package fakeaws ist ein Test-Double für driven.AWSReader.
package fakeaws

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Reader liefert eine vorbereitete Identität oder einen Fehler.
type Reader struct {
	Ident *model.AWSIdentity
	Err   error
}

// Identity implementiert driven.AWSReader.
func (r Reader) Identity() (*model.AWSIdentity, error) { return r.Ident, r.Err }
