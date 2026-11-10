package fs_test

import (
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/fs"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.Filesystem = fs.FS{}

func setup(t *testing.T) fs.FS {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"a/x.tf", "a/y.tf", "b/z.txt"} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("inhalt"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return fs.New(root)
}

func TestReadFile(t *testing.T) {
	f := setup(t)
	b, err := f.ReadFile("a/x.tf")
	if err != nil || string(b) != "inhalt" {
		t.Fatalf("got %q, %v", b, err)
	}
	if _, err := f.ReadFile("fehlt.txt"); !errors.Is(err, iofs.ErrNotExist) {
		t.Errorf("fehlende Datei: want ErrNotExist, got %v", err)
	}
}

func TestExists(t *testing.T) {
	f := setup(t)
	for path, want := range map[string]bool{"a/x.tf": true, "a": true, "nope": false} {
		got, err := f.Exists(path)
		if err != nil || got != want {
			t.Errorf("Exists(%q) = %v, %v; want %v", path, got, err, want)
		}
	}
}

func TestGlobSortedRelative(t *testing.T) {
	f := setup(t)
	got, err := f.Glob("a/*.tf")
	if err != nil || !reflect.DeepEqual(got, []string{"a/x.tf", "a/y.tf"}) {
		t.Errorf("got %v, %v", got, err)
	}
	none, err := f.Glob("c/*.tf")
	if err != nil || len(none) != 0 {
		t.Errorf("got %v, %v", none, err)
	}
}

func TestRejectsPathsOutsideRoot(t *testing.T) {
	f := setup(t)
	for _, p := range []string{"../etc/passwd", "a/../../x", "/etc/passwd", ".."} {
		if _, err := f.ReadFile(p); !errors.Is(err, fs.ErrOutsideRoot) {
			t.Errorf("ReadFile(%q): want ErrOutsideRoot, got %v", p, err)
		}
		if _, err := f.Exists(p); !errors.Is(err, fs.ErrOutsideRoot) {
			t.Errorf("Exists(%q): want ErrOutsideRoot, got %v", p, err)
		}
		if _, err := f.Glob(p); !errors.Is(err, fs.ErrOutsideRoot) {
			t.Errorf("Glob(%q): want ErrOutsideRoot, got %v", p, err)
		}
	}
}
