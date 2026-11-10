# Slice slice-018: Nutzerdokumentation, Release-Artefakte und Release-Ablauf

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

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

- [ ] Alle genannten Dokumente vorhanden, `make doc-check` grün.
- [ ] `make dist` baut beide Architekturen, verifiziert `SHA256SUMS`; zwei Läufe liefern identische Prüfsummen (Ergebnis in der Closure-Notiz).
- [ ] Workflow `release-artifacts` vorhanden, nur manuell, ohne Schreibrechte und ohne Veröffentlichungsschritt.
- [ ] `releasing.md` beschreibt den realen Ablauf einschließlich der Freigabepflicht und der nicht automatisierten Schritte.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; Welle-5-Abschluss (M5) erst nach Freigabe zur Veröffentlichung.

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

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
