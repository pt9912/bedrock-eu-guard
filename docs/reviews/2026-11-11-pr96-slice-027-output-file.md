# Review-Report: PR #96 (slice-027, `--output FILE`) — 2026-11-11

**Review-Art:** Code — geprüft wird der CLI-Teil von slice-027 (Ausgabe in eine Datei) gegen den Slice-Plan, die Spezifikation und die Konventionen. Zusätzlich wurde das Verhalten am gebauten Binary stichprobenartig beobachtet.

**Gegenstand:** Pull Request #96 (`feat/slice-027-output-file`, Autor w2), gemerged in `main`; geprüfter Stand: `main` bei `5a8b92b` (enthält #96) mit `bin/bedrock-eu-check`, gebaut per `make build`.

**Skill:** keiner. Das Repository enthält keine Reviewer-Skill-Datei (die Vorlage verweist auf `.harness/skills/reviewer.md`, die es hier nicht gibt); geprüft wurde nach den unten genannten Verträgen.
**Modell:** claude-sonnet-5-5 (Agent w1) · **Datum:** 2026-11-11

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- Slice-Plan `slice-027-output-file` (Ziel, Definition of Done, Risiken); er liegt zum Zeitpunkt dieses Reports noch in `docs/plan/planning/open/` und wandert bei der Closure nach `done/`, deshalb ohne Link
- [`SPEC-030`](../../spec/spezifikation.md#spec-030--exit-codes), Abschnitt „Ausgabeziel `--output FILE`“ (Vertrag für CI-Konsumenten, in PR #97 ergänzt)
- [`LH-MOD-005`](../../spec/lastenheft.md#lh-mod-005--ci-modus) (CI-Modus), [`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe) (keine Secret-Ausgabe)
- [`AGENTS.md`](../../AGENTS.md) (Hard Rules)

---

## Findings

### F-1 — Temporäre Datei liegt im Zielverzeichnis

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `internal/adapter/driving/cli/cli.go` (`writeFileAtomic`)
- `befund`: Die temporäre Datei `.bedrock-eu-check-*` wird im Zielverzeichnis angelegt und nach dem Umbenennen bzw. bei einem Fehler entfernt; bricht der Prozess dazwischen hart ab (zum Beispiel `kill -9`), bleibt sie dort zurück. Liegt das Ziel im Scan-Root, liegt sie dort.
- `verifizierbar`: nein — nur durch einen Abbruch-Test des Prozesses; die Gates prüfen den Normalfall und den Fehlerfall (Test „kein Temp-Rest bei Rename-Fehler“).

### F-2 — Datei mit Teilergebnis bei Exit 2

- `kategorie`: INFO
- `quelle`: SPEC-030 (Abschnitt „Ausgabeziel“)
- `pfad`: `internal/adapter/driving/cli/cli.go` (`finish`)
- `befund`: Bei einem späten Tool-Fehler (Beobachtung: `all` ohne `aws`-Binary) wird die Ausgabedatei mit dem Teilergebnis angelegt (5 Befunde, gültiges JSON) und der Lauf endet mit Exit 2; ein Konsument, der nur die Datei und nicht den Exit-Code auswertet, hält das Teilergebnis für vollständig.
- `verifizierbar`: ja — am Binary beobachtet; das Verhalten ist in SPEC-030 festgelegt und im Test abgedeckt.

### F-3 — Feste Fehlermeldung bei Schreibfehlern ohne Pfad und Ursache

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `internal/adapter/driving/cli/cli.go` (`finish`, Meldung „Ausgabedatei konnte nicht geschrieben werden (--output)“)
- `befund`: Die Meldung nennt weder den Pfad noch den Grund (Verzeichnis fehlt, Ziel ist ein Verzeichnis, keine Rechte); beobachtet in zwei verschiedenen Fehlerfällen mit identischem Text.
- `verifizierbar`: ja — am Binary beobachtet (Ziel ist ein Verzeichnis; Zielverzeichnis fehlt).

## Negativbefunde

- geprüft, ohne Befund: Standardverhalten ohne `--output` — Ausgabe auf stdout unverändert (JSON parsebar, gleiche Struktur wie vorher).
- geprüft, ohne Befund: Atomares Schreiben — Datei entsteht vollständig oder gar nicht; Rechte 644; nach den Fehlerfällen (Ziel ist Verzeichnis, Zielverzeichnis fehlt, Konfigurationsfehler) blieb kein Temp-Rest im beobachteten Verzeichnis.
- geprüft, ohne Befund: Tool-Fehler vor dem Ergebnis — bei ungültiger Konfiguration (Exit 2) bleibt eine vorhandene Zieldatei unverändert (beobachtet mit einer Sentinel-Datei).
- geprüft, ohne Befund: stdout bleibt mit `--output` leer; mit `--ci` erscheint die Zusammenfassung auf stderr; Exit-Codes unverändert.
- geprüft, ohne Befund: Symlink als Ziel — der Link selbst wird durch eine reguläre Datei ersetzt, das Ziel des Links bleibt unverändert (beobachtet; in der Referenz dokumentiert).
- geprüft, ohne Befund: Usage — leerer Wert für `--output` ergibt Exit 2 mit `usage`.
- geprüft, ohne Befund: Secrets — keine Dateiinhalte oder Werte in den Fehlermeldungen (feste Meldung).
- nicht geprüft: Verhalten unter Windows (Umbenennen über eine vorhandene Datei) und auf Dateisystemen ohne atomares Umbenennen (zum Beispiel manche Netzlaufwerke); gleichzeitige Läufe mit demselben Ziel; sehr große Ausgaben (der Reporter schreibt zuerst in einen Speicherpuffer).

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |
| INFO | 3 |

## Verdikt

**Merge-blockierend:** nein — kein HIGH, kein MEDIUM. Der PR wurde nach grüner CI und dem Gegenlesen durch w1 gemerged; dieser Report entstand danach.

**Übergabe:** Die drei INFO-Findings sind Beobachtungen ohne Handlungsbedarf für 1.1.0; über eine ausführlichere Fehlermeldung (F-3) kann w2 entscheiden. Dieser Report ersetzt keine Verifikation — Definition of Done und Spezifikations-Konformität prüft der Verifier separat.

**Hinweis zur Entstehung:** Der Report wurde nachträglich geschrieben. Das Gegenlesen des Diffs geschah vor dem Merge im Chat zwischen w1 und w2; die Beobachtungen am Binary (Abschnitt Negativbefunde und F-2/F-3) wurden danach am gebauten Stand von `main` gemacht, nicht am Diff allein. Der Report stammt von einem Agenten desselben Modells wie der Autor; eine unabhängige menschliche Prüfung ersetzt er nicht.
