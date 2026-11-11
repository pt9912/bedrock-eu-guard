# Handbuch

Dieses Handbuch ist nach **Aufgaben** gegliedert: Suchen Sie die Aufgabe, die Sie
erledigen wollen, und folgen Sie den Schritten. Nachschlagewerke sind verlinkt statt
wiederholt: die [Referenz](reference.md) (Befehle, Optionen, Konfiguration), der
[Regelkatalog](rules.md) (jede Prüfung) und die [Fehlersuche](troubleshooting.md).

`bedrock-eu-check` prüft, ob Claude Code über **Amazon Bedrock in einer zugelassenen
EU-Region** betrieben wird (Default `eu-central-1`) und nicht über die direkte
Anthropic-API, ein US-/Global-Inference-Profil oder eine fehlende Region. Es **liest nur**
und gibt **nie Secret-Werte** aus.

> **Stand:** Version 1.0.0 ist veröffentlicht (Binary, Docker-Images, Devcontainer-Feature, siehe
> [`README.md`](../../README.md)). Alternativ bauen Sie das Werkzeug selbst (Aufgabe 1). In den Beispielen steht `bedrock-eu-check` für das
> mit `make build` erzeugte Binary bzw. das Binary in Ihrem `PATH`.

## Inhalt

1. [Das Werkzeug bereitstellen](#aufgabe-1-das-werkzeug-bereitstellen)
2. [Mein Projekt prüfen](#aufgabe-2-mein-projekt-prüfen)
3. [Einen Befund verstehen und beheben](#aufgabe-3-einen-befund-verstehen-und-beheben)
4. [Terraform prüfen](#aufgabe-4-terraform-prüfen)
5. [Das AWS-Konto und die Berechtigungen prüfen](#aufgabe-5-das-aws-konto-und-die-berechtigungen-prüfen)
6. [Alles in einem Lauf prüfen](#aufgabe-6-alles-in-einem-lauf-prüfen)
7. [Die Regeln an meine Organisation anpassen](#aufgabe-7-die-regeln-an-meine-organisation-anpassen)
8. [Ergebnisse weiterverarbeiten](#aufgabe-8-ergebnisse-weiterverarbeiten)
9. [In CI einbinden](#aufgabe-9-in-ci-einbinden)
10. [Im Container oder Devcontainer nutzen](#aufgabe-10-im-container-oder-devcontainer-nutzen)
11. [Ein Problem lösen](#aufgabe-11-ein-problem-lösen)
12. [Grenzen und Sicherheit](#grenzen-und-sicherheit)

---

## Aufgabe 1: Das Werkzeug bereitstellen

**Ziel:** Ein lauffähiges `bedrock-eu-check`.

**Weg A — aus dem Quellcode** (Voraussetzung: `make` und Docker; die Go-Toolchain läuft in
einem gepinnten Container, ohne Netz):

```bash
make build
./bin/bedrock-eu-check version
```

**Weg B — Docker-Image:** `make image` baut `slim` (nur `local` und `terraform`) und `full`
(mit AWS-CLI, alle Modi). Aufruf und Besonderheiten: [`docker.md`](docker.md).

**Weg C — im Devcontainer:** als Feature oder vorinstalliert, siehe [`devcontainer.md`](devcontainer.md).

**Erfolg:** `version` gibt eine Zeile wie `bedrock-eu-check 1.0.0 (abc123def456)` aus
(ohne gesetzte Build-Informationen `dev` und `unknown`).

## Aufgabe 2: Mein Projekt prüfen

**Ziel:** Wissen, ob die Konfiguration in meinem Projekt EU-konform ist — ohne Netz und
ohne AWS-Zugang.

1. Wechseln Sie in das Wurzelverzeichnis des Projekts (oder geben Sie `--root DIR` an).
2. Starten Sie die lokale Prüfung:

   ```bash
   bedrock-eu-check local
   ```

3. Lesen Sie die Ausgabe. Beispiel mit mehreren Befunden:

   ```text
   [WARN] LH-FA-CLAUDE-001 ~/.claude/settings.json: Benutzer-Settings nicht vorhanden (FILE_MISSING)
   [FAIL] LH-FA-LENV-001 AWS_REGION: AWS_REGION=us-east-1 ist keine erlaubte Region (REGION_NOT_ALLOWED)
   [FAIL] LH-FA-LENV-003 CLAUDE_CODE_USE_BEDROCK: CLAUDE_CODE_USE_BEDROCK ist nicht auf 1 gesetzt (BEDROCK_NOT_ENABLED)
   [PASS] LH-FA-LENV-004 ANTHROPIC_API_KEY: ANTHROPIC_API_KEY: not set (ANTHROPIC_API_KEY_NOT_SET)
   [FAIL] LH-FA-LENV-006 ANTHROPIC_DEFAULT_SONNET_MODEL: ANTHROPIC_DEFAULT_SONNET_MODEL=us.anthropic.claude-sonnet-4-6 ist nicht erlaubt (MODEL_BLOCKED)
   Summary: 1 PASS, 1 WARN, 3 FAIL
   ```

**Eine Zeile lesen:**

| Teil | Bedeutung |
|---|---|
| `[PASS]` / `[WARN]` / `[FAIL]` | in Ordnung / nicht belegbar oder zu prüfen / nicht EU-konform |
| [`LH-FA-LENV-001`](../../spec/lastenheft.md#lh-fa-lenv-001--prüfung-von-aws_region) | die Regel; Beschreibung im [Regelkatalog](rules.md) |
| `AWS_REGION` oder `datei:zeile` | der Gegenstand; bei Dateien mit Fundstelle |
| Text | die Meldung; Secret-Werte erscheinen nie, nur „set“ oder „not set“ |
| `(REGION_NOT_ALLOWED)` | der Reason-Code, Grundlage für die Abhilfe (Aufgabe 3) |

**Was `local` prüft:** Umgebungsvariablen (`AWS_REGION`, `CLAUDE_CODE_USE_BEDROCK`,
`ANTHROPIC_API_KEY`, Modellvariablen, Mantle), `.env`-Dateien, die Claude-Code-Settings (Benutzer und
Projekt), Devcontainer, Compose-Dateien und Dockerfiles.

**Nicht geprüft:** eigenständige Policy-Dateien wie `iam-policy.json` werden derzeit **nicht**
ausgewertet; IAM-Policies prüft das Werkzeug in Terraform-Dateien (Aufgabe 4) und beim eigenen
Principal (Aufgabe 5). Gehen Sie nicht davon aus, dass losgelöste Policy-JSONs geprüft sind.

**Wo gesucht wird:** Claude-Projekt-Settings (`.claude/settings.json`,
`.claude/settings.local.json`), `.devcontainer/devcontainer.json`, `.devcontainer.json`,
`docker-compose*.yml|yaml`, `compose*.yml|yaml` und `Dockerfile*` werden **nur im Scan-Root**
gesucht, nicht in Unterverzeichnissen. **Rekursiv** gesucht werden Terraform-Dateien (Aufgabe 4) und
`.env`-Dateien: `.env` und `.env.<name>` (z. B. `.env.production`), nicht aber Beispieldateien
(`*.example`, `*.sample`, `*.template`). Die Verzeichnisse `.git`, `.terraform`, `vendor`, `node_modules` und `.tmp`
werden übersprungen, Symlinks nicht verfolgt.

**Erfolg:** Exit-Code `0` (kein `FAIL`). `WARN` lässt den Lauf nicht scheitern.

## Aufgabe 3: Einen Befund verstehen und beheben

**Ziel:** Aus einem `FAIL` oder `WARN` die passende Abhilfe ableiten.

1. Nehmen Sie den Reason-Code aus der Klammer (z. B. `REGION_NOT_ALLOWED`).
2. Suchen Sie ihn in der Tabelle; die vollständige Liste steht im [Regelkatalog](rules.md).
3. Beheben Sie die Ursache und führen Sie die Prüfung erneut aus.

| Reason | Typische Ursache | Abhilfe |
|---|---|---|
| `REGION_NOT_ALLOWED` | Region liegt nicht in `allowedRegions` (Default nur `eu-central-1`) | Eine erlaubte Region setzen, oder die Region bewusst erlauben (Aufgabe 7) |
| `REGION_MISSING` | Weder `AWS_REGION` noch `AWS_DEFAULT_REGION` gesetzt; auch: `env.AWS_REGION` fehlt in den Benutzer-Settings (`FAIL`) bzw. ein Terraform-AWS-Provider hat kein `region`-Attribut (`WARN`) | Region setzen: Umgebungsvariable, Eintrag unter `env` in den Settings, `region` im Provider |
| `REGION_FALLBACK_DEFAULT` | Nur `AWS_DEFAULT_REGION` gesetzt (`WARN`) | Zusätzlich `AWS_REGION` setzen |
| `BEDROCK_NOT_ENABLED` | `CLAUDE_CODE_USE_BEDROCK` ist nicht `1` | `CLAUDE_CODE_USE_BEDROCK=1` setzen |
| `ANTHROPIC_API_KEY_SET` | `ANTHROPIC_API_KEY` ist gesetzt, die Anfragen könnten direkt an Anthropic gehen | Variable entfernen |
| `DIRECT_ANTHROPIC` | Direkte Anthropic-Konfiguration (API-Key, Auth-Token oder `ANTHROPIC_BASE_URL` auf `api.anthropic.com`) in Settings, Devcontainer, Compose, Dockerfile oder `.env` | Eintrag entfernen und Bedrock verwenden |
| `MODEL_BLOCKED` | Modell-/Profil-ID entspricht einem Block-Muster (`us.*`, `global.*`, nicht erlaubte Region in der ID) | Ein EU-Profil (`eu.*`) verwenden |
| `MODEL_UNCLASSIFIED` | ID passt weder auf ein Block- noch auf ein Allow-Muster (`WARN`) | ID prüfen oder per Konfiguration erlauben (Aufgabe 7) |
| `SETTINGS_UNSAFE` | Projekt-Settings enthalten ein nicht erlaubtes Modell, eine nicht erlaubte Region oder einen API-Key | Aus den Settings entfernen |
| `MANTLE_UNVERIFIED` | `CLAUDE_CODE_USE_MANTLE=1`, EU-Verfügbarkeit nicht belegbar (`WARN`) | EU-Verfügbarkeit manuell prüfen oder Policy festlegen (Aufgabe 7) |
| `MANTLE_BLOCKED` | Mantle ist per Konfiguration verboten | `CLAUDE_CODE_USE_MANTLE` entfernen |
| `FILE_MISSING` | Benutzer-Settings (`~/.claude/settings.json`) fehlen (`WARN`) | In CI und Containern erwartbar; sonst Datei anlegen |
| `FILE_INVALID` | Eine geprüfte Datei ist kein gültiges JSON/YAML/HCL | Syntax reparieren; Zeile steht in der Ausgabe |
| `FILE_UNREADABLE` | Eine geprüfte Datei ist nicht lesbar | Dateirechte prüfen |
| `NO_CLAUDE_CONFIG` | Devcontainer enthält weder Bedrock-Flag noch Region noch Anthropic-Variable (`WARN`) | `CLAUDE_CODE_USE_BEDROCK=1` und `AWS_REGION` unter `containerEnv` bzw. `remoteEnv` setzen, falls Claude Code im Devcontainer läuft |
| `ENV_INCOMPLETE` | Devcontainer setzt nur eines von Bedrock-Flag und Region, oder das Flag ist nicht `1` (`WARN`) | Beides setzen |
| `REGION_NOT_SET` | Dockerfile setzt keine AWS-Region per `ENV` (`WARN`) | Region per `ENV` setzen, oder bewusst zur Laufzeit setzen und die Warnung akzeptieren |
| `ENDPOINT_FOREIGN` | `ANTHROPIC_BASE_URL` zeigt auf einen anderen Host als `api.anthropic.com` (`WARN`) | Prüfen, ob der Endpunkt (z. B. ein Proxy) Anfragen innerhalb der EU hält |
| `ENDPOINT_TEXT_ONLY` | `ANTHROPIC_API_KEY` oder `ANTHROPIC_BASE_URL` kommt nur als Text außerhalb von `env` vor (`WARN`) | Prüfen, ob es nur ein Verweis ist (z. B. in einer Sperrliste) oder eine echte Konfiguration |
| `APIKEY_HELPER` | `apiKeyHelper` ist in den Claude-Settings gesetzt (`WARN`) | Prüfen, ob der Helper einen Anthropic-Key liefert |

**Warum `WARN` statt `FAIL`?** Ein `WARN` heißt: Das Werkzeug kann die EU-Konformität weder
belegen noch widerlegen — etwa weil ein Wert erst zur Laufzeit feststeht, ein Helper-Skript
einen Key liefern könnte oder eine optionale Datei fehlt. Es ist bewusst nie ein stilles
`PASS`. Ein `FAIL` bedeutet einen belegten Verstoß oder eine Datei, die sich nicht prüfen
lässt. `apiKeyHelper` ist `WARN`, weil ein Helper allein keine direkte Anthropic-Verbindung
belegt; der Befehl wird dabei weder gelesen noch ausgegeben.

**Doppelte Befunde sind gewollt:** Dieselbe Ursache (z. B. ein API-Key in den
Benutzer-Settings) erscheint je betroffener Regel einmal.

## Aufgabe 4: Terraform prüfen

**Ziel:** Provider-Region, Bedrock-IAM-Policies und Modell-IDs in `.tf`-Dateien prüfen.

```bash
bedrock-eu-check terraform                  # aktuelles Verzeichnis, rekursiv
bedrock-eu-check terraform --root infra     # nur ein Teilbaum
```

Beispiel (absichtlich fehlerhafte Testdaten):

```text
[PASS] LH-FA-TF-001 terraform: 1 Terraform-Datei(en) gefunden und gelesen (TERRAFORM_FOUND)
[FAIL] LH-FA-TF-002 main.tf:2: AWS-Provider-Region us-east-1 ist nicht erlaubt (REGION_NOT_ALLOWED)
       Empfehlung: Provider-Region auf eine erlaubte EU-Region setzen
[FAIL] LH-FA-TF-003 main.tf:6: Bedrock-Rechte (bedrock:*/InvokeModel) auf "*" ohne Region-Condition (IAM_WILDCARD)
       Empfehlung: Resource einschränken oder aws:RequestedRegion-Condition setzen
[FAIL] LH-FA-TF-004 main.tf:20: 1 nicht erlaubte Modell-/Profil-ID(s), z. B. us.anthropic.claude-3-5-sonnet-20241022-v2:0 (MODEL_BLOCKED)
       Empfehlung: EU-Profil-IDs (eu.*) verwenden
```

Was Sie wissen sollten:

- **Rekursiv:** Es werden alle `*.tf`-Dateien unterhalb des Scan-Roots gelesen; die Verzeichnisse
  `.git`, `.terraform`, `vendor`, `node_modules` und `.tmp` werden übersprungen, Symlinks nicht verfolgt.
- **Keine Auswertung von Variablen:** Eine Region wie `var.region` ist nicht belegbar und
  ergibt `WARN` (`REGION_DYNAMIC`); ebenso eine nicht auflösbare Policy (`IAM_DYNAMIC`).
  Werte aus `variable`, `locals` und `.tfvars` werden nie gelesen oder ausgegeben.
- **Keine Terraform-Dateien:** `WARN` (`NO_TERRAFORM_FILES`) — prüfen Sie `--root`.
- **Testdaten ausnehmen:** absichtlich fehlerhafte Beispiele per `exclude` (Aufgabe 7).

## Aufgabe 5: Das AWS-Konto und die Berechtigungen prüfen

**Ziel:** Identität, Bedrock-Profile und IAM-Rechte des aktuellen Principals prüfen.

**Voraussetzungen:** die AWS-CLI (`aws`) im `PATH`, gültige Zugangsdaten (z. B. `AWS_PROFILE`)
und die Rechte aus [`aws-permissions.md`](aws-permissions.md). Der Modus arbeitet nur lesend.

```bash
bedrock-eu-check aws
bedrock-eu-check aws --allowed-account 123456789012     # erlaubtes Konto festlegen
```

`--allowed-account` ist wiederholbar und wird mit `allowedAccounts` aus der
Konfiguration vereinigt. Eine Konto-ID hat genau 12 Ziffern; andere Werte beenden den
Lauf mit Exit `2` (in allen Modi, auch wenn die Option nur in `aws` wirkt).

Was geprüft wird: STS-Identität und Konto, `AWS_PROFILE`, Art der Authentifizierung,
sichtbare und verwendete Bedrock-Inference-Profile, Modellverfügbarkeit in der Region und die
IAM-Rechte samt Policies des **eigenen** Principals.

| Situation | Ergebnis |
|---|---|
| keine oder abgelaufene Zugangsdaten | `FAIL` `AWS_NO_CREDENTIALS`, Exit `1` |
| AWS nicht erreichbar | `FAIL` `AWS_UNREACHABLE` |
| Zugriff verweigert | `FAIL` `AWS_ACCESS_DENIED`; bei der IAM-Prüfung nur `WARN` `IAM_UNKNOWN` |
| Konto nicht in der Allowlist | `FAIL` `ACCOUNT_NOT_ALLOWED` |
| Identität nicht ermittelbar | `WARN` `ACCOUNT_UNCHECKED` |
| Invoke-Recht fehlt | `FAIL` `IAM_INVOKE_DENIED` |
| zu breite Bedrock-Policy ohne Region-Condition | `FAIL` `IAM_WILDCARD` |
| sichtbare, aber nicht verwendete nicht erlaubte Profile | `WARN` `PROFILES_VISIBLE_NOT_ALLOWED` |
| verwendetes Modell in der Primärregion nicht verfügbar | `FAIL` `MODEL_UNAVAILABLE` |
| `aws`-Binary fehlt | Tool-Fehler `AWS_CLI_MISSING`, Exit `2` |

**„Verwendet“** heißt: konfiguriert über `ANTHROPIC_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`,
`ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_SMALL_FAST_MODEL`
in der Umgebung sowie `model` und `env` in den Claude-Settings (Benutzer und Projekt, ohne
ausgeschlossene Pfade).

**Grenzen der IAM-Prüfung:** Sie bewertet nur den eigenen Principal. Die Rechtesimulation kennt
Organisations-SCPs und Permission Boundaries nur eingeschränkt und ist eine **Näherung**;
„nicht ermittelbar“ ist ein `WARN`, nie ein stilles `PASS`. Als Region-Einschränkung wird nur
`aws:RequestedRegion` oder eine Bindung des Bedrock-ARN an eine erlaubte Region anerkannt.

## Aufgabe 6: Alles in einem Lauf prüfen

**Ziel:** `local`, `terraform` und `aws` zusammen.

```bash
bedrock-eu-check all
```

- Die Befunde aller Modi erscheinen gemeinsam, nach Regel-ID sortiert.
- Fehlt das `aws`-Binary, endet der Lauf mit Exit `2`, die offline ermittelten Befunde
  erscheinen aber trotzdem.
- Ohne AWS-Zugang nutzen Sie `local` und `terraform` getrennt.

## Aufgabe 7: Die Regeln an meine Organisation anpassen

**Ziel:** Defaults anpassen, ohne den Code zu ändern.

Die Konfiguration ist **optional** und liegt in `bedrock-eu-check.yml` im Scan-Root, oder an
einem anderen Ort über `--config DATEI` (diese Datei muss dann existieren). Fehlende Felder
nutzen die Defaults. Unbekannte Felder, falsche Typen oder ungültige Werte beenden den Lauf
mit Exit `2`; die Meldung nennt das Feld und den Grund, nie den Wert. Alle Felder und
Defaults: [Referenz](reference.md#konfigurationsdatei-bedrock-eu-checkyml).

### Eine weitere Region erlauben

```yaml
allowedRegions: [eu-central-1, eu-west-1]
```

Vorher: `[FAIL] … AWS_REGION=eu-west-1 ist keine erlaubte Region (REGION_NOT_ALLOWED)`,
danach: `[PASS] … AWS_REGION=eu-west-1 ist erlaubt (OK)`. `allowedRegions` darf nicht leer sein.

### Erlaubte Konten festlegen

```yaml
allowedAccounts: ["123456789012"]
```

Leer bedeutet: Das Konto wird nicht geprüft. Alternativ `--allowed-account` (Aufgabe 5).

### Modelle erlauben oder blockieren

```yaml
allowedModelPatterns: ['^eu\.', '^anthropic\.claude-sonnet-4-6$']
blockedModelPatterns: ['^us\.', '^global\.']
```

Muster sind reguläre Ausdrücke. **Block-Muster haben Vorrang.** Eine ID mit einer nicht
erlaubten Region wird zusätzlich als blockiert behandelt. Ein ungültiger Ausdruck beendet den
Lauf mit Exit `2`.

### Den Umgang mit Mantle festlegen

```yaml
mantle: block        # warn (Default) | allow | block
```

| Wert | Ergebnis bei `CLAUDE_CODE_USE_MANTLE=1` |
|---|---|
| `warn` | `WARN` `MANTLE_UNVERIFIED` — EU-Verfügbarkeit manuell prüfen |
| `allow` | `PASS` `MANTLE_ALLOWED` |
| `block` | `FAIL` `MANTLE_BLOCKED` |

Blockierte Modell-IDs bleiben in **jedem** Modus ein `FAIL`.

### Testdaten und Fremdcode ausnehmen (`exclude`)

```yaml
exclude:
  - testdata
  - vendor
  - docs/*/gen
```

- Muster werden mit `path.Match` gegen Pfade **relativ zum Scan-Root** (mit `/`) verglichen.
- Ein Muster trifft den Pfad selbst **oder ein Verzeichnis darüber**: `testdata` nimmt
  `testdata/a/b.tf` aus. `**` wird nicht unterstützt.
- Die Wirkung ist im Terraform-Beispiel sichtbar: ohne `exclude` drei `FAIL` aus einer
  Testdatei, mit `exclude: [testdata]` nur noch `WARN` `NO_TERRAFORM_FILES`.
- Ein ungültiges Muster beendet den Lauf mit Exit `2`.

## Aufgabe 8: Ergebnisse weiterverarbeiten

**Ziel:** Befunde maschinell auswerten.

| Format | Aufruf | Verwendung |
|---|---|---|
| Konsole | `--format console` (Default) | zum Lesen |
| JSON | `--format json` | Skripte, Dashboards |
| SARIF | `--format sarif` | GitHub/GitLab Code Scanning |

stdout trägt nur das Ergebnis; Meldungen und Fehler stehen auf stderr. JSON hat die Form
`{"version": 1, "findings": [...]}` mit `status`, `id`, `message`, `severity`, `reason`, `file`,
`line` und `recommendation` (`null`, wenn nicht vorhanden).

```bash
# Nur die Fehlbefunde mit Fundstelle
bedrock-eu-check terraform --format json \
  | jq -r '.findings[] | select(.status=="FAIL") | "\(.id)  \(.file):\(.line)"'
```

```text
LH-FA-TF-002  main.tf:2
LH-FA-TF-003  main.tf:6
LH-FA-TF-004  main.tf:20
```

```bash
# Anzahl je Status
bedrock-eu-check local --format json | jq -r '[.findings[].status] | group_by(.) | map("\(.[0]): \(length)") | .[]'
```

**SARIF:** `PASS` erzeugt kein Result. Befunde ohne Fundstelle in einer Datei des Repositorys
(Umgebung, AWS, Dateien außerhalb wie `~/.claude/settings.json`) erhalten eine **Ersatzposition**
(`bedrock-eu-check.yml`, sonst `README.md`, Zeile 1) und sind an `properties.synthetic`
erkennbar.

## Aufgabe 9: In CI einbinden

**Ziel:** Bei jedem Pull Request prüfen und Befunde in Code Scanning anzeigen.

- `--ci` ändert weder Ausgabe noch Exit-Codes, schreibt aber zusätzlich eine Zusammenfassung
  auf stderr (`bedrock-eu-check: 3 PASS, 1 WARN, 0 FAIL (Exit 0)`), damit Pipeline-Logs den
  Ausgang zeigen, auch wenn stdout in eine Datei umgeleitet ist.
- Die Pipeline scheitert bei Exit `1` (mindestens ein `FAIL`) und bei Exit `2` (Tool-Fehler),
  nicht bei `WARN`.
- Für Pull Requests aus Forks eignen sich meist nur `local` und `terraform`.

Ein vollständiger Beispiel-Workflow für GitHub Code Scanning (nicht auf GitHub erprobt) steht
in [`ci.md`](ci.md). Auch dieses Repository prüft sich selbst: `make self-check`.

## Aufgabe 10: Im Container oder Devcontainer nutzen

**Container (`slim` oder `full`):** Das Projekt wird nach `/work` eingebunden, Umgebungsvariablen
werden mit `-e` durchgereicht (der Container sieht die Umgebung des Hosts nicht):

```bash
docker run --rm --network none -v "$PWD:/work:ro" \
  -e AWS_REGION -e CLAUDE_CODE_USE_BEDROCK \
  bedrock-eu-check:slim local
```

- Der Container läuft als Nutzer `65532`. Ist das Projektverzeichnis dafür nicht lesbar, endet
  der Lauf mit Exit `2` (`FILE_UNREADABLE` der Konfiguration); Abhilfe: Rechte anpassen oder
  `--user "$(id -u):$(id -g)"`.
- Benutzer-Settings werden nur geprüft, wenn `~/.claude` eingebunden und `HOME` gesetzt ist.
- `aws` und `all` brauchen das `full`-Image.

Alle Einzelheiten: [`docker.md`](docker.md).

**Devcontainer:** Das Feature installiert ein Release-Binary mit geprüfter SHA256-Prüfsumme;
Details und Stand: [`devcontainer.md`](devcontainer.md).

## Aufgabe 11: Ein Problem lösen

| Symptom | Erste Schritte |
|---|---|
| Exit `2` mit `usage:` | Befehl oder Option falsch geschrieben — [Referenz](reference.md) |
| Exit `2` mit `FILE_INVALID (Zeile N)` | Eigene Konfiguration: unbekanntes Feld oder falscher Typ in Zeile N |
| Exit `2` mit `Konfiguration: <feld>: <GRUND>` | Ungültiger Wert in `<feld>` |
| `AWS_CLI_MISSING` | `aws`-CLI installieren oder das `full`-Image verwenden |
| `AWS_NO_CREDENTIALS` | anmelden bzw. SSO-Sitzung erneuern |
| Befunde aus Testdaten | `exclude` (Aufgabe 7) |
| Viele `WARN` ohne Bedeutung | prüfen, ob sie erwartbar sind: [Fehlersuche](troubleshooting.md) |

Mit `--format json` sehen Sie je Befund `reason`, `file` und `line` und können gezielt filtern.
Die vollständige Liste typischer Meldungen steht in der [Fehlersuche](troubleshooting.md).

## Grenzen und Sicherheit

- **Nur sichtbare Konfiguration.** Das Werkzeug ist keine Rechts- oder Datenschutzberatung und
  kein Beweis über das interne Routing von AWS; es prüft Konfigurationen, die es lesen kann.
- **Nicht alle Dateiarten.** Eigenständige Policy-JSONs (`iam-policy.json`) werden nicht geprüft (siehe
  Aufgabe 2); Direktzugriffe auf die Anthropic-API darin erkennt das Werkzeug daher nicht.
- **Nur lesend.** Es ändert weder Dateien noch AWS-Ressourcen und repariert nichts. Der Modus `aws`
  nutzt ausschließlich lesende Aufrufe.
- **Keine Secret-Werte.** Weder in der Konsole noch in JSON oder SARIF erscheinen Zugangsdaten
  oder Schlüsselwerte; die Authentifizierungsart wird nur aus dem Vorhandensein von Variablen
  abgeleitet. Fehlermeldungen der AWS-CLI werden nicht übernommen.
- **Näherungen sind markiert.** Nicht auswertbare Ausdrücke (Terraform-Variablen, `NotAction` in
  Policies), die eingeschränkte IAM-Simulation und nicht belegbare Einstellungen ergeben `WARN`.
- **Veröffentlichung.** Die Version 1.0.0 ist veröffentlicht; weitere Veröffentlichungen
  geschehen nur nach ausdrücklicher Freigabe ([`releasing.md`](releasing.md)).

## Weiterführend

| Thema | Dokument |
|---|---|
| Schneller Einstieg | [`quickstart.md`](quickstart.md) |
| Befehle, Optionen, Konfiguration, Exit-Codes | [`reference.md`](reference.md) |
| Jede Prüfung mit Bedingungen und Reason-Codes | [`rules.md`](rules.md) |
| CI und Code Scanning | [`ci.md`](ci.md) |
| Docker-Images | [`docker.md`](docker.md) |
| Devcontainer | [`devcontainer.md`](devcontainer.md) |
| Nötige AWS-Rechte | [`aws-permissions.md`](aws-permissions.md) |
| Typische Meldungen | [`troubleshooting.md`](troubleshooting.md) |
| Release-Ablauf | [`releasing.md`](releasing.md) |
