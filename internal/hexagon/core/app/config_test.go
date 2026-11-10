package app_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeconfig"
)

func TestLoadConfigMissingFileIsDefaults(t *testing.T) {
	c, err := app.LoadConfig(fakeconfig.Loader{})
	if err != nil || !reflect.DeepEqual(c, model.DefaultConfig()) {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	c, err := app.LoadConfig(fakeconfig.Loader{Config: &model.PartialConfig{Mantle: "block"}})
	if err != nil || c.Mantle != model.MantleBlock {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestLoadConfigErrorsAreToolErrors(t *testing.T) {
	for name, l := range map[string]fakeconfig.Loader{
		"kaputt":    {Err: &model.FileError{Path: "bedrock-eu-check.yml", Reason: model.ReasonFileInvalid, Line: 3}},
		"ungueltig": {Config: &model.PartialConfig{BlockedModelPatterns: []string{"("}}},
	} {
		_, err := app.LoadConfig(l)
		if err == nil {
			t.Fatalf("%s: kein Fehler", name)
		}
		if got := app.ExitCode(nil, err); got != 2 {
			t.Errorf("%s: exit=%d", name, got)
		}
	}
	var fe *model.FileError
	_, err := app.LoadConfig(fakeconfig.Loader{Err: &model.FileError{Reason: model.ReasonFileInvalid}})
	if !errors.As(err, &fe) {
		t.Errorf("FileError muss erhalten bleiben: %v", err)
	}
}
