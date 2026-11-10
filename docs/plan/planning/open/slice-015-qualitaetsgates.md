# Slice slice-015: Qualitätsgates (Architekturprüfung, Self-Check, Testdaten-Matrix)

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-05-release.

**Bezug:** [`LH-QA-001`](../../../../spec/lastenheft.md#lh-qa-001--unit-tests), [`LH-QA-002`](../../../../spec/lastenheft.md#lh-qa-002--integration-tests), [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten), [`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md), [`ARC-007`](../../../../spec/architecture.md#2-schichten-und-constraints).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Pull Requests führen die Prüfungen aus, die das Lastenheft verlangt
([`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung)): Unit-Tests und Linting (laufen bereits), eine **Architekturprüfung**
der Schichtenregeln und einen **Self-Check** des Tools gegen das eigene
Repository. Die Testdaten-Abdeckung nach [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) wird als Test festgehalten.

## 2. Bausteine

- **Architekturprüfung (w1):** Paket `internal/arch`. Die Regeln (`arch.Check(dir, import)`) bewerten jeden Import jeder Produktionsdatei; `TestArchitecture` wendet sie auf das ganze Repository an (ohne `vendor/`, `.tmp/`, `testdata/`, `_test.go`). Geprüft werden: `core/model` importiert nur die Stdlib; Kern, Ports und Test-Doubles importieren keine Drittbibliotheken und keine Seiteneffekt-Pakete (`os`, `os/exec`, `net`, `net/http`, …); `rules` nur `model`/`port`; `app` nur `model`/`rules`/`port`; `port/driven` nur `model`; driven-Adapter weder `rules` noch `app` noch driving-Adapter noch Test-Doubles; Produktionscode importiert keine Test-Doubles; **`os/exec` und ein AWS-SDK nur in `adapter/driven/awscli`** (`ARC-007`). `TestRulesDetectViolations` belegt per Tabelle, dass jede Regel anschlägt. Der Test läuft in `make test` und damit in `make gates` und CI.
- **Testdaten-Matrix (w1):** `TestTestdataCoverage` prüft je Dateiart (`devcontainer.json`, `docker-compose.yml`, `Dockerfile`, `main.tf`, `.claude/settings.json`) ein `good`- und ein `bad`-Beispiel unter `internal/adapter/driving/cli/testdata`.
- **Self-Check (w2):** `make self-check` baut das Binary und führt es mit kontrollierter Umgebung (`env -i`, gesetzte `AWS_REGION`/`CLAUDE_CODE_USE_BEDROCK`, leeres Home) im Modus `local --root .` und `terraform --root .` gegen das eigene Repository aus; Exit `0` ist Pflicht. Das Repository bekommt dafür eine eigene `bedrock-eu-check.yml` mit `exclude` für `internal/adapter/driving/cli/testdata` und `vendor` ([`slice-019`](slice-019-scan-exclude.md)). Läuft in CI als eigener Schritt; ohne AWS-Zugriff, daher ohne Modus `aws`.

## 3. Lücke bei [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten)

Die Anforderung nennt Testdaten für `.env` und `iam-policy.json`. Für beide Dateiarten
gibt es **keine** funktionale Anforderung und keinen Scanner; das Lastenheft ist
vertraglich bindend und wird nicht einseitig erweitert. Der Test loggt die Lücke
bei jedem Lauf. Entscheidung durch die Auftraggeber nötig:

1. **Scanner ergänzen** (Lastenheft um Anforderungen erweitern, neue Regel-IDs): `.env`/`dotenv` über `EnvFile`, `iam-policy.json` über die vorhandene Policy-Bewertung ( [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen) / [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) ). Aufwand klein, weil beide auf bestehenden Bausteinen aufsetzen.
2. **Anforderung streichen oder präzisieren** („Testdaten für alle gescannten Dateiarten“).

Bis dahin gilt: kein Scanner, keine Testdaten, Lücke dokumentiert.

## 4. Definition of Done

- [ ] `internal/arch` mit Regeln, Repository-Test und Tabellentest; Gates grün (w1, in diesem Slice umgesetzt).
- [ ] Testdaten-Matrix als Test (w1, umgesetzt).
- [ ] `make self-check` und `bedrock-eu-check.yml` im Repo, CI-Schritt (w2).
- [ ] Sensors-Tabellen in `AGENTS.md` und `harness/README.md` um `self-check` ergänzt, sobald es läuft (w2).
- [ ] Entscheidung zur Lücke `.env`/`iam-policy.json` dokumentiert (Auftraggeber).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/arch/` | neu | w1 |
| `Makefile`, `bedrock-eu-check.yml`, CI-Workflow, Sensors-Tabellen | update/neu | w2 |

## 6. Trigger

[`welle-05-release`](../welle-05-release.md) gestartet; [`slice-019`](slice-019-scan-exclude.md) gemerged (für den Self-Check).

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- Die Architekturregeln stehen im Code (`arch.go`) **und** in `spec/architecture.md`; bei Änderungen an der Spec muss der Test nachgezogen werden (Drift). Der Test ist die ausführbare Wahrheit für die Importregeln.
- Die Regel „Kern bleibt seiteneffektfrei“ prüft nur eine Liste von Stdlib-Paketen; sie erkennt keine Seiteneffekte durch erlaubte Pakete (z. B. `fmt` mit Ausgabe).
- Nur Produktionsdateien werden geprüft; Testdateien dürfen Test-Doubles importieren.
- Der Self-Check prüft nur, was im Repository liegt (derzeit kaum Zielartefakte); er belegt vor allem, dass das Binary im Build läuft und die Konfiguration lädt.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Prüfungen auf dem bestehenden Code.
