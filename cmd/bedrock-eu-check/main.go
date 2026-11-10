// Command bedrock-eu-check prüft lokale Umgebung auf EU-Konformität.
package main

import (
	"os"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driving/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
