# Welle welle-02-config-scanner: Config-Scanner (Claude Code, Devcontainer, Docker, Terraform)

**Status:** done

**Zielmeilenstein:** M2 — vollständiger statischer Scan

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Welle-Ziel

Der statische Scan: `bedrock-eu-check` liest Claude-Code-Settings,
Devcontainer-, Compose-, Dockerfile- und Terraform-Dateien (rein lesend,
offline) und liefert `PASS`/`WARN`/`FAIL` mit Datei und Zeile. Dazu kommt
die Konfigurationsdatei `bedrock-eu-check.yml` (Allow-/Blocklisten, `mantle`),
die in Welle 1 bewusst offen blieb (`RM-002` / v0.2).

## 2. Trigger (Welle startet)

- welle-01-lokaler-env-check done (erfüllt, 2026-11-10).
- **Abhängigkeitsstrategie entschieden** (neuer ADR): YAML- und HCL-Parsing
  brauchen Drittbibliotheken ([`ADR-0002`](../../adr/0002-implementierungssprache.md)), die Gates laufen aber
  mit `--network none`. Vendoring, Proxy oder anderer Weg muss vorher
  feststehen (siehe Risiken).

## 3. Closure-Trigger (Welle schließt)

- Alle Slices done.
- `make verify-clean` und CI grün.
- Scanner finden je Quelle ein bekanntes Negativ- und Positivbeispiel
  (Testdaten im Repo), Findings tragen Datei und Zeile.
- Kein Secret-Wert in der Ausgabe ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
- Closure-Notiz in `done/welle-02-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle (Entwurf, Reihenfolge)

| Slice | Titel | Bezug | Dependency |
|---|---|---|---|
| slice-004 | Dateisystem-Port + `fs`-Adapter, Scan-Gerüst; Claude-Code-Settings (JSON) | [`LH-FA-CLAUDE-001`](../../../../spec/lastenheft.md#lh-fa-claude-001--benutzerweite-claude-code-settings-prüfen), [`LH-FA-CLAUDE-002`](../../../../spec/lastenheft.md#lh-fa-claude-002--projektweite-claude-code-settings-prüfen), [`LH-FA-CLAUDE-003`](../../../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte), [`LH-FA-CLAUDE-004`](../../../../spec/lastenheft.md#lh-fa-claude-004--claude-code-wizard-konfiguration-prüfen) | nur Stdlib |
| slice-005 | Konfigurationsdatei `bedrock-eu-check.yml` (`configyaml`), Regex-Validierung, `mantle`-Feld | [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), [`LH-NF-007`](../../../../spec/lastenheft.md#lh-nf-007--konfigurierbarkeit) | YAML-Bibliothek |
| slice-006 | Devcontainer-, Compose- und Dockerfile-Scanner | [`LH-FA-DEVCON-001`](../../../../spec/lastenheft.md#lh-fa-devcon-001--prüfung-von-devcontainerdevcontainerjson), [`LH-FA-DOCKER-001`](../../../../spec/lastenheft.md#lh-fa-docker-001--prüfung-von-docker-composeyml), [`LH-FA-DOCKER-002`](../../../../spec/lastenheft.md#lh-fa-docker-002--prüfung-von-dockerfile) | YAML (Compose) |
| slice-007 | Terraform-Scanner | [`LH-FA-TF-001`](../../../../spec/lastenheft.md#lh-fa-tf-001--terraform-dateien-finden), [`LH-FA-TF-002`](../../../../spec/lastenheft.md#lh-fa-tf-002--aws-provider-region-prüfen), [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen), [`LH-FA-TF-004`](../../../../spec/lastenheft.md#lh-fa-tf-004--nicht-erlaubte-modell-ids-finden) | HCL-Bibliothek |

> Nur der Zuschnitt steht hier; die Slice-Dateien entstehen vor dem
> jeweiligen Start (Plan-vor-Code, Modul 5). slice-004 braucht keine
> Drittbibliothek und kann sofort starten, sobald die Welle beginnt.

## 5. Abhängigkeiten

- Blockiert: welle-03 (AWS-Checks nutzen Konfiguration und Findings-Modell).
- Wird blockiert von: Abhängigkeits-ADR (nur slice-005 bis slice-007).

## 5a. Architektur-Entscheidungen (Entwurf, in slice-004 zu bestätigen)

1. **Finding-Vertrag:** `model.Finding` bekommt additiv `File string`, `Line int`, `Recommendation string` (Nullwerte = bisheriges Verhalten; [`SPEC-021`](../../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) und [`LH-FA-REP-002`](../../../../spec/lastenheft.md#lh-fa-rep-002--json-ausgabe) verlangen sie). Eine einzige Vertragsänderung durch w1, vorher angekündigt. Der Konsolen-Reporter gibt optional `Datei:Zeile` aus (w2).
2. **Parser im driven-Adapter:** Wie in [`spec/architecture.md`](../../../../spec/architecture.md) vorgesehen, parsen die Adapter. Die Ports liefern geparste `model`-Typen (z. B. `driven.ClaudeSettingsReader` → `model.ClaudeSettings`); der Kern bleibt frei von Parsern, Regeln bleiben rein und mit Fakes testbar.
3. **Filesystem-Port:** `ReadFile(path) ([]byte, error)`, `Exists(path) (bool, error)`, `Glob(pattern) ([]string, error)`; Test-Double `coretest/memfs` (w1).
4. **Fehlerpolitik für gescannte Dateien** (Empfehlung): Datei fehlt → je nach Anforderung (z. B. `WARN` laut [`LH-FA-CLAUDE-001`](../../../../spec/lastenheft.md#lh-fa-claude-001--benutzerweite-claude-code-settings-prüfen)). Datei vorhanden, aber nicht lesbar oder syntaktisch ungültig → `FAIL` mit Reason `FILE_UNREADABLE`/`FILE_INVALID` und Scan läuft weiter, weil EU-Konformität nicht belegbar ist. Exit `2` (Tool-Fehler) bleibt Fehlern des Tools selbst vorbehalten (Argumente, ungültige eigene Config, [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes)).
   Die Konstanten `FILE_UNREADABLE`/`FILE_INVALID` legt w1 in `model` an. Meldungen zu ungültigen Dateien nennen nur Pfad, Zeile und Fehlertyp des Parsers, nie Dateiinhalt ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)). Pro Regel benennt der Slice-Plan das Verhalten bei fehlender Datei, damit sie eindeutig testbar ist.
5. **Aufteilung:** w1 = Kern (`model`, Ports, Regeln, Doubles); w2 = Adapter (`fs`, `configyaml`, Parser), Abhängigkeits-ADR, Makefile/CI, Reporter.

## 6. Out-of-Scope für diese Welle

- AWS-Live-Calls (Welle 3).
- JSON-/SARIF-Ausgabe und `--ci` (Welle 4).
- Automatische Korrektur von Dateien ([`LH-SEC-005`](../../../../spec/lastenheft.md#lh-sec-005--keine-automatischen-remediations)).

## 7. Risiken

- **Offline-Gates vs. Drittbibliotheken:** `GO_RUN` nutzt `--network none`. Entscheidung nötig (Vendoring unter `vendor/`, vorab gefülltes Modul-Cache-Volume oder ein Download-Schritt). Betrifft auch die Offline-Pflicht des Tools ([`LH-NF-005`](../../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)).
- **Zeilenangaben** ([`LH-FA-REP-001`](../../../../spec/lastenheft.md#lh-fa-rep-001--konsolenausgabe)) brauchen im Finding-Modell ein Feld für Datei und Zeile (Änderung am `model`, Kern).
- **JSONC** in Devcontainer-Dateien (Kommentare) ist mit Stdlib-`encoding/json` nicht lesbar.

## 8. Closure-Notiz

Abgeschlossen am 2026-11-10, siehe [`done/welle-02-results.md`](welle-02-results.md).
