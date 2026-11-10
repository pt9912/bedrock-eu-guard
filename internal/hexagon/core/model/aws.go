package model

// Reason-Codes der AWS-Regeln und -Fehler.
const (
	AWSNoCredentials = "AWS_NO_CREDENTIALS"
	AWSUnreachable   = "AWS_UNREACHABLE"
	AWSAccessDenied  = "AWS_ACCESS_DENIED"
	AWSCLIMissing    = "AWS_CLI_MISSING" // Tool-Fehler (Exit 2), kein Finding

	ReasonProfileSet        = "AWS_PROFILE_SET"
	ReasonProfileNotSet     = "AWS_PROFILE_NOT_SET"
	ReasonProfileUnusable   = "AWS_PROFILE_UNUSABLE"
	ReasonAccountAllowed    = "ACCOUNT_ALLOWED"
	ReasonAccountNotAllowed = "ACCOUNT_NOT_ALLOWED"
	ReasonAccountUnchecked  = "ACCOUNT_UNCHECKED"
	ReasonAuthProfile       = "AUTH_PROFILE"
	ReasonAuthTemporary     = "AUTH_TEMPORARY_CREDENTIALS"
	ReasonAuthStaticKeys    = "AUTH_STATIC_KEYS"
	ReasonAuthBedrockAPIKey = "AUTH_BEDROCK_API_KEY"
	ReasonAuthAnthropicKey  = "AUTH_ANTHROPIC_API_KEY"
	ReasonAuthNoneDetected  = "AUTH_NONE_DETECTED"
)

// AWSError ist ein typisierter Fehler des AWS-Zugriffs. Er enthält nie
// CLI-Ausgabe oder Credential-Werte (LH-SEC-001, LH-NF-004).
type AWSError struct {
	Op     string // z. B. "sts:GetCallerIdentity"
	Reason string // AWSNoCredentials | AWSUnreachable | AWSAccessDenied | AWSCLIMissing
}

// Error implementiert error ohne Fremdausgabe.
func (e *AWSError) Error() string { return "AWS " + e.Op + ": " + e.Reason }

// AWSIdentity ist das Ergebnis von sts:GetCallerIdentity.
type AWSIdentity struct {
	Account string
	ARN     string
	UserID  string
}

// InferenceProfile ist ein Bedrock-Inference-Profil (nur die Felder, die die
// Regeln brauchen).
type InferenceProfile struct {
	ID   string
	Name string
	Type string // z. B. SYSTEM_DEFINED | APPLICATION
}

// FoundationModel ist ein in einer Region verfügbares Basismodell.
type FoundationModel struct {
	ID string
}

// Reason-Codes der Bedrock-Regeln.
const (
	ReasonProfilesListed   = "PROFILES_LISTED"
	ReasonProfilesNone     = "PROFILES_NONE"
	ReasonProfilesVisible  = "PROFILES_VISIBLE_NOT_ALLOWED"
	ReasonProfileNotListed = "PROFILE_NOT_ALLOWLISTED"
	ReasonModelAvailable   = "MODEL_AVAILABLE"
	ReasonModelUnavailable = "MODEL_UNAVAILABLE"
	ReasonModelUnknown     = "MODEL_AVAILABILITY_UNKNOWN"
)
