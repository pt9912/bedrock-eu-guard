package report

import (
	"encoding/json"
	"io"
	"net/url"
	"sort"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Konstanten des SARIF-Dokuments (SPEC-022). $schema ist nur eine Zeichenkette;
// es gibt keinen Netzzugriff.
const (
	sarifVersion = "2.1.0"
	sarifSchema  = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"
	toolName     = "bedrock-eu-check"
	srcRootID    = "%SRCROOT%"
)

// SARIF schreibt Findings als SARIF 2.1.0 (SPEC-022).
type SARIF struct {
	w            io.Writer
	toolVersion  string
	fallbackFile string
}

// NewSARIF liefert einen SARIF-Reporter für w. toolVersion kommt aus einer
// Build-Variable (Default "dev"), fallbackFile ist die Datei, die als
// Ersatzposition für Findings ohne Fundstelle dient (z. B.
// "bedrock-eu-check.yml" oder "README.md").
func NewSARIF(w io.Writer, toolVersion, fallbackFile string) *SARIF {
	if toolVersion == "" {
		toolVersion = "dev"
	}
	return &SARIF{w: w, toolVersion: toolVersion, fallbackFile: fallbackFile}
}

type (
	sarifDoc struct {
		Schema  string     `json:"$schema"`
		Version string     `json:"version"`
		Runs    []sarifRun `json:"runs"`
	}
	sarifRun struct {
		Tool    sarifTool     `json:"tool"`
		Results []sarifResult `json:"results"`
	}
	sarifTool struct {
		Driver sarifDriver `json:"driver"`
	}
	sarifDriver struct {
		Name    string      `json:"name"`
		Version string      `json:"version"`
		Rules   []sarifRule `json:"rules"`
	}
	sarifRule struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	sarifResult struct {
		RuleID    string          `json:"ruleId"`
		RuleIndex int             `json:"ruleIndex"`
		Level     string          `json:"level"`
		Message   sarifMessage    `json:"message"`
		Locations []sarifLocation `json:"locations"`
		Props     map[string]any  `json:"properties"`
	}
	sarifMessage struct {
		Text string `json:"text"`
	}
	sarifLocation struct {
		Physical sarifPhysical `json:"physicalLocation"`
	}
	sarifPhysical struct {
		Artifact sarifArtifact `json:"artifactLocation"`
		Region   *sarifRegion  `json:"region,omitempty"`
	}
	sarifArtifact struct {
		URI       string `json:"uri"`
		URIBaseID string `json:"uriBaseId"`
	}
	sarifRegion struct {
		StartLine int `json:"startLine"`
	}
)

// Report gibt das SARIF-Dokument aus. PASS erzeugt kein Result; die Regelliste
// enthält jede Regel mit Result genau einmal, sortiert nach ID. Die Ausgabe
// ist deterministisch (keine Zeitstempel, keine GUIDs).
func (s *SARIF) Report(findings []model.Finding) error {
	var kept []model.Finding
	ids := map[string]bool{}
	for _, f := range findings {
		if f.Severity.SARIFLevel() == "none" {
			continue
		}
		kept = append(kept, f)
		ids[f.RuleID] = true
	}
	rules := make([]sarifRule, 0, len(ids))
	for id := range ids {
		rules = append(rules, sarifRule{ID: id, Name: id})
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	index := map[string]int{}
	for i, r := range rules {
		index[r.ID] = i
	}

	results := make([]sarifResult, 0, len(kept))
	for _, f := range kept {
		results = append(results, s.result(f, index[f.RuleID]))
	}
	doc := sarifDoc{
		Schema:  sarifSchema,
		Version: sarifVersion,
		Runs: []sarifRun{{
			Tool:    sarifTool{Driver: sarifDriver{Name: toolName, Version: s.toolVersion, Rules: rules}},
			Results: results,
		}},
	}
	enc := json.NewEncoder(s.w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}

func (s *SARIF) result(f model.Finding, ruleIndex int) sarifResult {
	text := f.Message
	if f.Recommendation != "" {
		text += " Empfehlung: " + f.Recommendation
	}
	props := map[string]any{"reason": f.Reason}
	var phys sarifPhysical
	if uri, ok := repoRelativeURI(f.File); ok {
		phys.Artifact = sarifArtifact{URI: uri, URIBaseID: srcRootID}
		if f.Line > 0 {
			phys.Region = &sarifRegion{StartLine: f.Line}
		}
	} else {
		// Ersatzposition: erkennbar als solche, täuscht keine Fundstelle vor.
		fallback, _ := repoRelativeURI(s.fallbackFile)
		phys.Artifact = sarifArtifact{URI: fallback, URIBaseID: srcRootID}
		phys.Region = &sarifRegion{StartLine: 1}
		props["synthetic"] = true
		props["source"] = sourceOf(f)
	}
	return sarifResult{
		RuleID: f.RuleID, RuleIndex: ruleIndex, Level: f.Severity.SARIFLevel(),
		Message: sarifMessage{Text: text}, Locations: []sarifLocation{{Physical: phys}}, Props: props,
	}
}

// repoRelativeURI normalisiert einen Dateipfad zu einem relativen, "/"-getrennten
// und je Segment URI-kodierten Pfad. Leere, absolute (auch Windows-Laufwerke),
// "~"-Pfade und Pfade mit ".." liegen nicht im Repository (ok=false).
func repoRelativeURI(p string) (string, bool) {
	p = strings.ReplaceAll(p, `\`, "/")
	if p == "" || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "~") || (len(p) > 1 && p[1] == ':') {
		return "", false
	}
	var segs []string
	for _, seg := range strings.Split(p, "/") {
		switch seg {
		case "", ".":
			continue
		case "..":
			return "", false
		}
		segs = append(segs, url.PathEscape(seg))
	}
	if len(segs) == 0 {
		return "", false
	}
	return strings.Join(segs, "/"), true
}

// sourceOf benennt die Herkunft eines Findings ohne Fundstelle im Repository:
// der ursprüngliche Pfad (z. B. "~/.claude/settings.json") oder env/aws.
func sourceOf(f model.Finding) string {
	if f.File != "" {
		return f.File
	}
	for _, p := range []string{"-AWS-", "-BEDROCK-", "-IAM-"} {
		if strings.Contains(f.RuleID, p) {
			return "aws"
		}
	}
	return "env"
}
