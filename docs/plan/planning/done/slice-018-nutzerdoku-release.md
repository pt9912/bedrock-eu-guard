# Slice slice-018: Nutzerdokumentation, Release-Artefakte und Release-Ablauf

**Status:** done.

**Welle:** welle-05-release.

**Bezug:** [`LH-RM-005`](../../../../spec/lastenheft.md#lh-rm-005--version-10), [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`LH-TECH-004`](../../../../spec/lastenheft.md#lh-tech-004--docker-image), [`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Eine vollständige Nutzerdokumentation und ein reproduzierbarer, dokumentierter
Release-Ablauf. Der Ablauf **baut** Artefakte, **veröffentlicht aber nichts**:
Tag, GitHub-Release, Image-Push und Feature-Publish sind außenwirksam und nur nach
ausdrücklicher Freigabe beider Nutzer erlaubt.

## 2. Verträge

- **Artefakte:** `make dist VERSION=x.y.z` baut statisch gelinkte Binaries für Linux (`amd64`, `arm64`) nach `dist/v<version>/bedrock-eu-check_<version>_linux_<arch>` und erzeugt `SHA256SUMS` (`<hash>  <datei>`), das direkt verifiziert wird. Die Namen sind der Vertrag mit dem Devcontainer-Feature.
- **Workflow `release-artifacts`:** nur `workflow_dispatch` (Eingabe: Version, validiert), nur lesende Rechte, ruft `make dist` und legt das Ergebnis als Workflow-Artefakt ab. Kein Tag, kein Release, kein Push.
- **Nutzerdokumentation** unter `docs/user/`: Schnellstart (`quickstart.md`), Referenz (`reference.md`), Regelkatalog (`rules.md`), CI (`ci.md`), Docker (`docker.md`), Devcontainer (`devcontainer.md`), AWS-Rechte (`aws-permissions.md`), Fehlersuche (`troubleshooting.md`) und Release-Ablauf (`releasing.md`). Die Dokumente verlinken einander statt Inhalte zu duplizieren.
- **Drift-Schutz:** Regelkatalog und Referenz werden per Test gegen die Reason-Codes und Regel-IDs im Code abgeglichen.

## 3. Definition of Done

- [x] Alle genannten Dokumente vorhanden, `make doc-check` grün.
- [x] `make dist` baut beide Architekturen, verifiziert `SHA256SUMS`; zwei Läufe liefern identische Prüfsummen (Ergebnis in der Closure-Notiz).
- [x] Workflow `release-artifacts` vorhanden, nur manuell, ohne Schreibrechte und ohne Veröffentlichungsschritt.
- [x] `releasing.md` beschreibt den realen Ablauf einschließlich der Freigabepflicht und der nicht automatisierten Schritte.
- [x] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag; Welle-5-Abschluss (M5) erst nach Freigabe zur Veröffentlichung.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `docs/user/rules.md`, `docs/user/reference.md`, Doku-Drift-Test | neu | w1 |
| `docs/user/quickstart.md`, `troubleshooting.md`, `releasing.md` | neu/update | w2 |
| `Makefile` (`dist`), `.github/workflows/release.yml` | neu/update | w2 |

## 5. Trigger

[`slice-014`](../done/slice-014-version-befehl.md) bis [`slice-017`](../done/slice-017-devcontainer-feature.md) gemerged.

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben. Der Meilenstein M5 (Release 1.0) gilt
erst mit dem freigegebenen, tatsächlichen Release als erreicht.

## 7. Risiken und offene Punkte

- **Nicht erprobt:** Der Workflow `release-artifacts` und der Beispiel-Workflow in `ci.md` laufen erst bei Bedarf auf GitHub; lokal wurde `make dist` getestet, das YAML nur syntaktisch geprüft.
- **Plattformen:** nur Linux; macOS/Windows sind nicht Teil von 1.0.
- **Veröffentlichung:** Registry-Namen für Image und Feature sind noch nicht festgelegt.
- **Vertrauensmodell:** Prüfsummen schützen vor Korruption, nicht vor einem kompromittierten Ursprung; Signaturen (Sigstore) und SBOM sind Folgepunkte.

## 8. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-10 (PRs #61/#62 Regelkatalog, Referenz, Doku-Abgleichstest/w1; #59 Docker-Doku, #65 Release/Schnellstart/Fehlersuche/w2).

- **Funktioniert:** `docs/user/` ist vollständig: `quickstart.md`, `reference.md`, `rules.md`, `ci.md`, `docker.md`, `devcontainer.md`, `aws-permissions.md`, `troubleshooting.md`, `releasing.md`. Ein Test gleicht Regelkatalog und Referenz mit den Reason-Codes und Regel-IDs im Code ab. `make dist VERSION=x.y.z` baut statisch gelinkte Linux-Binaries (`amd64`, `arm64`) mit `SHA256SUMS` nach `dist/v<version>/`, verifiziert die Prüfsummen und liefert bei zwei Läufen identische Prüfsummen. Der Workflow `release-artifacts` läuft nur manuell, mit lesenden Rechten, und **veröffentlicht nichts**; `releasing.md` beschreibt den realen Ablauf einschließlich Freigabe beider Nutzer und der manuellen, außenwirksamen Schritte.
- **Anders als geplant:** Beim Zusammenführen kollidierte das Verschieben von Slice-Dateien (`open/` → `done/`) mit einem Link im Plan dieses Slices: Beide PRs waren einzeln grün, `main` war kurz rot (Hotfix #68).
- **Offen:** Der Workflow `release-artifacts` und der Beispiel-Workflow in `ci.md` sind nicht auf GitHub gelaufen. Veröffentlichung jeder Art (Tag, GitHub-Release, Image-Push, Feature-Publish) steht aus und braucht die ausdrückliche Freigabe beider Nutzer; Registry-Namen sind nicht festgelegt. Nur Linux; Signaturen und SBOM sind Folgepunkte.
- **Steering-Loop:** Nach **jedem** Merge `main` im sauberen Checkout prüfen und Befehle mit `&&` verketten, nie mit `;`. Beim Verschieben einer Datei alle Verweise darauf per `grep` nachziehen. Dokumentation, die Code beschreibt, per Test gegen den Code abgleichen.
- **Folge-Slices:** keine; weiter mit der Freigabe zum Release und den Folgeaufgaben der Qualitätsrunde (siehe `welle-05-results.md`).

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
