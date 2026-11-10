package model_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

func TestResolveEmptyIsDefault(t *testing.T) {
	c, err := model.Resolve(model.PartialConfig{})
	if err != nil || !reflect.DeepEqual(c, model.DefaultConfig()) {
		t.Fatalf("got %+v, %v", c, err)
	}
	if c.Mantle != model.MantleWarn {
		t.Fatalf("mantle default: %q", c.Mantle)
	}
}

func TestResolveOverridesOnlyGivenFields(t *testing.T) {
	c, err := model.Resolve(model.PartialConfig{
		AllowedRegions:  []string{"eu-central-1", "eu-west-1"},
		AllowedAccounts: []string{"123456789012"},
		Mantle:          "allow",
	})
	if err != nil {
		t.Fatal(err)
	}
	d := model.DefaultConfig()
	if !reflect.DeepEqual(c.AllowedRegions, []string{"eu-central-1", "eu-west-1"}) ||
		!reflect.DeepEqual(c.AllowedAccounts, []string{"123456789012"}) ||
		c.Mantle != model.MantleAllow ||
		!reflect.DeepEqual(c.AllowedModelPatterns, d.AllowedModelPatterns) ||
		!reflect.DeepEqual(c.BlockedModelPatterns, d.BlockedModelPatterns) {
		t.Fatalf("got %+v", c)
	}
}

func TestResolveExplicitEmptyLists(t *testing.T) {
	c, err := model.Resolve(model.PartialConfig{AllowedModelPatterns: []string{}, AllowedAccounts: []string{}})
	if err != nil || len(c.AllowedModelPatterns) != 0 || c.AllowedModelPatterns == nil {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestResolveErrors(t *testing.T) {
	secret := "sk-geheim-(["
	cases := map[string]struct {
		p      model.PartialConfig
		field  string
		reason string
	}{
		"regex allow":   {model.PartialConfig{AllowedModelPatterns: []string{secret}}, "allowedModelPatterns", model.ConfigInvalidRegex},
		"regex block":   {model.PartialConfig{BlockedModelPatterns: []string{`ok`, `(`}}, "blockedModelPatterns", model.ConfigInvalidRegex},
		"regex profile": {model.PartialConfig{AllowedProfiles: []string{`[`}}, "allowedProfiles", model.ConfigInvalidRegex},
		"mantle":        {model.PartialConfig{Mantle: "maybe"}, "mantle", model.ConfigInvalidValue},
		"regions leer":  {model.PartialConfig{AllowedRegions: []string{}}, "allowedRegions", model.ConfigEmptyRegions},
		"region leer":   {model.PartialConfig{AllowedRegions: []string{""}}, "allowedRegions", model.ConfigInvalidValue},
		"account":       {model.PartialConfig{AllowedAccounts: []string{"123"}}, "allowedAccounts", model.ConfigInvalidAccount},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := model.Resolve(c.p)
			var ce *model.ConfigError
			if !errors.As(err, &ce) || ce.Field != c.field || ce.Reason != c.reason {
				t.Fatalf("got %v", err)
			}
			if strings.Contains(err.Error(), "geheim") {
				t.Fatalf("Feldwert im Fehler: %v", err)
			}
		})
	}
}

func TestResolveCollectsAllErrorsWithIndex(t *testing.T) {
	_, err := model.Resolve(model.PartialConfig{BlockedModelPatterns: []string{"ok", "("}, Mantle: "x"})
	if err == nil || !strings.Contains(err.Error(), "blockedModelPatterns[1]") || !strings.Contains(err.Error(), "mantle") {
		t.Fatalf("got %v", err)
	}
}

func TestExcluded(t *testing.T) {
	c := model.Config{Exclude: []string{"testdata", "docs/*/gen", "./build/", "*.tmp.tf", "  "}}
	for rel, want := range map[string]bool{
		"testdata":             true,
		"testdata/bad/main.tf": true,
		"a/testdata/main.tf":   false, // Muster ist relativ zum Scan-Root
		"docs/x/gen/file.md":   true,
		"docs/x/other/file.md": false,
		"build/out/Dockerfile": true,
		"./build/x":            true,
		"main.tmp.tf":          true,
		"sub/main.tmp.tf":      false, // '*' überschreitet kein '/'
		"src/main.tf":          false,
		`testdata\win\main.tf`: true,
		"":                     false,
	} {
		if got := c.Excluded(rel); got != want {
			t.Errorf("Excluded(%q)=%v want %v", rel, got, want)
		}
	}
	if (model.Config{}).Excluded("anything") {
		t.Error("ohne Muster nichts ausgeschlossen")
	}
}

func TestResolveExclude(t *testing.T) {
	c, err := model.Resolve(model.PartialConfig{Exclude: []string{"testdata", "docs/*/gen"}})
	if err != nil || len(c.Exclude) != 2 {
		t.Fatalf("got %+v, %v", c, err)
	}
	if c, err := model.Resolve(model.PartialConfig{}); err != nil || c.Exclude != nil {
		t.Fatalf("Default: %+v, %v", c, err)
	}
	for name, pat := range map[string]string{"leer": "", "nur Slash": "/", "kaputt": "a[b"} {
		_, err := model.Resolve(model.PartialConfig{Exclude: []string{"ok", pat}})
		var ce *model.ConfigError
		if !errors.As(err, &ce) || ce.Field != "exclude" || ce.Index != 1 {
			t.Errorf("%s: %v", name, err)
		}
	}
}
