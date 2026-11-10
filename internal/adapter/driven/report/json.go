package report

import (
	"encoding/json"
	"io"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// JSONVersion ist die Version des JSON-Schemas (SPEC-021). Eine Änderung des
// Schemas ist eine bewusste Änderung dieser Konstante; der Test fixiert sie.
const JSONVersion = 1

// JSON schreibt Findings als JSON-Dokument nach SPEC-021.
type JSON struct{ w io.Writer }

// NewJSON liefert einen JSON-Reporter für w.
func NewJSON(w io.Writer) *JSON { return &JSON{w: w} }

type jsonDocument struct {
	Version  int           `json:"version"`
	Findings []jsonFinding `json:"findings"`
}

type jsonFinding struct {
	Status         string  `json:"status"`
	ID             string  `json:"id"`
	Message        string  `json:"message"`
	Severity       string  `json:"severity"`
	Reason         string  `json:"reason"`
	File           *string `json:"file"`
	Line           *int    `json:"line"`
	Recommendation *string `json:"recommendation"`
}

// Report gibt das Dokument mit zwei Leerzeichen Einrückung und
// abschließendem Zeilenumbruch aus. Die Reihenfolge entspricht der der
// Findings; die Ausgabe enthält weder Zeitstempel noch Zufallswerte.
func (j *JSON) Report(findings []model.Finding) error {
	doc := jsonDocument{Version: JSONVersion, Findings: make([]jsonFinding, 0, len(findings))}
	for _, f := range findings {
		jf := jsonFinding{
			Status:   f.Severity.String(),
			ID:       f.RuleID,
			Message:  f.Message,
			Severity: f.Severity.Level(),
			Reason:   f.Reason,
		}
		if f.File != "" {
			file := f.File
			jf.File = &file
		}
		if f.Line > 0 {
			line := f.Line
			jf.Line = &line
		}
		if f.Recommendation != "" {
			rec := f.Recommendation
			jf.Recommendation = &rec
		}
		doc.Findings = append(doc.Findings, jf)
	}
	enc := json.NewEncoder(j.w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}
