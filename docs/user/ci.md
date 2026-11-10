# Einbindung in CI

`bedrock-eu-check` lässt sich in Pipelines einsetzen
([`LH-MOD-005`](../../spec/lastenheft.md#lh-mod-005--ci-modus)). Diese Seite beschreibt die
Optionen und einen Beispiel-Workflow für GitHub Code Scanning. Das Werkzeug lädt
nichts hoch; das Hochladen übernimmt der Workflow.

## Optionen und Exit-Codes

| Option | Wirkung |
|---|---|
| `--format console\|json\|sarif` | Ausgabeformat auf stdout (Default `console`) |
| `--ci` | stdout trägt weiterhin nur das Ergebnis; zusätzlich steht eine Zusammenfassung auf **stderr**, damit das Pipeline-Log auch bei umgeleitetem stdout den Ausgang zeigt |
| `--root DIR` | Wurzel der zu prüfenden Dateien (Default: aktuelles Verzeichnis) |
| `--config FILE` | Konfigurationsdatei (Default: `<root>/bedrock-eu-check.yml`) |

| Exit | Bedeutung |
|---|---|
| `0` | kein `FAIL` ([`LH-FA-EXIT-001`](../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung)) |
| `1` | mindestens ein `FAIL` ([`LH-FA-EXIT-002`](../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung)) |
| `2` | Tool-Fehler, z. B. ungültige Konfiguration, unbekannte Option oder fehlendes `aws`-Binary ([`LH-FA-EXIT-003`](../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler)) |

`--ci` ändert die Exit-Codes nicht: `WARN` lässt den Lauf nicht scheitern. Ein
Schwellenwert wie `--fail-on warn` ist ein Folgepunkt.

Bei einem Tool-Fehler bleibt stdout trotzdem im gewählten Format parsebar
(bereits ermittelte Findings), die Fehlermeldung steht auf stderr.

## Welcher Modus in welcher Pipeline

| Modus | Braucht | Geeignet für |
|---|---|---|
| `local` | nichts außer dem Repository | alle Pipelines, auch Pull Requests aus Forks |
| `terraform` | nichts außer dem Repository | alle Pipelines |
| `aws` | `aws`-CLI und Zugangsdaten ([`aws-permissions.md`](aws-permissions.md)) | nur Pipelines mit AWS-Zugang |
| `all` | wie `aws` | Pipelines mit AWS-Zugang |

Ohne `aws`-Binary meldet `aws` bzw. `all` einen Tool-Fehler (Exit `2`); ohne
Zugangsdaten ein `FAIL` (Exit `1`).

## Beispiel: GitHub Code Scanning (SARIF)

Der Workflow baut das Werkzeug aus dem Repository, führt die offline
Prüfungen aus, lädt das SARIF hoch und lässt den Job erst danach scheitern,
falls es Fehlfunde gab. Versionen der Actions sind Beispiele; für
reproduzierbare Pipelines die Actions auf einen Commit-SHA pinnen.

```yaml
name: bedrock-eu-check

on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read
  security-events: write   # nötig für den SARIF-Upload

jobs:
  eu-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Werkzeug bauen
        run: make build

      - name: Prüfen (local)
        id: scan_local
        continue-on-error: true   # Upload soll auch bei Fehlfunden laufen
        run: ./bin/bedrock-eu-check local --ci --format sarif > local.sarif

      - name: Prüfen (terraform)
        id: scan_terraform
        continue-on-error: true
        run: ./bin/bedrock-eu-check terraform --ci --format sarif > terraform.sarif

      - name: SARIF hochladen (local)
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: local.sarif
          category: bedrock-eu-check-local

      - name: SARIF hochladen (terraform)
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: terraform.sarif
          category: bedrock-eu-check-terraform

      - name: Bei Fehlfunden scheitern
        if: steps.scan_local.outcome == 'failure' || steps.scan_terraform.outcome == 'failure'
        run: exit 1
```

Hinweise:

- **Zwei Schritte, zwei Dateien:** Jeder Lauf schreibt ein eigenes SARIF; die
  `category` hält die Alerts getrennt. Zwei getrennte Schritte sorgen dafür,
  dass auch nach einem Fehlfund beide Dateien entstehen (GitHub bricht einen
  Schritt mit mehreren Befehlen beim ersten Fehler ab).
- **Findings ohne Fundstelle** (Umgebung, AWS, Dateien außerhalb des
  Repositorys wie `~/.claude/settings.json`) erhalten in SARIF eine als Ersatz
  gekennzeichnete Position (`bedrock-eu-check.yml`, sonst `README.md`, Zeile 1;
  `properties.synthetic` ist `true`). Code Scanning zeigt sie daher an dieser
  Datei an.
- **Keine Zugangsdaten in der Ausgabe:** Weder Credentials noch Werte aus
  Konfigurationsdateien erscheinen in einem Format
  ([`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
- **Nicht getestet:** Der Workflow lässt sich offline nicht ausführen; er folgt
  den dokumentierten Aufrufen, wurde aber nicht gegen GitHub erprobt.
