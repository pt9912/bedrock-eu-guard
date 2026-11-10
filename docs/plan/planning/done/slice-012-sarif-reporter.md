# Slice slice-012: SARIF-Reporter

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-04-output-ci.

**Bezug:** [`LH-FA-REP-003`](../../../../spec/lastenheft.md#lh-fa-rep-003--sarif-ausgabe), [`SPEC-022`](../../../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check <modus> --format sarif` gibt die Findings als SARIF 2.1.0
nach [`SPEC-022`](../../../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif) aus, sodass Code-Scanning-Systeme (GitHub, GitLab) sie
verarbeiten können.

## 2. Verträge

- **Adapter `report.SARIF(w io.Writer, version string)`** implementiert `driven.Reporter`. Aufbau: `$schema` (feste URL der SARIF-2.1.0-Schemas, nur als Zeichenkette, kein Netzzugriff), `version: "2.1.0"`, ein `run` mit `tool.driver` (`name: bedrock-eu-check`, `version`, `rules`).
- **Tool-Version:** Paketvariable, per `-ldflags` setzbar, Default `dev`; die CLI reicht sie an den Reporter.
- **Regeln:** `tool.driver.rules` enthält je vorkommender Regel-ID genau einen Eintrag (`id`, `name`), sortiert; kein Eintrag für Regeln ohne Result.
- **Results:** `PASS` erzeugt kein Result; `WARN` → `level: warning`, `FAIL` → `level: error` (`SARIFLevel()` im Kern, w1). `ruleId` = Regel-ID, `message.text` = Message (unverändert, ggf. mit Empfehlung als zweiter Satz), Reihenfolge wie die Aggregation.
- **Position:** Findings mit `File` innerhalb des Scan-Roots bekommen `locations[0].physicalLocation` mit `artifactLocation.uri` (relativ, `/`-getrennt, pro Segment URI-kodiert, `uriBaseId: "%SRCROOT%"`) und `region.startLine` nur bei `Line > 0`.
- **Findings ohne Position** (Umgebung, AWS, Pfade außerhalb des Scan-Roots wie `~/.claude/settings.json`): Code-Scanning-Systeme erwarten eine Datei-Position. Ersatz: `bedrock-eu-check.yml`, falls sie existiert, sonst `README.md`, jeweils mit `startLine: 1`; zusätzlich `properties.synthetic: true` und `properties.source` (der ursprüngliche Pfad bzw. `env`/`aws`). Der Ersatz ist damit erkennbar und täuscht keine Fundstelle vor.
- **CLI:** `--format sarif` in allen Modi; `all` und `terraform` sind die üblichen Kandidaten. stdout nur das Dokument.

## 3. Definition of Done

- [ ] `report.SARIF` mit strukturellen Tests (kein Schema im Repo, offline): `version`, `runs[0].tool.driver.name/rules`, `results[]` mit `ruleId`, `level`, `message.text`, `locations`; `PASS` ohne Result; leerer Lauf ergibt ein gültiges Dokument mit leerem `results`-Array.
- [ ] Pfadregeln getestet: relative `/`-Pfade, Sonderzeichen und Leerzeichen kodiert, Windows-Trenner normalisiert, Ersatzposition mit Kennzeichnung, Pfad außerhalb des Scan-Roots.
- [ ] Determinismus: zweimal identische Ausgabe; keine Zeitstempel oder GUIDs.
- [ ] Secret-Regel: Message und Empfehlung unverändert, keine Werte in `properties`; Test mit `bad`-Testdaten ohne Secret in stdout/stderr.
- [ ] CLI-Akzeptanztest: `terraform --format sarif` auf den `bad`-Testdaten liefert parsebares SARIF mit `error`-Results; Exit `1`.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/` (`Severity.SARIFLevel()`) | update | w1 |
| `internal/adapter/driven/report/sarif.go` (+ Test) | neu | w2 |
| `internal/adapter/driving/cli/` (`--format sarif`, Tool-Version) | update | w2 |

## 5. Trigger

[`slice-011`](slice-011-json-reporter.md) gemerged (Format-Auswahl und `Severity`-Abbildung vorhanden).

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- Das offizielle SARIF-Schema liegt nicht im Repo; die Tests prüfen strukturell. Ob GitHub die synthetische Ersatzposition akzeptiert, lässt sich offline nicht prüfen.
- Rauschen: `PASS` fehlt in SARIF bewusst; Warnungen ohne Position können Code-Scanning-Alerts auf einer Ersatzdatei erzeugen.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
