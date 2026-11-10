# Slice slice-011: JSON-Reporter und Format-Auswahl

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-04-output-ci.

**Bezug:** [`LH-FA-REP-002`](../../../../spec/lastenheft.md#lh-fa-rep-002--json-ausgabe), [`SPEC-021`](../../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json), [`SPEC-001`](../../../../spec/spezifikation.md#spec-001--severity--und-aggregationsmodell), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check <modus> --format json` gibt die Findings als JSON nach
[`SPEC-021`](../../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) aus. Die Auswahl des Formats trifft der Composition Root; der
Port `driven.Reporter` bleibt unverändert.

## 2. Verträge

- **Kern (w1):** `model.Severity.Level()` bildet `PASS|WARN|FAIL` auf `info|warning|error` ab (einzige Kernänderung der Welle, eigener kleiner PR).
- **Adapter `report.JSON(w io.Writer)`** implementiert `driven.Reporter`. Ausgabe: ein JSON-Objekt `{"version": 1, "findings": [...]}`, zwei Leerzeichen eingerückt, abschließender Zeilenumbruch.
- **Finding-Objekt:** `status` (`PASS|WARN|FAIL`), `id` (Regel-ID), `message`, `severity` (`info|warning|error`), `file` (String oder `null`), `line` (Zahl oder `null`; 0 = unbekannt), `recommendation` (String oder `null`). Zusätzlich das additive Feld `reason` (stabiler Reason-Code), damit Konsumenten die Ursache maschinell auswerten können; `subject` wird nicht ausgegeben (steckt in `message`/`file`).
- **Reihenfolge** wie die Aggregation ([`SPEC-001`](../../../../spec/spezifikation.md#spec-001--severity--und-aggregationsmodell)); keine Zeitstempel, keine zufälligen Werte (deterministisch). Ohne Findings ist `findings` ein leeres Array, nie `null`.
- **CLI:** `--format console|json` (Default `console`); ein unbekannter Wert ist ein Aufruffehler (Exit `2`, Hinweis auf stderr). stdout trägt ausschließlich das Ergebnis im gewählten Format.
- **Tool-Fehler:** Bereits ermittelte Findings werden im gewählten Format ausgegeben, die Fehlermeldung geht auf stderr, Exit `2`.

## 3. Definition of Done

- [ ] `report.JSON` mit Tests: Pflichtfelder, `null` für fehlende Datei/Zeile/Empfehlung, leeres Array, Umlaute und Sonderzeichen korrekt kodiert, deterministische Ausgabe (zweimal identisch).
- [ ] Secret-Regel: `message` und `recommendation` unverändert übernommen, keine zusätzlichen Felder mit Werten; Test mit Befunden aus der CLI (`bad`-Testdaten) ohne Secret in stdout/stderr.
- [ ] CLI: `--format` in allen Modi; unbekanntes Format = Exit `2`; Akzeptanztest, dass die Ausgabe als JSON parsebar ist und die Exit-Codes unverändert bleiben (`0` ohne FAIL, `1` mit FAIL).
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/` (`Severity.Level()`) | update | w1 |
| `internal/adapter/driven/report/json.go` (+ Test) | neu | w2 |
| `internal/adapter/driving/cli/` (`--format`, Reporter-Auswahl) | update | w2 |

## 5. Trigger

[`welle-04-output-ci`](../welle-04-output-ci.md) gestartet (Plan gemerged); `Severity.Level()` gemerged oder im selben Stapel.

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- Das Schema wird mit `version: 1` festgeschrieben; spätere Änderungen sind additiv oder erhöhen die Version.
- `reason` weicht additiv von [`SPEC-021`](../../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) ab (dort nicht aufgeführt); bei Einwand von w1 entfällt es, ohne dass andere Felder betroffen sind.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
