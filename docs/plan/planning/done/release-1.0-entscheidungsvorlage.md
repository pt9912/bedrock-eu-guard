# Entscheidungsvorlage: Release 1.0

**Status:** teilweise entschieden (Nutzer 1, 2026-11-11); **Freigabe zur Veröffentlichung noch nicht erteilt** (Commit steht noch nicht fest, Antwort von Nutzer 2 fehlt). **Datum:** 2026-11-11.
**Verfasser:** w2. **Gemeinsame Vorlage** mit w1: Die Entscheidung zur Lücke bei den Testdaten
(`.env`, `iam-policy.json`) steht in [`release-1.0-entscheidung-qa003.md`](release-1.0-entscheidung-qa003.md).

Dieses Dokument bereitet Entscheidungen vor und **führt nichts aus**. Tag, GitHub-Release,
Image-Push und das Veröffentlichen des Devcontainer-Features geschehen nur nach einem
ausdrücklichen „Ja“ **beider** Nutzer zu genau der hier genannten Version, dem Commit und den
Zielen (siehe [`releasing.md`](../../../user/releasing.md), Abschnitt „Release-Ablauf“).

## 1. Kurzfassung

`bedrock-eu-check` ist als Release-Kandidat technisch fertig
([`welle-05-results.md`](../done/welle-05-results.md)). Es ist **nichts veröffentlicht**: keine Tags,
keine Release-Artefakte, keine Images in einer Registry, kein veröffentlichtes Feature. Das
GitHub-Repository ist **öffentlich**.

| # | Entscheidung | Empfehlung | Abschnitt |
|---|---|---|---|
| A | Freigabe zur Veröffentlichung: Version, Commit, Ziele | stufenweise: zuerst Tag und GitHub-Release, danach Image, zuletzt Feature | 3 |
| B | Registry und Namen für Image und Feature | GitHub Container Registry; Namen siehe Abschnitt 5 | 5 |
| C | `apiKeyHelper`: `WARN` oder `FAIL` | `WARN` beibehalten | 6 |
| D | Schwelle `--fail-on warn` | Folgeslice nach 1.0 | 7 |
| E | Lücke `.env` / `iam-policy.json` | siehe [w1-Vorlage](release-1.0-entscheidung-qa003.md) (Empfehlung dort: mindestens `.env`); bei Option 1 zusätzlich deren Festlegungen 1 bis 4 und der Wortlaut der Lastenheft-Anforderungen | 8 |

Die Entscheidungen A und E hängen zusammen: Wird E mit „Scanner ergänzen“ beantwortet, verschiebt sich
der Zeitpunkt von A um die zusätzlichen Slices.

## 2. Stand (Messwerte)

Gemessen auf `main`, Commit `7e78decdd8c8`, am 2026-11-11 (**vor** dem `.env`-Scanner; der Release-Commit ist ein späterer, siehe Abschnitt 3).

| Gegenstand | Stand |
|---|---|
| Gates | `make verify-clean` (doc-check, lint, test inkl. Architektur- und Doku-Abgleichstests), `make self-check`, `make deps-verify`, `make image-test`, `make feature-test`: grün; CI-Workflows `gates`, `image`, `deps-verify` grün |
| Binaries | Linux `amd64` (5,1 MB) und `arm64` (4,9 MB), statisch gelinkt; zwei Läufe von `make dist` liefern identische Prüfsummen |
| Docker-Images | `slim` 7,3 MB (`local`, `terraform`), `full` 659 MB (mit AWS-CLI 2.39.7, alle Modi); Nutzer `65532` |
| Gepinnte Basen | Go-Toolchain `go1.25.14` per Digest (`golang@sha256:699337d6…`), AWS-CLI-Image `amazon/aws-cli:2.39.7@sha256:bba26234…` |
| Devcontainer-Feature | `bedrock-eu-check` Version `1.0.0`; erwartet `bedrock-eu-check_1.0.0_linux_<arch>` und `SHA256SUMS` unter `<baseUrl>/v1.0.0/` |
| Tags / Releases | keine |
| Repository | öffentlich |

**Vorschau der Prüfsummen** (`make dist VERSION=1.0.0 COMMIT=7e78decdd8c8`):

```text
b0494016e33b7d3f5456a3e1468bb9ee0822925370e7ced451b5d6821556c3cb  bedrock-eu-check_1.0.0_linux_amd64
e83aba16a734ccb28227ce98c51ce98bd33a2c13074f2a7c27c60fc1259dc352  bedrock-eu-check_1.0.0_linux_arm64
```

Die Vorschau gilt für den älteren Commit und ändert sich mit dem Release-Commit. Die Prüfsummen hängen von Version **und Commit** ab (beides steckt im Binary). **Maßgeblich** sind die
Werte des Release-Commits; die Vorschau zeigt nur, dass der Build reproduzierbar ist.

## 3. Entscheidung A: Freigabe zur Veröffentlichung

Jedes Ziel ist ein eigener, getrennt freizugebender Schritt.

| Ziel | Was geschieht | Rückholbar? | Voraussetzung |
|---|---|---|---|
| 1. Git-Tag `v1.0.0` | Tag wird auf den Release-Commit gesetzt und gepusht | **praktisch nein**: Das Repository ist öffentlich, der Modulpfad `github.com/pt9912/bedrock-eu-guard` ist über den Go-Modul-Proxy abrufbar, der Versionen dauerhaft zwischenspeichern kann. Ein Tag macht das Werkzeug damit auch per `go install github.com/pt9912/bedrock-eu-guard/cmd/bedrock-eu-check@v1.0.0` installierbar (erwartet, nicht geprüft); ein so gebautes Binary trägt keine Build-Informationen (`version` zeigt `dev`) | Entscheidung E geklärt, Gates grün |
| 2. GitHub-Release | Release mit Binaries, `SHA256SUMS` und Release-Notizen | Release lässt sich löschen oder bearbeiten, **bereits geladene Kopien nicht** | Tag (1) |
| 3. Docker-Images | `slim` und `full` werden in eine Registry gepusht | Tags lassen sich je nach Registry löschen, **bereits gezogene Images nicht** | Registry und Namen (B) |
| 4. Devcontainer-Feature | Feature wird als OCI-Artefakt veröffentlicht | wie bei Images; versionierte Tags gelten als dauerhaft | Release-Artefakte (2), Registry (B) |

**Empfehlung:** stufenweise freigeben und nach jeder Stufe nachprüfen (Abschnitt 4, Schritt 6):
zuerst 1 und 2, danach 3, zuletzt 4. Das Feature braucht die Release-Artefakte, die Images sind
davon unabhängig. Wer nur das Binary veröffentlichen möchte, gibt nur 1 und 2 frei; 3 und 4 können später folgen.

**Wortlaut der Freigabe** (beide Nutzer, getrennt): „Ja zu Release **1.0.0** auf Commit **`<SHA>`** für die
Ziele **`<Liste aus 1–4>`**.“ Ohne diesen Wortlaut führen die Agenten keinen der Schritte aus.

## 4. Release-Checkliste (nur nach Freigabe)

Alle Schritte sind **manuell**; kein Workflow veröffentlicht etwas. `<SHA>` ist der freigegebene Commit.

1. **Stand prüfen** (sauberer Checkout von `<SHA>`):
   `make verify-clean`, `make self-check`, `make deps-verify`, `make image-test`, `make feature-test`
   und die CI auf `<SHA>` grün. Außerdem prüfen, dass `LICENSE` und `THIRD_PARTY_NOTICES.md`
   (MPL-2.0 für HCL, `yaml.v3` u. a.) zum Release-Stand passen: Lizenzhinweise gehören in ein öffentliches Release.
2. **Release-Notizen** schreiben: Highlights, bekannte Grenzen (nicht geprüfte Dateiarten, Entscheidung E,
   nicht erprobte Workflows), Verweis auf [`docs/user/handbuch.md`](../../../user/handbuch.md).
3. **Artefakte bauen:** `make dist VERSION=1.0.0 COMMIT=<SHA kurz>` lokal, und den Workflow
   `release-artifacts` mit Version `1.0.0` auf `<SHA>` laufen lassen. Die Prüfsummen aus beiden Läufen
   müssen übereinstimmen (Reproduzierbarkeit); `sha256sum -c SHA256SUMS` meldet „OK“.
4. **Tag setzen** *(Ziel 1)*: `git tag -a v1.0.0 -m "bedrock-eu-check 1.0.0" <SHA>`, danach
   `git push origin v1.0.0`. Der Tag-Push löst **keinen** Workflow aus: `gates` reagiert nur auf Pull Requests und
   auf Pushes nach `main`, `release-artifacts` nur auf manuellen Start.
5. **GitHub-Release anlegen** *(Ziel 2)*:
   `gh release create v1.0.0 dist/v1.0.0/* --title "bedrock-eu-check 1.0.0" --notes-file <datei>`.
6. **Nachprüfen:** Artefakte frisch herunterladen, `sha256sum -c`, `bedrock-eu-check version` meldet `1.0.0`
   und den Commit.
7. **Images** *(Ziel 3, nach Registry-Entscheidung B)*: `make image VERSION=1.0.0 COMMIT=<SHA kurz>`,
   dann taggen und pushen, z. B. `docker tag bedrock-eu-check:slim <registry>/<name>:1.0.0-slim` und
   `docker push …`; ebenso `full`. *Nicht erprobt.* Danach ein Pull in einer frischen Umgebung und
   `tools/image-smoke.sh` gegen das gezogene Image.
8. **Feature** *(Ziel 4)*: als OCI-Feature mit der Devcontainer-CLI oder der zugehörigen GitHub-Action
   veröffentlichen. *Nicht erprobt; Werkzeug und Namensraum sind noch festzulegen.* Danach in einem
   frischen Devcontainer mit der Standard-`baseUrl` installieren und `bedrock-eu-check version` prüfen.
9. **Dokumentation und Roadmap:** Stand in README, `releasing.md` und Roadmap von „Release-Kandidat“ auf
   „veröffentlicht“ ändern, `welle-05` nach `done/` abschließen (reiner `git mv`, danach Verweise per `grep` nachziehen,
   macht w1), Meilenstein M5 eintragen (eigener PR).

## 5. Entscheidung B: Registry und Namen

Aktuell ist nichts festgelegt. Vorschlag zur Entscheidung (keiner der Namen ist reserviert oder geprüft):

| Gegenstand | Vorschlag | Anmerkung |
|---|---|---|
| Registry | GitHub Container Registry (`ghcr.io`) | passt zum Repository; Docker Hub wäre die Alternative |
| Image | `ghcr.io/pt9912/bedrock-eu-check` mit Tags `1.0.0-slim`, `1.0.0-full` (und später `latest` nur bewusst) | `latest` bewusst nicht in 1.0 |
| Feature | Namensraum unter dem Repository, z. B. `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check:1` | Namensschema der Feature-Veröffentlichung **nicht verifiziert** |

Wer Images und Feature zunächst nicht veröffentlichen möchte, überspringt diese Entscheidung; die Docker-Nutzung
und das Feature bleiben dann lokal (Build aus dem Quellcode), wie in der [Dokumentation](../../../user/docker.md) beschrieben.

## 6. Entscheidung C: `apiKeyHelper`

Ist in den Claude-Settings ein `apiKeyHelper` gesetzt, meldet das Werkzeug ein `WARN`
([`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings)); der Befehl wird nie gelesen oder ausgegeben.

| Option | Wirkung | Bewertung |
|---|---|---|
| `WARN` (Ist) | Lauf scheitert nicht, Befund sichtbar | Ein Helper ist ein Skript; ob es einen Anthropic-Key liefert, ist aus der Konfiguration nicht belegbar. Ein `FAIL` wäre bei Bedrock-Nutzung mit legitimem Helper ein Fehlalarm |
| `FAIL` | Lauf scheitert | strenger, aber unzuverlässig; würde Teams zwingen, den Helper zu entfernen, obwohl er evtl. ein Bedrock-Token liefert |
| `FAIL` nur, wenn `apiKeyHelper` gesetzt **und** `CLAUDE_CODE_USE_BEDROCK=1` im selben Geltungsbereich fehlt | Lauf scheitert nur ohne aktives Bedrock | genauer, aber Benutzer- und Projekt-Settings müssten zusammen bewertet werden: mehr Kernlogik und eine Änderung an [`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings) |

**Empfehlung:** `WARN` beibehalten; wer streng sein will, wertet `APIKEY_HELPER` in der JSON-Ausgabe aus.

## 7. Entscheidung D: Schwelle `--fail-on warn`

Heute lässt nur `FAIL` den Lauf scheitern (Exit 1). Eine Option, auch bei `WARN` zu scheitern, wäre
eine Erweiterung der Exit-Code-Regeln ([`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes)) und braucht Spec und Tests.

**Empfehlung:** Folgeslice nach 1.0. Bis dahin: `--format json` und eine Auswertung wie
`jq -e '[.findings[] | select(.status=="WARN")] | length == 0'` in der Pipeline (siehe [Handbuch](../../../user/handbuch.md)).

## 8. Entscheidung E: Lücke `.env` / `iam-policy.json`

Das Lastenheft verlangt Testdaten für beide Dateiarten, beschreibt aber keine Scanner dafür. Optionen, Aufwand und
Empfehlung stehen in [`release-1.0-entscheidung-qa003.md`](release-1.0-entscheidung-qa003.md). Auswirkung auf diesen Plan:

- **Option 1 (Scanner ergänzen):** Release verschiebt sich um die zusätzlichen Slices; A wird danach freigegeben.
- **Option 2 (Anforderung präzisieren):** Release 1.0 kann zeitnah erfolgen; die Grenze („`.env` und Policy-JSON werden
  nicht geprüft“) steht bereits im [Handbuch](../../../user/handbuch.md) und im README und gehört in die Release-Notizen.

**Umsetzungsstand (2026-11-11):** Option 1b ist umgesetzt: Der `.env`-Scanner ist in `main`
([`slice-022`](../done/slice-022-dotenv-scanner.md), PRs #76–#79, Matrix-Test #78). Offen bleibt nur die
eigenständige Policy-Datei `iam-policy.json` (Soll nach 1.0, gehört in die Release-Notizen).

## 9. Nicht erprobt und Risiken

Nicht auf der Zielplattform ausgeführt oder gegen externe Systeme getestet:

- der Workflow `release-artifacts` (nur als YAML geprüft; `make dist` ist lokal getestet),
- der Beispiel-Workflow für Code Scanning in [`ci.md`](../../../user/ci.md),
- das Devcontainer-Feature mit der Devcontainer-CLI und mit der Standard-`baseUrl` (erst nach einem Release möglich),
- die Beispiel-IAM-Policy gegen ein echtes AWS-Konto,
- das Pushen von Images und das Veröffentlichen des Features.

Weitere Risiken: Prüfsummen schützen vor Korruption, nicht vor einem kompromittierten Ursprung (keine Signaturen, keine
SBOM in 1.0); nur Linux (`amd64`, `arm64`); gepinnte Basen erhalten keine automatischen Sicherheitsupdates
(Aktualisierung nach [`releasing.md`](../../../user/releasing.md)); `full` ist mit 659 MB groß.

## 10. Antwortblatt

Bitte je Zeile eintragen (beide Nutzer getrennt):

| # | Frage | Antwort Nutzer 1 | Antwort Nutzer 2 |
|---|---|---|---|
| A | Freigabe für Version 1.0.0 auf Commit `<SHA>`: welche Ziele (1 Tag, 2 Release, 3 Images, 4 Feature)? | Grundsätzlich **ja**; Ziele: **1 Tag + 2 GitHub-Release, 3 Docker-Images, 4 Devcontainer-Feature**. Commit noch nicht festgelegt, daher **erneute Bestätigung** mit dem endgültigen `<SHA>` nach dem `.env`-Scanner | |
| B | Registry und Namen (oder „zunächst nicht veröffentlichen“) | **GitHub** (`ghcr.io`); endgültige Namen noch zu bestätigen | |
| C | `apiKeyHelper`: `WARN` behalten oder `FAIL` | nach Belieben der Agenten → **`WARN` bleibt** | |
| D | `--fail-on warn`: Folgeslice nach 1.0 (ja/nein) | **ja** | |
| E | `.env` / `iam-policy.json`: Option 1a, 1b oder 2 (siehe w1-Vorlage); bei Option 1 zusätzlich Festlegungen 1 bis 4 (Dateimuster, Modus, Suchmuster für Policy-Dateien, Suchtiefe für `.env`) und Wortlaut der Lastenheft-Anforderungen | **Option 1b**: erst `.env`-Scanner, dann Release; Lastenheft wird um die `.env`-Anforderung ergänzt. Festlegungen 1 bis 4 und Wortlaut nach Belieben der Agenten, orientiert an der w1-Vorlage | |

## 11. Beschlossenes Vorgehen (Stand 2026-11-11)

1. `.env`-Scanner als eigener Slice (Plan, Kern, Adapter, Tests, Lastenheft-Eintrag) vor dem Release.
2. Danach erneute Abfrage beider Nutzer mit dem endgültigen Commit im Wortlaut aus Abschnitt 3.
3. Erst nach dem Ja **beider** Nutzer: Checkliste aus Abschnitt 4, stufenweise (Tag und Release, Images, Feature).
4. Folgeslice nach 1.0: `--fail-on warn`.
