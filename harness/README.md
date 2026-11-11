# Harness

## Purpose

Dieser Harness verbindet die Spezifikationen, ADRs, Planning-Dokumente und
Gates von **bedrock-eu-guard**. Er ist **kein Ersatz** für `spec/` oder
`docs/`, sondern ein **Einstiegspunkt** für Menschen und AI-Code-Agenten.

Wenn diese Datei einer kanonischen Quelle widerspricht, **gewinnt die
kanonische Quelle**, und diese Datei wird angepasst.

Strukturregeln (Verzeichniskonvention, ID-Schemata, Modus-Deklarationen
pro Sub-Area, Repo-Klasse, Adaptionen ggü. Baseline) leben in
[`conventions.md`](conventions.md). Diese Datei dupliziert sie nicht.

## Source precedence

| Rang | Datei | Charakter |
|---|---|---|
| 1 | [`spec/lastenheft.md`](../spec/lastenheft.md) | vertraglich abnahmebindend |
| 2 | [`spec/spezifikation.md`](../spec/spezifikation.md) | technisch fortschreibbar (3. Spec-Stratum, siehe `conventions.md` [`MR-001`](conventions.md#mr-001--source-precedence-mit-eigener-spezifikations-schicht)) |
| 3 | [`spec/architecture.md`](../spec/architecture.md) | Komponenten/Sequenzen, meilensteinfrei |
| 4 | [`docs/plan/adr/`](../docs/plan/adr/) | Architekturentscheidungen |
| 5 | [`docs/plan/planning/in-progress/roadmap.md`](../docs/plan/planning/in-progress/roadmap.md) | aktuelle Welle |
| 6 | [`docs/user/releasing.md`](../docs/user/releasing.md) | Operations, Releasing |
| 7 | [`README.md`](../README.md) | Projekt-Überblick |
| 8 | [`AGENTS.md`](../AGENTS.md) | Agent-Briefing |
| 9 | diese Datei | Harness-Einstieg |

Wenn diese Datei einer kanonischen Quelle widerspricht, **gewinnt die
kanonische Quelle**, und diese Datei wird angepasst.

## Guides (Feedforward-Quellen)

| Quelle | Inhalt |
|---|---|
| [`spec/lastenheft.md`](../spec/lastenheft.md) | Anforderungen (`LH-*`), IDs, Akzeptanzkriterien |
| [`spec/spezifikation.md`](../spec/spezifikation.md) | Regel-Engine, Klassifikation, Config-/Finding-Schema (`SPEC-*`) |
| [`spec/architecture.md`](../spec/architecture.md) | Komponenten, Schichten, Constraints (`ARC-*`) |
| [`docs/plan/adr/`](../docs/plan/adr/) | Architekturentscheidungen |
| [`docs/plan/planning/`](../docs/plan/planning/) | Slice-Pläne und Roadmap |
| [`AGENTS.md`](../AGENTS.md) | Hard Rules, Source Precedence, Workflow |
| [`conventions.md`](conventions.md) | repo-lokale Strukturregeln, Adaptions-Block (`MR-*`), Modus-Deklarationen |
| [Regelwerk v1.4.0 (Modul-Bundle, ZIP)](https://github.com/pt9912/ai-harness-course/releases/download/v1.4.0/lab-regelwerk.zip) | adoptiertes Regelwerk, self-navigierbares Modul-Bundle; derivativ, Stand siehe [`conventions.md`](conventions.md) §Baseline |

## Sensors (Feedback-Gates)

| Target | Vertrag | Bindung |
|---|---|---|
| `make doc-check` | Doku-Referenzen lösen auf (Links + Anker; via d-check, netzlos `--network none`) | Reproduzierbarkeit (`:v0.91.0` in `d-check.mk`; `DCHECK_DIGEST` für Digest-Pin) |
| `make lint` | `gofmt -l .` (leer) + `go vet ./...` | Go-Image per Digest gepinnt (`GO_IMAGE` im `Makefile`), netzlos |
| `make test` | `go test ./...` (Unit-Tests je Regel/Adapter, [`LH-QA-001`](../spec/lastenheft.md#lh-qa-001--unit-tests)) | wie `lint` |
| `make build` | Binary `bin/bedrock-eu-check` bauen | wie `lint` |
| `make gates` | alle lauffähigen Gates: `doc-check`, `lint`, `test` | — |
| `make verify-clean` | `make gates` im frischen `git worktree` des committeten HEAD (fängt ungetrackte/ignorierte Dateien ab) | Handoff-Gate; CI führt `make gates` im frischen Checkout aus (`.github/workflows/gates.yml`) |
| `make deps` | `go mod tidy` + `vendor/` neu erzeugen (mit Netz, nicht Teil der Gates) | [`ADR-0004`](../docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md): `vendor/` nie von Hand editieren |
| `make deps-verify` | `go mod verify` + `vendor/` entspricht Upstream (mit Netz) | CI-Job `deps-verify`, nur bei Änderungen an `vendor/`, `go.mod`, `go.sum` |
| `make self-check` | `bedrock-eu-check local` und `terraform` gegen das eigene Repository ([`LH-QA-005`](../spec/lastenheft.md#lh-qa-005--ci-prüfung)); Exit 0 Pflicht | `env -i` mit EU-Minimalumgebung; Ausschlüsse (Testdaten, `vendor/`) in `bedrock-eu-check.yml`; CI-Schritt in `gates.yml` |
| `make image` | Docker-Images `slim` und `full` bauen (Basis-Images per Digest) | braucht Netz zum Ziehen der Basen; nicht Teil von `make gates` |
| `make image-test` | `make image` plus Smoke-Tests (`tools/image-smoke.sh`: kein Root, `version`, `local` gut/schlecht, `terraform --format json`, `aws` je Variante) | CI-Job `image`, nur bei Änderungen, die das Image betreffen |
| `make feature-test` | Devcontainer-Feature (`devcontainer-feature/`) gegen ein lokal gebautes „Release“ testen: Struktur, Syntax, Installation per `file://`, Negativfälle (Prüfsumme, Version) | CI-Job `image`; nicht Teil von `make gates` |
| `make dist` | Release-Artefakte (Linux `amd64`/`arm64`) und `SHA256SUMS` nach `dist/v<version>/` bauen und verifizieren (`VERSION=x.y.z` Pflicht) | Workflow `release-artifacts` (nur manuell); veröffentlicht nichts |

**Aktueller Lauf-Status:** lokal `make gates` / `make help`; in CI das
Pipeline-Ergebnis (Rang höher als diese Datei).
**Nicht behauptet** (geplant, entstehen in späteren Slices): `make coverage-gate`
(bootstrap-aware), `make self-check` (`bedrock-eu-check` gegen das eigene
Repo, [`LH-QA-005`](../spec/lastenheft.md#lh-qa-005--ci-prüfung)). Diese Targets existieren **noch nicht** im
`Makefile` und dürfen erst eingetragen werden, wenn sie real laufen
(keine halluzinierten Gates, Modul 13).

## Traceability rules

- PRs/Commits **müssen** mindestens eine `LH-*`- oder `ADR-*`-ID nennen.
- Neue oder geänderte Anforderungen brauchen einen Beleg: Test, Gate, Demo oder ADR.
- Neue ADRs müssen im [ADR-Index](../docs/plan/adr/README.md) ergänzt werden.
- Änderungen an Planning-Dokumenten beachten die Lifecycle-Regeln
  (`open → next → in-progress → done`; reine `git mv`-Commits, siehe
  [`AGENTS.md`](../AGENTS.md) §3.3).

## Safety and scope boundaries

Repo-Klasse *Tooling/Referenz* — die Grenzen ergeben sich direkt aus dem
Lastenheft (§4 Nicht-Ziele, §10 Sicherheit):

- **Keine Rechtsberatung.** `bedrock-eu-check` prüft *technische
  Konfigurationen*, ist kein Datenschutzgutachten und keine vertragliche
  Zusicherung ([`LH-NZ-001`](../spec/lastenheft.md#lh-nz-001--kein-rechtlicher-nachweis), [`LH-RISK-004`](../spec/lastenheft.md#lh-risk-004--scheinbare-sicherheit)).
- **Kein Beweis über AWS-internes Routing.** Geprüft werden nur sichtbare,
  steuerbare Konfigurationen ([`LH-NZ-002`](../spec/lastenheft.md#lh-nz-002--keine-garantie-interner-aws-routings)).
- **Niemals Secrets im Klartext ausgeben.** `ANTHROPIC_API_KEY` &
  Co. werden nur als „set"/„not set" gemeldet ([`LH-NF-004`](../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe),
  [`LH-SEC-002`](../spec/lastenheft.md#lh-sec-002--keine-speicherung-von-claude-anthropic-secrets)).
- **Read-only gegenüber AWS.** Version 1 nutzt ausschließlich lesende
  AWS-Operationen, keine automatischen Remediations
  ([`LH-SEC-003`](../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff), [`LH-SEC-005`](../spec/lastenheft.md#lh-sec-005--keine-automatischen-remediations), [`LH-NZ-003`](../spec/lastenheft.md#lh-nz-003--keine-automatische-änderung-produktiver-aws-konfigurationen)).
- **Offline-Pfad muss offline bleiben.** Lokale und Terraform-Prüfungen
  funktionieren ohne Internet/AWS ([`LH-NF-005`](../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)).

## Minimal agent workflow

1. Diese Datei lesen.
2. Relevante kanonische Quelle lesen (Source Precedence beachten).
3. Betroffene `LH-*`/`ADR-*`-IDs identifizieren.
4. Kleinste sinnvolle Änderung planen.
5. Engsten nützlichen Sensor laufen lassen.
6. Repo-weiten Gate-Lauf vor Handoff (`make gates`).
7. Doku/Indizes aktualisieren, falls ein öffentlicher Vertrag berührt ist.
8. Ausgeführte Sensors und verbleibende Risiken berichten — keine
   Erfolgsmeldung ohne Gate-Ausführung.
