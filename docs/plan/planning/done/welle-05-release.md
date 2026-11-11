# Welle welle-05-release: Release 1.0

**Status:** done (Release 1.0.0 veröffentlicht am 2026-11-11)

**Zielmeilenstein:** M5 — Release 1.0

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Welle-Ziel

Eine stabile Team-Version ([`LH-RM-005`](../../../../spec/lastenheft.md#lh-rm-005--version-10)): `bedrock-eu-check` ist als Binary und als
Docker-Image ([`LH-TECH-004`](../../../../spec/lastenheft.md#lh-tech-004--docker-image)) reproduzierbar baubar ([`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)), im Devcontainer nutzbar
([`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)), hat den Befehl `version` ([`LH-TECH-002`](../../../../spec/lastenheft.md#lh-tech-002--cli-struktur)), eine vollständige Nutzerdokumentation
und die Qualitätsgates, die das Lastenheft für Pull Requests verlangt ([`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung)).

## 2. Trigger (Welle startet)

- welle-04-output-ci done (M4): JSON, SARIF und `--ci` stehen.
- Das CLI-Verhalten ist stabil genug, um es zu dokumentieren und zu verteilen.

## 3. Closure-Trigger (Welle schließt)

- Alle Slices done; `make verify-clean` und CI grün.
- `bedrock-eu-check version` zeigt Version und Commit; die SARIF-Tool-Version stammt aus derselben Quelle.
- Das Docker-Image baut reproduzierbar aus gepinnten Basen und führt `bedrock-eu-check all --ci` auf einem Beispiel-Repository aus.
- Der Self-Check (`bedrock-eu-check` gegen das eigene Repo) und die Architekturprüfung laufen in CI.
- Positiv- und Negativbeispiele decken die Dateiarten aus [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) ab (oder Lücken sind begründet dokumentiert).
- Nutzerdokumentation in `docs/user/` vollständig; `releasing.md` beschreibt den realen Ablauf.
- Release-Tag `v1.0.0` ist gesetzt **nach ausdrücklicher Freigabe durch die Nutzer**; Veröffentlichung von Artefakten ist ein eigener, freigegebener Schritt.
- Closure-Notiz in `done/welle-05-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle (Entwurf, Reihenfolge)

| Slice | Titel | Bezug | Dependency |
|---|---|---|---|
| slice-014 | Befehl `version`, Build-Informationen | [`LH-TECH-002`](../../../../spec/lastenheft.md#lh-tech-002--cli-struktur) | nur Stdlib |
| slice-019 | Vorlauf: Scan-Root-Ausschluss (`exclude`) in Konfiguration und Discovery | [`LH-NF-007`](../../../../spec/lastenheft.md#lh-nf-007--konfigurierbarkeit), [`LH-TECH-003`](../../../../spec/lastenheft.md#lh-tech-003--konfigurationsdatei) | slice-005 |
| slice-015 | Qualitätsgates: Architekturprüfung, Self-Check, Testdaten-Abdeckung | [`LH-QA-001`](../../../../spec/lastenheft.md#lh-qa-001--unit-tests), [`LH-QA-002`](../../../../spec/lastenheft.md#lh-qa-002--integration-tests), [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten), [`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung) | slice-014, slice-019 |
| slice-016 | Docker-Image und reproduzierbarer Build | [`LH-TECH-004`](../../../../spec/lastenheft.md#lh-tech-004--docker-image), [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build) | slice-014 |
| slice-017 | Devcontainer-Feature | [`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration) | slice-016 |
| slice-018 | Nutzerdokumentation, Release-Artefakte und Release-Ablauf | [`LH-RM-005`](../../../../spec/lastenheft.md#lh-rm-005--version-10) | slice-014–017 |

> Nur der Zuschnitt steht hier; die Slice-Dateien entstehen vor dem
> jeweiligen Start (Plan-vor-Code, Modul 5).

## 5. Abhängigkeiten

- Blockiert: nichts (Endpunkt der aktuellen Roadmap).
- Wird blockiert von: welle-04 (stabiles Ausgabeformat).

## 5a. Architektur-Entscheidungen (Entwurf, in den Slices zu bestätigen)

1. **Version (slice-014):** Version und Commit kommen über `-ldflags` zur Buildzeit (Default `dev`); ein Paket im CLI-Rahmen hält sie. `version` gibt eine Zeile aus (`bedrock-eu-check <version> (<commit>)`), kein Netz, Exit `0`. Die SARIF-Tool-Version nutzt dieselbe Quelle. Kein Kernanteil.
2. **Architekturprüfung (slice-015, Kern/w1):** Die Schichtenregeln aus [`spec/architecture.md`](../../../../spec/architecture.md) §2 werden als Go-Test erzwungen (Importgraph per `go list`): `core/model` importiert nichts Internes; `rules` und `app` importieren keinen Adapter; `awscli` ist der einzige AWS-Zugang ([`ARC-007`](../../../../spec/architecture.md#2-schichten-und-constraints)); Adapter importieren weder `rules` noch `app`. Ein Verstoß ist ein Testfehler und damit in `make test` bzw. einem eigenen Target `make arch-check` enthalten.
3. **Self-Check (slice-015):** `bedrock-eu-check local --root .` gegen das eigene Repository in CI; das Repo muss dafür selbst EU-konform sein (keine Funde, sonst begründete Testdaten-Ausnahmen über Konfiguration). Testdaten mit absichtlich schlechten Beispielen liegen unter `testdata/` und werden vom Self-Check ausgenommen. Dafür braucht es ein neues Konfigurationsfeld `exclude` (Glob-Muster relativ zum Scan-Root) in [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), `PartialConfig`/`Resolve` (Kern, w1) und die Anwendung in der Datei-Discovery von `fs`, `terraform` und `envsource` (Adapter, w2); das ist ein eigener Vorlauf-Slice (slice-019) vor slice-015.
4. **Testdaten ([`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten)):** Eine Abdeckungsmatrix prüft je Dateiart (`.env`, `devcontainer.json`, `docker-compose.yml`, `Dockerfile`, `main.tf`, `iam-policy.json`, `.claude/settings.json`) Positiv- und Negativbeispiel. Für `.env` und `iam-policy.json` gibt es derzeit keinen Scanner; ob sie Teil von 1.0 sind oder als Folgepunkt dokumentiert werden, entscheidet slice-015 (nicht stillschweigend weglassen).
5. **Docker-Image (slice-016, w2):** Mehrstufiger Build mit gepinnten Basis-Images (Digest), statisch gelinktes Binary, Laufzeit-Image ohne Shell-Zwang, Nutzer ohne Root, Arbeitsverzeichnis `/work`. Entscheidung, ob das `aws`-Binary im Image enthalten ist (für den Modus `aws` nötig, für `local`/`terraform` nicht): zwei Varianten (`slim` ohne, `aws` mit) oder nur eine. Das Image enthält keine Secrets und speichert keine Credentials. Basis-Images werden per Digest gepinnt; Bau und Test des Images brauchen Netz (Basis-Images ziehen) und laufen deshalb als eigener CI-Job mit Netz, **nicht** in `make gates` (analog zu `deps-verify`).
6. **Devcontainer-Feature (slice-017, w2):** `devcontainer-feature.json` und `install.sh` im Repo (Feature-Format der Devcontainer-Spezifikation), lädt das Release-Binary mit geprüfter Prüfsumme oder nutzt das Image; Test per Beispiel-Devcontainer.
7. **Release (slice-018):** Reproduzierbarer Build über `make` und Image-Digests (`-trimpath`, `CGO_ENABLED=0`, `-buildvcs=false`, feste `-ldflags`; zwei Builds müssen dieselbe SHA256 liefern); Release-Artefakte (Binaries je Plattform, Prüfsummen, ggf. SBOM) werden von einem Workflow bei einem Tag gebaut. **Tag und Veröffentlichung sind außenwirksam und nur mit ausdrücklicher Freigabe der Nutzer.** Die Dokumentation (`docs/user/`): Schnellstart, Befehle und Optionen, Konfigurationsreferenz, Regelkatalog (Regel-ID → Bedeutung → Reasons), CI-Einbindung, AWS-Rechte, Fehlersuche.
8. **Aufteilung:** w1 = Architekturprüfung, Self-Check-Konzept, Testdaten-Matrix, Regelkatalog-Doku, Plan/Review; w2 = Version, Docker-Image, Devcontainer-Feature, Release-Workflow, Makefile/CI. Beide teilen sich die Nutzerdokumentation.

## 6. Out-of-Scope für diese Welle

- Neue Prüfregeln (kommen in Folgeslices der Qualitätsrunde, siehe §7).
- Windows-/macOS-spezifische Installer oder Paketmanager-Einträge.
- Signierte Releases (Sigstore/GPG) — als Folgepunkt prüfen.
- Schwellenwert-Gates wie `coverage-gate` (jede Schwelle braucht einen ADR, `AGENTS.md` §3.6).

## 7. Folgeslices der Qualitätsrunde (nach oder parallel, nicht Teil von 1.0)

- `apiKeyHelper` als direkte Anthropic-Konfiguration ([`LH-FA-CLAUDE-003`](../../../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte); Spec-Präzisierung nötig, kein stilles FAIL).
- „Verwendet“ in den Bedrock-Regeln auch aus Claude-Settings und Konfiguration, nicht nur aus der Umgebung ([`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen)).
- Abfrage zusätzlicher Nicht-EU-Regionen für „sichtbare“ Profile.
- `--fail-on warn` und `--output FILE` für den CI-Modus.
- `coverage-gate` mit begründeter Schwelle (eigener ADR).

## 8. Risiken und offene Punkte

- **Außenwirkung:** Image-Veröffentlichung, Tag und Release sind nicht rückholbar; sie brauchen eine ausdrückliche Freigabe der Nutzer und laufen nie automatisch beim Merge.
- **`aws`-Binary im Image:** Größe, Lizenz und Updates der AWS-CLI; gepinnte Version nötig ([`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)).
- **Reproduzierbarkeit:** Build-Zeitstempel, Pfade und Modulcache dürfen die Binary-Prüfsumme nicht verändern (`-trimpath`, feste Umgebung); ob bitgleiche Builds erreichbar sind, klärt slice-016.
- **Devcontainer-Feature-Test:** ohne Netz und ohne Container-Laufzeit in den Gates nur strukturell prüfbar.
- **Self-Check auf dem eigenen Repo** kann an absichtlich schlechten Testdaten scheitern; Ausschluss muss explizit und sichtbar sein.

## 9. Closure-Notiz

Zwischenstand 2026-11-10: Release-Kandidat, siehe [`done/welle-05-results.md`](welle-05-results.md). Die Welle schließt erst mit dem ausdrücklich freigegebenen Release (M5).
