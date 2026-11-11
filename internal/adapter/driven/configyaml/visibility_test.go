package configyaml_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/configyaml"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// slice-028: visibilityRegions — abwesend (nil) und explizit leer ([]) bleiben
// unterscheidbar; Einträge, die keine Strings sind, ergeben FILE_INVALID ohne Wertzitat.
func TestVisibilityRegions(t *testing.T) {
	p, err := configyaml.Parse("c.yml", []byte("mantle: warn\n"))
	if err != nil || p.VisibilityRegions != nil {
		t.Errorf("abwesend: %+v, %v", p, err)
	}
	p, err = configyaml.Parse("c.yml", []byte("visibilityRegions: []\n"))
	if err != nil || p.VisibilityRegions == nil || len(p.VisibilityRegions) != 0 {
		t.Errorf("leer: %#v, %v", p, err)
	}
	p, err = configyaml.Parse("c.yml", []byte("visibilityRegions:\n  - us-east-1\n  - us-west-2\n"))
	if err != nil || len(p.VisibilityRegions) != 2 || p.VisibilityRegions[0] != "us-east-1" || p.VisibilityRegions[1] != "us-west-2" {
		t.Errorf("Liste: %#v, %v", p, err)
	}
	for _, in := range []string{"visibilityRegions: GEHEIM\n", "visibilityRegions:\n  - {a: GEHEIM}\n", "visibilityRegions:\n  - [GEHEIM]\n"} {
		_, err := configyaml.Parse("c.yml", []byte(in))
		var fe *model.FileError
		if !errors.As(err, &fe) || fe.Reason != model.ReasonFileInvalid {
			t.Errorf("%q: want FILE_INVALID, got %v", in, err)
		} else if strings.Contains(err.Error(), "GEHEIM") {
			t.Errorf("%q: Fehlertext zitiert Wert: %v", in, err)
		}
	}
}
