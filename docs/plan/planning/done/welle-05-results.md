# Welle welle-05-release: Ergebnisse

**Abgeschlossen:** 2026-11-11. **Status:** Release 1.0.0 **veröffentlicht**.
**Meilenstein:** M5 — Release 1.0 **erreicht** (2026-11-11).

## Closure-Trigger

| Kriterium | Stand |
|---|---|
| Alle Slices done (014 bis 022) | erfüllt |
| `make verify-clean` und CI grün | erfüllt (`gates`, `image`, `deps-verify`) |
| `version` zeigt Version und Commit; SARIF-Tool-Version aus derselben Quelle | erfüllt; das veröffentlichte Binary meldet `bedrock-eu-check 1.0.0 (5f1b35cca776)` |
| Image baut aus gepinnten Basen und führt `all --ci` auf einem Beispiel-Repository aus | erfüllt (Smoke-Test je Image) |
| Self-Check und Architekturprüfung in CI | erfüllt |
| Testdaten decken die Dateiarten aus [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) ab | erfüllt für alle geprüften Dateiarten einschließlich `.env`; `iam-policy.json` ist bis zu einer Anforderung zur Prüfung eigenständiger Policy-Dateien ausgenommen (Präzisierung im Lastenheft) |
| Nutzerdokumentation vollständig, `releasing.md` mit realem Ablauf | erfüllt (Regelkatalog, Referenz, Handbuch, Schnellstart, Fehlersuche, CI, Docker, Devcontainer, AWS-Rechte, Release-Ablauf) |
| Release-Tag `v1.0.0` nach ausdrücklicher Freigabe beider Nutzer | erfüllt |

## Freigabe

Beide Nutzer haben im Wortlaut der Entscheidungsvorlage zugestimmt (festgehalten am 2026-11-11):

> Ja zu Release 1.0.0 auf Commit 5f1b35cca776 für die Ziele Tag + GitHub-Release, Docker-Images nach ghcr.io, Devcontainer-Feature

Das Ja des Nutzers von w1 liegt im Dialog vor. Das Ja des Nutzers von w2 hat w2 gemeldet (per Antwort auf eine Anfrage von w1); w1 hat es nicht selbst gesehen. w1 hat von der Ausführung erst im Nachhinein über GitHub erfahren (Nachrichten von w2 kamen bei w1 nicht an) und das Ergebnis danach unabhängig nachgeprüft (siehe Steering-Loop).

## Veröffentlicht

| Ziel | Ergebnis |
|---|---|
| Git-Tag | `v1.0.0` (annotiert) auf `5f1b35cca776` |
| GitHub-Release | `1.0.0` mit `bedrock-eu-check_1.0.0_linux_amd64`, `…_linux_arm64` und `SHA256SUMS` |
| Docker-Images | `ghcr.io/pt9912/bedrock-eu-check:1.0.0-slim` (`sha256:82c14361…`) und `:1.0.0-full` (`sha256:c82b6271…`) |
| Devcontainer-Feature | `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check:1` (`sha256:cd61b5bc…`) |

Prüfsummen der Binaries:

```text
32928970eca539a319aa2ba3a1f9bbb0662ee9c731f93ebae16dfd72da4f046f  bedrock-eu-check_1.0.0_linux_amd64
8f91f8e78051505ddbb99dfae301fe21bf1ce4900bd0e1f22f4d43a44d3066a0  bedrock-eu-check_1.0.0_linux_arm64
```

**Herkunft der Release-Dateien (ehrlich):** Die Binaries und `SHA256SUMS` im GitHub-Release stammen aus einem **lokalen** `make dist` (reproduzierbar, die Prüfsummen sind identisch zur Vorschau in der Entscheidungsvorlage). Der Workflow `release-artifacts` ist auf GitHub **nie gelaufen** und damit nicht erprobt; nur der Workflow `publish` lief.

**Bekannter Mangel im veröffentlichten Feature:** Die Option-Beschreibung in `devcontainer-feature.json` („der Default zeigt auf ein Release, das noch nicht existiert“) ist im veröffentlichten Feature veraltet, weil das Release inzwischen existiert. Die Korrektur gibt es erst mit der nächsten Feature-Version (Neuveröffentlichung nötig).

**Nachprüfung:** w1 hat das Release frisch heruntergeladen: `sha256sum -c` meldet für beide Dateien OK, `version` meldet `bedrock-eu-check 1.0.0 (5f1b35cca776)`. Die Existenz der Images und des Features auf `ghcr.io` wurde über die Manifeste bestätigt. w2 hat die Images gestartet und getestet; das Feature wurde nur über das Manifest geprüft (nicht in einem Devcontainer installiert).

**Fehllauf:** Der erste Lauf des Workflows `publish` (04:25 UTC) scheiterte, weil das Paket `bedrock-eu-check` schon ohne Zugriff für Actions existierte und der Push verweigert wurde; es wurde nichts veröffentlicht. Nach der Zugriffsfreigabe durch den Nutzer liefen die Wiederholungen grün.

## Entscheidungen der Auftraggeber

1. **`.env`/`iam-policy.json`:** Option 1b — Scanner für `.env` (neue Anforderung [`LH-FA-LENV-007`](../../../../spec/lastenheft.md#lh-fa-lenv-007--prüfung-von-env-dateien), slice-022), `iam-policy.json` bleibt Soll-Anforderung nach 1.0.
2. **Veröffentlichung:** alle drei Ziele, Registry GitHub (`ghcr.io`).
3. **`apiKeyHelper`:** bleibt `WARN`. **`--fail-on warn`:** Folgeslice nach 1.0.

## Lieferumfang

`version`, Self-Check, Architekturprüfung und Doku-Abgleichstest, Docker-Images `slim` und `full`, Devcontainer-Feature, `make dist`, die manuellen Workflows `release-artifacts` und `publish`, vollständige Nutzerdokumentation, der `.env`-Scanner sowie die Vorlauf- und Folgeslices `exclude`, `apiKeyHelper` und „verwendet“ aus Claude-Settings. Das Werkzeug `d-check` wurde nach dem Release auf die aktuelle Version gehoben (PR #82).

## Steering-Loop

1. Nach jedem Merge `main` im sauberen Checkout prüfen; Befehle mit `&&` verketten. Ein Verschiebe-PR hat mit einem anderen PR einen Link gebrochen, ohne dass einer der beiden einzeln rot war; beim Verschieben vorher alle Verweise per `grep` suchen.
2. Aussagen in der Nutzerdoku am echten Artefakt prüfen (zum Beispiel führt ein nicht lesbares Mount im Image zu Exit 2, nicht zu Findings).
3. Außenwirksame Schritte (Tag, Release, Push, Publish) sind von Build-Schritten getrennt und laufen nur manuell nach dem Ja beider Nutzer. Der Workflow `publish` prüft zusätzlich, dass der Tag auf den freigegebenen Commit zeigt.
4. Gemeinsame Konflikte entstehen dort, wo beide Seiten ans Ende derselben Datei anhängen: Änderungen früh abstimmen oder in getrennte Dateien legen.
5. **Nachrichten zwischen den Agenten müssen bestätigt werden.** Das Sendewerkzeug meldet nie den Empfang; Nachrichten gingen verloren (Freigabe, Statusmeldungen), und das Release lief, ohne dass w1 es erfuhr. Seitdem gilt: Anfragen mit Antwort (`request`/`reply`) statt einseitigem `send`, jede Nachricht wird mit einem kurzen ACK bestätigt, und der Zustand wird auf GitHub nachgeprüft statt dem Chat zu glauben.
6. Vor jeder Freigabe den freigegebenen Commit festhalten und nach der Freigabe nichts mehr auf `main` ändern, das den Release-Stand betrifft; später gemergte Änderungen (hier der Workflow `publish`) laufen nur, weil der Tag auf den freigegebenen Commit zeigt.

## Folgeaufgaben

- `--fail-on warn` und `--output FILE` für den CI-Modus; Zusatzregionen für „sichtbare“ Profile; `coverage-gate` mit begründeter Schwelle (eigener ADR).
- Eigenständige IAM-Policy-Dateien (`iam-policy.json`) als Soll-Anforderung mit Scanner.
- Den Workflow `release-artifacts` auf GitHub erproben (bisher nur der lokale `make dist` im Einsatz); die Option-Beschreibung des Devcontainer-Features korrigieren und eine neue Feature-Version veröffentlichen (nur nach Freigabe beider Nutzer).
- Signierte Releases (Sigstore), SBOM, bitgleiche Image-IDs; Erprobung des Feature-Installationswegs und der Workflows gegen die Devcontainer-CLI bzw. GitHub; Plattformen außer Linux.
- Die neuen Prüfziele von `d-check` (`doc-planning`, `doc-targets`, …) sind nicht in `make gates` aufgenommen (Folgeentscheidung).
