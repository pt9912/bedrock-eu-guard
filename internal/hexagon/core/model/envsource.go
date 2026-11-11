package model

// EnvSourceKind benennt die Art einer Env-Quelle.
type EnvSourceKind string

// Unterstützte Env-Quellen (LH-FA-DEVCON-001, LH-FA-DOCKER-001/-002).
const (
	EnvDevcontainer EnvSourceKind = "devcontainer"
	EnvCompose      EnvSourceKind = "compose"
	EnvDockerfile   EnvSourceKind = "dockerfile"
	EnvDotenv       EnvSourceKind = "dotenv"
)

// Reason-Codes der Env-Quellen-Regeln.
const (
	ReasonNoClaudeConfig  = "NO_CLAUDE_CONFIG"
	ReasonEnvIncomplete   = "ENV_INCOMPLETE"
	ReasonRegionNotSetEnv = "REGION_NOT_SET"
)

// EnvFile ist der geparste Env-Ausschnitt einer Quelle: alle dort gesetzten
// Variablen mit Zeile (und Scope), Secret-Werte bereits geschwärzt.
type EnvFile struct {
	Path string
	Kind EnvSourceKind
	Env  []EnvEntry
}

// All liefert alle Einträge zu key (über alle Scopes).
func (f *EnvFile) All(key string) []EnvEntry {
	var out []EnvEntry
	for _, e := range f.Env {
		if e.Key == key {
			out = append(out, e)
		}
	}
	return out
}
