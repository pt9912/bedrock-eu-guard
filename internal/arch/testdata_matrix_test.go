package arch_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTestdataCoverage hält die Abdeckungsmatrix für LH-QA-003 fest: je
// Dateiart ein Positiv- (good) und ein Negativbeispiel (bad). Für Dateiarten
// ohne Scanner (.env, iam-policy.json) ist die Lücke ausdrücklich benannt
// und wird geloggt, nicht stillschweigend übergangen.
func TestTestdataCoverage(t *testing.T) {
	root := moduleRoot(t)
	base := "internal/adapter/driving/cli/testdata"
	kinds := []struct {
		name string
		rel  string // relativ zu good/project bzw. bad/project
	}{
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
	for _, gap := range []string{".env", "iam-policy.json"} {
		t.Logf("LÜCKE (LH-QA-003): für %s gibt es keinen Scanner und keine Testdaten; Klärung offen", gap)
	}
}
