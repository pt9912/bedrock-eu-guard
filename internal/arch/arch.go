// Package arch erzwingt die Schichtenregeln aus spec/architecture.md §2 als
// ausführbare Fitness Function. Die Regeln sind reine Funktionen über
// (Paketverzeichnis, Importpfad); der Test in arch_test.go wendet sie auf
// alle Produktionsdateien des Repositorys an.
package arch

import (
	"fmt"
	"strings"
)

// Modul ist der Modulpfad (go.mod).
const Modul = "github.com/pt9912/bedrock-eu-guard"

// Schichten als Verzeichnispräfix relativ zum Modulwurzelverzeichnis.
const (
	Model    = "internal/hexagon/core/model"
	Rules    = "internal/hexagon/core/rules"
	App      = "internal/hexagon/core/app"
	Port     = "internal/hexagon/port/driven"
	CoreTest = "internal/hexagon/coretest"
	Driven   = "internal/adapter/driven"
	Driving  = "internal/adapter/driving"
	Cmd      = "cmd"

	// AWSDoor ist die einzige Schicht, die den Prozess `aws` starten darf (ARC-007).
	AWSDoor = Driven + "/awscli"
)

// Violation beschreibt einen Verstoß gegen eine Architekturregel.
type Violation struct {
	Dir    string // Paketverzeichnis relativ zum Modulwurzelverzeichnis
	Import string
	Rule   string // z. B. "ARC-003"
	Reason string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s importiert %q: %s (%s)", v.Dir, v.Import, v.Reason, v.Rule)
}

func under(dir, prefix string) bool { return dir == prefix || strings.HasPrefix(dir, prefix+"/") }

// internalDir liefert das modulrelative Verzeichnis eines Importpfads, falls
// er zum Modul gehört.
func internalDir(imp string) (string, bool) {
	if imp == Modul {
		return "", true
	}
	if strings.HasPrefix(imp, Modul+"/") {
		return strings.TrimPrefix(imp, Modul+"/"), true
	}
	return "", false
}

// isStdlib: Importpfade der Standardbibliothek haben keinen Punkt im ersten Element.
func isStdlib(imp string) bool {
	first := imp
	if i := strings.Index(imp, "/"); i >= 0 {
		first = imp[:i]
	}
	return !strings.Contains(first, ".")
}

// Seiteneffekt-Pakete, die der reine Kern nicht importieren darf.
var impureStdlib = map[string]bool{
	"os": true, "os/exec": true, "os/signal": true, "net": true, "net/http": true,
	"io/ioutil": true, "syscall": true, "database/sql": true,
}

func inCore(dir string) bool {
	return under(dir, Model) || under(dir, Rules) || under(dir, App)
}

// Check bewertet einen Import einer Produktionsdatei in Verzeichnis dir.
func Check(dir, imp string) []Violation {
	var out []Violation
	add := func(rule, reason string) { out = append(out, Violation{dir, imp, rule, reason}) }

	// ARC-007: nur awscli startet Prozesse (os/exec) bzw. nutzt ein AWS-SDK.
	if imp == "os/exec" && !under(dir, AWSDoor) {
		// Weitere Prozessstarts außerhalb von awscli sind nicht vorgesehen.
		add("ARC-007", "os/exec nur in adapter/driven/awscli erlaubt")
	}
	if strings.HasPrefix(imp, "github.com/aws/") && !under(dir, AWSDoor) {
		add("ARC-007", "AWS-SDK nur in adapter/driven/awscli erlaubt")
	}

	// Reiner Kern: nur Stdlib ohne Seiteneffekt-Pakete, keine Drittbibliotheken.
	if inCore(dir) || under(dir, Port) || under(dir, CoreTest) {
		if !isStdlib(imp) {
			if _, ok := internalDir(imp); !ok {
				add("ARC-003", "Kern, Ports und Test-Doubles importieren keine Drittbibliotheken")
			}
		}
		if impureStdlib[imp] {
			add("ARC-003", "Kern, Ports und Test-Doubles bleiben seiteneffektfrei (kein "+imp+")")
		}
	}

	d, internal := internalDir(imp)
	if !internal {
		return out
	}

	switch {
	case under(dir, Model):
		add("ARC-003", "core/model ist der importfreie Ring (nur Stdlib)")
	case under(dir, Port):
		if !under(d, Model) {
			add("ARC-005", "port/driven darf nur model importieren")
		}
	case under(dir, Rules):
		if !(under(d, Model) || under(d, Port)) {
			add("ARC-004", "core/rules darf nur model und port/driven importieren")
		}
	case under(dir, App):
		if !(under(d, Model) || under(d, Rules) || under(d, Port)) {
			add("ARC-002", "core/app darf nur model, rules und port/driven importieren")
		}
	case under(dir, CoreTest):
		if !(under(d, Model) || under(d, Port)) {
			add("ARC-003", "coretest darf nur model und port/driven importieren")
		}
	case under(dir, Driven):
		if under(d, Rules) || under(d, App) || under(d, Driving) || under(d, CoreTest) {
			add("ARC-006", "driven-Adapter importieren weder rules noch app, driving-Adapter noch Test-Doubles")
		}
	case under(dir, Driving):
		if under(d, CoreTest) {
			add("ARC-001", "Produktionscode importiert keine Test-Doubles")
		}
	case under(dir, Cmd):
		if under(d, CoreTest) {
			add("ARC-001", "Produktionscode importiert keine Test-Doubles")
		}
	}
	return out
}
