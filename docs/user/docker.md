# Docker-Image

`bedrock-eu-check` gibt es als Docker-Image in zwei Varianten
([`LH-TECH-004`](../../spec/lastenheft.md#lh-tech-004--docker-image)). Die Images sind
**noch nicht veröffentlicht**; bis dahin lassen sie sich lokal bauen:

```bash
make image        # baut bedrock-eu-check:slim und bedrock-eu-check:full
make image-test   # baut und führt die Smoke-Tests aus (Docker, kein Netz zur Laufzeit)
```

Das Bauen braucht Netz zum Ziehen der Basis-Images (per Digest gepinnt); der
Build selbst nutzt `vendor/` und braucht kein Netz.

## Varianten

| Image | Inhalt | Modi |
|---|---|---|
| `slim` | nur das statische Binary (`FROM scratch`, rund 7 MB, ohne Shell) | `local`, `terraform`, `version` |
| `full` | zusätzlich die AWS-CLI 2.39.7 (gepinnt) | alle Modi, auch `aws` und `all` |

Mit `slim` enden `aws` und `all` mit einem Tool-Fehler (Exit `2`,
`AWS_CLI_MISSING`), weil das `aws`-Binary fehlt.

Beide Images laufen als Nutzer `65532:65532` (kein Root) mit dem Arbeitsverzeichnis
`/work`; das Binary ist der `ENTRYPOINT`.

## Aufruf

Das Projekt wird nach `/work` eingebunden (lesend genügt):

```bash
docker run --rm --network none -v "$PWD:/work:ro" bedrock-eu-check:slim local
docker run --rm --network none -v "$PWD:/work:ro" bedrock-eu-check:slim terraform --ci --format sarif > terraform.sarif
```

`--network none` ist für `local` und `terraform` möglich und empfehlenswert: Diese
Modi arbeiten offline ([`LH-NF-005`](../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)).
Die Optionen und Exit-Codes stehen in [`ci.md`](ci.md).

### Umgebungsvariablen

Die Prüfungen von `local` lesen die Umgebung **im Container**, nicht die des
Hosts. Variablen müssen daher durchgereicht werden:

```bash
docker run --rm -v "$PWD:/work:ro" \
  -e AWS_REGION -e CLAUDE_CODE_USE_BEDROCK \
  bedrock-eu-check:slim local
```

Ohne `-e` meldet der Lauf fehlende Variablen als Befund.

### Benutzer-Settings von Claude Code

`~/.claude/settings.json` liegt auf dem Host, nicht im Container. Ohne `HOME`
im Container meldet `local` die Benutzer-Settings als nicht vorhanden (`WARN`,
`FILE_MISSING`). Zum Prüfen wird das Verzeichnis eingebunden und `HOME` gesetzt:

```bash
docker run --rm --network none \
  -e HOME=/home/nonroot -v "$HOME/.claude:/home/nonroot/.claude:ro" \
  -v "$PWD:/work:ro" \
  bedrock-eu-check:slim local
```

Im `full`-Image ist `HOME` bereits `/home/nonroot`.

### Modus `aws` (nur `full`)

```bash
docker run --rm --user "$(id -u):$(id -g)" \
  -e HOME=/home/nonroot -v "$HOME/.aws:/home/nonroot/.aws:ro" -e AWS_PROFILE \
  bedrock-eu-check:full aws
```

Dieser Modus braucht Netz und Zugangsdaten; die benötigten Rechte stehen in
[`aws-permissions.md`](aws-permissions.md). Ohne Zugangsdaten endet der Modus mit
`FAIL` (Exit `1`). Bei SSO-Profilen kann die AWS-CLI Cache-Dateien unter `~/.aws`
schreiben; dort funktioniert gegebenenfalls nur ein Mount ohne `:ro` (nicht
erprobt).

## Dateirechte bei Bind-Mounts

Der Container läuft als Nutzer `65532`, nicht als der Nutzer des Hosts. Das
eingebundene Verzeichnis muss für „andere“ lesbar sein (üblich: `755`/`644`).
Ist es das nicht, scheitert schon das Lesen der Konfigurationsdatei, und der Lauf
endet mit Exit `2` und der Meldung `bedrock-eu-check.yml: FILE_UNREADABLE`.
Einzelne nicht lesbare Projektdateien erscheinen dagegen als `FAIL` mit
`FILE_UNREADABLE`.

Zwei Auswege: die Rechte anpassen oder den Container als eigener Nutzer
starten (`--user "$(id -u):$(id -g)"`). Letzteres ist auch für
`~/.aws` nötig, dessen Dateien meist nur dem Besitzer gehören (`600`).

## Hinweise

- **Kein Root, keine Secrets im Image:** Das Image speichert keine Zugangsdaten; sie
  werden zur Laufzeit eingebunden. Das Werkzeug gibt keine Secret-Werte aus
  ([`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
- **Read-only:** Beide Images laufen mit `--read-only`.
- **Reproduzierbarkeit:** Das Binary im Image ist bitgleich mit dem aus `make build`
  ([`LH-QA-004`](../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)); die Image-ID
  selbst kann wegen Zeitstempeln der Ebenen abweichen.
- **Aktualisierung:** Die Basis-Images sind per Digest gepinnt und erhalten keine
  automatischen Updates; die AWS-CLI-Version wird bewusst erhöht (Tag und Digest im
  `Dockerfile`).
