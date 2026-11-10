// Package cli ist der Composition Root (ARC-001): Arg-Parsing, Verdrahtung
// Adapter → Kern und Exit-Code-Mapping (SPEC-030).
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/claudesettings"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/envos"
	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/report"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/app"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

const exitToolError = 2

const usage = "usage: bedrock-eu-check local\n       bedrock-eu-check scan [--root DIR]\n"

// Run führt die CLI aus und liefert den Prozess-Exit-Code.
func Run(args []string, stdout, stderr io.Writer) int {
	home, _ := os.UserHomeDir() // leer → User-Settings gelten als nicht vorhanden
	return run(args, envos.New(), home, stdout, stderr)
}

func run(args []string, env driven.Environment, home string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return exitToolError
	}
	switch args[0] {
	case "local":
		if len(args) != 1 {
			break
		}
		findings, err := app.CheckLocal{}.Run(env, model.DefaultConfig())
		return finish(findings, err, stdout, stderr)
	case "scan":
		root, ok := scanRoot(args[1:])
		if !ok {
			break
		}
		scan := app.Scan{Settings: claudesettings.New(root, home)}
		findings, err := scan.Run(model.DefaultConfig())
		return finish(findings, err, stdout, stderr)
	}
	fmt.Fprint(stderr, usage)
	return exitToolError
}

// scanRoot wertet "[--root DIR]" aus; Default ist das aktuelle Verzeichnis.
func scanRoot(args []string) (string, bool) {
	switch {
	case len(args) == 0:
		return ".", true
	case len(args) == 2 && args[0] == "--root" && args[1] != "":
		return args[1], true
	}
	return "", false
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
