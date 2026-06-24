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

Das **Regelwerk der adoptierten Baseline** liegt als self-navigierbares
**Modul-Bundle (ZIP)** unter
<https://github.com/pt9912/ai-harness-course/releases/download/v1.4.0/lab-regelwerk.zip>
— einmal pro Session die relevanten Module lesen, bevor der Workflow (§6)
startet. Derivativ: bei Konflikt gelten die kanonischen Quellen;
adoptierter Stand steht in
[`harness/conventions.md`](harness/conventions.md) §Baseline.

Die **Skelett-Vorlagen** der Baseline liegen als ZIP unter
<https://github.com/pt9912/ai-harness-course/releases/download/v1.4.0/lab-templates.zip>;
die wiederkehrenden Varianten (ADR, Slice, Welle, Carveout, Review-Report)
sind als `*.template.md` co-located im Repo. Beim Anlegen neuer Artefakte
das passende Template kopieren statt frei zu formulieren.

## 2. Kanonische Quellen (Source Precedence)

In dieser Reihenfolge:

1. [`spec/lastenheft.md`](spec/lastenheft.md) — vertraglich abnahmebindend.
2. [`spec/spezifikation.md`](spec/spezifikation.md) — technisch verbindlich, fortschreibbar (3. Spec-Stratum, `conventions.md` MR-001).
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
auf `:v0.29.0` gepinntes, netzloses Docker-Image (`d-check.mk`, erzeugt mit
`d-check --print-mk`; `DCHECK_DIGEST` für strikten Digest-Pin). **Aber:** Das Werkzeug
`bedrock-eu-check` selbst muss lokal **ohne** Docker und ohne Internet
lauffähig sein (`LH-NF-001`, `LH-NF-005`) — es gibt also keinen
Docker-only-Zwang für die Anwendung, nur für die reproduzierbaren Gates.

### 3.2 Suppression-Verbot

Inline-Suppression von Lint-/Typecheck-Findings ist verboten; Ausnahmen
leben mit Begründung in einer zentralen Konfiguration. Die konkrete
Marker-Syntax (`# noqa` / `//nolint` / …) wird mit der Sprachwahl in
**ADR-0002** festgelegt; bis dahin ist die Regel deklariert, aber noch
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
geloggt oder in JSON/SARIF geschrieben (`LH-NF-004`,
`LH-SEC-002`).

### 3.8 Read-only gegenüber AWS, keine Remediation (Domäne)

Version 1 nutzt ausschließlich lesende AWS-Operationen. Keine
automatische Änderung von AWS-Ressourcen oder lokalen Configs
(`LH-SEC-003`, `LH-SEC-005`, `LH-NZ-003`).

### 3.9 Allow-/Blocklisten sind konfigurierbar, nicht hartcodiert (Domäne)

Erlaubte Regionen, Accounts, Modell-IDs und Profile sind über
`bedrock-eu-check.yml` steuerbar (`LH-NF-007`). Defaults dürfen im
Code stehen, müssen aber überschreibbar bleiben.

## 4. Quality Gates

Nur real existierende Targets (keine halluzinierten Gates, Modul 13):

| Target | Zweck |
|---|---|
| `make doc-check` | Doku-Referenzen prüfen (Links/Anker, via d-check v0.29.0, netzlos) |
| `make gates` | alle aktuell lauffähigen Gates (heute nur `doc-check`) |

`make lint`, `make test`, `make coverage-gate`, `make self-check`
entstehen mit dem ersten Code-Slice (Welle 1) und **ADR-0002**; sie
werden erst hier eingetragen, wenn sie laufen.

## 5. Dokumentations-Regeln

- Requirement- und Architektur-IDs müssen in PRs/Commits referenziert
  sein. IDs werden beim Spec-/ADR-Schreiben nach dem in
  [`harness/conventions.md`](harness/conventions.md) MR-002 deklarierten
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
