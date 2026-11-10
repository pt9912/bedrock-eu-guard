// Package report implementiert den Reporter-Port als Konsolenausgabe
// (LH-FA-REP-001). Message und Subject werden unverändert ausgegeben; der
// Kern garantiert, dass sie keinen Secret-Wert enthalten (LH-NF-004).
package report

import (
	"fmt"
	"io"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Console schreibt Findings zeilenweise auf einen Writer.
type Console struct{ w io.Writer }

// New liefert einen Konsolen-Reporter für w.
func New(w io.Writer) *Console { return &Console{w: w} }

// Report gibt je Finding eine Zeile und eine Zusammenfassung aus.
func (c *Console) Report(findings []model.Finding) error {
	var pass, warn, fail int
	for _, f := range findings {
		switch f.Severity {
		case model.FAIL:
			fail++
		case model.WARN:
			warn++
		default:
			pass++
		}
		subject, loc := f.Subject, location(f)
		if f.File != "" && f.Subject == f.File { // Pfad nicht doppelt ausgeben
			subject, loc = place(f), ""
		}
		if _, err := fmt.Fprintf(c.w, "[%s] %s %s%s: %s (%s)\n",
			f.Severity, f.RuleID, subject, loc, f.Message, f.Reason); err != nil {
			return err
		}
		if f.Recommendation != "" {
			if _, err := fmt.Fprintf(c.w, "       Empfehlung: %s\n", f.Recommendation); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(c.w, "Summary: %d PASS, %d WARN, %d FAIL\n", pass, warn, fail)
	return err
}

// place liefert "Datei:Zeile" bzw. "Datei".
func place(f model.Finding) string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return f.File
}

// location liefert " [Datei:Zeile]" bzw. " [Datei]" für dateibezogene Findings.
func location(f model.Finding) string {
	if f.File == "" {
		return ""
	}
	return " [" + place(f) + "]"
}
