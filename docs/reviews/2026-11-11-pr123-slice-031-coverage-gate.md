# Review-Report: PR #123 (slice-031, `make coverage-gate`) — 2026-11-11

**Review-Art:** Code — geprüft wird die Umsetzung von slice-031 (Skript, Makefile, ADR-0005, Doku) gegen den Plan, die Harten Regeln und die ADR. Zusätzlich wurde das Verhalten des Skripts im gepinnten Image beobachtet.

**Gegenstand:** Pull Request #123 (`feat/slice-031-coverage-gate`, Autor w1), geprüfter Stand: Branch-Kopf `a81205c` auf `main` bei `908866c`.

**Skill:** [`.harness/skills/reviewer.md`](../../.harness/skills/reviewer.md) @ `e03b883` (PR #118)
**Modell:** claude-sonnet-5-5 (Agent w2) · **Datum:** 2026-11-11

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- der Plan [`slice-031-coverage-gate.md`](../plan/planning/done/slice-031-coverage-gate.md) (Definition of Done, Entscheidungen)
- [`ADR-0005`](../plan/adr/0005-coverage-gate.md) (Proposed) und [`ADR-0004`](../plan/adr/0004-abhaengigkeiten-und-offline-gates.md)
- [`AGENTS.md`](../../AGENTS.md) §3.5, §3.6 und §4
- [`LH-QA-001`](../../spec/lastenheft.md#lh-qa-001--unit-tests)
- `Makefile` (Targets `test`, `gates`, `GO_RUN`), `tools/coverage-gate.sh`

---

## Findings

### F-1 — Mindestwert wird nicht nach unten begrenzt oder als Zahl geprüft

- `kategorie`: MEDIUM
- `quelle`: [`AGENTS.md`](../../AGENTS.md) §3.6
- `pfad`: `tools/coverage-gate.sh` · `''|*[!0-9.]*) echo "coverage-gate: ungueltiger Mindestwert`
- `befund`: Das Skript lehnt nur Zeichen außerhalb von Ziffern und Punkt ab. Die Werte `0`, `.` und `9.9.9` werden angenommen, und das Gate meldet bei 95,4 % „OK“ (im Image nachgestellt, Exit 0); `COVERAGE_MIN` ist im Makefile mit `?=` gesetzt und lässt sich über die Umgebung überschreiben. Ein Senken auf `1` im Makefile oder per Umgebung besteht das Gate, obwohl ADR-0005 für jedes Senken eine neue ADR verlangt; erkannt würde es nur beim menschlichen Lesen des Diffs.
- `verifizierbar`: ja — `sh tools/coverage-gate.sh 0` im gepinnten Image.
- `klasse`: Schwellenwert-Eingabe unbegrenzt und nicht als Zahl validiert

### F-2 — Fehlschlagende Tests erzeugen im Gate keine Meldung

- `kategorie`: LOW
- `quelle`: Maintainability
- `pfad`: `tools/coverage-gate.sh` · `go test -coverpkg="$pkgs" -coverprofile="$profile" ./... >/dev/null`
- `befund`: Die Ausgabe von `go test` geht nach `/dev/null`. Bei einem fehlschlagenden Test endet `sh tools/coverage-gate.sh 90` mit Exit 1 und ohne jede Zeile Ausgabe (mit einem absichtlich fehlschlagenden Test im Image nachgestellt); ein Aufruf von `make coverage-gate` allein zeigt keinen Grund. In `make gates` läuft `test` vorher und meldet den Fehler.
- `verifizierbar`: ja — ein temporärer fehlschlagender Test und derselbe Aufruf.
- `klasse`: Gate bricht ohne Meldung ab

### F-3 — Messwerte und Negativproben des PR reproduziert

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `docs/plan/adr/0005-coverage-gate.md` · `paketübergreifend gemessen sind es 95,1 %`
- `befund`: Gesamt 95,4 % und `model` 95,1 % stimmen mit dem Lauf des Skripts überein; `COVERAGE_MIN=99` scheitert mit Exit 1, `abc` und `-1` mit Exit 2, `95.4` besteht, `95.5` scheitert; ein Lauf des Skripts dauert rund 4 s; `make gates` endet mit Exit 0.
- `verifizierbar`: ja — dieselben Aufrufe im Image.
- `klasse`: Behauptungen des PR belegt

## Negativbefunde

- geprüft, ohne Befund: Berechnung je Paket — die Summe der Anweisungen je Paket (höchster Zähler je Block) stimmt mit `go tool cover -func` überein (Gesamt 95,4 %, `model` 95,1 %).
- geprüft, ohne Befund: Paketauswahl — Ausschlüsse nur `coretest/*` und `cmd/`, wie in ADR-0005 benannt; `internal/arch` ist als Produktionspaket enthalten (95,6 %).
- geprüft, ohne Befund: Makefile — `.PHONY`, Aufnahme in `gates`, `GO_RUN` (offline, gepinnt), Profil nach `.tmp/` (ignoriert).
- geprüft, ohne Befund: ADR-0005 — Status `Proposed` (Sign-off der Auftraggeber steht aus), Entscheidung, Alternativen, Fitness Function, Re-Evaluierungs-Trigger und Geschichte vorhanden; Änderungsweg (Anheben per PR, Senken per neuer ADR) entspricht §3.5 und §3.6.
- geprüft, ohne Befund: Lebenszyklus und Roadmap — reiner Move nach `in-progress/` (`6bbcaad`), Roadmap-Block ohne Ruhe-Marker; der Probelauf des Moduls `planning` meldet 0 Befunde.
- geprüft, ohne Befund: Doku — `AGENTS.md` §4, `harness/README.md` und README tragen das Target ein, `make doc-targets` ist grün.
- nicht geprüft: Verhalten auf anderen Architekturen und in anderen Shells als dem Image; die Laufzeit des Gates in der CI; ob ein späteres neues Produktionspaket mit sehr niedriger Abdeckung den Gesamtwert verdecken würde (steht als Re-Evaluierungs-Trigger in der ADR).

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 1 |
| INFO | 1 |

## Verdikt

**Merge-blockierend:** nein — kein HIGH. F-1 (MEDIUM) sollte vor dem Merge geklärt werden, weil es die Durchsetzung von §3.6 durch das Gate selbst betrifft; F-2 ist eine Verbesserung der Diagnose.

**Übergabe:** Die Findings gehen an w1 (Autor). Der Report ersetzt weder die Verifikation noch die Prüfung der Definition of Done; die Annahme von ADR-0005 liegt bei den Auftraggebern.

**Hinweis zur Entstehung:** Das Verhalten des Skripts habe ich im gepinnten Go-Image selbst beobachtet (Messwerte, Grenzfälle des Mindestwerts, ein absichtlich fehlschlagender Test, der danach wieder entfernt wurde); die Prüfung gegen die Regeln stützt sich auf das Lesen von `AGENTS.md`, der ADR und des Plans.
