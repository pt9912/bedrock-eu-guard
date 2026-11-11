# Slice slice-030: Scanner für eigenständige IAM-Policy-Dateien (`iam-policy.json`)

**Status:** next. **Verantwortlich:** Kern w1, Adapter/CLI/Doku w2. **Priorisierung:** Beide Nutzer haben diesen Slice am 2026-11-11 priorisiert und die Festlegungen in Abschnitt 3 bestätigt.

**Welle:** [`welle-08-erweiterungen`](../welle-08-erweiterungen.md).

**Bezug:** [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) (Testdaten für `iam-policy.json` erst bei bestehender Anforderung), [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen), [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen), [`LH-MOD-003`](../../../../spec/lastenheft.md#lh-mod-003--terraform-modus), [`release-1.0-entscheidung-qa003.md`](../done/release-1.0-entscheidung-qa003.md) (Optionen und Festlegungen).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Eigenständige IAM-Policy-Dokumente im JSON-Format (zum Beispiel `iam-policy.json`) werden auf riskante Bedrock-Rechte geprüft, mit derselben Bewertung wie Policies in Terraform und im AWS-Konto: `FAIL`, wenn `bedrock:*` oder `InvokeModel` auf `*` ohne Region-Condition erlaubt ist; `WARN` bei `*` mit Region-Condition; `PASS` bei eng gefassten Rechten. Damit entfällt die letzte bekannte Lücke aus [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten).

## 2. Ausgangslage

- Das Lastenheft nennt `iam-policy.json` bei den Testdaten, beschreibt aber keinen Scanner dafür; [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) ist präzisiert: Für diese Dateiart gilt die Pflicht erst bei bestehender Anforderung.
- Die Bewertung von Policy-Statements gibt es im Kern (gemeinsame Funktion für Terraform und den AWS-Modus); das Paket `iampolicy` im Adapter liest Policy-Dokumente (AWS und Terraform).
- Im Handbuch steht als Grenze, dass eigenständige Policy-Dateien nicht geprüft werden.

## 3. Entscheidungen der Auftraggeber (vor dem Start)

Beide Nutzer haben die Priorisierung und diese Festlegungen bestätigt (siehe unten). Vorschlag in Klammern.

**Bestätigungen** (Stand 2026-11-11): Nutzer von w2: „ja stimme zu“ zu den Festlegungen 1 bis 5 in der Fassung nach dem Plan-Review (die Bestätigung bezieht sich auf die Fassung dieses Abschnitts nach Commit `Plan-Review F-1 bis F-3 eingearbeitet`). Nutzer von w1: „Ja, Priorisierung und Festlegungen 1–5“ (nachdem ihm Abschnitt 3 vorgelegt wurde).

1. **Neue Anforderung** (Nummer und Wortlaut legen die Auftraggeber fest). Vorschlag: `IAM-004` im Bereich der IAM-Anforderungen (die ID ist noch nicht vergeben), „Prüfung von IAM-Policy-Dokumenten (JSON)“, Akzeptanzkriterium: „Das Tool findet eigenständige IAM-Policy-Dateien (das Policy-Dokument selbst oder die Ausgabe von `aws iam get-policy-version` bzw. `get-role-policy`) und bewertet sie wie [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen) und [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen): `FAIL`, wenn `bedrock:*` oder `InvokeModel` auf `*` ohne Region-Condition erlaubt ist; `WARN`, wenn `*` verwendet wird, aber eine Region-Condition vorhanden ist, oder wenn die Wirkung nicht auswertbar ist; `PASS`, wenn die Policy eng eingeschränkt ist. Eine nicht lesbare oder ungültige Datei mit eindeutigem Namen (`iam-policy*.json`) ist ein `FAIL` mit Grund `FILE_UNREADABLE` bzw. `FILE_INVALID`; bei einer Datei, die nur das breite Muster `*-policy.json` trifft, ist dies ein `WARN` (Fremdformate lösen keinen `FAIL` aus), und ein gültiges JSON ohne Policy-Statement oder ohne Bedrock-Aktion erzeugt kein Finding. Dateiinhalte erscheinen nie in der Ausgabe.“
2. **Modus.** Vorschlag: `terraform` ([`LH-MOD-003`](../../../../spec/lastenheft.md#lh-mod-003--terraform-modus): IAM-Konfigurationen in Infrastrukturdateien). Alternative `local`: beide Modi arbeiten offline und ohne AWS-Zugriff; `local` prüft Umgebung, Settings und Container-Dateien, und wer nur `local` ausführt, würde die Policy-Datei dann mitprüfen, aber die Modusgrenzen der Anforderungen verschieben sich (IAM-Konfigurationen gehören bisher zu `terraform`). Der Vorschlag `terraform` folgt der bestehenden Zuordnung; wer nur `local` ausführt, prüft die Policy-Datei nicht (als Grenze im Handbuch zu nennen).
3. **Dateimuster und Fehlerregel.** Vorschlag: rekursiv `iam-policy*.json` (eindeutiger Name: Fehler = `FAIL`) und `*-policy.json` (breites Muster: Fehler = `WARN`, Fremdformate ohne Finding); Ausschlüsse wie bei Terraform (`.git`, `.terraform`, `vendor`, `node_modules`, `.tmp`, keine Symlinks) und über `exclude`. Policies ohne Bedrock-Aktion ergeben kein Finding (kein Fehlalarm bei fremden Policies wie S3).
4. **Dokumentformen.** Vorschlag: das nackte Policy-Dokument (`Version`, `Statement`) und die Ausgabe von `aws iam get-policy-version` (`PolicyVersion.Document`) bzw. `get-role-policy` (`PolicyDocument`); nicht unterstützt sind CloudFormation- und andere Vorlagen.
5. **Testdaten.** `testdata/good` und `testdata/bad` enthalten je eine Policy-Datei; die Matrix zu [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) wird um `iam-policy.json` erweitert (Teil von w1).

## 4. Definition of Done

- [x] Beide Nutzer haben Priorisierung und die Festlegungen aus Abschnitt 3 bestätigt (2026-11-11).
- [ ] Der Lastenheft-Eintrag steht mit dem bestätigten Wortlaut, [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) ist angepasst (w1, eigener PR).
- [ ] Kern (w1): Port für die Suche und das Lesen von Policy-Dateien (eigene kleine Schnittstelle nach dem Muster der `.env`-Suche), Regel auf der vorhandenen Bewertung, Einbindung in `app.Scan` (nur Modus `terraform`), `exclude` im Kern, Unit-Tests; Regelkatalog und Referenz ergänzt (Doku-Drift-Test grün).
- [ ] Adapter (w2): JSON-Reader auf dem Paket `iampolicy`, rekursive Suche (sortiert, relative Pfade), Testdaten `good`/`bad` (nur offensichtlich synthetische Werte), CLI-Verdrahtung, CLI-Tests ohne Dateiinhalt in der Ausgabe.
- [ ] Doku (w2): Handbuch und Fehlersuche; die Grenze „eigenständige Policy-Dateien werden nicht geprüft“ entfällt.
- [ ] Matrix-Test zu [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) um `iam-policy.json` erweitert; das Self-Check-Gate (`make self-check`) bleibt grün.
- [ ] `make verify-clean` und CI grün; Review-Report durch den anderen Agenten.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `spec/lastenheft.md` (neue Anforderung, [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten)) | update nach Wortlaut der Auftraggeber | w1 |
| Kern: Port, Regel, `app.Scan` | neu/update | w1 |
| `internal/adapter/driven/iampolicy` (Datei-Reader, Suche) | update/neu | w2 |
| `internal/adapter/driving/cli` (Verdrahtung `terraform`) | update | w2 |
| `internal/adapter/driving/cli/testdata/{good,bad}/…/iam-policy.json` | neu | w2 |
| `docs/user/rules.md`, `reference.md` | update | w1 |
| `docs/user/handbuch.md`, `troubleshooting.md`, README | update | w2 |

## 6. Trigger

Bestätigung beider Nutzer zu Priorisierung und Festlegungen (Abschnitt 3): erteilt am 2026-11-11.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- **Fehlalarme** bei fremden Policy-Dateien: nur Policies mit Bedrock-Aktion erzeugen Findings; das Dateimuster `*-policy.json` ist breit und im Plan-Review kritisch zu prüfen.
- **Datei-Discovery** erfasst nur, was zum Muster passt; Policies in anderen Dateien (zum Beispiel Vorlagen) bleiben außen vor und sind als Grenze zu dokumentieren.
- **Lastenheft ist bindend:** Ohne bestätigten Wortlaut gibt es keinen Eintrag und keinen Code.
- Dateiinhalte können sensible Konto-IDs enthalten; die Ausgabe nennt nur Pfad, Zeile und Grund.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regel und Reader auf dem bestehenden Kern, nach dem Muster der `.env`-Prüfung.
