package model

import (
	"errors"
	"regexp"
	"strconv"
)

// MantleMode steuert den Umgang mit Mantle (SPEC-020, LH-OFF-003).
type MantleMode string

// Zulässige Mantle-Modi; leer gilt als MantleWarn.
const (
	MantleWarn  MantleMode = "warn"
	MantleAllow MantleMode = "allow"
	MantleBlock MantleMode = "block"
)

// Reason-Codes für Mantle-Konfiguration und Konfigurationsfehler.
const (
	ReasonMantleAllowed = "MANTLE_ALLOWED"
	ReasonMantleBlocked = "MANTLE_BLOCKED"

	ConfigInvalidRegex   = "INVALID_REGEX"
	ConfigInvalidValue   = "INVALID_VALUE"
	ConfigInvalidAccount = "INVALID_ACCOUNT"
	ConfigEmptyRegions   = "EMPTY_REGIONS"
)

// PartialConfig ist der gelesene Inhalt von bedrock-eu-check.yml. Ein
// nil-Slice / leerer String bedeutet "nicht angegeben" (Default gilt);
// ein nicht-nil leerer Slice bedeutet "explizit leer".
type PartialConfig struct {
	AllowedRegions       []string
	AllowedAccounts      []string
	AllowedModelPatterns []string
	BlockedModelPatterns []string
	AllowedProfiles      []string
	Mantle               string
}

// ConfigError beschreibt ein ungültiges Konfigurationsfeld. Sie zitiert
// nie den Feldwert (nur Feldname, Index und Grund).
type ConfigError struct {
	Field  string
	Index  int // -1 = ganzes Feld
	Reason string
}

// Error implementiert error ohne Feldwert.
func (e *ConfigError) Error() string {
	if e.Index >= 0 {
		return "Konfiguration: " + e.Field + "[" + strconv.Itoa(e.Index) + "]: " + e.Reason
	}
	return "Konfiguration: " + e.Field + ": " + e.Reason
}

var accountRE = regexp.MustCompile(`^[0-9]{12}$`)

// Resolve legt p über DefaultConfig (LH-NF-007: alle Felder optional) und
// validiert das Ergebnis. Alle gefundenen Fehler werden als *ConfigError
// zusammengefasst (errors.Join; errors.As funktioniert).
func Resolve(p PartialConfig) (Config, error) {
	c := DefaultConfig()
	var errs []error
	bad := func(field string, idx int, reason string) {
		errs = append(errs, &ConfigError{Field: field, Index: idx, Reason: reason})
	}

	if p.AllowedRegions != nil {
		c.AllowedRegions = p.AllowedRegions
		if len(p.AllowedRegions) == 0 {
			bad("allowedRegions", -1, ConfigEmptyRegions)
		}
		for i, r := range p.AllowedRegions {
			if r == "" {
				bad("allowedRegions", i, ConfigInvalidValue)
			}
		}
	}
	if p.AllowedAccounts != nil {
		c.AllowedAccounts = p.AllowedAccounts
		for i, a := range p.AllowedAccounts {
			if !accountRE.MatchString(a) {
				bad("allowedAccounts", i, ConfigInvalidAccount)
			}
		}
	}
	if p.AllowedModelPatterns != nil {
		c.AllowedModelPatterns = p.AllowedModelPatterns
	}
	if p.BlockedModelPatterns != nil {
		c.BlockedModelPatterns = p.BlockedModelPatterns
	}
	if p.AllowedProfiles != nil {
		c.AllowedProfiles = p.AllowedProfiles
	}
	for _, f := range []struct {
		name string
		pats []string
	}{
		{"allowedModelPatterns", c.AllowedModelPatterns},
		{"blockedModelPatterns", c.BlockedModelPatterns},
		{"allowedProfiles", c.AllowedProfiles},
	} {
		for i, pat := range f.pats {
			if _, err := regexp.Compile(pat); err != nil {
				bad(f.name, i, ConfigInvalidRegex)
			}
		}
	}
	if p.Mantle != "" {
		switch m := MantleMode(p.Mantle); m {
		case MantleWarn, MantleAllow, MantleBlock:
			c.Mantle = m
		default:
			bad("mantle", -1, ConfigInvalidValue)
		}
	}
	return c, errors.Join(errs...)
}
