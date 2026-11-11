package arch_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestTestdataCoverage hält die Abdeckungsmatrix für LH-QA-003 fest: je
// Dateiart ein Positiv- (good) und ein Negativbeispiel (bad). Für Dateiarten
// ohne Scanner (iam-policy.json) ist die Lücke ausdrücklich benannt und wird
// geloggt, nicht stillschweigend übergangen.
func TestTestdataCoverage(t *testing.T) {
	root := moduleRoot(t)
	base := "internal/adapter/driving/cli/testdata"
	kinds := []struct {
		name string
		rel  string // relativ zu good/project bzw. bad/project
	}{
		{".env", ".env"},
		{"devcontainer.json", ".devcontainer/devcontainer.json"},
		{"docker-compose.yml", "docker-compose.yml"},
		{"Dockerfile", "Dockerfile"},
		{"main.tf", "main.tf"},
		{".claude/settings.json", ".claude/settings.json"},
	}
	for _, k := range kinds {
		for _, variant := range []string{"good", "bad"} {
			p := filepath.Join(root, base, variant, "project", filepath.FromSlash(k.rel))
			if st, err := os.Stat(p); err != nil || st.IsDir() {
				t.Errorf("Testdaten fehlen (LH-QA-003): %s/%s (%s)", variant, k.name, k.rel)
			}
		}
	}
	// Verschachtelte und ausgenommene Varianten des .env-Scanners (LH-FA-LENV-007).
	for _, rel := range []string{"bad/project/apps/api/.env.production", "bad/project/.env.example"} {
		p := filepath.Join(root, base, filepath.FromSlash(rel))
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			t.Errorf("Testdaten fehlen (LH-FA-LENV-007): %s", rel)
		}
	}
	t.Logf("LÜCKE (LH-QA-003): für iam-policy.json gibt es keinen Scanner und keine Testdaten (Soll-Anforderung nach 1.0)")
}

// secretKey erkennt Schlüsselnamen, deren Werte geheim sein könnten.
var secretKey = regexp.MustCompile(`(?i)(KEY|TOKEN|SECRET|PASSWORD|CREDENTIAL)`)

// TestDotenvTestdataIsSynthetic stellt sicher, dass .env-Testdaten nur
// offensichtlich synthetische Werte enthalten (Marker "GEHEIM"), damit
// Secret-Scanner und Reviews sie nicht für echte Schlüssel halten.
func TestDotenvTestdataIsSynthetic(t *testing.T) {
	root := moduleRoot(t)
	base := filepath.Join(root, "internal", "adapter", "driving", "cli", "testdata")
	found := 0
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if n := d.Name(); n != ".env" && !strings.HasPrefix(n, ".env.") {
			return nil
		}
		found++
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "export "))
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			if secretKey.MatchString(k) && !strings.Contains(v, "GEHEIM") {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("%s:%d: Wert von %s ohne Marker GEHEIM (nur synthetische Testwerte erlaubt)", rel, i+1, strings.TrimSpace(k))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if found < 4 {
		t.Fatalf("nur %d .env-Testdateien gefunden; Pfad prüfen", found)
	}
}
