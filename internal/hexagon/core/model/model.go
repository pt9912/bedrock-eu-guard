// Package model ist der importfreie Ring des Kerns (ARC-003): Finding,
// Severity, Reason-Codes und Config. Es importiert nur die Stdlib.
package model

// Severity ist der Schweregrad eines Findings (SPEC-001). Die Ordnung
// entspricht dem Schweregrad; der Maximalwert bestimmt den Exit-Code.
type Severity int

const (
	PASS Severity = iota
	WARN
	FAIL
)

// String liefert "PASS", "WARN" oder "FAIL".
func (s Severity) String() string {
	switch s {
	case PASS:
		return "PASS"
	case WARN:
		return "WARN"
	case FAIL:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// Reason-Codes der Regeln.
const (
	ReasonOK                 = "OK"
	ReasonRegionMissing      = "REGION_MISSING"
	ReasonRegionNotAllowed   = "REGION_NOT_ALLOWED"
	ReasonRegionFallback     = "REGION_FALLBACK_DEFAULT"
	ReasonBedrockNotEnabled  = "BEDROCK_NOT_ENABLED"
	ReasonAnthropicKeySet    = "ANTHROPIC_API_KEY_SET"
	ReasonAnthropicKeyNotSet = "ANTHROPIC_API_KEY_NOT_SET"
	ReasonModelAllowed       = "MODEL_ALLOWED"
	ReasonModelBlocked       = "MODEL_BLOCKED"
	ReasonModelUnclassified  = "MODEL_UNCLASSIFIED"
	ReasonMantleUnverified   = "MANTLE_UNVERIFIED"
	ReasonMantleVerified     = "MANTLE_VERIFIED"
)

// ModelClass ist das Ergebnis der Modell-/Profil-ID-Klassifikation (SPEC-010).
type ModelClass int

const (
	ModelUnclassified ModelClass = iota
	ModelAllowed
	ModelBlocked
)

// String liefert "allowed", "blocked" oder "unclassified".
func (c ModelClass) String() string {
	switch c {
	case ModelAllowed:
		return "allowed"
	case ModelBlocked:
		return "blocked"
	default:
		return "unclassified"
	}
}

// Finding ist das Ergebnis einer Regel. Message darf nie einen
// Secret-Wert enthalten (LH-NF-004).
type Finding struct {
	RuleID   string
	Severity Severity
	Reason   string
	Message  string
	Subject  string
}

// Config ist die Laufzeitkonfiguration (SPEC-020, hier nur der
// Slice-001-Ausschnitt).
type Config struct {
	AllowedRegions       []string
	AllowedModelPatterns []string // reguläre Ausdrücke (SPEC-020)
	BlockedModelPatterns []string // reguläre Ausdrücke; Vorrang vor Allow (SPEC-010)
}

// DefaultConfig liefert die Defaults (LH-NF-007: überschreibbar).
func DefaultConfig() Config {
	return Config{
		AllowedRegions:       []string{"eu-central-1"},
		AllowedModelPatterns: []string{`^eu\.`, `^anthropic\.claude-sonnet-4-6$`, `^anthropic\.claude-haiku-4-5$`},
		BlockedModelPatterns: []string{`^us\.`, `^global\.`, `us-east-1`, `us-west-2`},
	}
}

// RegionAllowed meldet, ob region in AllowedRegions enthalten ist (SPEC-011).
func (c Config) RegionAllowed(region string) bool {
	for _, r := range c.AllowedRegions {
		if r == region {
			return true
		}
	}
	return false
}

// MaxSeverity liefert den höchsten Schweregrad (PASS bei leerer Liste).
func MaxSeverity(fs []Finding) Severity {
	m := PASS
	for _, f := range fs {
		if f.Severity > m {
			m = f.Severity
		}
	}
	return m
}
