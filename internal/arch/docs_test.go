package arch_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// stringConsts liefert alle String-Konstanten (Name → Wert) der
// Produktionsdateien eines Pakets.
func stringConsts(t *testing.T, dir string) map[string]string {
	t.Helper()
	root := moduleRoot(t)
	files, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(dir), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(token.NewFileSet(), f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range af.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, sp := range gd.Specs {
				vs := sp.(*ast.ValueSpec)
				for i, n := range vs.Names {
					if i >= len(vs.Values) {
						continue
					}
					if bl, ok := vs.Values[i].(*ast.BasicLit); ok && bl.Kind == token.STRING {
						v, err := strconv.Unquote(bl.Value)
						if err != nil {
							t.Fatal(err)
						}
						out[n.Name] = v
					}
				}
			}
		}
	}
	return out
}

func readDoc(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(moduleRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// referenceOnly sind Reason-Codes, die keine Regel-Findings sind, sondern
// Fehler der eigenen Konfiguration bzw. Tool-Fehler. Sie stehen
// stattdessen in docs/user/reference.md (und werden dort geprüft).
func referenceOnly(name string) bool {
	return strings.HasPrefix(name, "Config") || name == "AWSCLIMissing"
}

var codeToken = regexp.MustCompile("`([A-Z][A-Z0-9_]*[A-Z0-9])`")

// TestDocsMatchReasonCodes verhindert Doku-Drift: jeder Reason-Code im Kern
// steht im Regelkatalog (bzw. die Konfigurations- und Tool-Fehler-Gründe in
// der Referenz), und jeder Code in der Spalte „Reason-Codes“ des Katalogs
// existiert im Code.
func TestDocsMatchReasonCodes(t *testing.T) {
	consts := stringConsts(t, "internal/hexagon/core/model")
	rulesDoc := readDoc(t, "docs/user/rules.md")
	refDoc := readDoc(t, "docs/user/reference.md")

	known := map[string]bool{}
	for name, val := range consts {
		isReason := strings.HasPrefix(name, "Reason") || strings.HasPrefix(name, "AWS") || strings.HasPrefix(name, "Config")
		if !isReason || !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(val) {
			continue
		}
		known[val] = true
		doc, where := rulesDoc, "docs/user/rules.md"
		if referenceOnly(name) {
			doc, where = refDoc, "docs/user/reference.md"
		}
		if !strings.Contains(doc, "`"+val+"`") {
			t.Errorf("Reason-Code %s (%s) fehlt in %s", val, name, where)
		}
	}
	if len(known) < 40 {
		t.Fatalf("nur %d Reason-Codes im Code gefunden; Parser/Pfad prüfen", len(known))
	}

	// Rückrichtung: Codes in der letzten Spalte der Katalogtabellen
	for _, line := range strings.Split(rulesDoc, "\n") {
		if !strings.HasPrefix(line, "| [`LH-") {
			continue
		}
		cells := strings.Split(line, "|")
		last := cells[len(cells)-2] // letzte Zelle vor dem abschließenden "|"
		for _, m := range codeToken.FindAllStringSubmatch(last, -1) {
			if !known[m[1]] {
				t.Errorf("Katalog nennt unbekannten Reason-Code %s (Zeile: %.60s…)", m[1], line)
			}
		}
	}
}

// TestDocsListAllRules: jede Regel-ID aus core/rules steht im Katalog.
func TestDocsListAllRules(t *testing.T) {
	ids := stringConsts(t, "internal/hexagon/core/rules")
	rulesDoc := readDoc(t, "docs/user/rules.md")
	var missing []string
	count := 0
	for name, val := range ids {
		if !strings.HasPrefix(name, "ID") || !strings.HasPrefix(val, "LH-") {
			continue
		}
		count++
		if !strings.Contains(rulesDoc, "`"+val+"`") {
			missing = append(missing, val+" ("+name+")")
		}
	}
	if count < 25 {
		t.Fatalf("nur %d Regel-IDs im Code gefunden; Parser/Pfad prüfen", count)
	}
	sort.Strings(missing)
	for _, m := range missing {
		t.Errorf("Regel %s fehlt in docs/user/rules.md", m)
	}
}
