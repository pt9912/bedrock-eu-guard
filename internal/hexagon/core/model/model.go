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
)

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
	AllowedRegions []string
}

// DefaultConfig liefert die Defaults (LH-NF-007: überschreibbar).
func DefaultConfig() Config {
	return Config{AllowedRegions: []string{"eu-central-1"}}
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
