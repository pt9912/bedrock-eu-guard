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

## Stand

`bedrock-eu-check` ist **funktionsfähig**; **Version 1.1.0** ist veröffentlicht
(Commit `5274f3dce55f`): als [GitHub-Release](https://github.com/pt9912/bedrock-eu-guard/releases/tag/v1.1.0)
mit Linux-Binaries und `SHA256SUMS` und als Docker-Images `ghcr.io/pt9912/bedrock-eu-check:1.1.0-slim`
und `:1.1.0-full`. Neu gegenüber 1.0.0 sind `--fail-on warn`, `--output FILE` und die Konfiguration
`visibilityRegions` (siehe [`docs/user/reference.md`](docs/user/reference.md)); Exit-Codes und
Standardverhalten sind unverändert. Das Devcontainer-Feature liegt in der Version 1.0.1 vor
(Tag `feature-v1.0.1`, Commit `5274f3dce55f`); es installiert standardmäßig weiterhin das Werkzeug 1.0.0.
Version 1.0.0 (Commit `5f1b35cca776`) bleibt unverändert abrufbar. Alle Prüfmodi, die Ausgabeformate
(Konsole, JSON, SARIF) und der CI-Modus sind implementiert und getestet. Der Ablauf steht in
[`docs/user/releasing.md`](docs/user/releasing.md).

Nicht auf externen Plattformen erprobt ist der Beispiel-Workflow
für GitHub Code Scanning; die Beispiel-IAM-Policy wurde nicht gegen ein echtes AWS-Konto getestet.
Der Release-Ablauf (`release-create`, `publish`) ist scharf durchlaufen. Das Devcontainer-Feature
ist mit der Devcontainer-CLI 0.90.0 erprobt (`devcontainer up` und `exec`); ältere CLI-Versionen (0.72.0)
hängen mit `up` auf GitHub-Runnern mit Docker 29.4.2.
Den Fahrplan zeigt die [Roadmap](docs/plan/planning/in-progress/roadmap.md).

## Schnellstart

Voraussetzung sind GNU `make` und Docker (die Go-Toolchain läuft in einem gepinnten
Container, ohne Netz):

```bash
make build                          # erzeugt bin/bedrock-eu-check
./bin/bedrock-eu-check version

./bin/bedrock-eu-check local        # Umgebung, .env, Claude-Settings, Devcontainer, Compose, Dockerfile
./bin/bedrock-eu-check terraform    # Terraform-Dateien und IAM-Policy-Dateien (rekursiv)
```

Beispielausgabe:

```text
[FAIL] LH-FA-LENV-001 AWS_REGION: AWS_REGION=us-east-1 ist keine erlaubte Region (REGION_NOT_ALLOWED)
[PASS] LH-FA-LENV-003 CLAUDE_CODE_USE_BEDROCK: CLAUDE_CODE_USE_BEDROCK=1 (OK)
Summary: 1 PASS, 0 WARN, 1 FAIL
```

Weiter geht es im [Schnellstart](docs/user/quickstart.md); für typische Aufgaben gibt es
das [Handbuch](docs/user/handbuch.md).

## Befehle

| Befehl | Prüft | Netz |
|---|---|---|
| `bedrock-eu-check local` | Umgebungsvariablen, `.env`-Dateien, Claude-Code-Settings, Devcontainer, Compose, Dockerfile | nein |
| `bedrock-eu-check terraform` | Provider-Region, IAM-Policies, Modell-IDs in `.tf`-Dateien; eigenständige IAM-Policy-Dateien (`iam-policy*.json`, `*-policy.json`) | nein |
| `bedrock-eu-check aws` | STS-Identität, Bedrock-Profile, IAM-Rechte (nur lesend, über die AWS-CLI) | ja |
| `bedrock-eu-check all` | `local` + `terraform` + `aws` | ja |
| `bedrock-eu-check version` | Version und Commit | nein |

Wichtige Optionen: `--root DIR`, `--config FILE`, `--format console|json|sarif`, `--ci`,
`--allowed-account ID`, `--fail-on fail|warn`, `--output FILE`. Exit-Codes: `0` kein `FAIL`, `1` mindestens ein `FAIL`
(mit `--fail-on warn` auch bei `WARN`), `2` Tool-Fehler.
Alle Details stehen in der [Referenz](docs/user/reference.md); jede Prüfung ist im
[Regelkatalog](docs/user/rules.md) beschrieben.

Beispiel für CI (SARIF für Code Scanning, Zusammenfassung auf stderr):

```bash
bedrock-eu-check terraform --ci --format sarif --output terraform.sarif
```

## Konfiguration

Ohne Konfiguration gelten sichere Defaults (nur `eu-central-1`). Abweichungen und
Ausnahmen stehen in `bedrock-eu-check.yml` im Projektstamm:

```yaml
allowedRegions: [eu-central-1, eu-west-1]
exclude:
  - testdata        # absichtlich fehlerhafte Beispiele nicht prüfen
```

Alle Felder und Defaults: [Referenz](docs/user/reference.md#konfigurationsdatei-bedrock-eu-checkyml).

## Bereitstellung

| Weg | Stand | Dokumentation |
|---|---|---|
| Binary (Linux `amd64`/`arm64`, `make dist VERSION=x.y.z`) | [GitHub-Release v1.1.0](https://github.com/pt9912/bedrock-eu-guard/releases/tag/v1.1.0) | [`releasing.md`](docs/user/releasing.md) |
| Docker-Image `slim` (`local`, `terraform`) und `full` (mit AWS-CLI, alle Modi) | `ghcr.io/pt9912/bedrock-eu-check:1.1.0-slim` / `:1.1.0-full` (lokal: `make image`) | [`docker.md`](docs/user/docker.md) |
| Devcontainer-Feature | `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check:1` (Version 1.0.1) | [`devcontainer.md`](docs/user/devcontainer.md) |
| CI-Einbindung | Beispiel-Workflow | [`ci.md`](docs/user/ci.md) |

## Dokumentation

Nutzerdokumentation unter [`docs/user/`](docs/user/):
[Handbuch](docs/user/handbuch.md) · [Schnellstart](docs/user/quickstart.md) ·
[Referenz](docs/user/reference.md) · [Regelkatalog](docs/user/rules.md) ·
[CI](docs/user/ci.md) · [Docker](docs/user/docker.md) ·
[Devcontainer](docs/user/devcontainer.md) · [AWS-Rechte](docs/user/aws-permissions.md) ·
[Fehlersuche](docs/user/troubleshooting.md) · [Release-Ablauf](docs/user/releasing.md).

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

## Entwicklung

Die Architektur ist **hexagonal** ([ADR-0003](docs/plan/adr/0003-hexagonale-architektur.md),
Vorbild [d-check](https://github.com/pt9912/d-check)); Implementierungssprache ist Go
([ADR-0002](docs/plan/adr/0002-implementierungssprache.md)). Abhängigkeiten liegen unter
`vendor/` ([ADR-0004](docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md)); die
Gates laufen netzlos in gepinnten Containern.

```bash
make help          # alle Targets
make gates         # doc-check, doc-targets, lint, test, coverage-gate (netzlos)
make self-check    # bedrock-eu-check gegen das eigene Repository
make verify-clean  # make gates im frischen Worktree des committeten Stands
```

Weitere Targets mit Netz oder Docker: `make image-test`, `make feature-test`,
`make deps-verify`. Die Gates im Überblick: [`harness/README.md`](harness/README.md).
Lizenzen der Abhängigkeiten: [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

## Was macht es vertrauenswürdig?

- **Prozess:** [`AGENTS.md`](AGENTS.md) (Hard Rules), [`harness/README.md`](harness/README.md) (Source Precedence, Gates).
- **Verträge:** [`spec/lastenheft.md`](spec/lastenheft.md) (`LH-*`-IDs mit Akzeptanzkriterien), [`spec/spezifikation.md`](spec/spezifikation.md), [`spec/architecture.md`](spec/architecture.md).
- **Gates:** Doku-Referenzen (`make doc-check`), Lint, Unit-Tests inklusive Architekturprüfung und Abgleich Doku ↔ Code, Self-Check; in CI bei jedem Pull Request ([`LH-QA-001`](spec/lastenheft.md#lh-qa-001--unit-tests)/[`LH-QA-005`](spec/lastenheft.md#lh-qa-005--ci-prüfung)).
- **Auditierbarkeit:** Entscheidungen in [`docs/plan/adr/`](docs/plan/adr/), Planung und Abschlussnotizen in [`docs/plan/planning/`](docs/plan/planning/).

> **Grenzen.** Das Tool ist **keine** Rechts-/Datenschutzberatung und kein
> Beweis über AWS-internes Routing; es prüft nur sichtbare, steuerbare
> Konfigurationen ([`LH-NZ-001`](spec/lastenheft.md#lh-nz-001--kein-rechtlicher-nachweis)/[`LH-NZ-002`](spec/lastenheft.md#lh-nz-002--keine-garantie-interner-aws-routings), [`LH-RISK-004`](spec/lastenheft.md#lh-risk-004--scheinbare-sicherheit)).

## Lizenz

[MIT](LICENSE) — Copyright (c) 2026 pt9912.
