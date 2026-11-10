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
