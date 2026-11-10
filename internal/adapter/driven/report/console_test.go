package report_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/report"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.Reporter = (*report.Console)(nil)

func TestReport(t *testing.T) {
	var buf bytes.Buffer
	err := report.New(&buf).Report([]model.Finding{
		{RuleID: "LENV-001", Severity: model.PASS, Reason: "REGION_OK", Message: "ok", Subject: "AWS_REGION"},
		{RuleID: "LENV-004", Severity: model.WARN, Reason: "KEY_SET", Message: "ANTHROPIC_API_KEY is set", Subject: "ANTHROPIC_API_KEY"},
		{RuleID: "LENV-002", Severity: model.FAIL, Reason: "REGION_MISSING", Message: "not set", Subject: "AWS_DEFAULT_REGION"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "[PASS] LENV-001 AWS_REGION: ok (REGION_OK)\n" +
		"[WARN] LENV-004 ANTHROPIC_API_KEY: ANTHROPIC_API_KEY is set (KEY_SET)\n" +
		"[FAIL] LENV-002 AWS_DEFAULT_REGION: not set (REGION_MISSING)\n" +
		"Summary: 1 PASS, 1 WARN, 1 FAIL\n"
	if got := buf.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestReportEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := report.New(&buf).Report(nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Summary: 0 PASS, 0 WARN, 0 FAIL") {
		t.Errorf("unexpected output %q", buf.String())
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }

func TestReportWriteError(t *testing.T) {
	if err := report.New(failWriter{}).Report(nil); err == nil {
		t.Error("expected write error")
	}
}

func TestReportFileLineAndRecommendation(t *testing.T) {
	var buf bytes.Buffer
	err := report.New(&buf).Report([]model.Finding{
		{RuleID: "R1", Severity: model.FAIL, Reason: "X", Message: "m", Subject: "S", File: ".claude/settings.json", Line: 7, Recommendation: "Region ändern"},
		{RuleID: "R2", Severity: model.WARN, Reason: "Y", Message: "n", Subject: "T", File: "a.json"},
		{RuleID: "R3", Severity: model.PASS, Reason: "OK", Message: "o", Subject: "b.json", File: "b.json", Line: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "[FAIL] R1 S [.claude/settings.json:7]: m (X)\n" +
		"       Empfehlung: Region ändern\n" +
		"[WARN] R2 T [a.json]: n (Y)\n" +
		"[PASS] R3 b.json:3: o (OK)\n" +
		"Summary: 1 PASS, 1 WARN, 1 FAIL\n"
	if got := buf.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
