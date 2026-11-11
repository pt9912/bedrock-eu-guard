package envsource

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// skipDirs werden bei der rekursiven Suche nicht betreten. ".tmp" enthält Scratch
// und die Worktrees von "make verify-clean".
var skipDirs = map[string]bool{
	".git": true, ".terraform": true, "vendor": true, "node_modules": true, ".tmp": true,
}

// exampleSuffixes kennzeichnen Beispieldateien, die keine echte Konfiguration sind.
var exampleSuffixes = []string{".example", ".sample", ".template"}

// isDotenvName meldet, ob ein Dateiname eine zu prüfende .env-Datei ist: ".env" oder
// ".env.<name>" (z. B. ".env.production"), aber keine Beispieldatei.
func isDotenvName(name string) bool {
	if name != ".env" && !strings.HasPrefix(name, ".env.") {
		return false
	}
	for _, suf := range exampleSuffixes {
		if strings.HasSuffix(name, suf) {
			return false
		}
	}
	return true
}

// Find sucht Dateien der angegebenen Art rekursiv unter dem Scan-Root und liefert
// sortierte, relative Pfade mit "/". Unterstützt wird nur die Art dotenv; für alle
// anderen Arten ist die Dateisuche Sache des Kerns (Glob), Find liefert dort nichts.
// Symlinks werden nicht verfolgt.
func (r Reader) Find(kind model.EnvSourceKind) ([]string, error) {
	if kind != dotenvKind {
		return nil, nil
	}
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
		if d.Type().IsRegular() && isDotenvName(d.Name()) {
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
