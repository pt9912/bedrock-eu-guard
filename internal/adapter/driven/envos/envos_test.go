package envos_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/envos"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.Environment = envos.Env{}

func TestLookup(t *testing.T) {
	t.Setenv("BEDROCK_EU_TEST_SET", "x")
	t.Setenv("BEDROCK_EU_TEST_EMPTY", "")

	cases := []struct {
		key    string
		want   string
		wantOK bool
	}{
		{"BEDROCK_EU_TEST_SET", "x", true},
		{"BEDROCK_EU_TEST_EMPTY", "", true},
		{"BEDROCK_EU_TEST_UNSET_XYZ", "", false},
	}
	for _, c := range cases {
		got, ok := envos.New().Lookup(c.key)
		if got != c.want || ok != c.wantOK {
			t.Errorf("Lookup(%q) = (%q, %v), want (%q, %v)", c.key, got, ok, c.want, c.wantOK)
		}
	}
}
