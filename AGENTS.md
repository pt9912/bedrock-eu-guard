# AGENTS.md — Briefing für AI-Coding-Agenten

## 1. Was diese Datei ist

Onboarding-Briefing für jede AI-Session, die in **bedrock-eu-guard** Code
oder Dokumentation ändert. Sie verweist auf die kanonischen Quellen und
formuliert die Hard Rules, die der Implementation-Agent immer einhalten
muss.

**Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt
die kanonische Quelle** (Source Precedence — siehe
[`harness/README.md`](harness/README.md)).

Strukturregeln (ID-Schemata, Verzeichniskonvention, Repo-Klasse,
Adaptionen ggü. Baseline, Modus-Deklarationen pro Sub-Area) leben in
[`harness/conventions.md`](harness/conventions.md).

Das Betriebsregelwerk der adoptierten Baseline in Agenten-Kurzform
einmal pro Session lesen, bevor der Workflow (§6) startet. **Lese-Form ist
ausschließlich das nach Modulen und Grundlagen-Abschnitten aufgeteilte
Release-Bundle**
[`lab-regelwerk.zip`](https://github.com/pt9912/ai-harness-course/releases/download/v1.4.0/lab-regelwerk.zip)
(`v1.4.0`): herunterladen, entpacken und **nur den benötigten Abschnitt**
laden (die Bundle-`README.md` ist der Index), ohne das gesamte Regelwerk
im Kontext zu halten. Konventioneller Ablageort des entpackten **Regelwerks** ist
`.harness/cache/<tag>/regelwerk/` (gitignored; Pfadschema siehe
[`harness/conventions.md`](harness/conventions.md) §Adoptierte
Konventions-Quellen); materialisieren per
`tools/harness/fetch-baseline-cache.sh` (zieht + entpackt, Tag aus §Baseline).
**Templates** werden dagegen nicht gefetcht, sondern **verkörpert**: die
wiederkehrenden Skelette (ADR, Slice, Welle, Carveout, Review-Report) liegen
co-located als `*.template.md` im Repo — beim Anlegen eines Artefakts das
passende kopieren. Singleton-Skelette (AGENTS, conventions, README, spec,
roadmap) werden beim Bootstrap einmalig gefüllt; dasselbe Skript entpackt
`lab-templates.zip` nach `.harness/cache/<tag>/templates/` nur als
Adoptions-/Drift-Audit-Staging, nicht als Autorenquelle
([`MR-003`](harness/conventions.md#mr-003--template-verkörperung-nur-für-wiederkehrende-skelette)).
Das Bundle ist derivativ; bei Konflikt sticht die
Quelldatei das Bundle, über ihr die kanonischen Quellen (Source
Precedence). Den Link auf die Quelldatei und den adoptierten Stand führt
[`harness/conventions.md`](harness/conventions.md) (§Adoptierte
Konventions-Quellen bzw. §Baseline).

## 2. Kanonische Quellen (Source Precedence)

In dieser Reihenfolge:

1. [`spec/lastenheft.md`](spec/lastenheft.md) — vertraglich abnahmebindend.
2. [`spec/spezifikation.md`](spec/spezifikation.md) — technisch verbindlich, fortschreibbar (3. Spec-Stratum, `conventions.md` [`MR-001`](harness/conventions.md#mr-001--source-precedence-mit-eigener-spezifikations-schicht)).
3. [`spec/architecture.md`](spec/architecture.md) — Komponenten- und Sequenzsicht.
4. [`docs/plan/adr/`](docs/plan/adr/) — ADR-Verzeichnis und -Index.
5. [`docs/plan/planning/in-progress/roadmap.md`](docs/plan/planning/in-progress/roadmap.md) — aktuelle Welle.
6. [`docs/user/releasing.md`](docs/user/releasing.md) — Operations, Releasing.
7. [`README.md`](README.md) — Projekt-Überblick.
8. **AGENTS.md (diese Datei).**
9. [`harness/README.md`](harness/README.md) — Harness-Einstieg.

## 3. Harte Regeln

### 3.1 Gates laufen über `make`

Quality-Gates werden über `make` ausgeführt; `make doc-check` nutzt ein
auf `:v0.91.0` gepinntes, netzloses Docker-Image (`d-check.mk`, erzeugt mit
`d-check --print-mk`; `DCHECK_DIGEST` für strikten Digest-Pin). **Aber:** Das Werkzeug
`bedrock-eu-check` selbst muss lokal **ohne** Docker und ohne Internet
lauffähig sein ([`LH-NF-001`](spec/lastenheft.md#lh-nf-001--plattform), [`LH-NF-005`](spec/lastenheft.md#lh-nf-005--offline-fähigkeit)) — es gibt also keinen
Docker-only-Zwang für die Anwendung, nur für die reproduzierbaren Gates.

### 3.2 Suppression-Verbot

Inline-Suppression von Lint-/Typecheck-Findings ist verboten; Ausnahmen
leben mit Begründung in einer zentralen Konfiguration. Die konkrete
Marker-Syntax (`# noqa` / `//nolint` / …) wird mit der Sprachwahl in
**[`ADR-0002`](docs/plan/adr/0002-implementierungssprache.md)** festgelegt; bis dahin ist die Regel deklariert, aber noch
ohne Gate gebunden.

### 3.3 git mv + Inhaltsänderung = zwei Commits

Wird eine Datei verschoben **und** ihr Inhalt umgeschrieben:

1. `git mv source target` → eigener Commit (reiner Move, R-Rename).
2. Inhalt umschreiben → zweiter Commit.

**Begründung:** Sonst fällt die Rename-Detection unter die
50 %-Similarity-Schwelle und `git log --follow` wird unzuverlässig.
Gilt insbesondere für Slice-Bewegungen `open → next → in-progress → done`.

### 3.4 Architektur ist sprach- und meilensteinfrei

[`spec/architecture.md`](spec/architecture.md) referenziert ADRs und
Komponenten, aber **keine** Wellen, Slices, Commit-Hashes oder
Closure-Daten. Die zeitliche Schicht lebt in `docs/plan/planning/`.

### 3.5 ADRs sind nach `Accepted` immutable

Eine ADR mit Status `Accepted` wird nicht inhaltlich überschrieben.
Korrekturen entstehen als neue ADR mit `Supersedes ADR-NNNN`.

### 3.6 Gates dürfen nicht ohne ADR gelockert werden

Jede Schwellen-Senkung (Coverage, Linter-Strenge, Architekturregel) ist
ein ADR, kein PR-Kommentar.

### 3.7 Niemals Secret-Werte ausgeben (Domäne)

`bedrock-eu-check` meldet Secrets ausschließlich als „set"/„not set"; der
Wert von `ANTHROPIC_API_KEY`, AWS-Credentials o. Ä. wird nie ausgegeben,
geloggt oder in JSON/SARIF geschrieben ([`LH-NF-004`](spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe),
[`LH-SEC-002`](spec/lastenheft.md#lh-sec-002--keine-speicherung-von-claude-anthropic-secrets)).

### 3.8 Read-only gegenüber AWS, keine Remediation (Domäne)

Version 1 nutzt ausschließlich lesende AWS-Operationen. Keine
automatische Änderung von AWS-Ressourcen oder lokalen Configs
([`LH-SEC-003`](spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff), [`LH-SEC-005`](spec/lastenheft.md#lh-sec-005--keine-automatischen-remediations), [`LH-NZ-003`](spec/lastenheft.md#lh-nz-003--keine-automatische-änderung-produktiver-aws-konfigurationen)).

### 3.9 Allow-/Blocklisten sind konfigurierbar, nicht hartcodiert (Domäne)

Erlaubte Regionen, Accounts, Modell-IDs und Profile sind über
`bedrock-eu-check.yml` steuerbar ([`LH-NF-007`](spec/lastenheft.md#lh-nf-007--konfigurierbarkeit)). Defaults dürfen im
Code stehen, müssen aber überschreibbar bleiben.

## 4. Quality Gates

Nur real existierende Targets (keine halluzinierten Gates, Modul 13):

| Target           | Zweck                                                              |
| ---------------- | ------------------------------------------------------------------ |
| `make doc-check` | Doku-Referenzen prüfen (Links/Anker, via d-check v0.91.0, netzlos) |
| `make lint`      | `gofmt -l .` + `go vet ./...` (containerisiert, netzlos)           |
| `make test`      | Unit-Tests (`go test ./...`, containerisiert, netzlos)             |
| `make build`     | Binary `bin/bedrock-eu-check` bauen                                |
| `make doc-targets` | Doku ↔ Makefile: jedes Target aus `Makefile` steht in diesen Tabellen und umgekehrt (d-check, Modul `targets`, netzlos) |
| `make gates`     | alle lauffähigen Gates: `doc-check`, `doc-targets`, `lint`, `test` |
| `make verify-clean` | `make gates` im frischen `git worktree` des committeten HEAD     |
| `make deps`      | `go mod tidy` + `vendor/` neu erzeugen (mit Netz, nicht Teil der Gates, [`ADR-0004`](docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md)) |
| `make deps-verify` | `go mod verify` + `vendor/` entspricht Upstream (mit Netz; CI-Job nur bei Änderungen an `vendor/`, `go.mod`, `go.sum`) |
| `make self-check` | `bedrock-eu-check local` und `terraform` gegen das eigene Repo (kontrollierte Umgebung, `bedrock-eu-check.yml` im Root); Exit 0 Pflicht |
| `make image`     | Docker-Images `slim` und `full` bauen (mit Netz, nicht Teil der Gates)                 |
| `make image-test` | Images bauen und Smoke-Tests (`tools/image-smoke.sh`, Laufzeit ohne Netz; CI-Job `image`) |
| `make feature-test` | Devcontainer-Feature gegen ein lokal gebautes Release testen (Docker, Laufzeit ohne Netz; CI-Job `image`) |
| `make dist`      | Release-Artefakte (Linux amd64/arm64) und `SHA256SUMS` nach `dist/v<version>/` bauen und verifizieren (`VERSION=x.y.z`); veröffentlicht nichts |

`make coverage-gate` entsteht in einem späteren Slice; es wird erst hier
eingetragen, wenn es läuft.

## 5. Dokumentations-Regeln

- Requirement- und Architektur-IDs müssen in PRs/Commits referenziert
  sein. IDs werden beim Spec-/ADR-Schreiben nach dem in
  [`harness/conventions.md`](harness/conventions.md) [`MR-002`](harness/conventions.md#mr-002--id-schema-kanon-präfix-lh--mit-bereichscodes) deklarierten
  Schema vergeben (`LH-*`, `SPEC-*`, `ARC-*`, ADR-Nummern
  über den Index) — nie ad hoc im PR.
- Neue ADRs müssen den [ADR-Index](docs/plan/adr/README.md) aktualisieren.
- Roadmap/Status-Geschichte lebt in `docs/plan/planning/`, nicht in
  `spec/architecture.md`.
- Release-/Betriebs-Definitionen leben in
  [`docs/user/releasing.md`](docs/user/releasing.md).

## 6. Minimal Agent Workflow

Pro Slice:

1. [`harness/README.md`](harness/README.md) lesen.
2. Relevante kanonische Quelle lesen (Source Precedence beachten).
3. Betroffene `LH-*`/`ADR-*`-IDs identifizieren.
4. Kleinste sinnvolle Änderung planen.
5. Engsten nützlichen Sensor laufen lassen.
6. Repo-weiten Gate-Lauf vor Handoff (`make gates`).
7. Doku/Indizes aktualisieren, falls ein öffentlicher Vertrag berührt ist.
8. Ausgeführte Sensors und verbleibende Risiken berichten — keine
   Erfolgsmeldung ohne Gate-Ausführung.
