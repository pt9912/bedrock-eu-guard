// Package cli ist der Composition Root (ARC-001): Arg-Parsing, Verdrahtung
// Adapter → Kern und Exit-Code-Mapping (SPEC-030).
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/awscli"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/claudesettings"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/configyaml"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/envos"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/envsource"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/fs"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/report"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/terraform"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

const exitToolError = 2

const usage = "usage: bedrock-eu-check version | (local|terraform|aws|all) [--root DIR] [--config FILE] [--allowed-account ID]... [--format console|json|sarif] [--ci]\n"

// Build-Informationen, zur Buildzeit per -ldflags gesetzt (siehe Makefile):
//
//	-X github.com/pt9912/bedrock-eu-guard/internal/adapter/driving/cli.version=1.0.0
//	-X github.com/pt9912/bedrock-eu-guard/internal/adapter/driving/cli.commit=abc123
//
// version erscheint im Befehl "version" und als Tool-Version in SARIF.
var (
	version = "dev"
	commit  = "unknown"
)

// awsBinary ist das aws-Binary (Name oder Pfad); Variable nur für Tests.
var awsBinary = awscli.DefaultBinary

// Run führt die CLI aus und liefert den Prozess-Exit-Code.
func Run(args []string, stdout, stderr io.Writer) int {
	home, _ := os.UserHomeDir() // leer → User-Settings gelten als nicht vorhanden
	return run(args, envos.New(), home, stdout, stderr)
}

func run(args []string, env driven.Environment, home string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "version" {
		if len(args) != 1 {
			fmt.Fprint(stderr, usage)
			return exitToolError
		}
		fmt.Fprintf(stdout, "bedrock-eu-check %s (%s)\n", version, commit)
		return 0
	}
	if len(args) == 0 || (args[0] != "local" && args[0] != "terraform" && args[0] != "aws" && args[0] != "all") {
		fmt.Fprint(stderr, usage)
		return exitToolError
	}
	mode := args[0]
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
	cfg, err := app.LoadConfig(accountOverlay{configyaml.New(opts.configPath()), opts.accounts})
	if err != nil {
		return finish(nil, err, reporterFor(opts, stdout), opts.ci, stdout, stderr)
	}
	var findings []model.Finding
	if mode == "local" || mode == "all" {
		localFindings, lerr := runLocal(env, cfg, opts, home)
		findings, err = append(findings, localFindings...), lerr
	}
	if err == nil && (mode == "terraform" || mode == "all") {
		var tf []model.Finding
		tf, err = app.Scan{Terraform: terraform.New(opts.root)}.Run(cfg)
		findings = append(findings, tf...)
	}
	if err == nil && (mode == "aws" || mode == "all") {
		var aws []model.Finding
		reader := awscli.New(awsBinary, cfg.AllowedRegions[0])
		aws, err = app.CheckAWS{AWS: reader, Bedrock: reader, IAM: reader}.Run(env, cfg)
		findings = append(findings, aws...)
	}
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].RuleID < findings[j].RuleID })
	return finish(findings, err, reporterFor(opts, stdout), opts.ci, stdout, stderr)
}

// runLocal ist der Modus local (LH-MOD-001): Env, Claude-Settings,
// Devcontainer, Compose und Dockerfile — offline. Terraform gehört zum Modus
// terraform (LH-MOD-003).
func runLocal(env driven.Environment, cfg model.Config, opts options, home string) ([]model.Finding, error) {
	findings, err := app.CheckLocal{}.Run(env, cfg)
	if err != nil {
		return nil, err
	}
	files, err := app.Scan{
		Settings:   claudesettings.New(opts.root, home),
		EnvSources: envsource.New(opts.root),
		Files:      fs.New(opts.root),
	}.Run(cfg)
	return append(findings, files...), err
}

type options struct {
	root, config string
	format       string // "" = console
	ci           bool   // CI-Modus: Zusammenfassung auf stderr (LH-MOD-005)
	accounts     []string
}

// reporterFor wählt den Reporter für das Ausgabeformat (stdout trägt nur das
// Ergebnis). Das Format wurde bereits in parseOptions geprüft.
func reporterFor(opts options, w io.Writer) driven.Reporter {
	switch opts.format {
	case "json":
		return report.NewJSON(w)
	case "sarif":
		return report.NewSARIF(w, version, opts.sarifFallback())
	default:
		return report.New(w)
	}
}

// sarifFallback ist die Datei, die in SARIF als Ersatzposition für Findings
// ohne Fundstelle dient: die Konfigurationsdatei, falls vorhanden, sonst README.md.
func (o options) sarifFallback() string {
	if _, err := os.Stat(filepath.Join(o.root, "bedrock-eu-check.yml")); err == nil {
		return "bedrock-eu-check.yml"
	}
	return "README.md"
}

// accountOverlay ergänzt die Konfiguration um --allowed-account (Vereinigung
// mit allowedAccounts aus der Datei, Format wird in model.Resolve geprüft).
type accountOverlay struct {
	driven.ConfigLoader
	accounts []string
}

func (o accountOverlay) Load() (*model.PartialConfig, error) {
	p, err := o.ConfigLoader.Load()
	if err != nil || len(o.accounts) == 0 {
		return p, err
	}
	if p == nil {
		p = &model.PartialConfig{}
	}
	p.AllowedAccounts = append(append([]string{}, p.AllowedAccounts...), o.accounts...)
	return p, nil
}

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
		if args[0] == "--ci" { // Schalter ohne Wert
			opts.ci = true
			args = args[1:]
			continue
		}
		if len(args) < 2 || args[1] == "" {
			return options{}, false
		}
		switch args[0] {
		case "--root":
			opts.root = args[1]
		case "--config":
			opts.config = args[1]
		case "--format":
			if args[1] != "console" && args[1] != "json" && args[1] != "sarif" {
				return options{}, false
			}
			opts.format = args[1]
		case "--allowed-account":
			opts.accounts = append(opts.accounts, args[1])
		default:
			return options{}, false
		}
		args = args[2:]
	}
	return opts, true
}

// finish gibt das Ergebnis über den Reporter aus und liefert den Exit-Code
// (SPEC-030). Im CI-Modus steht zusätzlich eine Zusammenfassung auf stderr,
// damit die Pipeline-Logs auch bei umgeleitetem stdout (z. B. SARIF-Datei)
// den Ausgang zeigen; stdout trägt weiterhin nur das Ergebnis.
func finish(findings []model.Finding, err error, rep driven.Reporter, ci bool, stdout, stderr io.Writer) int {
	code := exitToolError
	if err != nil {
		if len(findings) > 0 { // bereits ermittelte Befunde trotzdem zeigen
			_ = rep.Report(findings)
		}
		fmt.Fprintf(stderr, "error: %v\n", err)
		code = app.ExitCode(nil, err)
	} else if rerr := rep.Report(findings); rerr != nil {
		fmt.Fprintf(stderr, "error: %v\n", rerr)
	} else {
		code = app.ExitCode(findings, nil)
	}
	if ci {
		fmt.Fprintln(stderr, summaryLine(findings, code))
	}
	return code
}

func summaryLine(findings []model.Finding, code int) string {
	var pass, warn, fail int
	for _, f := range findings {
		switch f.Severity {
		case model.FAIL:
			fail++
		case model.WARN:
			warn++
		default:
			pass++
		}
	}
	return fmt.Sprintf("bedrock-eu-check: %d PASS, %d WARN, %d FAIL (Exit %d)", pass, warn, fail, code)
}
