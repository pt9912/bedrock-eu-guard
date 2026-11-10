// Package terraform implementiert den TerraformReader-Port (ARC-006): findet
// *.tf-Dateien und liest daraus nur Fakten (Provider-Regionen, IAM-Statements,
// Modell-IDs). Ausdrücke werden nicht ausgewertet, es gibt keine .tfvars,
// Variablenwerte oder übrigen Literale im Ergebnis (LH-NF-004).
package terraform

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// skipDirs werden bei der Suche nicht betreten. ".tmp" enthält Scratch und
// die Worktrees von "make verify-clean".
var skipDirs = map[string]bool{
	".git": true, ".terraform": true, "vendor": true, "node_modules": true, ".tmp": true,
}

// Reader liest Terraform-Dateien unterhalb eines Scan-Roots.
type Reader struct{ root string }

// New liefert einen Reader für root.
func New(root string) Reader { return Reader{root: root} }

// Find liefert alle *.tf-Dateien rekursiv als sortierte, relative Pfade mit
// "/". Symlinks werden nicht verfolgt.
func (r Reader) Find() ([]string, error) {
	var out []string
	err := filepath.WalkDir(r.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != r.root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".tf") {
			rel, relErr := filepath.Rel(r.root, path)
			if relErr != nil {
				return relErr
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// Read liest eine Datei. Fehlt sie, gilt (nil, nil); kaputt oder nicht lesbar
// ergibt ein *model.FileError ohne Dateiinhalt.
func (r Reader) Read(rel string) (*model.TerraformFile, error) {
	clean := filepath.Clean(rel)
	display := filepath.ToSlash(clean)
	if filepath.IsAbs(rel) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, &model.FileError{Path: rel, Reason: model.ReasonFileUnreadable}
	}
	data, err := os.ReadFile(filepath.Join(r.root, clean))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, &model.FileError{Path: display, Reason: model.ReasonFileUnreadable}
	}
	f, ferr := Parse(display, data)
	if ferr != nil {
		return nil, ferr
	}
	return f, nil
}

// Parse extrahiert die Fakten aus dem Inhalt einer .tf-Datei.
func Parse(path string, data []byte) (*model.TerraformFile, *model.FileError) {
	file, diags := hclsyntax.ParseConfig(data, path, hcl.InitialPos)
	if diags.HasErrors() {
		line := 0
		for _, d := range diags {
			if d.Severity == hcl.DiagError && d.Subject != nil {
				line = d.Subject.Start.Line
				break
			}
		}
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid, Line: line}
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid}
	}
	f := &model.TerraformFile{Path: path}
	for _, b := range body.Blocks {
		switch {
		case b.Type == "provider" && len(b.Labels) == 1 && b.Labels[0] == "aws":
			f.Providers = append(f.Providers, providerFrom(b))
		case b.Type == "data" && len(b.Labels) == 2 && b.Labels[0] == "aws_iam_policy_document":
			f.Statements = append(f.Statements, documentStatements(b)...)
		case b.Type == "resource" && len(b.Labels) == 2 && policyResources[b.Labels[0]]:
			f.Statements = append(f.Statements, resourcePolicy(b)...)
		}
	}
	f.ModelRefs = modelRefs(body)
	return f, nil
}
