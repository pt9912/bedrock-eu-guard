package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenvsource"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakesettings"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/faketerraform"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/memfs"
)

func TestScanAllFilesMissing(t *testing.T) {
	fs, err := app.Scan{Settings: fakesettings.Reader{}}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 || fs[0].Reason != model.ReasonFileMissing || fs[0].Severity != model.WARN {
		t.Fatalf("got %+v", fs)
	}
	if app.ExitCode(fs, nil) != 0 {
		t.Fatal("WARN darf nicht Exit 1 sein")
	}
}

func TestScanDirectAnthropicInUserSettings(t *testing.T) {
	u := &model.ClaudeSettings{Env: []model.EnvEntry{
		{Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1", Line: 2},
		{Key: "ANTHROPIC_API_KEY", Secret: true, Line: 3}}}
	fs, _ := app.Scan{Settings: fakesettings.Reader{User: u}}.Run(model.DefaultConfig())
	ids := map[string]bool{}
	for _, f := range fs {
		if f.Severity == model.FAIL {
			ids[f.RuleID] = true
		}
	}
	for _, id := range []string{"LH-FA-CLAUDE-001", "LH-FA-CLAUDE-003", "LH-FA-CLAUDE-004"} {
		if !ids[id] {
			t.Errorf("fehlendes FAIL %s: %+v", id, fs)
		}
	}
	if app.ExitCode(fs, nil) != 1 {
		t.Fatal("exit != 1")
	}
}

func TestScanFileErrorsBecomeFindingsNotToolErrors(t *testing.T) {
	r := fakesettings.Reader{
		UserErr:    &model.FileError{Path: "~/.claude/settings.json", Reason: model.ReasonFileInvalid, Line: 4},
		ProjectErr: map[string]error{".claude/settings.json": errors.New("permission denied")},
		Project: map[string]*model.ClaudeSettings{
			".claude/settings.local.json": {Env: []model.EnvEntry{{Key: "AWS_REGION", Value: "eu-central-1"}}}},
	}
	fs, err := app.Scan{Settings: r}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatalf("Dateifehler sind kein Tool-Fehler: %v", err)
	}
	var invalid, unreadable, passed bool
	for _, f := range fs {
		switch {
		case f.Reason == model.ReasonFileInvalid && f.Line == 4:
			invalid = true
		case f.Reason == model.ReasonFileUnreadable && f.File == ".claude/settings.json":
			unreadable = true
		case f.RuleID == "LH-FA-CLAUDE-002" && f.Severity == model.PASS && f.File == ".claude/settings.local.json":
			passed = true
		}
	}
	if !invalid || !unreadable || !passed {
		t.Fatalf("invalid=%v unreadable=%v passed=%v: %+v", invalid, unreadable, passed, fs)
	}
	if app.ExitCode(fs, nil) != 1 {
		t.Fatal("exit != 1")
	}
}

func TestScanStableOrder(t *testing.T) {
	u := &model.ClaudeSettings{Env: []model.EnvEntry{{Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1"}, {Key: "AWS_REGION", Value: "eu-central-1"}}}
	fs, _ := app.Scan{Settings: fakesettings.Reader{User: u}}.Run(model.DefaultConfig())
	for i := 1; i < len(fs); i++ {
		if fs[i-1].RuleID > fs[i].RuleID {
			t.Fatalf("nicht sortiert: %+v", fs)
		}
	}
}

func TestMemFS(t *testing.T) {
	m := memfs.FS{".claude/settings.json": "{}", ".claude/settings.local.json": "{}", "a.txt": "x"}
	if ok, _ := m.Exists("a.txt"); !ok {
		t.Error("Exists")
	}
	if _, err := m.ReadFile("nope"); err == nil {
		t.Error("ReadFile nope")
	}
	g, _ := m.Glob(".claude/*.json")
	if len(g) != 2 {
		t.Errorf("Glob: %v", g)
	}
}

func TestScanEnvSourcesDiscoveryAndRules(t *testing.T) {
	files := memfs.FS{
		".devcontainer/devcontainer.json": "{}",
		"docker-compose.yml":              "x",
		"docker-compose.override.yml":     "x",
		"Dockerfile":                      "x",
		"Dockerfile.dev":                  "x",
		".dockerignore":                   "x",
		"README.md":                       "x",
	}
	env := func(kvs ...model.EnvEntry) *model.EnvFile { return &model.EnvFile{Env: kvs} }
	r := fakeenvsource.Reader{
		Files: map[string]*model.EnvFile{
			".devcontainer/devcontainer.json": env(model.EnvEntry{Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1"}, model.EnvEntry{Key: "AWS_REGION", Value: "eu-central-1"}),
			"docker-compose.yml":              env(model.EnvEntry{Key: "AWS_REGION", Value: "us-east-1", Line: 7}),
			"Dockerfile":                      env(model.EnvEntry{Key: "AWS_REGION", Value: "eu-central-1"}),
		},
		Errs: map[string]error{"Dockerfile.dev": &model.FileError{Path: "Dockerfile.dev", Reason: model.ReasonFileInvalid, Line: 2}},
	}
	fs, err := app.Scan{Settings: fakesettings.Reader{}, EnvSources: r, Files: files}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range fs {
		if f.File != "" {
			got[f.RuleID+"|"+f.File] = f.Severity.String()
		}
	}
	want := map[string]string{
		"LH-FA-DEVCON-001|.devcontainer/devcontainer.json": "PASS",
		"LH-FA-DOCKER-001|docker-compose.yml":              "FAIL",
		"LH-FA-DOCKER-002|Dockerfile":                      "PASS",
		"LH-FA-DOCKER-002|Dockerfile.dev":                  "FAIL",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q want %q (alle: %v)", k, got[k], v, got)
		}
	}
	if _, ok := got["LH-FA-DOCKER-001|docker-compose.override.yml"]; ok {
		t.Errorf("Override ohne relevante Variablen darf kein Finding erzeugen: %v", got)
	}
}

func TestScanWithoutEnvSourcesIsUnchanged(t *testing.T) {
	a, _ := app.Scan{Settings: fakesettings.Reader{}}.Run(model.DefaultConfig())
	b, _ := app.Scan{Settings: fakesettings.Reader{}, EnvSources: fakeenvsource.Reader{}}.Run(model.DefaultConfig())
	if len(a) != len(b) {
		t.Fatalf("%d vs %d", len(a), len(b))
	}
}

func TestScanTerraform(t *testing.T) {
	r := faketerraform.Reader{
		Files: map[string]*model.TerraformFile{
			"main.tf":      {Providers: []model.AWSProvider{{Region: "us-east-1", Line: 3}}},
			"mod/iam.tf":   {Statements: []model.IAMStatement{{Effect: "Allow", Actions: []string{"bedrock:*"}, Resources: []string{"*"}, Line: 7}}},
			"mod/model.tf": {ModelRefs: []model.ModelRef{{Value: "eu.anthropic.claude-sonnet-4-6", Line: 1}}},
		},
		Errs: map[string]error{"broken.tf": &model.FileError{Path: "broken.tf", Reason: model.ReasonFileInvalid, Line: 2}},
	}
	fs, err := app.Scan{Settings: fakesettings.Reader{}, Terraform: r}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range fs {
		if f.File != "" && f.RuleID[:9] == "LH-FA-TF-" {
			got[f.RuleID+"|"+f.File] = f.Severity.String()
		}
	}
	want := map[string]string{
		"LH-FA-TF-002|main.tf":      "FAIL",
		"LH-FA-TF-003|mod/iam.tf":   "FAIL",
		"LH-FA-TF-004|mod/model.tf": "PASS",
		"LH-FA-TF-001|broken.tf":    "FAIL",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q want %q (alle: %v)", k, got[k], v, got)
		}
	}
	if app.ExitCode(fs, nil) != 1 {
		t.Error("exit != 1")
	}
}

func TestScanTerraformNoFilesAndFindError(t *testing.T) {
	fs, _ := app.Scan{Settings: fakesettings.Reader{}, Terraform: faketerraform.Reader{}}.Run(model.DefaultConfig())
	var warn bool
	for _, f := range fs {
		if f.RuleID == "LH-FA-TF-001" && f.Severity == model.WARN {
			warn = true
		}
	}
	if !warn {
		t.Errorf("WARN erwartet: %+v", fs)
	}
	fs, err := app.Scan{Settings: fakesettings.Reader{}, Terraform: faketerraform.Reader{FindErr: errors.New("boom")}}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatalf("Suchfehler ist kein Tool-Fehler: %v", err)
	}
	var fail bool
	for _, f := range fs {
		if f.RuleID == "LH-FA-TF-001" && f.Severity == model.FAIL {
			fail = true
		}
	}
	if !fail {
		t.Errorf("FAIL erwartet: %+v", fs)
	}
}

func TestScanOnlyTerraformSkipsClaudeRules(t *testing.T) {
	r := faketerraform.Reader{Files: map[string]*model.TerraformFile{
		"main.tf": {Providers: []model.AWSProvider{{Region: "eu-central-1", Line: 2}}}}}
	fs, err := app.Scan{Terraform: r}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fs {
		if len(f.RuleID) >= 12 && f.RuleID[:12] == "LH-FA-CLAUDE" {
			t.Errorf("Claude-Regel trotz nil-Reader: %+v", f)
		}
	}
	if len(fs) == 0 {
		t.Fatal("Terraform-Findings erwartet")
	}
}

func TestScanAllReadersNilIsEmpty(t *testing.T) {
	fs, err := app.Scan{}.Run(model.DefaultConfig())
	if err != nil || len(fs) != 0 {
		t.Fatalf("got %+v, %v", fs, err)
	}
}

func TestScanExcludeFiltersAllDiscoveries(t *testing.T) {
	env := func(kvs ...model.EnvEntry) *model.EnvFile { return &model.EnvFile{Env: kvs} }
	bad := env(model.EnvEntry{Key: "AWS_REGION", Value: "us-east-1", Line: 1})
	good := env(model.EnvEntry{Key: "AWS_REGION", Value: "eu-central-1", Line: 1})
	r := fakeenvsource.Reader{Files: map[string]*model.EnvFile{"Dockerfile": good, "Dockerfile.bad": bad, "testdata/Dockerfile": bad}}
	fsys := memfs.FS{"Dockerfile": "x", "Dockerfile.bad": "x", "testdata/Dockerfile": "x"}
	tf := faketerraform.Reader{Files: map[string]*model.TerraformFile{
		"main.tf":         {Providers: []model.AWSProvider{{Region: "eu-central-1"}}},
		"testdata/bad.tf": {Providers: []model.AWSProvider{{Region: "us-east-1"}}},
	}}
	st := fakesettings.Reader{Project: map[string]*model.ClaudeSettings{
		".claude/settings.json": {Env: []model.EnvEntry{{Key: "AWS_REGION", Value: "us-east-1"}}}}}

	run := func(cfg model.Config) map[string]string {
		fs, err := app.Scan{Settings: st, EnvSources: r, Files: fsys, Terraform: tf}.Run(cfg)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, f := range fs {
			if f.File != "" {
				got[f.RuleID+"|"+f.File] = f.Severity.String()
			}
		}
		return got
	}
	cfg := model.DefaultConfig()
	all := run(cfg)
	if all["LH-FA-DOCKER-002|Dockerfile.bad"] != "FAIL" || all["LH-FA-TF-002|testdata/bad.tf"] != "FAIL" || all["LH-FA-CLAUDE-002|.claude/settings.json"] != "FAIL" {
		t.Fatalf("ohne exclude: %v", all)
	}
	cfg.Exclude = []string{"testdata", ".claude", "Dockerfile.bad"}
	ex := run(cfg)
	for k := range ex {
		if strings.Contains(k, "testdata") || strings.HasSuffix(k, "|.claude/settings.json") || strings.Contains(k, "Dockerfile.bad") {
			t.Errorf("ausgeschlossen, aber gemeldet: %s (%v)", k, ex)
		}
	}
	if ex["LH-FA-DOCKER-002|Dockerfile"] != "PASS" || ex["LH-FA-TF-002|main.tf"] != "PASS" {
		t.Errorf("Rest bleibt geprüft: %v", ex)
	}
}

func TestScanDotenvDiscoveryExcludeAndErrors(t *testing.T) {
	good := &model.EnvFile{Env: []model.EnvEntry{{Key: "AWS_REGION", Value: "eu-central-1", Line: 1}}}
	bad := &model.EnvFile{Env: []model.EnvEntry{{Key: "ANTHROPIC_API_KEY", Secret: true, Line: 2}}}
	r := fakeenvsource.Reader{
		Files: map[string]*model.EnvFile{".env": good, "apps/api/.env": bad, "testdata/.env": bad},
		Errs:  map[string]error{"apps/web/.env.local": &model.FileError{Path: "apps/web/.env.local", Reason: model.ReasonFileInvalid, Line: 3}},
	}
	finder := fakeenvsource.Finder{Paths: map[model.EnvSourceKind][]string{
		model.EnvDotenv: {"testdata/.env", "apps/web/.env.local", ".env", "apps/api/.env"}}}
	cfg := model.DefaultConfig()
	cfg.Exclude = []string{"testdata"}
	fs, err := app.Scan{EnvSources: r, EnvFinder: finder}.Run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range fs {
		if f.RuleID == "LH-FA-LENV-007" {
			got[f.File] = f.Severity.String() + "/" + f.Reason
		}
	}
	want := map[string]string{".env": "PASS/OK", "apps/api/.env": "FAIL/DIRECT_ANTHROPIC", "apps/web/.env.local": "FAIL/FILE_INVALID"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q want %q", k, got[k], v)
		}
	}
	if app.ExitCode(fs, nil) != 1 {
		t.Error("exit != 1")
	}

	// Suchfehler: Finding, kein Tool-Fehler
	fs, err = app.Scan{EnvSources: r, EnvFinder: fakeenvsource.Finder{Err: errors.New("boom")}}.Run(model.DefaultConfig())
	if err != nil {
		t.Fatalf("Suchfehler ist kein Tool-Fehler: %v", err)
	}
	if len(fs) != 1 || fs[0].Severity != model.FAIL || fs[0].Reason != model.ReasonFileUnreadable || fs[0].RuleID != "LH-FA-LENV-007" {
		t.Errorf("Suchfehler: %+v", fs)
	}

	// ohne EnvFinder (oder ohne EnvSources): keine .env-Prüfung
	for name, s := range map[string]app.Scan{
		"ohne Finder": {EnvSources: r},
		"ohne Reader": {EnvFinder: finder},
		"beides leer": {},
	} {
		fs, _ := s.Run(model.DefaultConfig())
		for _, f := range fs {
			if f.RuleID == "LH-FA-LENV-007" {
				t.Errorf("%s: unerwartetes Finding %+v", name, f)
			}
		}
	}
}
