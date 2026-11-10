package arch_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/arch"
)

// moduleRoot sucht go.mod oberhalb des Testverzeichnisses.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod nicht gefunden")
		}
		dir = parent
	}
}

// TestArchitecture wendet die Regeln auf alle Produktionsdateien an
// (ARC-001 bis ARC-007). Testdateien (_test.go), vendor/, .tmp/ und
// testdata/ werden nicht geprüft.
func TestArchitecture(t *testing.T) {
	root := moduleRoot(t)
	var violations []arch.Violation
	files := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", ".tmp", ".git", "testdata", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		dir := filepath.ToSlash(rel)
		files++
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			violations = append(violations, arch.Check(dir, p)...)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files < 20 {
		t.Fatalf("nur %d Produktionsdateien gefunden; Modulwurzel falsch?", files)
	}
	var msgs []string
	for _, v := range violations {
		msgs = append(msgs, v.String())
	}
	sort.Strings(msgs)
	for _, m := range msgs {
		t.Error(m)
	}
}

// TestRulesDetectViolations stellt sicher, dass die Regeln tatsächlich
// anschlagen (sonst wäre der Architekturtest wirkungslos).
func TestRulesDetectViolations(t *testing.T) {
	m := arch.Modul
	cases := []struct {
		name string
		dir  string
		imp  string
		rule string // erwartete Regel; "" = kein Verstoß
	}{
		{"model importiert Intern", arch.Model, m + "/internal/hexagon/port/driven", "ARC-003"},
		{"model importiert Stdlib", arch.Model, "regexp", ""},
		{"model importiert os", arch.Model, "os", "ARC-003"},
		{"rules importiert Adapter", arch.Rules, m + "/internal/adapter/driven/fs", "ARC-004"},
		{"rules importiert app", arch.Rules, m + "/internal/hexagon/core/app", "ARC-004"},
		{"rules importiert model", arch.Rules, m + "/internal/hexagon/core/model", ""},
		{"rules importiert Port", arch.Rules, m + "/internal/hexagon/port/driven", ""},
		{"rules importiert Drittbibliothek", arch.Rules, "gopkg.in/yaml.v3", "ARC-003"},
		{"app importiert Adapter", arch.App, m + "/internal/adapter/driving/cli", "ARC-002"},
		{"app importiert rules", arch.App, m + "/internal/hexagon/core/rules", ""},
		{"app importiert net/http", arch.App, "net/http", "ARC-003"},
		{"port importiert rules", arch.Port, m + "/internal/hexagon/core/rules", "ARC-005"},
		{"port importiert model", arch.Port, m + "/internal/hexagon/core/model", ""},
		{"coretest importiert app", arch.CoreTest + "/fakeenv", m + "/internal/hexagon/core/app", "ARC-003"},
		{"adapter importiert rules", arch.Driven + "/report", m + "/internal/hexagon/core/rules", "ARC-006"},
		{"adapter importiert app", arch.Driven + "/report", m + "/internal/hexagon/core/app", "ARC-006"},
		{"adapter importiert Test-Double", arch.Driven + "/fs", m + "/internal/hexagon/coretest/memfs", "ARC-006"},
		{"adapter importiert Hilfspaket", arch.Driven + "/claudesettings", m + "/internal/adapter/driven/secretkey", ""},
		{"adapter importiert model", arch.Driven + "/fs", m + "/internal/hexagon/core/model", ""},
		{"adapter importiert yaml", arch.Driven + "/configyaml", "gopkg.in/yaml.v3", ""},
		{"os/exec in fs", arch.Driven + "/fs", "os/exec", "ARC-007"},
		{"os/exec in awscli", arch.AWSDoor, "os/exec", ""},
		{"os/exec im Kern", arch.Rules, "os/exec", "ARC-003"},
		{"AWS-SDK im Adapter fs", arch.Driven + "/fs", "github.com/aws/aws-sdk-go-v2/aws", "ARC-007"},
		{"AWS-SDK in awscli", arch.AWSDoor, "github.com/aws/aws-sdk-go-v2/aws", ""},
		{"cli importiert alles", arch.Driving + "/cli", m + "/internal/hexagon/core/app", ""},
		{"cli importiert Test-Double", arch.Driving + "/cli", m + "/internal/hexagon/coretest/fakeenv", "ARC-001"},
		{"cmd importiert Test-Double", arch.Cmd + "/bedrock-eu-check", m + "/internal/hexagon/coretest/fakeenv", "ARC-001"},
		{"cmd importiert cli", arch.Cmd + "/bedrock-eu-check", m + "/internal/adapter/driving/cli", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := arch.Check(c.dir, c.imp)
			if c.rule == "" {
				if len(vs) != 0 {
					t.Fatalf("kein Verstoß erwartet: %v", vs)
				}
				return
			}
			for _, v := range vs {
				if v.Rule == c.rule {
					return
				}
			}
			t.Fatalf("Verstoß %s erwartet, got %v", c.rule, vs)
		})
	}
}
