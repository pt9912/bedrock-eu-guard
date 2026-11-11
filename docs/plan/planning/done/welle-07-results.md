# Welle welle-07-ci-komfort: Ergebnisse

**Abgeschlossen:** 2026-11-11. **Status:** done; Werkzeug 1.1.0 **veröffentlicht**.
**Meilenstein:** M7 — CI-Komfort (1.1.0) **erreicht** (2026-11-11).

## Closure-Trigger

| Kriterium | Stand |
|---|---|
| Alle Slices done (026 bis 029) | erfüllt |
| `make verify-clean` und CI grün | erfüllt (Stand `5274f3dce55f`) |
| Exit-Codes und Standardverhalten unverändert ([`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes)) | erfüllt (Regressionstests ohne Option; [`LH-FA-EXIT-001`](../../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung) und [`LH-FA-EXIT-002`](../../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung) bleiben erfüllt) |
| Spezifikation angepasst, Lastenheft unverändert | erfüllt ([`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes); `git diff v1.0.0..HEAD -- spec` zeigt nur `spezifikation.md`) |
| Veröffentlichung nur nach dem Ja beider Nutzer | erfüllt (siehe Freigabe) |

## Freigabe

Beide Nutzer haben im Wortlaut der Entscheidungsvorlage ([`release-1.1.0-entscheidungsvorlage.md`](../in-progress/release-1.1.0-entscheidungsvorlage.md)) zugestimmt (2026-11-11):

> A: Ja zu Release 1.1.0 auf Commit 5274f3dce55f für die Ziele Tag + GitHub-Release, Docker-Images nach ghcr.io

> B: Ja zu Feature 1.0.1 auf Commit 5274f3dce55f für das Ziel Devcontainer-Feature nach ghcr.io

Das Ja des Nutzers von w1 liegt im Dialog vor. Das Ja des Nutzers von w2 hat w2 wörtlich gemeldet; w1 hat es nicht selbst gesehen. Die Freigabe der Umgebung `release` für `release-create` erteilte ein Mensch in der Weboberfläche (Prüfer `pt9912`, siehe Welle 6).

## Veröffentlicht

| Ziel | Ergebnis |
|---|---|
| Git-Tag | `v1.1.0` (annotiert) auf `5274f3dce55f4423e2937978d3c4d05c888a13cb`, `release-create` scharf: Run 42990440491 (Trockenlauf: Run 42989181539) |
| GitHub-Release | `1.1.0` mit `bedrock-eu-check_1.1.0_linux_amd64`, `…_linux_arm64` und `SHA256SUMS` |
| Docker-Images | `ghcr.io/pt9912/bedrock-eu-check:1.1.0-slim` (`sha256:224b9dbe…a057`) und `:1.1.0-full` (`sha256:0792b9ac…e578`); `publish` Trockenlauf Run 42993109110, scharf Run 42993271052 |

Prüfsummen der Binaries (identisch zur Vorschau und zum Lauf von `release-artifacts`, Run 42989306278):

```text
fb7272144d3575d5a6e12a4844ad7770488f00156d2adcf4152213245976ce6e  bedrock-eu-check_1.1.0_linux_amd64
2475305aa24fcb0a3ebece98cf5df8e83f13b136488fc3dd2341e9f9aa060db2  bedrock-eu-check_1.1.0_linux_arm64
```

**Nachprüfung durch w1 (unabhängig):** Release frisch heruntergeladen, `sha256sum -c` OK für beide Dateien, `bedrock-eu-check version` meldet `bedrock-eu-check 1.1.0 (5274f3dce55f)`. Beide Images anonym von `ghcr.io` gezogen und gestartet: Digests wie gemeldet, Nutzer 65532, Labels `version=1.1.0` und `revision=5274f3dce55f…`, `version` meldet dieselbe Zeile; `1.0.0-slim` ist unverändert (`sha256:82c14361…`).

## Lieferumfang

| Slice | Änderung | Bedienung |
|---|---|---|
| 026 | Exit 1 schon bei `WARN` | `--fail-on fail` (Standard) oder `--fail-on warn`; Exit 2 hat Vorrang |
| 027 | Ergebnis in eine Datei (vollständig oder gar nicht geschrieben) | `--output FILE`; Schreibfehler nennen nur die Fehlerklasse (PR #103) |
| 028 | Sichtbarkeit nicht erlaubter Profile in Zusatzregionen, nur `WARN` | Konfiguration `visibilityRegions`; Grund `VISIBILITY_UNKNOWN`, Konfigurationsfehler `REGION_ALSO_ALLOWED` |
| 029 | Gate-Bewertung der d-check-Ziele | `doc-targets` in `make gates` aufgenommen; `planning` bleibt als Folgepunkt |

## Offen

1. **`planning` als Gate** (d-check-Modul): als Folgeslice vorgemerkt; die Konvention (Block „Aktuelle Welle“) wird bereits eingehalten.
2. Folgepunkte aus der Roadmap (welle-08): `iam-policy.json`-Scanner, `coverage-gate`, Sigstore/SBOM, macOS/Windows; je nach Entscheidung der Auftraggeber.
3. Offene Review-Beobachtungen zu PR #96 (F-1: Temp-Datei im Zielverzeichnis bleibt bei hartem Abbruch liegen; F-2: Datei mit Teilergebnis bei Exit 2, in [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes) festgelegt) bleiben bekannte Grenzen.

## Steering-Loop

Ein Release ist nur so gut wie die unabhängige Nachprüfung: w1 hat Tag, Artefakte, Digests und Images selbst gegen GitHub und `ghcr.io` geprüft, statt sich auf Meldungen zu stützen. Und der Lifecycle (`open → next → in-progress → done`) mit getrennten Move-Commits hat ab slice-024 gehalten, weil die Roadmap den Ruhezustand mit einem eigenen Marker belegt.
