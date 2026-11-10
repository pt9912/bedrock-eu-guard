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

// ConfigLoader liefert den Inhalt von bedrock-eu-check.yml.
//
// Vertrag: Datei fehlt → (nil, nil) (Defaults gelten); vorhanden, aber
// unlesbar oder ungültig (YAML, unbekannte Felder) → (nil, *model.FileError).
type ConfigLoader interface {
	Load() (*model.PartialConfig, error)
}

// EnvSourceReader liest Umgebungsvariablen aus Devcontainer-, Compose- und
// Dockerfile-Quellen (rel relativ zum Scan-Root). Vertrag wie bei
// ClaudeSettingsReader: fehlt → (nil, nil); kaputt → (nil, *model.FileError);
// Secret-Werte sind geschwärzt (EnvEntry.Secret, Value="").
type EnvSourceReader interface {
	ReadEnvSource(kind model.EnvSourceKind, rel string) (*model.EnvFile, error)
}

// TerraformReader findet und liest Terraform-Dateien (LH-FA-TF-001..004).
//
// Find sucht `*.tf` rekursiv unter dem Scan-Root (ohne .git, .terraform,
// vendor, node_modules) und liefert sortierte relative Pfade; ein Fehler der
// Suche selbst ist ein Fehler. Read: Datei fehlt → (nil, nil); kaputt →
// (nil, *model.FileError). Es werden nur Fakten geliefert, keine
// Variablenwerte.
type TerraformReader interface {
	Find() ([]string, error)
	Read(rel string) (*model.TerraformFile, error)
}

// AWSReader ist die einzige Tür zu AWS (ARC-007), ausschließlich lesend
// (LH-SEC-003). Fehler sind *model.AWSError und enthalten nie CLI-Ausgabe
// oder Credential-Werte. Weitere Methoden (Bedrock, IAM) kommen mit
// slice-009 und slice-010.
type AWSReader interface {
	// Identity entspricht `aws sts get-caller-identity`.
	Identity() (*model.AWSIdentity, error)
}

// BedrockReader liest Bedrock-Metadaten einer Region (ARC-007, lesend).
// Fehler sind *model.AWSError (nie CLI-Ausgabe oder Credential-Werte).
// Die Regionen bestimmt der Kern; der Adapter bekommt jeweils eine.
type BedrockReader interface {
	// ListInferenceProfiles entspricht `aws bedrock list-inference-profiles`.
	ListInferenceProfiles(region string) ([]model.InferenceProfile, error)
	// ListFoundationModels entspricht `aws bedrock list-foundation-models`.
	ListFoundationModels(region string) ([]model.FoundationModel, error)
}

// IAMReader liest die IAM-Sicht auf den aktuellen Principal (ARC-007,
// lesend). principalARN ist der von STS gemeldete ARN; der Adapter löst ihn
// bei Bedarf in den IAM-ARN auf (z. B. assumed-role → role). Fehler sind
// *model.AWSError.
type IAMReader interface {
	// SimulateBedrock entspricht `aws iam simulate-principal-policy`.
	SimulateBedrock(principalARN string, actions []string) ([]model.ActionDecision, error)
	// PrincipalPolicies liefert die Statements der angehängten und Inline-Policies.
	PrincipalPolicies(principalARN string) ([]model.IAMStatement, error)
}
