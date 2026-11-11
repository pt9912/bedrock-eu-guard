# Referenz — Befehle, Optionen, Konfiguration

**Status:** Stand v1.0-Kandidat. Dieses Dokument trägt keine eigenen Anforderungen
(Lastenheft: [`LH-TECH-002`](../../spec/lastenheft.md#lh-tech-002--cli-struktur), [`LH-TECH-003`](../../spec/lastenheft.md#lh-tech-003--konfigurationsdatei); Schema:
[`SPEC-020`](../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml)).

## Befehle

| Befehl | Prüft | Netz |
|---|---|---|
| `bedrock-eu-check local` | Umgebungsvariablen, `.env`-Dateien (rekursiv), Claude-Code-Settings, Devcontainer, Compose, Dockerfile | nein |
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
| `--allowed-account ID` | `aws`, `all` (wirkt nur dort; das Format wird in allen Modi geprüft) | erlaubte Account-ID (12 Ziffern); mehrfach möglich, wird mit `allowedAccounts` vereint; ungültiges Format ist ein Tool-Fehler (Exit `2`), auch bei `local` |

## Exit-Codes

| Exit | Bedeutung |
|---|---|
| `0` | kein `FAIL`-Finding |
| `1` | mindestens ein `FAIL` |
| `2` | Tool-Fehler: ungültige Argumente, ungültige oder nicht lesbare eigene Konfiguration, fehlendes `aws`-Binary (Grund `AWS_CLI_MISSING`) |

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

## Fehlermeldungen der Konfiguration

Eine ungültige eigene Konfiguration ist ein Tool-Fehler (Exit `2`). Die Meldung
steht auf stderr als `Konfiguration: <feld>[<index>]: <GRUND>` und nennt nie den
Feldwert; bei Syntaxfehlern der Datei als `FILE_INVALID` mit Zeile.

| Grund | Bedeutung |
|---|---|
| `EMPTY_REGIONS` | `allowedRegions` ist leer |
| `INVALID_ACCOUNT` | Account-ID hat nicht genau 12 Ziffern |
| `INVALID_REGEX` | ungültiger regulärer Ausdruck in `allowedModelPatterns`, `blockedModelPatterns` oder `allowedProfiles` |
| `INVALID_PATTERN` | ungültiges Glob-Muster in `exclude` |
| `INVALID_VALUE` | leerer oder unzulässiger Wert (z. B. `mantle` außerhalb von `warn`/`allow`/`block`, leerer Regionsname, leeres `exclude`-Muster) |
| `FILE_INVALID` | Datei nicht als YAML lesbar, unbekanntes Feld oder falscher Typ (mit Zeile) |

## Ausgabeformate

- `console`: eine Zeile je Finding (`PASS|WARN|FAIL`, Regel-ID, Meldung, ggf. `Datei:Zeile`).
- `json`: `{"version": 1, "findings": [...]}` nach [`SPEC-021`](../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) mit `status`, `id`, `message`, `severity`, `reason`, `file`, `line`, `recommendation`.
- `sarif`: SARIF 2.1.0 nach [`SPEC-022`](../../spec/spezifikation.md#spec-022--sarif-abbildung---format-sarif); `PASS` erzeugt kein Result; Findings ohne Datei erhalten eine Ersatzposition mit `properties.synthetic` (siehe `ci.md`).
