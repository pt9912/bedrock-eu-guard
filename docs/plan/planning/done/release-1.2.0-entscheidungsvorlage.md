# Entscheidungsvorlage: Release 1.2.0

**Status:** Entwurf zur Entscheidung durch beide Nutzer. **Stand:** 2026-11-11.

**Zweck:** Eine gemeinsame Vorlage für beide Nutzer. Sie bereitet **eine Veröffentlichung** und **eine Grundsatzfrage** vor; es wird
nichts veröffentlicht, bevor beide Nutzer der Veröffentlichung ausdrücklich zugestimmt haben. Versionsnummern nach 1.0 sind
projektintern (das Lastenheft kennt Versionen nur bis 1.0).

## 1. Kurzfassung

| # | Entscheidung | Inhalt | Empfehlung |
|---|---|---|---|
| A | Werkzeug **1.2.0** veröffentlichen | Tag `v1.2.0`, GitHub-Release mit Binaries, Images `1.2.0-slim` und `1.2.0-full` auf `ghcr.io` | ja, nach Entscheidung C; Ablauf wie bei 1.1.0 (Trockenläufe, dann scharf) |
| C | Neue Prüfung **ohne Opt-in** für Bestandsnutzer | Eigenständige IAM-Policy-Dateien werden im Modus `terraform` und `all` immer geprüft | ja, mit deutlichem Hinweis in den Release-Notizen (siehe Abschnitt 5) |

Das Devcontainer-Feature ändert sich nicht (Feature 1.0.1 bleibt veröffentlicht, der Standard der Option `version` bleibt 1.0.0);
eine Anhebung dieses Standards ist **nicht Teil** dieser Vorlage.

## 2. Änderungen seit 1.1.0

**Werkzeug 1.2.0** bringt gegenüber 1.1.0 eine neue Prüfung und eine Angleichung an das Handbuch:

| Änderung | Wirkung | Bedienung |
|---|---|---|
| Eigenständige IAM-Policy-Dateien ([`LH-FA-IAM-004`](../../../../spec/lastenheft.md#lh-fa-iam-004--iam-policy-dokumente-json-prüfen), slice-030) | Der Modus `terraform` (und `all`) findet rekursiv `iam-policy*.json` und `*-policy.json` und bewertet sie wie [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) und [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen): `FAIL` bei `bedrock:*` oder `InvokeModel` auf `*` ohne Region-Condition, `WARN` mit Region-Condition oder nicht auswertbar, sonst `PASS`. Erkannt werden das Policy-Dokument sowie die Ausgabe von `aws iam get-policy-version` und `get-role-policy`. | Immer aktiv in `terraform` und `all`; ausnehmen mit `exclude` je Pfad. Nicht in `local`. |
| `Allow` nur mit `NotAction` ergibt `WARN` (`IAM_DYNAMIC`) | Gilt für [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen), [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen) und die neue Prüfung; bisher entstand dafür kein Finding, das Handbuch sagte aber `WARN` zu. | Keine; ein `WARN` ändert den Standard-Exit-Code nicht. |

**Fehlerregel der neuen Prüfung:** Eine nicht lesbare oder ungültige Datei mit eindeutigem Namen (`iam-policy*.json`) ist ein `FAIL` (`FILE_UNREADABLE`, `FILE_INVALID`), bei einer Datei, die nur das breite Muster `*-policy.json` trifft, ein `WARN`. Gültiges JSON ohne Policy-Statement oder ohne Bedrock-Aktion erzeugt kein Finding. Dateiinhalte erscheinen nie in der Ausgabe.

**Kompatibilität.**

- Exit-Codes `0`, `1` und `2`, Optionen, Ausgabeformate und die Konfiguration sind unverändert ([`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes) und [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) ohne Änderung seit 1.1.0).
- **Verhaltensänderung für Bestandsnutzer (Entscheidung C):** Ein Repository mit einer eigenständigen Policy-Datei, die Bedrock breit erlaubt, ergibt im Modus `terraform` oder `all` künftig `FAIL` und damit Exit `1`, wo 1.1.0 nichts gefunden hat. Es gibt keinen Schalter zum Abschalten; ausnehmen lässt sich je Pfad mit `exclude`. Ebenso kann ein vorhandenes `NotAction`-Statement neu ein `WARN` erzeugen (kein neuer Exit `1`, außer mit `--fail-on warn`).
- Neue Befunde tragen die Regel-ID [`LH-FA-IAM-004`](../../../../spec/lastenheft.md#lh-fa-iam-004--iam-policy-dokumente-json-prüfen) und vorhandene Reason-Codes; es gibt keinen neuen Reason-Code und kein neues Konfigurationsfeld.

**Spezifikation und Lastenheft.** Geändert wurde nur das Lastenheft: die neue Anforderung [`LH-FA-IAM-004`](../../../../spec/lastenheft.md#lh-fa-iam-004--iam-policy-dokumente-json-prüfen) und [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) (Testdaten für `iam-policy.json`), jeweils mit dem von beiden Nutzern bestätigten Wortlaut. `spec/spezifikation.md` ist seit `v1.1.0` unverändert (`git diff v1.1.0..HEAD -- spec` zeigt nur `lastenheft.md`).

**Seit 1.1.0 geändert, ohne Wirkung auf das Binary:** das `coverage-gate` in `make gates` (Mindestwert 90 %, [ADR-0005](../../adr/0005-coverage-gate.md)), der Reviewer-Skill `.harness/skills/reviewer.md` und die Review-Reports, die devcontainer-CLI 0.90.0 in `feature-cli` und `publish`, die Korrektur der Prüfsummen-Negativprobe in `tools/feature-test.sh`, Dokumentation und Planung.

**Nicht enthalten:** Änderungen am Devcontainer-Feature und welle-08 jenseits der Slices 030 und 031 (Sigstore/SBOM, macOS/Windows).

## 3. Stand (Messwerte)

Gemessen auf `main`, Commit `a0c1f5e8f6c2`, am 2026-11-11, **vor** der Aufnahme dieser Vorlage; der Release-Commit ist ein späterer
(siehe Abschnitt 4).

| Gegenstand | Stand |
|---|---|
| Gates | `make verify-clean` (doc-check, doc-targets, lint, test mit 16 Test-Paketen, coverage-gate 95,4 % bei Mindestwert 90 %), `make self-check`, `make image-test` und `make feature-test`: grün; CI-Workflows `gates` und `image` grün |
| Binaries (Vorschau `make dist VERSION=1.2.0`) | Linux `amd64` 5,21 MB und `arm64` 4,98 MB, statisch gelinkt |
| Docker-Images | `slim` 7,43 MB (`local`, `terraform`), `full` 659 MB (mit AWS-CLI, alle Modi); beide bestehen die Smoke-Tests |
| Tags / Releases | `v1.0.0`, `v1.1.0`, `feature-v1.0.1`; Releases 1.0.0 und 1.1.0 |

**Vorschau der Prüfsummen** (`make dist VERSION=1.2.0 COMMIT=a0c1f5e8f6c2`, nur Beleg der Reproduzierbarkeit):

```text
7a907117261d46945ff29ccd62283ad68ead5380ab00249c5e1d26d41af8c637  bedrock-eu-check_1.2.0_linux_amd64
62e754d22f25b21f3cfce219752815ba5594ce6e11427f3c8f31807c40797310  bedrock-eu-check_1.2.0_linux_arm64
```

Die Prüfsummen hängen von Version **und Commit** ab; **maßgeblich** sind die Werte des Release-Commits aus dem Trockenlauf
von `release-create` (Abschnitt 6, Schritt 2).

## 4. Der freigegebene Commit

Der Release-Commit steht erst fest, wenn diese Vorlage gemerged ist: Jede Änderung auf `main`, auch reine Dokumentation, ändert den
Commit. Wie bei 1.1.0 wird die Vorlage gemerged, dann nennen wir den Commit (`main` nach dem Merge) **ausdrücklich im Wortlaut der
Freigabe**, und danach ändert niemand mehr `main`, bis die Veröffentlichung abgeschlossen ist.

## 5. Entscheidung C: Neue Prüfung ohne Opt-in

**Sachverhalt.** Seit 1.1.0 prüft der Modus `terraform` (und damit `all`) zusätzlich eigenständige IAM-Policy-Dateien
(`iam-policy*.json`, `*-policy.json`). Ein Repository mit einer solchen Datei, die `bedrock:*` auf `*` ohne Region-Condition erlaubt, liefert
jetzt `FAIL` und **Exit 1**, wo 1.1.0 nichts gefunden hat. Außerdem ergibt ein `Allow` nur mit `NotAction` jetzt `WARN`
(`IAM_DYNAMIC`) statt keines Findings, auch für Terraform und den AWS-Modus; der Standard-Exit ändert sich dadurch nicht.
Die Exit-Codes selbst sind unverändert, und es gibt keinen Schalter, der die neue Prüfung abschaltet.

| Option | Wirkung | Bewertung |
|---|---|---|
| **C1: ohne Opt-in als 1.2.0** (Empfehlung) | Die Prüfung gilt sofort; Hinweis in den Release-Notizen; Ausnahmen mit `exclude` je Pfad | Das Lastenheft ([`LH-FA-IAM-004`](../../../../spec/lastenheft.md#lh-fa-iam-004--iam-policy-dokumente-json-prüfen)) verlangt die Prüfung ohne Schalter; die `.env`-Prüfung kam in 1.0.0 genauso. Pipelines mit schlechten Policies werden rot, genau dafür ist das Werkzeug da |
| C2: mit Schalter (Opt-in) | Eine neue Konfigurationsoption schaltet die Prüfung ein | weicht vom Wortlaut der Anforderung ab, verlangt eine Änderung von Lastenheft und [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) und neuen Code; später aufwändiger zurückzunehmen |
| C3: als 2.0.0 kennzeichnen | Die Versionsnummer drückt die Verhaltensänderung aus | Versionen nach 1.0 sind projektintern; der Aufwand steht in keinem Verhältnis, solange `exclude` die Ausnahme je Pfad erlaubt |

## 6. Ablauf Entscheidung A: Werkzeug 1.2.0

**Voraussetzung (besteht):** Die Umgebung `release` mit Pflicht-Prüfer ist angelegt; die Freigabe eines scharfen Laufs erteilt der Nutzer in
GitHub. Sie erzwingt das Ja beider Nutzer nicht (Selbstfreigabe ist erlaubt); das bleibt Prozessregel.

1. **Stand prüfen** (sauberer Checkout des Commits): `make verify-clean`, `make self-check`, `make image-test`, `make feature-test`;
   CI auf dem Commit grün; `LICENSE` und `THIRD_PARTY_NOTICES.md` passen zum Stand.
2. **Trockenlauf von `release-create`:** `target=tool`, `version=1.2.0`, `commit=<SHA>`, `dry_run=true` (Standard); die Prüfsummen vergleichen wir
   mit einem zweiten Lauf von `release-artifacts` auf demselben Commit.
3. **Trockenlauf von `publish`:** `images=true`, `dry_run=true`: baut und testet beide Images, pusht nichts.
4. **Freigabe einholen** (Wortlaut siehe Abschnitt 7) und festhalten.
5. **Scharfer Lauf von `release-create`** (`dry_run=false`): legt den annotierten Tag `v1.2.0` und das GitHub-Release an (Freigabe in der Umgebung `release`).
6. **Nachprüfen:** Artefakte frisch herunterladen, `sha256sum -c`, `bedrock-eu-check version` meldet `1.2.0` und den Commit.
7. **Scharfer Lauf von `publish`** (`images=true`, `dry_run=false`): pusht `1.2.0-slim` und `1.2.0-full`; danach anonym ziehen, starten, `version` prüfen.
8. **Dokumentation und Roadmap:** README und `releasing.md` auf „1.2.0 veröffentlicht“, `welle-08` bleibt offen (Slices 032 und 033).

Alle Veröffentlichungsschritte sind **nicht rückholbar** (Tag, Release, Images; der Tag ist über den Go-Proxy abrufbar). Ein Release 1.2.0 ändert die
veröffentlichten Versionen 1.0.0 und 1.1.0 nicht.

## 7. Wortlaut der Freigabe

Jeder Nutzer antwortet getrennt, je Entscheidung, mit Version, Commit und Zielen:

- A: „Ja zu Release 1.2.0 auf Commit `<SHA>` für die Ziele Tag + GitHub-Release, Docker-Images nach ghcr.io“
- C: „C1, 1.2.0 ohne Opt-in“ (oder C2 beziehungsweise C3)

A setzt eine Antwort zu C voraus; eine Antwort zu C ist noch keine Freigabe zur Veröffentlichung.

## 8. Nicht erprobt und Risiken

- **Hüllformate der Policy-Dateien** (`get-policy-version`, `get-role-policy`) sind nachgebaut und nicht gegen die Ausgabe einer echten AWS-CLI geprüft.
- **Kein Größenlimit** beim Lesen von Policy-Dateien (wie bei den anderen Readern).
- **Bestandsrepos mit Policy-Dateien** können nach dem Upgrade ein `FAIL` erhalten (Entscheidung C); `exclude` nimmt Pfade aus.
- **Scharfe Läufe** von `release-create` und `publish` wurden für 1.1.0 und Feature 1.0.1 erfolgreich ausgeführt; das Risiko ist damit kleiner als bei 1.1.0.
- **Nicht Teil dieser Entscheidung:** Signierung und SBOM, weitere Plattformen (Welle 08), das Anheben des Standards der Feature-Option `version`.

## 9. Antwortblatt

| # | Frage | Antwort |
|---|---|---|
| C | Neue Prüfung ohne Opt-in als 1.2.0 (C1), mit Schalter (C2) oder als 2.0.0 (C3)? | |
| A | Werkzeug 1.2.0 veröffentlichen (Ziele Tag + Release, Images)? | |
