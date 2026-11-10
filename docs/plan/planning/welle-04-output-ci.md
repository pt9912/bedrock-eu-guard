# Welle welle-04-output-ci: Ausgabeformate und CI-Modus

**Status:** open

**Zielmeilenstein:** M4 — CI-fähig

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Welle-Ziel

`bedrock-eu-check` liefert dieselben Findings wahlweise als Konsolentext
([`LH-FA-REP-001`](../../../spec/lastenheft.md#lh-fa-rep-001--konsolenausgabe)), als JSON ([`LH-FA-REP-002`](../../../spec/lastenheft.md#lh-fa-rep-002--json-ausgabe), [`SPEC-021`](../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json)) und als SARIF 2.1.0 für
Code-Scanning ([`LH-FA-REP-003`](../../../spec/lastenheft.md#lh-fa-rep-003--sarif-ausgabe), [`SPEC-022`](../../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif)). Der CI-Modus ([`LH-MOD-005`](../../../spec/lastenheft.md#lh-mod-005--ci-modus)) macht die Ausgabe
maschinenfreundlich (stdout nur Ergebnis, stabile Reihenfolge, keine
Farben) und beendet den Lauf mit den definierten Exit-Codes ([`SPEC-030`](../../../spec/spezifikation.md#spec-030--exit-codes)).

## 2. Trigger (Welle startet)

- welle-03-aws-checks done (M3), mindestens der Kern von slice-010 gemerged.
- Das Findings-Modell ist stabil (`Finding` hat bereits `File`, `Line`, `Recommendation`).

## 3. Closure-Trigger (Welle schließt)

- Alle Slices done; `make verify-clean` und CI grün.
- `bedrock-eu-check all --format json` und `--format sarif` erzeugen gültige, deterministische Ausgaben; Tests prüfen Pflichtfelder und Struktur offline.
- `all --ci` endet mit Exit `0` ohne FAIL, `1` mit FAIL, `2` bei Tool-Fehler; stdout enthält nur das Ergebnis.
- Kein Secret-Wert in keinem Format ([`LH-NF-004`](../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
- Beispiel-Workflow für Code-Scanning in `docs/user/` dokumentiert.
- Closure-Notiz in `done/welle-04-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle (Entwurf, Reihenfolge)

| Slice | Titel | Bezug | Dependency |
|---|---|---|---|
| slice-011 | JSON-Reporter und Format-Auswahl (`--format console\|json`) | [`LH-FA-REP-002`](../../../spec/lastenheft.md#lh-fa-rep-002--json-ausgabe), [`SPEC-021`](../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) | nur Stdlib |
| slice-012 | SARIF-Reporter (`--format sarif`) | [`LH-FA-REP-003`](../../../spec/lastenheft.md#lh-fa-rep-003--sarif-ausgabe), [`SPEC-022`](../../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif) | slice-011 |
| slice-013 | CI-Modus `--ci`, Ausgabedisziplin, Beispiel-Workflow | [`LH-MOD-005`](../../../spec/lastenheft.md#lh-mod-005--ci-modus), [`LH-FA-EXIT-001`](../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung), [`LH-FA-EXIT-002`](../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung), [`LH-FA-EXIT-003`](../../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler) | slice-011 |

> Nur der Zuschnitt steht hier; die Slice-Dateien entstehen vor dem
> jeweiligen Start (Plan-vor-Code, Modul 5).

## 5. Abhängigkeiten

- Blockiert: Release/Distribution (Docker-Image, Releasing), die ein stabiles CLI-Verhalten voraussetzen.
- Wird blockiert von: welle-03 (vollständiges Findings-Modell aus allen Modi).

## 5a. Architektur-Entscheidungen (Entwurf, in slice-011 zu bestätigen)

1. **Reporter bleibt Adapter, Port bleibt:** `driven.Reporter.Report(findings)` ändert sich nicht. Es entstehen drei Adapter (`report.Console`, `report.JSON`, `report.SARIF`); die Auswahl trifft der Composition Root anhand von `--format`. Ein Format-Parameter im Port ist nicht nötig. Die Adapter bekommen ihren `io.Writer` im Konstruktor.
2. **Kern-Anteil klein:** `model.Severity.Level()` bildet `PASS/WARN/FAIL` auf die JSON-Schweregrade `info|warning|error` ab ([`SPEC-021`](../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json)) und für SARIF auf `none|warning|error` ([`SPEC-022`](../../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif)); das ist die einzige Kernänderung (w1), die Tabelle lebt an einer Stelle.
3. **JSON ([`SPEC-021`](../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json)):** je Finding `status` (`PASS|WARN|FAIL`), `id` (Regel-ID), `message`, `severity`, `file` (oder `null`), `line` (0 = unbekannt, dann `null`), `recommendation` (oder `null`). Top-Level: Objekt mit `version` des Schemas und `findings`-Array (Erweiterbarkeit); Reihenfolge wie die Aggregation ([`SPEC-001`](../../../spec/spezifikation.md#spec-001--severity--und-aggregationsmodell)).
4. **SARIF ([`SPEC-022`](../../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif)):** SARIF 2.1.0; `tool.driver` mit Name, Version und `rules` (eindeutige Regel-IDs); `PASS` erzeugt **kein** Result; `ruleId` = Regel-ID; `level` `warning|error`; `locations[].physicalLocation` aus `File` (relativ, `/`-getrennt, URI-kodiert) und `region.startLine` bei `Line > 0`. Findings ohne Datei (Umgebung, AWS) und Findings mit Pfaden außerhalb des Scan-Roots (z. B. `~/.claude/settings.json`) bekommen als Ersatzposition die Konfigurationsdatei `bedrock-eu-check.yml` (fehlt sie: `README.md`) mit `startLine` 1 und `properties.synthetic=true`; der Originalpfad steht nur in der Meldung. Ob Code-Scanning-Systeme das wie gewünscht darstellen, lässt sich offline nicht prüfen (siehe §7).
5. **Ausgabedisziplin:** stdout trägt ausschließlich das Ergebnis im gewählten Format; Hinweise, Warnungen und Fehler gehen auf stderr. Keine Zeitstempel, Pfade außerhalb des Scan-Roots oder zufällige IDs in der Ausgabe (Determinismus).
6. **`--ci`:** schaltet Farben und Fortschrittsanzeigen ab, setzt stabile Sortierung und ändert die Exit-Codes nicht (Exit `1` bei mindestens einem FAIL, Exit `2` bei Tool-Fehler, [`SPEC-030`](../../../spec/spezifikation.md#spec-030--exit-codes)). `--ci` ohne `--format` behält das Default-Format (Konsole); der Unterschied zum Normalbetrieb ist vor allem die Ausgabedisziplin (stdout nur Ergebnis), da der Konsolen-Reporter ohnehin weder Farben noch Fortschritt ausgibt. Ob WARN in CI den Lauf scheitern lassen darf (`--fail-on warn`) und eine Option `--output FILE` (statt Umleitung) sind Folgepunkte.
7. **Secrets:** Message, Subject und Recommendation werden unverändert übernommen; die Adapter fügen keine Felder mit Werten hinzu ([`LH-NF-004`](../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
8. **Aufteilung:** w1 = `Severity.Level()`, Plan, Review; w2 = drei Reporter-Adapter, CLI-Optionen, Schema-/Strukturtests, Beispiel-Workflow, Makefile/CI.

## 6. Out-of-Scope für diese Welle

- Docker-Image und Releasing ([`LH-TECH-004`](../../../spec/lastenheft.md#lh-tech-004--docker-image) ist ein eigener Block).
- Schwellen wie `--fail-on warn` oder Baseline-Dateien.
- HTML- oder Markdown-Berichte.
- Automatisches Hochladen nach Code-Scanning (nur dokumentierter Workflow).

## 7. Risiken und offene Punkte

- **Offline-Validierung von SARIF:** Das offizielle JSON-Schema liegt nicht im Repo; die Tests prüfen strukturell (Version, Runs, Pflichtfelder). Ein Schema im Repo wäre ein eigener Beschaffungsschritt ([`ADR-0004`](../adr/0004-abhaengigkeiten-und-offline-gates.md) gilt für Code, nicht für Testdaten).
- **Findings ohne Position:** GitHub Code Scanning erwartet pro Alert eine Datei-Position; Env-/AWS-Findings haben keine. Festgelegt (§5a Punkt 4): Ersatzposition mit `properties.synthetic=true`. Das Verhalten lässt sich offline nicht prüfen und wird beim ersten echten Upload verifiziert.
- **Rauschen:** Viele PASS/WARN-Findings blähen JSON auf; PASS bleibt im JSON (vollständiges Bild), fehlt aber in SARIF.
- **Pfad-Normalisierung:** Windows-Pfade und Pfade außerhalb des Scan-Roots (z. B. `~/.claude/settings.json`) brauchen eine definierte Darstellung im SARIF-URI.

## 8. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-04-results.md. -->
