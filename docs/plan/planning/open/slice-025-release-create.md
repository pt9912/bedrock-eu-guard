# Slice slice-025: Workflow `release-create` für Tag und GitHub-Release

**Status:** open.

**Welle:** [`welle-06-release-haertung`](../welle-06-release-haertung.md).

**Bezug:** [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`LH-RM-005`](../../../../spec/lastenheft.md#lh-rm-005--version-10), `docs/user/releasing.md`.

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Ein neuer Workflow `release-create` legt Tag und GitHub-Release aus CI an, statt lokal. Er läuft **nur** per `workflow_dispatch`, nie auf Push, Merge oder Zeitplan, und hat einen Trockenlauf als Standard.

## 2. Definition of Done

- [ ] Eingaben: `version` (x.y.z), `commit` (mindestens 12 Hex-Zeichen), `dry_run` (Standard: wahr).
- [ ] Prüft Version und Commit; der Commit muss auf `main` liegen und dem ausgecheckten HEAD entsprechen; der Tag `v<version>` darf noch nicht existieren.
- [ ] Trockenlauf: baut mit `make dist`, prüft `SHA256SUMS`, gibt Namen und Prüfsummen aus; schreibt nichts.
- [ ] Scharfer Lauf (`dry_run=false`): legt den annotierten Tag auf dem Commit an und das Release mit Binaries und `SHA256SUMS`; `contents: write` nur im Job, oben `permissions: {}`.
- [ ] Im Trockenlauf erprobt (Run-ID in der Closure-Notiz); der scharfe Lauf geschieht erst bei der nächsten freigegebenen Veröffentlichung.
- [ ] `releasing.md` beschreibt den Ablauf mit dem Workflow; das Ja beider Nutzer zur Version, zum Commit und zu den Zielen bleibt Voraussetzung.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.github/workflows/release-create.yml` | neu | Tag und Release reproduzierbar aus CI; Actions per SHA gepinnt (Vorbild `publish.yml`) |
| `docs/user/releasing.md` | update | Ablauf Schritt 4 mit dem Workflow; Freigabepflicht unverändert |

## 4. Trigger

slice-023 done (Erkenntnisse zum Build im Workflow). Die neue Veröffentlichungsfähigkeit selbst braucht das Ja beider Nutzer (Schreibrecht für Tags und Releases), bevor der Workflow gemerged wird.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Ein Schreibrecht (`contents: write`) im Workflow ist eine neue Fähigkeit; bewusst nur im Job und nur per Dispatch.
- Ein scharfer Lauf ist nicht rückholbar (Tag, Release, Go-Proxy); der Standard `dry_run` und die Prüfungen sollen Fehlbedienung abfangen.
- Branch-Schutz oder Berechtigungen im Repository können den Tag-Push verhindern; Befund an die Nutzer, nicht umgehen.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
