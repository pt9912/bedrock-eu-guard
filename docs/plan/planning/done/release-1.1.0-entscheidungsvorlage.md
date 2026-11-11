# Entscheidungsvorlage: Release 1.1.0 und Feature 1.0.1

**Status:** abgeschlossen: beide Entscheidungen sind von beiden Nutzern getroffen und umgesetzt (siehe Abschnitt 10). **Stand:** 2026-11-11.

**Zweck:** Eine gemeinsame Vorlage für beide Nutzer. Sie bereitet **zwei getrennte Entscheidungen** vor; es wird
nichts veröffentlicht, bevor beide Nutzer der jeweiligen Entscheidung ausdrücklich zugestimmt haben.
Versionsnummern nach 1.0 sind projektintern (das Lastenheft kennt Versionen nur bis 1.0).

## 1. Kurzfassung

| # | Entscheidung | Inhalt | Empfehlung |
|---|---|---|---|
| A | Werkzeug **1.1.0** veröffentlichen | Tag `v1.1.0`, GitHub-Release mit Binaries, Images `1.1.0-slim` und `1.1.0-full` auf `ghcr.io` | ja, wenn die Voraussetzung in Abschnitt 5 erfüllt ist; Reihenfolge: Trockenlauf, dann scharf |
| B | **Feature 1.0.1** veröffentlichen | Korrigierte Option-Beschreibung, Tag `feature-v1.0.1`, `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check` | getrennt von A; erst nach dem Trockenlauf des Tags |

Die beiden Entscheidungen sind unabhängig: Das Werkzeug 1.1.0 ändert das Feature nicht, und das Feature 1.0.1 ändert das
Werkzeug nicht (der Standard der Feature-Option `version` bleibt 1.0.0, siehe Abschnitt 6, Frage B2).

## 2. Änderungen seit 1.0.0

**Werkzeug 1.1.0** enthält drei additive Erweiterungen, die jeweils nur auf ausdrückliche Anforderung wirken. Ohne neue Optionen verhält sich das Werkzeug wie 1.0.0.

| Slice | Änderung | Bedienung |
|---|---|---|
| 026 | Exit 1 schon bei `WARN` | `--fail-on fail` (Standard) oder `--fail-on warn`; Exit 2 hat immer Vorrang |
| 027 | Ergebnis in eine Datei statt nach stdout (vollständig oder gar nicht geschrieben) | `--output FILE`; stdout bleibt leer, Meldungen und die Zusammenfassung des CI-Modus gehen nach stderr |
| 028 | Sichtbarkeit nicht erlaubter Profile in Zusatzregionen (nur lesend, `WARN`, nie `FAIL`) | Konfiguration `visibilityRegions`; neuer Grund `VISIBILITY_UNKNOWN`, Konfigurationsfehler `REGION_ALSO_ALLOWED` |

**Kompatibilität.**

- Exit-Codes 0, 1 und 2 sowie das Standardverhalten sind unverändert ([`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes); [`LH-FA-EXIT-001`](../../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung) und [`LH-FA-EXIT-002`](../../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung) bleiben erfüllt). Ein bestehender Aufruf liefert dieselben Befunde und denselben Exit-Code.
- `visibilityRegions` ist optional (Standard leer, keine zusätzlichen AWS-Aufrufe). Die einzige Verhaltensänderung für Bestandsnutzer entsteht nur, wenn der Schlüssel gesetzt ist: Steht eine Region sowohl dort als auch in `allowedRegions`, endet der Lauf mit Exit 2.
- Neue Befund- und Fehlercodes stehen in [`docs/user/rules.md`](../../../user/rules.md) bzw. [`docs/user/reference.md`](../../../user/reference.md); ein Test prüft, dass jeder Code dokumentiert ist.

**Spezifikation und Lastenheft.** Geändert wurden nur [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) (Schlüssel `visibilityRegions`) und [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes) (Schwelle `--fail-on`, Ausgabeziel `--output FILE`) in `spezifikation.md`, jeweils additiv (35 Zeilen hinzugefügt, keine entfernt). Das Lastenheft ist seit `v1.0.0` unverändert (`git diff v1.0.0..HEAD -- spec` zeigt nur `spezifikation.md`).

**Seit 1.0.0 geändert, ohne Wirkung auf das Binary:** der Workflow `release-create` (slice-025; scharfer Lauf noch nicht erprobt), die d-check-Ziele in `make gates` (slice-029), Dokumentation, Review-Reports und Lifecycle-Regeln.

**Nicht enthalten:** das Devcontainer-Feature (siehe Entscheidung B) und welle-08 (Sigstore/SBOM, macOS/Windows, `iam-policy.json`).

## 3. Stand (Messwerte)

Gemessen auf `main`, Commit `af911c900500`, am 2026-11-11, **vor** der Aufnahme dieser Vorlage; der Release-Commit ist ein
späterer (siehe Abschnitt 4).

| Gegenstand | Stand |
|---|---|
| Gates | `make verify-clean` (doc-check, doc-targets, lint, test, 15 Test-Pakete), `make self-check`, `make image-test` und `make feature-test`: grün; CI-Workflows `gates` und `image` grün |
| Binaries (Vorschau `make dist VERSION=1.1.0`) | Linux `amd64` 5,19 MB und `arm64` 4,92 MB, statisch gelinkt |
| Docker-Images | `slim` 7,41 MB (`local`, `terraform`), `full` 659 MB (mit AWS-CLI, alle Modi); beide bestehen die Smoke-Tests |
| Feature | Version 1.0.1 im Repository (korrigierte Option-Beschreibung); veröffentlicht ist 1.0.0 |
| Tags / Releases | `v1.0.0` und das Release 1.0.0 |

**Vorschau der Prüfsummen** (`make dist VERSION=1.1.0 COMMIT=af911c900500`, nur Beleg der Reproduzierbarkeit):

```text
fbbbb54e19a13cb5827c1699276b8bfdd7c33ab03e17dd884aa98f1dd414e692  bedrock-eu-check_1.1.0_linux_amd64
5025d54f3a19e946ce139e9e5f2eef4a38b2372d9bf18fb9641c9d152d2c79a8  bedrock-eu-check_1.1.0_linux_arm64
```

Die Prüfsummen hängen von Version **und Commit** ab; **maßgeblich** sind die Werte des Release-Commits aus dem Trockenlauf
von `release-create` (Abschnitt 5, Schritt 2).

## 4. Der freigegebene Commit

Der Release-Commit steht erst fest, wenn diese Vorlage gemerged ist: Jede Änderung auf `main`, auch reine Dokumentation, ändert
den Commit. Der Ablauf wie bei 1.0.0: Die Vorlage wird gemerged, dann nennen wir den Commit (`main` nach dem Merge) **ausdrücklich im
Wortlaut der Freigabe**, und danach ändert niemand mehr `main`, bis die Veröffentlichung abgeschlossen ist.

## 5. Entscheidung A: Werkzeug 1.1.0

**Voraussetzung (Aufgabe der Nutzer, nicht der Agenten):** In den Repository-Einstellungen unter *Settings → Environments*
die Umgebung `release` mit *Required reviewers* anlegen. Sie schützt den ersten scharfen Lauf von `release-create`; ohne sie
legt GitHub die Umgebung ungeschützt an, und jede Person mit Schreibrecht könnte den Lauf starten.

**Ablauf** (jeder Schritt nennt den Beleg; erst Trockenlauf, dann scharf; Vorbild [`releasing.md`](../../../user/releasing.md)):

1. **Stand prüfen** (sauberer Checkout des Commits): `make verify-clean`, `make self-check`, `make deps-verify`,
   `make image-test`, `make feature-test`; CI auf dem Commit grün; `LICENSE` und `THIRD_PARTY_NOTICES.md` passen zum Stand.
2. **Trockenlauf von `release-create`:** `target=tool`, `version=1.1.0`, `commit=<SHA>`, `dry_run=true` (Standard). Er prüft Version,
   Commit (Vorfahre von `main`) und dass der Tag frei ist, baut mit `make dist` und zeigt die Prüfsummen; er schreibt nichts.
   Die Prüfsummen vergleichen wir mit einem zweiten Lauf von `release-artifacts` auf demselben Commit.
3. **Trockenlauf von `publish`:** `images=true`, `dry_run=true`: baut und testet beide Images, pusht nichts.
4. **Freigabe einholen** (Wortlaut siehe Abschnitt 7) und festhalten.
5. **Scharfer Lauf von `release-create`** (`dry_run=false`): legt den annotierten Tag `v1.1.0` und das GitHub-Release mit Binaries und
   `SHA256SUMS` an. *Dies ist die erste scharfe Nutzung des Workflows; sie ist noch nicht erprobt.*
6. **Nachprüfen:** Artefakte frisch herunterladen, `sha256sum -c`, `bedrock-eu-check version` meldet `1.1.0` und den Commit.
7. **Scharfer Lauf von `publish`** (`images=true`, `dry_run=false`): pusht `1.1.0-slim` und `1.1.0-full`. Danach die Images anonym
   ziehen, starten und `version` prüfen.
8. **Dokumentation und Roadmap:** README und `releasing.md` auf „1.1.0 veröffentlicht“, Meilenstein M7, `welle-07` schließen (macht w1).

Alle Veröffentlichungsschritte sind **nicht rückholbar** (Tag, Release, Images; der Tag ist über den Go-Proxy abrufbar).

## 6. Entscheidung B: Feature 1.0.1

| Frage | Optionen | Empfehlung |
|---|---|---|
| B1: Feature 1.0.1 veröffentlichen? | ja / später | ja, nach A; die veraltete Option-Beschreibung im veröffentlichten Feature wird korrigiert |
| B2: Standard der Option `version` des Features | bleibt `1.0.0` / auf `1.1.0` anheben | bleibt `1.0.0` (eine Änderung des Standards ist eine Verhaltensänderung für alle, die das Feature ohne Version einbinden; sie wäre mit eigener Feature-Version nachzuholen) |

**Ablauf:** (1) `release-create` mit `target=feature`, `version=1.0.1`, `commit=<SHA>` im Trockenlauf: prüft, dass der Tag
`feature-v1.0.1` frei ist und die Version in `devcontainer-feature.json` passt; (2) nach der Freigabe scharf: legt den Tag
`feature-v1.0.1` an; (3) `publish` mit `feature=true`, `ref=feature-v1.0.1`, `version=1.0.1` im **Trockenlauf** (packt das Feature, pusht nichts;
dies ist der erste Lauf des `ref`-Pfads, er war vorher ohne Tag nicht möglich); (4) scharf; (5) `feature-cli` Fall B per Dispatch
prüft das veröffentlichte Feature, danach ist `1.0.1` auch anonym ladbar.

## 7. Wortlaut der Freigabe

Jeder Nutzer antwortet getrennt, je Entscheidung, mit Version, Commit und Zielen:

- A: „Ja zu Release 1.1.0 auf Commit `<SHA>` für die Ziele Tag + GitHub-Release, Docker-Images nach ghcr.io“
- B: „Ja zu Feature 1.0.1 auf Commit `<SHA>` für das Ziel Devcontainer-Feature nach ghcr.io“

Die Freigabe zu A deckt nicht B ab und umgekehrt. Ein scharfer Lauf braucht beide Nutzer.

## 8. Nicht erprobt und Risiken

- **Scharfer Lauf von `release-create`** (Tag und Release aus CI): nur im Trockenlauf erprobt; Fehler (zum Beispiel Rechte für den Tag-Push,
  Branch-Schutz) zeigen sich erst scharf. Der Tag-Push ohne Force scheitert, statt zu überschreiben.
- **`ref`-Pfad von `publish`** (Feature-Tag): erst mit existierendem Tag im Trockenlauf prüfbar.
- **Umgebung `release`:** Ihr Schutz hängt an einer Einstellung der Nutzer; sie ersetzt das Ja beider Nutzer nicht.
- **`devcontainer up`** blieb auf dem GitHub-Runner hängen (Ursache ungeklärt); erprobt ist `devcontainer build` mit `docker run`.
- **Nicht Teil dieser Entscheidungen:** der Scanner für eigenständige Policy-Dateien, signierte Releases und weitere Plattformen (Welle 08).

## 9. Antwortblatt

| # | Frage | Antwort |
|---|---|---|
| A | Werkzeug 1.1.0 veröffentlichen (Ziele Tag + Release, Images)? | |
| B1 | Feature 1.0.1 veröffentlichen? | |
| B2 | Standard der Feature-Option `version` | |
| C | Umgebung `release` mit Pflicht-Prüfern angelegt? | |

## 10. Ergebnis (Stand 2026-11-11)

Beide Nutzer haben beiden Entscheidungen auf Commit `5274f3dce55f4423e2937978d3c4d05c888a13cb` zugestimmt, getrennt je Entscheidung:

- **A:** „Ja zu Release 1.1.0 auf Commit 5274f3dce55f für die Ziele Tag + GitHub-Release, Docker-Images nach ghcr.io“ (Nutzer von w2, wörtlich; Nutzer von w1: „Ja zu Release 1.1.0 auf Commit 5274f3dce55f“, Ziele laut Rückfrage Tag + GitHub-Release und Docker-Images nach ghcr.io).
- **B:** „Ja zu Feature 1.0.1 auf Commit 5274f3dce55f für das Ziel Devcontainer-Feature nach ghcr.io“ (Nutzer von w2, wörtlich; Nutzer von w1: „Ja zu Feature 1.0.1 auf Commit 5274f3dce55f“, Ziel Devcontainer-Feature).
- **Voraussetzung:** Die Umgebung `release` mit Pflicht-Prüfer wurde vor dem ersten scharfen Lauf angelegt; die Freigabe der scharfen Läufe erteilte der Nutzer in GitHub.

| Schritt | Beleg |
|---|---|
| Trockenlauf `release-create` (tool, 1.1.0) | Run 42989181539 |
| `release-artifacts` auf dem Commit, Prüfsummen bitgleich zum lokalen Build | Run 42989306278 |
| `release-create` scharf: Tag `v1.1.0` und GitHub-Release | Run 42990440491 |
| `publish` Images, Trockenlauf und scharf | Runs 42993109110 und 42993271052; `1.1.0-slim` `sha256:224b9dbe8a616feb3cda03b3827778edf9855c376983fa15ba1c087cfc74a057`, `1.1.0-full` `sha256:0792b9ac0896c36a93e9f4a32281c15de17e7d315b9f4304d1bdb5da98ece578` |
| `release-create` feature, Trockenlauf und scharf: Tag `feature-v1.0.1` | Runs 42993606450 und 42993689182 |
| `publish` Feature, Trockenlauf und scharf | Runs 42996795891 und 42996881739; Digest `sha256:34aa4f35d7704d07c21e4955ffc343091f679df267da91e6472139190e47f37a` |
| Nachprüfung `feature-cli` | Run 42996952937 |

Beide Agenten haben die Ergebnisse unabhängig nachgeprüft (frischer Download, `sha256sum -c`, anonymer Pull und Start der Images, Manifeste und Metadaten des Features). Die Abschnitte 3 bis 9 zeigen den Stand zum Zeitpunkt der Vorlage.
