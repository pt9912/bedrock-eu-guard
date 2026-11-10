package app

import (
	"errors"
	"sort"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// Scan ist der Use-Case der Datei-Prüfungen (Claude-Settings, Devcontainer,
// Compose, Dockerfile, Terraform). Alle Reader sind optional: nil heißt, die
// zugehörigen Regeln laufen nicht (kein Finding, kein „Datei fehlt“-WARN). Dateifehler
// werden zu Findings (nie zu Exit 2, SPEC-030 bleibt unberührt).
type Scan struct {
	Settings driven.ClaudeSettingsReader
	// EnvSources und Files sind optional (nil = keine Devcontainer-/Compose-/
	// Dockerfile-Prüfung). Files dient der Dateisuche per Glob.
	EnvSources driven.EnvSourceReader
	Files      driven.Filesystem
	// Terraform ist optional (nil = keine Terraform-Prüfung).
	Terraform driven.TerraformReader
}

// Run lädt die Settings einmal und wertet alle Scan-Regeln aus; Findings
// stabil nach RuleID (SPEC-001), innerhalb einer Regel in Ladereihenfolge.
func (u Scan) Run(cfg model.Config) ([]model.Finding, error) {
	var sc rules.ScanContext
	if u.Settings != nil {
		sc.User = load(rules.UserSettingsPath, u.Settings.ReadUser)
		for _, rel := range rules.ProjectSettingsFiles {
			rel := rel
			sc.Projects = append(sc.Projects, load(rel, func() (*model.ClaudeSettings, error) {
				return u.Settings.ReadProject(rel)
			}))
		}
	} else {
		sc.SkipClaude = true
	}
	if u.EnvSources != nil && u.Files != nil {
		sc.Devcontainers = u.loadEnv(model.EnvDevcontainer, rules.DevcontainerPatterns)
		sc.Composes = u.loadEnv(model.EnvCompose, rules.ComposePatterns)
		sc.Dockerfiles = u.loadEnv(model.EnvDockerfile, rules.DockerfilePatterns)
	}
	if u.Terraform != nil {
		sc.Terraform = u.loadTerraform()
	}
	var out []model.Finding
	for _, r := range rules.ScanRules() {
		out = append(out, r(sc, cfg)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out, nil
}

func load(path string, read func() (*model.ClaudeSettings, error)) rules.SettingsFile {
	s, err := read()
	f := rules.SettingsFile{Path: path, S: s}
	if err != nil {
		var fe *model.FileError
		if !errors.As(err, &fe) {
			fe = &model.FileError{Path: path, Reason: model.ReasonFileUnreadable}
		}
		f.S, f.Err = nil, fe
	}
	return f
}

// loadEnv sucht Dateien zu den Mustern (sortiert, ohne Duplikate) und lädt sie.
// Ein Fehler der Dateisuche selbst wird als nicht lesbare Quelle gemeldet.
func (u Scan) loadEnv(kind model.EnvSourceKind, patterns []string) []rules.EnvSourceFile {
	var out []rules.EnvSourceFile
	seen := map[string]bool{}
	var rels []string
	for _, pat := range patterns {
		found, err := u.Files.Glob(pat)
		if err != nil {
			out = append(out, rules.EnvSourceFile{Path: pat,
				Err: &model.FileError{Path: pat, Reason: model.ReasonFileUnreadable}})
			continue
		}
		for _, rel := range found {
			if !seen[rel] {
				seen[rel] = true
				rels = append(rels, rel)
			}
		}
	}
	sort.Strings(rels)
	for _, rel := range rels {
		f, err := u.EnvSources.ReadEnvSource(kind, rel)
		sf := rules.EnvSourceFile{Path: rel, F: f}
		if err != nil {
			var fe *model.FileError
			if !errors.As(err, &fe) {
				fe = &model.FileError{Path: rel, Reason: model.ReasonFileUnreadable}
			}
			sf.F, sf.Err = nil, fe
		}
		out = append(out, sf)
	}
	return out
}

func (u Scan) loadTerraform() rules.TerraformScan {
	t := rules.TerraformScan{Searched: true}
	rels, err := u.Terraform.Find()
	if err != nil {
		var fe *model.FileError
		if !errors.As(err, &fe) {
			fe = &model.FileError{Path: ".", Reason: model.ReasonFileUnreadable}
		}
		t.FindErr = fe
		return t
	}
	sort.Strings(rels)
	for _, rel := range rels {
		f, err := u.Terraform.Read(rel)
		st := rules.TerraformFileState{Path: rel, F: f}
		if err != nil {
			var fe *model.FileError
			if !errors.As(err, &fe) {
				fe = &model.FileError{Path: rel, Reason: model.ReasonFileUnreadable}
			}
			st.F, st.Err = nil, fe
		}
		t.Files = append(t.Files, st)
	}
	return t
}
