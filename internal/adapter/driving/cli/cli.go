// Package cli ist der Composition Root (ARC-001): Arg-Parsing, Verdrahtung
// Adapter → Kern und Exit-Code-Mapping (SPEC-030).
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/claudesettings"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/configyaml"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/envos"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/report"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

const exitToolError = 2

const usage = "usage: bedrock-eu-check local [--root DIR] [--config FILE]\n"

// Run führt die CLI aus und liefert den Prozess-Exit-Code.
func Run(args []string, stdout, stderr io.Writer) int {
	home, _ := os.UserHomeDir() // leer → User-Settings gelten als nicht vorhanden
	return run(args, envos.New(), home, stdout, stderr)
}

func run(args []string, env driven.Environment, home string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "local" {
		fmt.Fprint(stderr, usage)
		return exitToolError
	}
	opts, ok := parseOptions(args[1:])
	if !ok {
		fmt.Fprint(stderr, usage)
		return exitToolError
	}
	// Ein ausdrücklich angegebenes --config muss existieren; die Default-Datei
	// im Root ist optional (Datei fehlt → Defaults).
	if opts.config != "" {
		if _, statErr := os.Stat(opts.config); statErr != nil {
			fmt.Fprintln(stderr, "error: Konfigurationsdatei nicht lesbar oder nicht vorhanden (--config)")
			return exitToolError
		}
	}
	// Eigene Konfiguration ungültig → Tool-Fehler, Exit 2 (SPEC-030).
	cfg, err := app.LoadConfig(configyaml.New(opts.configPath()))
	if err != nil {
		return finish(nil, err, stdout, stderr)
	}
	// Modus local = Env-Regeln + Datei-Regeln (offline, SPEC-002).
	findings, err := app.CheckLocal{}.Run(env, cfg)
	if err == nil {
		var files []model.Finding
		files, err = app.Scan{Settings: claudesettings.New(opts.root, home)}.Run(cfg)
		findings = append(findings, files...)
	}
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].RuleID < findings[j].RuleID })
	return finish(findings, err, stdout, stderr)
}

type options struct{ root, config string }

// configPath liefert --config oder <root>/bedrock-eu-check.yml.
func (o options) configPath() string {
	if o.config != "" {
		return o.config
	}
	return filepath.Join(o.root, "bedrock-eu-check.yml")
}

// parseOptions wertet "[--root DIR] [--config FILE]" aus; Default-Root ist
// das aktuelle Verzeichnis.
func parseOptions(args []string) (options, bool) {
	opts := options{root: "."}
	for len(args) > 0 {
		if len(args) < 2 || args[1] == "" {
			return options{}, false
		}
		switch args[0] {
		case "--root":
			opts.root = args[1]
		case "--config":
			opts.config = args[1]
		default:
			return options{}, false
		}
		args = args[2:]
	}
	return opts, true
}

func finish(findings []model.Finding, err error, stdout, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return app.ExitCode(findings, err)
	}
	if rerr := report.New(stdout).Report(findings); rerr != nil {
		fmt.Fprintf(stderr, "error: %v\n", rerr)
		return exitToolError
	}
	return app.ExitCode(findings, nil)
}
