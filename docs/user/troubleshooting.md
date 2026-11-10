# Fehlersuche

Typische Meldungen und was dahintersteckt. Die Bedeutung der Reason-Codes steht im
[Regelkatalog](rules.md), Optionen und Exit-Codes in der [Referenz](reference.md).

## Exit 2 — Tool-Fehler

| Meldung | Ursache und Abhilfe |
|---|---|
| `usage: bedrock-eu-check …` | Unbekannter Befehl oder unbekannte Option, fehlender Wert (z. B. `--root` ohne Verzeichnis), ungültiges `--format`. Gültige Aufrufe: [Referenz](reference.md). |
| `error: <datei>: FILE_INVALID (Zeile N)` | Die eigene Konfigurationsdatei ist kein gültiges YAML, enthält ein **unbekanntes Feld** oder einen falschen Typ (z. B. `allowedRegions` als Text statt Liste). Zeile N prüfen. |
| `error: Konfiguration: <feld>: <GRUND>` | Ein Feld hat einen ungültigen Wert (z. B. ungültiger regulärer Ausdruck, unbekannter `mantle`-Wert, leere `allowedRegions`, Konto-ID ohne 12 Ziffern). Die Gründe stehen in der [Referenz](reference.md). |
| `error: bedrock-eu-check.yml: FILE_UNREADABLE` | Die Konfigurationsdatei ist nicht lesbar, im Docker-Image meist wegen Dateirechten des Mounts (Nutzer `65532`). Rechte anpassen oder `--user "$(id -u):$(id -g)"` verwenden, siehe [`docker.md`](docker.md). |
| `error: Konfigurationsdatei nicht lesbar oder nicht vorhanden (--config)` | Eine mit `--config` angegebene Datei muss existieren. Nur die Default-Datei im Scan-Root ist optional. |
| `AWS sts:GetCallerIdentity: AWS_CLI_MISSING` | Der Modus `aws`/`all` braucht das `aws`-Binary im `PATH`. Im `slim`-Image gibt es keines: das `full`-Image verwenden. |

## Exit 1 — Befunde

| Befund | Ursache und Abhilfe |
|---|---|
| `AWS_NO_CREDENTIALS` | Keine oder abgelaufene Zugangsdaten (auch abgelaufene SSO-Sitzung). Anmelden und erneut ausführen. |
| `AWS_ACCESS_DENIED` | Dem aktuellen Principal fehlt ein Recht der Prüfung; benötigte Rechte: [`aws-permissions.md`](aws-permissions.md). Bei der IAM-Prüfung ergibt das nur `WARN` (`IAM_UNKNOWN`). |
| `AWS_UNREACHABLE` | AWS nicht erreichbar (Netz, Proxy, falsche Region/Endpoint) oder unerwartete Antwort der AWS-CLI. |
| `FILE_INVALID` / `FILE_UNREADABLE` bei einer **geprüften** Datei | Die Datei ist kein gültiges JSON/YAML/HCL bzw. nicht lesbar. Das ist ein Befund (`FAIL`), kein Tool-Fehler: Die EU-Konformität lässt sich nicht belegen. |
| Befunde aus Testdaten | Absichtlich fehlerhafte Beispiele im Repository werden mitgeprüft. Mit `exclude` in `bedrock-eu-check.yml` ausnehmen (z. B. `testdata`). |

## Warnungen, die oft erwartet sind

| Warnung | Bedeutung |
|---|---|
| `FILE_MISSING` für `~/.claude/settings.json` | In CI und Containern gibt es meist keine Benutzer-Settings. Im Docker-Image per Mount bereitstellen ([`docker.md`](docker.md)). |
| `NO_TERRAFORM_FILES` | Im Scan-Root liegt keine `.tf`-Datei (Verzeichnisse wie `.terraform`, `vendor`, `node_modules` werden übersprungen). `--root` prüfen. |
| `REGION_DYNAMIC` | Die Region eines Terraform-Providers ist ein Ausdruck (`var.region`, `local.x`), der sich ohne Auswertung nicht belegen lässt. |
| `REGION_NOT_SET` bei einem Dockerfile | Das Dockerfile setzt keine AWS-Region per `ENV`. Gewollt, wenn die Region zur Laufzeit gesetzt wird. |
| `ACCOUNT_UNCHECKED` | Das Konto wurde nicht geprüft, weil die Identität nicht ermittelbar war. |

## Ausgabeformate

| Beobachtung | Erklärung |
|---|---|
| SARIF-Alerts erscheinen an `bedrock-eu-check.yml` oder `README.md`, Zeile 1 | Befunde ohne Fundstelle (Umgebung, AWS, Dateien außerhalb des Repositorys) erhalten eine Ersatzposition; erkennbar an `properties.synthetic`. |
| `PASS` fehlt in SARIF | Gewollt: `PASS` erzeugt dort kein Result. |
| stdout enthält nur JSON/SARIF, aber der Lauf endete mit Fehler | Das ist beabsichtigt: bereits ermittelte Befunde bleiben auf stdout parsebar, die Fehlermeldung steht auf stderr. Mit `--ci` zusätzlich eine Zusammenfassung auf stderr ([`ci.md`](ci.md)). |

## Weiter

Bleibt ein Problem ungeklärt, hilft die Ausgabe mit `--format json` (enthält `reason`,
`file`, `line` je Befund) beim Eingrenzen. Zugangsdaten und Secret-Werte gehören nicht
in Fehlerberichte; das Werkzeug gibt sie ohnehin nie aus.
