# Slice slice-019: Scan-Root-Ausschluss (`exclude`)

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-05-release (Vorlauf für slice-015).

**Bezug:** [`LH-NF-007`](../../../../spec/lastenheft.md#lh-nf-007--konfigurierbarkeit), [`LH-TECH-003`](../../../../spec/lastenheft.md#lh-tech-003--konfigurationsdatei), [`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung), [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Pfade lassen sich über `exclude` in `bedrock-eu-check.yml` von der
Datei-Discovery ausnehmen. Das ist Voraussetzung für den Self-Check
(slice-015), weil das Repository absichtlich schlechte Testdaten enthält.

## 2. Verträge

- `model.Config.Exclude []string`, `model.PartialConfig.Exclude []string` (`nil` = nicht angegeben, Default: nichts ausgeschlossen). `Resolve` validiert die Muster: leer oder nur `/` ist `INVALID_VALUE`, ein syntaktisch ungültiges Muster `INVALID_PATTERN` (jeweils mit Index, ohne Muster im Fehlertext).
- `model.Config.Excluded(rel) bool` ist die **einzige** Semantik: Muster nach `path.Match`, relativ zum Scan-Root, `/`-getrennt (Backslashes werden normalisiert); ein Muster trifft den Pfad selbst oder ein Verzeichnis darüber; kein `**`.
- Anwendung im **Kern** (`app.Scan`), nicht in den Adaptern: Treffer von `Files.Glob` (Devcontainer, Compose, Dockerfile), `TerraformReader.Find()` und die Projekt-Claude-Settings werden vor dem Lesen gefiltert. Die Benutzer-Settings sind nicht betroffen.
- Adapter (w2): `configyaml` mappt `exclude` (Liste) auf `PartialConfig.Exclude`; fehlt das Feld, bleibt es `nil`. Die Adapter dürfen weiter durch ausgeschlossene Verzeichnisse laufen (Kosten, kein Fehler); frühes Überspringen in `terraform.Find()` ist eine spätere Optimierung.

## 3. Verhalten

| Fall | Ergebnis |
|---|---|
| Datei passt auf ein Exclude-Muster | wird nicht gelesen, kein Finding (auch kein „Datei fehlt“) |
| alle Terraform-Dateien ausgeschlossen | wie „keine Terraform-Dateien gefunden“: `WARN` `NO_TERRAFORM_FILES` |
| ungültiges Muster in der Konfiguration | Tool-Fehler Exit `2` (eigene Config ungültig, [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes)) |

## 4. Definition of Done

- [ ] Kern (w1): `Config.Exclude`, `PartialConfig.Exclude`, `Resolve`-Validierung, `Excluded`, Filter in `app.Scan`, Unit-Tests; [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) fortgeschrieben.
- [ ] Adapter (w2): `configyaml` mappt `exclude` (inkl. strikter Prüfung des Typs), Test mit Konfigurationsdatei; CLI-Test, dass ausgeschlossene Pfade nicht gemeldet werden.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/{model,config}.go`, `core/app/scan.go`, `spec/spezifikation.md` | update | w1 |
| `internal/adapter/driven/configyaml/`, `internal/adapter/driving/cli/` (Test) | update | w2 |

## 6. Trigger

[`welle-05-release`](../welle-05-release.md) gestartet (Plan gemerged).

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- `path.Match` kennt kein `**`; tiefe Muster wie `**/testdata` müssen als `testdata` (Verzeichnis überall unter dem Scan-Root ist **nicht** gemeint, nur relativ zum Root) oder mit expliziten Ebenen (`*/testdata`) geschrieben werden. Eine Erweiterung um `**` wäre eine Spec-Änderung.
- `Files.Glob` ist nicht rekursiv; Dateien in Unterverzeichnissen (z. B. `services/api/Dockerfile`) werden von der Env-Discovery ohnehin nicht gefunden. Das ist ein bekannter Unterschied zur rekursiven Terraform-Suche.
- Ein zu großzügiges Muster kann echte Befunde verbergen; die Ausgabe weist ausgeschlossene Pfade nicht aus (Folgepunkt: Hinweis auf stderr).

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung der Konfiguration auf dem Kern
der Wellen 1 bis 3.
