// Package envos implementiert den Environment-Port (ARC-006) über os.LookupEnv.
package envos

import "os"

// Env liest Umgebungsvariablen des Prozesses.
type Env struct{}

// New liefert den os-basierten Environment-Adapter.
func New() Env { return Env{} }

// Lookup entspricht os.LookupEnv.
func (Env) Lookup(key string) (string, bool) { return os.LookupEnv(key) }
