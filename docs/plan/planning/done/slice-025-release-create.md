# Slice slice-025: Workflow `release-create` für Tag und GitHub-Release

**Status:** done.

**Welle:** [`welle-06-release-haertung`](../welle-06-release-haertung.md).

**Bezug:** [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`LH-RM-005`](../../../../spec/lastenheft.md#lh-rm-005--version-10), `docs/user/releasing.md`.

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Ein neuer Workflow `release-create` legt Tag und GitHub-Release aus CI an, statt lokal. Er läuft **nur** per `workflow_dispatch`, nie auf Push, Merge oder Zeitplan, und hat einen Trockenlauf als Standard.

## 2. Definition of Done

- [x] Eingaben: `version` (x.y.z), `commit` (mindestens 12 Hex-Zeichen), `dry_run` (Standard: wahr).
- [x] Prüft Version und Commit: Checkout mit `ref=<commit>`; der Commit muss ein Vorfahre von `main` sein (auch ein älterer freigegebener Commit ist zulässig, nicht nur der Kopf von `main`); der Tag darf noch nicht existieren. Neben `v<version>` (Werkzeug) unterstützt der Workflow den Feature-Tag `feature-v<x.y.z>` (siehe slice-024).
- [x] Trockenlauf: baut mit `make dist`, prüft `SHA256SUMS`, gibt Namen und Prüfsummen aus; schreibt nichts.
- [x] Scharfer Lauf (`dry_run=false`): legt den annotierten Tag auf dem Commit an und das Release mit Binaries und `SHA256SUMS`; `contents: write` nur im Job, oben `permissions: {}`.
- [x] Im Trockenlauf erprobt (Run-ID in der Closure-Notiz); der scharfe Lauf geschieht erst bei der nächsten freigegebenen Veröffentlichung.
- [x] Optionale technische Sicherung geprüft: GitHub-Umgebung (`environment`) mit Required Reviewers für den scharfen Lauf. Sie erzwingt eine zweite Bestätigung in GitHub, kann aber das „Ja beider Nutzer“ nicht prüfen (Reviewer sind GitHub-Konten, die Einstellung nimmt der Nutzer im Repository vor). Entscheidung und Begründung in der Closure-Notiz.
- [x] `releasing.md` beschreibt den Ablauf mit dem Workflow; das Ja beider Nutzer zur Version, zum Commit und zu den Zielen bleibt Voraussetzung.
- [x] `make verify-clean` und CI grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

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

- **Funktioniert:** Der Workflow `release-create` (PR #92, Korrektur #93) ist gemergt. Trockenlauf auf GitHub: Run 42972279064 (`target=tool`, Version 1.0.1, Commit `5f1b35cca776`) war erfolgreich: Eingaben geprüft, Commit ist Vorfahre von `main`, Tag `v1.0.1` frei, `make dist` baut beide Binaries, `sha256sum -c` OK; der Job `create` wurde übersprungen. Es entstand kein Tag und kein Release (Tags und Releases vorher und nachher unverändert).
- **Gegenproben:** Version 1.0.0 (Tag existiert) bricht ab (Run 42972414227, „Tag v1.0.0 existiert bereits“); ein unbekannter Commit bricht ab (Run 42972444200).
- **Anders als geplant:** Der erste Trockenlauf (Run 42972041118) scheiterte, weil `actions/checkout` keine verkürzte SHA als `ref` akzeptiert; behoben durch Checkout von `main` mit voller Historie und lokalen Wechsel auf den Commit (#93). Außerdem trägt der Workflow die Tag-frei-Prüfung im Job `create` ein zweites Mal (Race zwischen `check` und `create`).
- **Nicht erprobt:** Der scharfe Lauf (`dry_run=false`), die Umgebung `release` und das Ziel `feature` (Tag `feature-v<version>`) wurden nicht ausgeführt; dafür braucht es das Ja beider Nutzer zu Version, Commit und Ziel. Vor dem ersten scharfen Lauf legen die Nutzer die Umgebung `release` mit Pflicht-Prüfern in den Repository-Einstellungen an (siehe `releasing.md`); sonst wird sie ungeschützt erzeugt.
- **Steering-Loop:** Eine neue Workflow-Fähigkeit erst im Trockenlauf auf der echten Plattform erproben, bevor sie gebraucht wird; Eingaben, die Nutzer abkürzen (SHA), gegen die Eigenheiten der Plattform-Actions prüfen. Gegenproben (Tag vorhanden, Commit unbekannt) gehören zur Erprobung.

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
