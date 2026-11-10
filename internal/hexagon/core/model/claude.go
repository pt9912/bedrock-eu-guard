package model

import "fmt"

// Reason-Codes der Datei-/Settings-Regeln.
const (
	ReasonFileMissing      = "FILE_MISSING"
	ReasonFileUnreadable   = "FILE_UNREADABLE"
	ReasonFileInvalid      = "FILE_INVALID"
	ReasonDirectAnthropic  = "DIRECT_ANTHROPIC"
	ReasonEndpointForeign  = "ENDPOINT_FOREIGN"
	ReasonEndpointTextOnly = "ENDPOINT_TEXT_ONLY"
	ReasonSettingsUnsafe   = "SETTINGS_UNSAFE"
	ReasonSettingsReview   = "SETTINGS_REVIEW"
)

// Muster, nach denen der Settings-Adapter den Rohinhalt durchsucht
// (LH-FA-CLAUDE-003). Treffer tragen nur Muster und Zeile.
const (
	HitAnthropicHost = "api.anthropic.com"
	HitBaseURL       = "ANTHROPIC_BASE_URL"
	HitAPIKey        = "ANTHROPIC_API_KEY"
)

// FileError beschreibt eine vorhandene, aber nicht nutzbare Datei. Sie
// enthält nie Dateiinhalt (LH-NF-004); Reason ist ReasonFileUnreadable
// oder ReasonFileInvalid.
type FileError struct {
	Path   string
	Reason string
	Line   int
}

// Error implementiert error, ohne Dateiinhalt zu zitieren.
func (e *FileError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s: %s (Zeile %d)", e.Path, e.Reason, e.Line)
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Reason)
}

// EnvEntry ist ein Eintrag aus dem "env"-Block der Claude-Settings. Bei
// Secret=true ist Value immer leer (Schwärzung im Adapter); der Eintrag
// zählt trotzdem als gesetzt.
type EnvEntry struct {
	Key    string
	Value  string
	Secret bool
	Line   int
}

// Hit ist ein Rohinhalt-Fund eines Hit*-Musters, nur mit Zeile.
type Hit struct {
	Pattern string
	Line    int
}

// ClaudeSettings ist der geparste Ausschnitt einer settings.json.
type ClaudeSettings struct {
	Path  string
	Model string
	Env   []EnvEntry
	Hits  []Hit
}

// EnvEntry liefert den Eintrag zu key (und ob er existiert).
func (s *ClaudeSettings) EnvEntry(key string) (EnvEntry, bool) {
	for _, e := range s.Env {
		if e.Key == key {
			return e, true
		}
	}
	return EnvEntry{}, false
}
