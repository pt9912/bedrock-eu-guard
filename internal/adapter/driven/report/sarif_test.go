package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/report"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.Reporter = (*report.SARIF)(nil)

type sarifRoot struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []struct {
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Rules   []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Results []struct {
			RuleID    string `json:"ruleId"`
			RuleIndex int    `json:"ruleIndex"`
			Level     string `json:"level"`
			Message   struct {
				Text string `json:"text"`
			} `json:"message"`
			Locations []struct {
				Physical struct {
					Artifact struct {
						URI       string `json:"uri"`
						URIBaseID string `json:"uriBaseId"`
					} `json:"artifactLocation"`
					Region *struct {
						StartLine int `json:"startLine"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
			Props map[string]any `json:"properties"`
		} `json:"results"`
	} `json:"runs"`
}

func renderSARIF(t *testing.T, fs []model.Finding) (string, sarifRoot) {
	t.Helper()
	var buf bytes.Buffer
	if err := report.NewSARIF(&buf, "1.2.3", "bedrock-eu-check.yml").Report(fs); err != nil {
		t.Fatal(err)
	}
	var doc sarifRoot
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("kein gültiges SARIF-JSON: %v\n%s", err, buf.String())
	}
	return buf.String(), doc
}

func TestSARIFStructure(t *testing.T) {
	_, doc := renderSARIF(t, []model.Finding{
		{RuleID: "LH-FA-TF-002", Severity: model.FAIL, Reason: "REGION_NOT_ALLOWED", Message: "Region nicht erlaubt", File: "infra/main.tf", Line: 7, Recommendation: "EU-Region setzen"},
		{RuleID: "LH-FA-TF-001", Severity: model.PASS, Reason: "TERRAFORM_FOUND", Message: "ok"},
		{RuleID: "LH-FA-CLAUDE-001", Severity: model.WARN, Reason: "FILE_MISSING", Message: "fehlt", File: "~/.claude/settings.json"},
		{RuleID: "LH-FA-TF-002", Severity: model.WARN, Reason: "REGION_DYNAMIC", Message: "dynamisch", File: "infra/main.tf"},
	})
	if doc.Version != "2.1.0" || !strings.Contains(doc.Schema, "sarif-schema-2.1.0") || len(doc.Runs) != 1 {
		t.Fatalf("Kopf: %+v", doc)
	}
	d := doc.Runs[0].Tool.Driver
	if d.Name != "bedrock-eu-check" || d.Version != "1.2.3" {
		t.Errorf("driver: %+v", d)
	}
	if len(d.Rules) != 2 || d.Rules[0].ID != "LH-FA-CLAUDE-001" || d.Rules[1].ID != "LH-FA-TF-002" {
		t.Errorf("rules (eindeutig, sortiert, ohne PASS-only): %+v", d.Rules)
	}
	rs := doc.Runs[0].Results
	if len(rs) != 3 {
		t.Fatalf("PASS erzeugt kein Result; results: %+v", rs)
	}
	r0 := rs[0]
	if r0.RuleID != "LH-FA-TF-002" || r0.Level != "error" || r0.RuleIndex != 1 ||
		r0.Message.Text != "Region nicht erlaubt Empfehlung: EU-Region setzen" || r0.Props["reason"] != "REGION_NOT_ALLOWED" {
		t.Errorf("r0: %+v", r0)
	}
	loc := r0.Locations[0].Physical
	if loc.Artifact.URI != "infra/main.tf" || loc.Artifact.URIBaseID != "%SRCROOT%" || loc.Region == nil || loc.Region.StartLine != 7 {
		t.Errorf("r0 location: %+v", loc)
	}
	if _, ok := r0.Props["synthetic"]; ok {
		t.Error("echte Position darf nicht als synthetic markiert sein")
	}
	// Datei ohne Zeile: keine Region.
	if rs[2].Locations[0].Physical.Region != nil || rs[2].Level != "warning" {
		t.Errorf("r2: %+v", rs[2])
	}
}

func TestSARIFSyntheticLocation(t *testing.T) {
	_, doc := renderSARIF(t, []model.Finding{
		{RuleID: "LH-FA-LENV-001", Severity: model.FAIL, Reason: "REGION_MISSING", Message: "env"},
		{RuleID: "LH-FA-AWS-001", Severity: model.FAIL, Reason: "AWS_NO_CREDENTIALS", Message: "aws"},
		{RuleID: "LH-FA-CLAUDE-001", Severity: model.WARN, Reason: "FILE_MISSING", Message: "home", File: "~/.claude/settings.json"},
		{RuleID: "LH-FA-X", Severity: model.WARN, Reason: "X", Message: "außerhalb", File: "../outside/x.tf", Line: 3},
	})
	want := []string{"env", "aws", "~/.claude/settings.json", "../outside/x.tf"}
	for i, r := range doc.Runs[0].Results {
		loc := r.Locations[0].Physical
		if loc.Artifact.URI != "bedrock-eu-check.yml" || loc.Artifact.URIBaseID != "%SRCROOT%" || loc.Region == nil || loc.Region.StartLine != 1 {
			t.Errorf("r%d Ersatzposition: %+v", i, loc)
		}
		if r.Props["synthetic"] != true || r.Props["source"] != want[i] {
			t.Errorf("r%d properties: %+v (want source %q)", i, r.Props, want[i])
		}
	}
}

func TestSARIFPathNormalisation(t *testing.T) {
	cases := map[string]string{
		`infra\sub\main.tf`: "infra/sub/main.tf",
		"./a//b.tf":         "a/b.tf",
		"mit leer/ä#?.tf":   "mit%20leer/%C3%A4%23%3F.tf",
		"docs/README.md":    "docs/README.md",
	}
	for in, want := range cases {
		_, doc := renderSARIF(t, []model.Finding{{RuleID: "R", Severity: model.FAIL, Message: "m", File: in, Line: 1}})
		if got := doc.Runs[0].Results[0].Locations[0].Physical.Artifact.URI; got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
	for _, out := range []string{"/etc/x", `C:\x.tf`, "~/x", "../x.tf", "a/../../x.tf"} {
		_, doc := renderSARIF(t, []model.Finding{{RuleID: "R", Severity: model.FAIL, Message: "m", File: out, Line: 1}})
		if doc.Runs[0].Results[0].Props["synthetic"] != true {
			t.Errorf("%q muss als Ersatzposition behandelt werden", out)
		}
	}
}

func TestSARIFEmptyAndDeterministic(t *testing.T) {
	out, doc := renderSARIF(t, []model.Finding{{RuleID: "R", Severity: model.PASS, Message: "ok"}})
	if len(doc.Runs) != 1 || doc.Runs[0].Results == nil || len(doc.Runs[0].Results) != 0 || len(doc.Runs[0].Tool.Driver.Rules) != 0 {
		t.Errorf("nur PASS → gültiges Dokument ohne Results:\n%s", out)
	}
	if !strings.Contains(out, `"results": []`) || !strings.Contains(out, `"rules": []`) {
		t.Errorf("leere Arrays erwartet:\n%s", out)
	}
	fs := []model.Finding{{RuleID: "B", Severity: model.WARN, Message: "b"}, {RuleID: "A", Severity: model.FAIL, Message: "a", File: "x.tf", Line: 2}}
	a, _ := renderSARIF(t, fs)
	b, _ := renderSARIF(t, fs)
	if a != b {
		t.Error("Ausgabe ist nicht deterministisch")
	}
}

func TestSARIFDefaultsAndSecrets(t *testing.T) {
	var buf bytes.Buffer
	if err := report.NewSARIF(&buf, "", "README.md").Report([]model.Finding{{RuleID: "R", Severity: model.FAIL, Message: "m", Subject: "SUBJECT-NICHT-AUSGEBEN"}}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"version": "dev"`) || !strings.Contains(out, `"uri": "README.md"`) {
		t.Errorf("Defaults:\n%s", out)
	}
	if strings.Contains(out, "SUBJECT-NICHT-AUSGEBEN") {
		t.Error("subject darf nicht ausgegeben werden")
	}
	if err := report.NewSARIF(failWriter{}, "", "").Report(nil); err == nil {
		t.Error("expected write error")
	}
}
