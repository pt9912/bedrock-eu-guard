# Welle welle-05-release: Ergebnisse (Zwischenstand)

**Stand:** 2026-11-10. **Status:** Release-Kandidat, **Veröffentlichung ausstehend** (Freigabe beider Nutzer).
**Meilenstein:** M5 — Release 1.0 ist **noch nicht erreicht**: Er gilt erst mit dem tatsächlich freigegebenen und veröffentlichten Release.

## Closure-Trigger

| Kriterium | Stand |
|---|---|
| Alle Slices done (014 bis 021) | erfüllt; die Closure von slice-021 liegt bei w1 |
| `make verify-clean` und CI grün | erfüllt (`gates`, `image`, `deps-verify`) |
| `version` zeigt Version und Commit; SARIF-Tool-Version aus derselben Quelle | erfüllt |
| Image baut aus gepinnten Basen und führt `all --ci` auf einem Beispiel-Repository aus | erfüllt (Smoke-Test je Image, `all --ci` im Test) |
| Self-Check und Architekturprüfung in CI | erfüllt |
| Testdaten decken die Dateiarten aus [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) ab | **teilweise**: für `.env` und `iam-policy.json` gibt es keine Scanner-Anforderung (Frage unten) |
| Nutzerdokumentation vollständig, `releasing.md` mit realem Ablauf | erfüllt |
| Release-Tag `v1.0.0` nach ausdrücklicher Freigabe | **offen** |

## Lieferumfang

`version`, Self-Check, Architekturprüfung und Doku-Abgleichstest, Docker-Images `slim` und `full`, Devcontainer-Feature, `make dist` und der manuelle Workflow `release-artifacts`, vollständige Nutzerdokumentation; dazu die Vorlauf- und Folgeslices `exclude`, `apiKeyHelper` und „verwendet“ aus Claude-Settings.

## Entscheidungen, die bei den Auftraggebern liegen

1. **`.env` und `iam-policy.json`** ([`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten)): Das Lastenheft verlangt Testdaten für beide Dateiarten, beschreibt aber keine Scanner dafür. Entweder Anforderungen und Scanner ergänzen oder die Testdaten-Anforderung präzisieren.
2. **Freigabe zur Veröffentlichung:** Tag `v1.0.0`, GitHub-Release mit Artefakten, Image-Push (Registry und Namen festlegen) und Veröffentlichung des Devcontainer-Features. Nichts davon geschieht automatisch.
3. **Optionale Verschärfungen:** `apiKeyHelper` als `FAIL` statt `WARN` ([`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings)) und eine Schwelle `--fail-on warn`.

## Steering-Loop

1. Nach jedem Merge `main` im sauberen Checkout prüfen; Befehle mit `&&` verketten. Ein Verschiebe-PR hat mit einem anderen PR einen Link gebrochen, ohne dass einer der beiden einzeln rot war.
2. Aussagen in der Nutzerdoku am echten Artefakt prüfen (ein nicht lesbares Mount im Image führt zu Exit 2, nicht zu Findings).
3. Außenwirksame Schritte (Tag, Release, Push, Publish) sind von Build-Schritten getrennt und nie automatisiert.
4. Gemeinsame Konflikte entstehen dort, wo beide Seiten ans Ende derselben Datei anhängen (Tests, Tabellen): Änderungen früh abstimmen oder in getrennte Dateien legen.

## Folgeaufgaben

- Qualitätsrunde: `--fail-on warn`, `--output FILE`, Zusatzregionen für „sichtbare“ Profile, `coverage-gate` mit begründeter Schwelle (eigener ADR).
- Signierte Releases (Sigstore), SBOM, bitgleiche Image-IDs, Erprobung von Workflows und Feature gegen GitHub und die Devcontainer-CLI.
- Entscheidung zu `.env`/`iam-policy.json` abwarten; danach gegebenenfalls Scanner-Slice.
