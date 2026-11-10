# Referenz — Befehle, Optionen, Konfiguration

**Status:** Stand v1.0-Kandidat. Dieses Dokument trägt keine eigenen Anforderungen
(Lastenheft: [`LH-TECH-002`](../../spec/lastenheft.md#lh-tech-002--cli-struktur), [`LH-TECH-003`](../../spec/lastenheft.md#lh-tech-003--konfigurationsdatei); Schema:
[`SPEC-020`](../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml)).

## Befehle

| Befehl | Prüft | Netz |
|---|---|---|
| `bedrock-eu-check local` | Umgebungsvariablen, Claude-Code-Settings, Devcontainer, Compose, Dockerfile | nein |
| `bedrock-eu-check terraform` | Terraform-Dateien (rekursiv) | nein |
| `bedrock-eu-check aws` | STS-Identität, Bedrock-Profile, IAM-Rechte (lesend, über das `aws`-Binary) | ja |
| `bedrock-eu-check all` | `local` + `terraform` + `aws` | ja |
| `bedrock-eu-check version` | gibt Version und Commit aus | nein |

Die Regeln je Modus stehen in `rules.md`.

## Optionen

| Option | Gilt für | Bedeutung |
|---|---|---|
| `--root DIR` | alle Prüfbefehle | Scan-Root für die Datei-Prüfungen (Default: aktuelles Verzeichnis) |
| `--config FILE` | alle Prüfbefehle | Konfigurationsdatei; muss existieren. Ohne Option wird `<root>/bedrock-eu-check.yml` gelesen, falls vorhanden |
| `--format console\|json\|sarif` | alle Prüfbefehle | Ausgabeformat (Default `console`); stdout trägt nur das Ergebnis |
| `--ci` | alle Prüfbefehle | CI-Modus: schreibt zusätzlich eine Zusammenfassung auf stderr; Format und Exit-Codes bleiben unverändert (siehe `ci.md`) |
| `--allowed-account ID` | `aws`, `all` | erlaubte Account-ID (12 Ziffern); mehrfach möglich, wird mit `allowedAccounts` vereint |

## Exit-Codes

| Exit | Bedeutung |
|---|---|
| `0` | kein `FAIL`-Finding |
| `1` | mindestens ein `FAIL` |
| `2` | Tool-Fehler: ungültige Argumente, ungültige oder nicht lesbare eigene Konfiguration, fehlendes `aws`-Binary |

## Konfigurationsdatei `bedrock-eu-check.yml`

Alle Felder sind optional; fehlende Felder fallen auf die Defaults zurück. Unbekannte
Felder oder falsche Typen sind ein Fehler (Exit `2`).

| Feld | Default | Bedeutung |
|---|---|---|
| `allowedRegions` | `[eu-central-1]` | erlaubte AWS-Regionen (nicht leer) |
| `allowedAccounts` | `[]` | erlaubte Account-IDs (12 Ziffern); leer = nicht geprüft |
| `allowedModelPatterns` | `['^eu\.', '^anthropic\.claude-sonnet-4-6$', '^anthropic\.claude-haiku-4-5$']` | reguläre Ausdrücke für erlaubte Modell-/Profil-IDs |
| `blockedModelPatterns` | `['^us\.', '^global\.', 'us-east-1', 'us-west-2']` | reguläre Ausdrücke für blockierte IDs; haben Vorrang |
| `allowedProfiles` | `['eu.*', 'eu-central-1.*']` | Allowlist für verwendete Profile (verankerte reguläre Ausdrücke) |
| `mantle` | `warn` | Umgang mit `CLAUDE_CODE_USE_MANTLE=1`: `warn`, `allow` oder `block` |
| `exclude` | `[]` | Glob-Muster relativ zum Scan-Root, die von der Datei-Discovery ausgenommen werden (`path.Match`, ein Muster trifft auch alles unterhalb eines passenden Verzeichnisses, kein `**`) |

Beispiel:

```yaml
allowedRegions: [eu-central-1, eu-west-1]
allowedAccounts: ["123456789012"]
mantle: block
exclude:
  - testdata
  - vendor
```

## Ausgabeformate

- `console`: eine Zeile je Finding (`PASS|WARN|FAIL`, Regel-ID, Meldung, ggf. `Datei:Zeile`).
- `json`: `{"version": 1, "findings": [...]}` nach [`SPEC-021`](../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) mit `status`, `id`, `message`, `severity`, `reason`, `file`, `line`, `recommendation`.
- `sarif`: SARIF 2.1.0 nach [`SPEC-022`](../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif); `PASS` erzeugt kein Result; Findings ohne Datei erhalten eine Ersatzposition mit `properties.synthetic` (siehe `ci.md`).
