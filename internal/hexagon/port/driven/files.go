package driven

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Filesystem liest Dateien relativ zum Scan-Root (nur lesend). Eine
// fehlende Datei meldet ReadFile mit einem Fehler, für den
// errors.Is(err, fs.ErrNotExist) gilt.
type Filesystem interface {
	ReadFile(path string) ([]byte, error)
	Exists(path string) (bool, error)
	Glob(pattern string) ([]string, error)
}

// ClaudeSettingsReader liefert geparste Claude-Settings.
//
// Vertrag: Datei fehlt → (nil, nil); vorhanden, aber nicht lesbar oder
// ungültig → (nil, *model.FileError). Secret-Werte sind bereits
// geschwärzt (EnvEntry.Secret, Value="").
type ClaudeSettingsReader interface {
	// ReadUser liest ~/.claude/settings.json.
	ReadUser() (*model.ClaudeSettings, error)
	// ReadProject liest eine Projektdatei relativ zum Scan-Root,
	// z. B. ".claude/settings.json".
	ReadProject(rel string) (*model.ClaudeSettings, error)
}
