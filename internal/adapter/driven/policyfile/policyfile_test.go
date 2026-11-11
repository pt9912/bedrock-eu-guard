package policyfile

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.IAMPolicyFileReader = Reader{}

const wildcardDoc = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"bedrock:*","Resource":"*"}]}`

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestIsPolicyName(t *testing.T) {
	yes := []string{"iam-policy.json", "iam-policy-bedrock.json", "iam-policy2.json", "bedrock-policy.json", "role-policy.json", "-policy.json"}
	no := []string{"policy.json", "iam-policy.yaml", "iam-policy", "my-policy.json.bak", "policies.json", "IAM-POLICY.json", "x-policy.jsonl", "iam_policy.json"}
	for _, n := range yes {
		if !isPolicyName(n) {
			t.Errorf("%s soll erkannt werden", n)
		}
	}
	for _, n := range no {
		if isPolicyName(n) {
			t.Errorf("%s soll nicht erkannt werden", n)
		}
	}
}

func TestFind(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"iam-policy.json", "infra/role-policy.json", "infra/deep/iam-policy-x.json", "docs/policy.json", "notes.txt"} {
		write(t, root, rel, "{}")
	}
	for _, skip := range []string{".git", ".terraform", "vendor", "node_modules", ".tmp"} {
		write(t, root, skip+"/iam-policy.json", "{}")
	}
	if err := os.Symlink(filepath.Join(root, "infra"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "iam-policy.json"), filepath.Join(root, "linked-policy.json")); err != nil {
		t.Fatal(err)
	}
	got, err := New(root).Find()
	want := []string{"iam-policy.json", "infra/deep/iam-policy-x.json", "infra/role-policy.json"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Find: got %v, %v; want %v", got, err, want)
	}
}

func TestFindFailsOnMissingRoot(t *testing.T) {
	if _, err := New(filepath.Join(t.TempDir(), "gibt-es-nicht")).Find(); err == nil {
		t.Error("fehlendes Root soll ein Fehler der Suche sein")
	}
}

func TestReadDocumentForms(t *testing.T) {
	envelopes := map[string]string{
		"nacktes Dokument":   wildcardDoc,
		"get-policy-version": `{"PolicyVersion":{"Document":` + wildcardDoc + `,"VersionId":"v1","IsDefaultVersion":true}}`,
		"get-role-policy":    `{"RoleName":"r","PolicyName":"p","PolicyDocument":` + wildcardDoc + `}`,
	}
	for name, content := range envelopes {
		root := t.TempDir()
		write(t, root, "iam-policy.json", content)
		f, err := New(root).Read("iam-policy.json")
		if err != nil || f == nil || len(f.Statements) != 1 {
			t.Errorf("%s: %+v, %v", name, f, err)
			continue
		}
		s := f.Statements[0]
		if s.Effect != "Allow" || len(s.Actions) != 1 || s.Actions[0] != "bedrock:*" || len(s.Resources) != 1 || s.Resources[0] != "*" || s.Line != 0 {
			t.Errorf("%s: Statement %+v", name, s)
		}
	}
}

func TestReadRegionConditionAndDynamic(t *testing.T) {
	root := t.TempDir()
	write(t, root, "iam-policy.json", `{"Statement":[
	  {"Effect":"Allow","Action":"bedrock:InvokeModel","Resource":"*","Condition":{"StringEquals":{"aws:RequestedRegion":["eu-central-1"]}}},
	  {"Effect":"Allow","NotAction":"s3:*","Resource":"*"}]}`)
	f, err := New(root).Read("iam-policy.json")
	if err != nil || f == nil || len(f.Statements) != 2 {
		t.Fatalf("%+v, %v", f, err)
	}
	if !f.Statements[0].HasRegionCondition || len(f.Statements[0].ConditionRegions) != 1 || f.Statements[0].ConditionRegions[0] != "eu-central-1" {
		t.Errorf("Region-Condition: %+v", f.Statements[0])
	}
	if !f.Statements[1].Dynamic {
		t.Errorf("NotAction soll Dynamic sein: %+v", f.Statements[1])
	}
}

// Gültiges JSON ohne erkennbares Policy-Dokument (Fremdformat) ergibt eine leere Liste.
func TestReadForeignFormats(t *testing.T) {
	cases := map[string]string{
		"anderes Objekt":                    `{"Name":"x","Rules":[1,2,3]}`,
		"Array":                             `[1,2,3]`,
		"Zahl":                              `42`,
		"leeres Objekt":                     `{}`,
		"URL-kodiertes Dokument (String)":   `{"PolicyDocument":"%7B%22Statement%22%3A%5B%5D%7D"}`,
		"PolicyVersion ohne Document":       `{"PolicyVersion":{"VersionId":"v1"}}`,
		"PolicyVersion mit String-Dokument": `{"PolicyVersion":{"Document":"abc"}}`,
		"Statement leer":                    `{"Statement":[]}`,
	}
	for name, content := range cases {
		root := t.TempDir()
		write(t, root, "x-policy.json", content)
		f, err := New(root).Read("x-policy.json")
		if err != nil || f == nil || len(f.Statements) != 0 {
			t.Errorf("%s: %+v, %v", name, f, err)
		}
	}
}

func TestReadInvalidAndMissingAndUnreadable(t *testing.T) {
	root := t.TempDir()
	write(t, root, "iam-policy.json", "{\n  \"Statement\": [\n    GEHEIM-nicht-zitieren\n  ]\n}\n")
	_, err := New(root).Read("iam-policy.json")
	var fe *model.FileError
	if !errors.As(err, &fe) || fe.Reason != model.ReasonFileInvalid || fe.Path != "iam-policy.json" || fe.Line != 3 {
		t.Fatalf("FILE_INVALID mit Zeile 3 erwartet, got %v", err)
	}
	if strings.Contains(err.Error(), "GEHEIM") {
		t.Errorf("Fehlertext zitiert Dateiinhalt: %v", err)
	}
	if f, err := New(root).Read("nicht/da.json"); f != nil || err != nil {
		t.Errorf("fehlend: %+v, %v", f, err)
	}
	if err := os.Mkdir(filepath.Join(root, "dir-policy.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := New(root).Read("dir-policy.json"); !errors.As(err, &fe) || fe.Reason != model.ReasonFileUnreadable {
		t.Errorf("Verzeichnis statt Datei soll FILE_UNREADABLE sein: %v", err)
	}
	for _, bad := range []string{"../etwas.json", "/etc/x-policy.json"} {
		if _, err := New(root).Read(bad); !errors.As(err, &fe) || fe.Reason != model.ReasonFileUnreadable {
			t.Errorf("%s soll abgewiesen werden: %v", bad, err)
		}
	}
}

func TestReadToleratesBOM(t *testing.T) {
	root := t.TempDir()
	write(t, root, "iam-policy.json", "\ufeff"+wildcardDoc)
	if f, err := New(root).Read("iam-policy.json"); err != nil || f == nil || len(f.Statements) != 1 {
		t.Errorf("BOM: %+v, %v", f, err)
	}
}
