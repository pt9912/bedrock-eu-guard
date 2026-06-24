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
| 2 | [`spec/spezifikation.md`](../spec/spezifikation.md) | technisch fortschreibbar (3. Spec-Stratum, siehe `conventions.md` MR-001) |
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
| `make doc-check` | Doku-Referenzen lösen auf (Links + Anker; via d-check, netzlos `--network none`) | Reproduzierbarkeit (`:v0.29.0` in `d-check.mk`; `DCHECK_DIGEST` für Digest-Pin) |
| `make gates` | alle aktuell lauffähigen Gates (heute nur `doc-check`) | — |

**Aktueller Lauf-Status:** lokal `make gates` / `make help`; in CI das
Pipeline-Ergebnis (Rang höher als diese Datei).
**Nicht behauptet** (geplant, entstehen mit dem ersten Code-Slice in
Welle 1 und ADR-0002): `make lint`, `make test`, `make coverage-gate`
(bootstrap-aware), `make self-check` (`bedrock-eu-check` gegen das eigene
Repo, `LH-QA-005`). Diese Targets existieren **noch nicht** im
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
  Zusicherung (`LH-NZ-001`, `LH-RISK-004`).
- **Kein Beweis über AWS-internes Routing.** Geprüft werden nur sichtbare,
  steuerbare Konfigurationen (`LH-NZ-002`).
- **Niemals Secrets im Klartext ausgeben.** `ANTHROPIC_API_KEY` &
  Co. werden nur als „set"/„not set" gemeldet (`LH-NF-004`,
  `LH-SEC-002`).
- **Read-only gegenüber AWS.** Version 1 nutzt ausschließlich lesende
  AWS-Operationen, keine automatischen Remediations
  (`LH-SEC-003`, `LH-SEC-005`, `LH-NZ-003`).
- **Offline-Pfad muss offline bleiben.** Lokale und Terraform-Prüfungen
  funktionieren ohne Internet/AWS (`LH-NF-005`).

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
