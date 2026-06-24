# bedrock-eu-guard

## Was ist bedrock-eu-guard?

`bedrock-eu-guard` liefert **`bedrock-eu-check`** — ein CLI-Werkzeug, das
lokale, AWS-, Bedrock-, Claude-Code-, Devcontainer-, Docker- und
Terraform-Konfigurationen darauf prüft, ob Claude Code **über AWS Bedrock
in einer zugelassenen EU-Region** (Default `eu-central-1` / Frankfurt)
betrieben wird — und nicht versehentlich über die direkte Anthropic-API,
ein US-/Global-Inference-Profil oder eine fehlende Region.

Es richtet sich an Softwareentwicklung, DevOps, Cloud Engineering,
Datenschutz und Security und arbeitet gegen die Annahme „die Region passt
schon" — die in der Praxis oft nicht stimmt.

## Was kann ich heute tun?

Ehrlicher Ist-Stand: Das Repo ist **frisch ge-bootstrappt** (Greenfield-
Harness nach dem [ai-harness-course](https://github.com/pt9912/ai-harness-course) v1.4.0).
Es gibt **noch keinen Anwendungscode**. Die Gates brauchen nur **Docker**
und GNU `make`:

- `make help` — verfügbare Targets anzeigen.
- `make doc-check` — Doku-Referenzen prüfen (Links, Anker, ID-Linkpflicht, Referenzmatrix; via d-check v0.29.0, netzlos im Container).
- `make gates` — alle aktuell lauffähigen Gates (heute nur `doc-check`).

Die erste lauffähige Tool-Version (`bedrock-eu-check local` mit
Env-/Modell-ID-Checks und Exit-Codes) ist als **Welle 1** geplant, siehe
[Roadmap](docs/plan/planning/in-progress/roadmap.md). Die Architektur ist
**hexagonal** ([ADR-0003](docs/plan/adr/0003-hexagonale-architektur.md),
Vorbild [d-check](https://github.com/pt9912/d-check)); die Sprachwahl ist
in [ADR-0002](docs/plan/adr/0002-implementierungssprache.md)
(Status *Proposed*, Go empfohlen) offen.

## Geplante Nutzung

> Zielbild (noch nicht implementiert — ab Welle 1):

```bash
bedrock-eu-check local      # Env, Claude Code, Devcontainer/Docker prüfen
bedrock-eu-check aws        # STS-Identität, Bedrock-Profile, IAM (read-only)
bedrock-eu-check terraform  # .tf: Provider-Region, IAM-Policies, Modell-IDs
bedrock-eu-check all --ci   # alles in einem Lauf, CI-Ausgabe, Exit 1 bei FAIL
```

## Warum bedrock-eu-guard?

Durch falsche Umgebungsvariablen, AWS-Profile, Claude-Code-Settings,
Inference-Profile oder Terraform/IAM-Definitionen kann unbeabsichtigt eine
US-Region oder ein globales Profil verwendet werden. Manuelle Kontrolle
ist fehleranfällig und skaliert nicht über mehrere Entwickler,
Devcontainer, CI-Pipelines und AWS-Accounts ([`LH-AUSG-003`](spec/lastenheft.md#lh-ausg-003--fehlkonfigurationsrisiko)/[`LH-AUSG-004`](spec/lastenheft.md#lh-ausg-004--manuelle-prüfung-reicht-nicht-aus)).
`bedrock-eu-check` macht diese Prüfung automatisierbar und CI-tauglich.

## Kerngedanke

**Fehlkonfiguration früh und deterministisch sichtbar machen — `PASS` /
`WARN` / `FAIL` mit klarer Kennung, Exit-Code `1` bei jedem `FAIL`, ohne je
ein Secret auszugeben.** Jede Prüfung hängt an einer `LH-*`-ID und ist
über `bedrock-eu-check.yml` konfigurierbar.

## Was macht es vertrauenswürdig?

- **Prozess:** [`AGENTS.md`](AGENTS.md) (Hard Rules), [`harness/README.md`](harness/README.md) (Source Precedence, Gates).
- **Verträge:** [`spec/lastenheft.md`](spec/lastenheft.md) (`LH-*`-IDs mit Akzeptanzkriterien).
- **Gates:** `make doc-check` (Doku-Referenz-Gate); Code-Gates folgen mit Welle 1 (`LH-QA-*`).
- **Auditierbarkeit:** Entscheidungen in [`docs/plan/adr/`](docs/plan/adr/), Planung in [`docs/plan/planning/`](docs/plan/planning/).

> **Grenzen.** Das Tool ist **keine** Rechts-/Datenschutzberatung und kein
> Beweis über AWS-internes Routing; es prüft nur sichtbare, steuerbare
> Konfigurationen ([`LH-NZ-001`](spec/lastenheft.md#lh-nz-001--kein-rechtlicher-nachweis)/[`LH-NZ-002`](spec/lastenheft.md#lh-nz-002--keine-garantie-interner-aws-routings), [`LH-RISK-004`](spec/lastenheft.md#lh-risk-004--scheinbare-sicherheit)).

## Lizenz

[MIT](LICENSE) — Copyright (c) 2026 pt9912.
