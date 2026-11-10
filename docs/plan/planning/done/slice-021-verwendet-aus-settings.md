# Slice slice-021: „Verwendet“ in den Bedrock-Regeln auch aus Claude-Settings

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** Qualitätsrunde nach welle-04-output-ci (Folgepunkt aus slice-009).

**Bezug:** [`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen), [`LH-FA-BEDROCK-003`](../../../../spec/lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen), [`LH-FA-BEDROCK-004`](../../../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen), [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Als „verwendet“ gelten für die Bedrock-Regeln (`aws`-Modus) nicht nur Modell-IDs
aus Umgebungsvariablen, sondern auch die in Claude-Settings konfigurierten:
der Schlüssel `model` und die `env`-Einträge `ANTHROPIC_MODEL`,
`ANTHROPIC_DEFAULT_{SONNET,HAIKU,OPUS}_MODEL`, `ANTHROPIC_SMALL_FAST_MODEL`
(Benutzer- und Projekt-Settings).

## 2. Verträge

- `app.CheckAWS` bekommt das optionale Feld `Settings` (`driven.ClaudeSettingsReader`, derselbe Port wie in `app.Scan`). `nil` = Verhalten wie bisher (nur Umgebung). Fehlende oder kaputte Settings-Dateien werden im `aws`-Modus nicht gemeldet (das tut `local`); ausgeschlossene Pfade (`exclude`, [`slice-019`](../done/slice-019-scan-exclude.md)) zählen nicht.
- `rules.SettingsModel{Source, Var, ID}` (Quelle = Dateipfad, `Var` = `model` bzw. Variablenname); `rules.AWSScan.Settings`. Secret-Einträge und leere Werte werden übergangen.
- Zusammenführung: Umgebung zuerst, dann Settings; gleiche Variable mit gleicher ID nur einmal, abweichende IDs zusätzlich geprüft (beide sind potenziell wirksam).
- Findings aus Settings tragen `File` = Quelle und nennen die Herkunft in der Meldung („… (in .claude/settings.json)“); die Zeile ist unbekannt (0).
- Keine Änderung an den Regel-IDs, Reason-Codes oder am Adapter-Vertrag.

**Adapter (w2):** Der Composition Root übergibt dem Modus `aws` bzw. `all` zusätzlich den vorhandenen `ClaudeSettingsReader` (`CheckAWS{…, Settings: reader}`).

## 3. Verhalten

| Regel | Wirkung |
|---|---|
| [`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen) | blockierte ID aus Settings → `FAIL` (wie aus der Umgebung) |
| [`LH-FA-BEDROCK-003`](../../../../spec/lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen) | ID aus Settings nicht durch Allowlist abgedeckt → `FAIL` |
| [`LH-FA-BEDROCK-004`](../../../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen) | Verfügbarkeit in der Primärregion wird auch für Settings-IDs geprüft |

## 4. Definition of Done

- [ ] Kern (w1): `CheckAWS.Settings`, `settingsModels`, Zusammenführung in `usedModels`, Herkunft in `File`/Meldung, Tests (alle drei Regeln, Dedup, Exclude, Secret-Einträge, Reader ohne/mit Fehler).
- [ ] CLI (w2): `aws`/`all` übergeben den Settings-Reader; Akzeptanztest über die CLI mit Settings-Modell und Fake-`aws`.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/rules/{aws,bedrock}.go`, `core/app/checkaws.go` | update | w1 |
| `internal/adapter/driving/cli/` | update | w2 |

## 6. Risiken und offene Punkte

- Settings-Datei und Umgebung können dieselbe Variable unterschiedlich setzen; die tatsächliche Rangfolge in Claude Code wird nicht nachgebildet, beide IDs werden geprüft.
- Settings der Projektebene liegen relativ zum Scan-Root; im `aws`-Modus ohne `--root` ist das das aktuelle Verzeichnis.
- Modelle in anderen Konfigurationsquellen (Devcontainer-, Compose-Env, `.env`) zählen weiterhin nicht als „verwendet“ (Folgepunkt, abhängig von der Klärung zu `.env`).

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->
