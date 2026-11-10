// Package fakeenv ist ein map-basiertes Test-Double für driven.Environment.
package fakeenv

// Env implementiert driven.Environment.
type Env map[string]string

// Lookup liefert den Wert und ob die Variable gesetzt ist.
func (e Env) Lookup(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}
