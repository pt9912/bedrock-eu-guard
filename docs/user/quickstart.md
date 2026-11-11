# Schnellstart

`bedrock-eu-check` prüft, ob Claude Code, Container-Dateien und Terraform für
Amazon Bedrock in einer EU-Region konfiguriert sind, ohne dass Daten über
`us.*`-Profile oder die direkte Anthropic-API abfließen. Dieser Schnellstart führt
vom Installieren bis zum ersten Ergebnis. Details stehen in der
[Referenz](reference.md) und im [Regelkatalog](rules.md).

## 1. Installieren

Es gibt noch kein veröffentlichtes Release ([`releasing.md`](releasing.md)); bis dahin
wird aus dem Quellcode gebaut. Voraussetzung ist `make` und Docker (die Go-Toolchain
läuft in einem gepinnten Container, ohne Netz):

```bash
make build                    # erzeugt bin/bedrock-eu-check
./bin/bedrock-eu-check version
```

Mit einer lokalen Go-Installation (Version laut `go.mod`) geht es auch ohne Docker:
`go build -mod=vendor -o bin/bedrock-eu-check ./cmd/bedrock-eu-check`.

Alternativen: das [Docker-Image](docker.md) oder das [Devcontainer-Feature](devcontainer.md).

## 2. Erste Prüfung

Im Wurzelverzeichnis des zu prüfenden Projekts:

```bash
bedrock-eu-check local        # Umgebung, .env, Claude-Settings, Devcontainer, Compose, Dockerfile
bedrock-eu-check terraform    # Terraform-Dateien und IAM-Policy-Dateien (rekursiv)
```

Beide Modi arbeiten offline. Die Ausgabe hat eine Zeile je Befund:

```text
[FAIL] LH-FA-LENV-001 AWS_REGION: AWS_REGION=us-east-1 ist keine erlaubte Region (REGION_NOT_ALLOWED)
[PASS] LH-FA-LENV-003 CLAUDE_CODE_USE_BEDROCK: CLAUDE_CODE_USE_BEDROCK=1 (OK)
[WARN] LH-FA-CLAUDE-001 ~/.claude/settings.json: Benutzer-Settings nicht vorhanden (FILE_MISSING)
Summary: 1 PASS, 1 WARN, 1 FAIL
```

- `PASS`: in Ordnung. `WARN`: nicht belegbar oder zu prüfen. `FAIL`: nicht EU-konform.
- Hinter dem Befund steht der Reason-Code; was er bedeutet, steht im [Regelkatalog](rules.md).
- Befunde mit Datei zeigen `[Datei:Zeile]`.
- Es werden nie Secret-Werte ausgegeben, nur „set“ oder „not set“.

**Exit-Codes:** `0` kein `FAIL`, `1` mindestens ein `FAIL` (mit `--fail-on warn` auch bei `WARN`), `2` Tool-Fehler
(z. B. ungültige Konfiguration). Eine Übersicht steht in der [Referenz](reference.md).

## 3. AWS-Prüfung (optional)

Der Modus `aws` fragt AWS lesend ab (Identität, Bedrock-Profile, IAM-Rechte). Er braucht
die AWS-CLI und Zugangsdaten; die nötigen Rechte stehen in
[`aws-permissions.md`](aws-permissions.md).

```bash
bedrock-eu-check aws --allowed-account 123456789012
bedrock-eu-check all          # local + terraform + aws
```

## 4. Konfigurieren

Ohne Konfiguration gelten sichere Defaults (nur `eu-central-1`). Abweichungen stehen
in `bedrock-eu-check.yml` im Projektstamm, zum Beispiel:

```yaml
allowedRegions: [eu-central-1, eu-west-1]
exclude:
  - testdata        # absichtlich fehlerhafte Beispiele nicht prüfen
```

Alle Felder und Defaults: [Referenz](reference.md#konfigurationsdatei-bedrock-eu-checkyml).

## 5. In CI einsetzen

```bash
bedrock-eu-check local --ci --format sarif > local.sarif
```

Ein vollständiger Workflow für GitHub Code Scanning steht in [`ci.md`](ci.md).

## Bei Problemen

[`troubleshooting.md`](troubleshooting.md) listet typische Meldungen und ihre Ursachen.
