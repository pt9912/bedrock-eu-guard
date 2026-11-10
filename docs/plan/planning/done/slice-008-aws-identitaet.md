# Slice slice-008: AWS-Port, STS-Identität, Profil, Account und Authentifizierungsart

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-03-aws-checks.

**Bezug:** [`LH-FA-AWS-001`](../../../../spec/lastenheft.md#lh-fa-aws-001--sts-identität-prüfen), [`LH-FA-AWS-002`](../../../../spec/lastenheft.md#lh-fa-aws-002--aws-profil-anzeigen), [`LH-FA-AWS-003`](../../../../spec/lastenheft.md#lh-fa-aws-003--erwartete-account-id-prüfen), [`LH-FA-AUTH-001`](../../../../spec/lastenheft.md#lh-fa-auth-001--authentifizierungsart-klassifizieren), [`LH-MOD-002`](../../../../spec/lastenheft.md#lh-mod-002--aws-modus), [`LH-SEC-001`](../../../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials), [`LH-SEC-003`](../../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check aws` ermittelt die STS-Identität über den Adapter `awscli`
(einzige AWS-Tür, lesend), zeigt Profil, Account und Principal und klassifiziert
die Authentifizierungsart. Keine Credentials oder AWS nicht erreichbar sind
Findings, nie stille Erfolge und nie Exit `2`.

## 2. Verträge

- `model.AWSIdentity{Account, ARN, UserID}`; `model.AWSError{Op, Reason}` mit den Reasons `AWS_NO_CREDENTIALS`, `AWS_UNREACHABLE`, `AWS_ACCESS_DENIED` und `AWS_CLI_MISSING` (Tool-Fehler). Die Fehlermeldung enthält nie CLI-Ausgabe oder Credential-Werte.
- Port `driven.AWSReader` mit der Methode `Identity() (*model.AWSIdentity, error)`; die Methoden für Bedrock und IAM kommen additiv mit slice-009 und slice-010. Fehler sind `*model.AWSError`; unbekannte Fehler behandelt der Kern als `AWS_UNREACHABLE` ohne Fremdtext.
- Use-Case `app.CheckAWS{AWS}.Run(env, cfg)`: ermittelt die Identität einmal, wertet `rules.AWSRules()` aus, sortiert stabil nach Regel-ID. `AWS_CLI_MISSING` ist ein Tool-Fehler (Exit `2`).
- Konfiguration: `allowedAccounts` aus [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml); die CLI-Option `--allowed-account` (w2) ergänzt sie.
- Double `coretest/fakeaws`.

**Adapter-Pflichten (w2):** `awscli` startet das `aws`-Binary ohne Shell (`exec.LookPath`, Pfad injizierbar, Default `aws`) mit festen Argumenten (`sts get-caller-identity --output json`, `--region` aus `allowedRegions`); die Umgebung wird nur durchgereicht (PATH, HOME, `AWS_*`) und nie geloggt; Fehlerklassifikation nach Exit-Code und stderr-Mustern (NoCredentialsError, Could not connect, AccessDenied, ExpiredToken); stderr wird nie in Meldungen übernommen. Tests mit einem Fake-`aws`-Skript in `testdata` (per `PATH`), nie gegen echtes AWS.

## 3. Verhalten je Regel

| Regel | Verhalten |
|---|---|
| [`LH-FA-AWS-001`](../../../../spec/lastenheft.md#lh-fa-aws-001--sts-identität-prüfen) | `PASS` mit Account und Principal-ARN. `FAIL` bei `AWS_NO_CREDENTIALS`, `AWS_ACCESS_DENIED` oder `AWS_UNREACHABLE` (mit Empfehlung). |
| [`LH-FA-AWS-002`](../../../../spec/lastenheft.md#lh-fa-aws-002--aws-profil-anzeigen) | `PASS`: `AWS_PROFILE` gesetzt und Identität ermittelbar. `WARN`: nicht gesetzt (`AWS_PROFILE_NOT_SET`) oder gesetzt, aber nicht verwendbar (`AWS_PROFILE_UNUSABLE`). |
| [`LH-FA-AWS-003`](../../../../spec/lastenheft.md#lh-fa-aws-003--erwartete-account-id-prüfen) | Leere `allowedAccounts`: kein Finding. `PASS`: Account enthalten. `FAIL`: Account nicht enthalten. `WARN` `ACCOUNT_UNCHECKED`: Identität nicht ermittelbar. |
| [`LH-FA-AUTH-001`](../../../../spec/lastenheft.md#lh-fa-auth-001--authentifizierungsart-klassifizieren) | Aus der Existenz von Variablen, nie aus Werten; ein Finding je erkannter Art: `ANTHROPIC_API_KEY` `FAIL`; `AWS_BEARER_TOKEN_BEDROCK` `WARN`; `AWS_ACCESS_KEY_ID` mit `AWS_SESSION_TOKEN` `PASS` (temporär), ohne Token `WARN` (statisch); `AWS_PROFILE` ohne Access Key `PASS`; nichts erkannt: `WARN` (Default-Credential-Chain). Läuft auch ohne AWS-Erreichbarkeit. |

## 4. Definition of Done

- [ ] Kern (w1): `model/aws.go`, `driven.AWSReader`, `rules/aws.go`, `app.CheckAWS`, Double `coretest/fakeaws`, Unit-Tests je Zweig und Fehlerart, Test „kein Wert in Meldungen“.
- [ ] Adapter (w2): `adapter/driven/awscli` (Identity, Fehlerklassifikation, Fake-`aws` in Tests).
- [ ] CLI (w2): Befehl `aws` (mit `--allowed-account`); Exit `1` bei fehlenden Credentials, Exit `2` bei fehlendem `aws`-Binary; Akzeptanztest ohne Credential-Werte in der Ausgabe.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/aws.go`, `port/driven/files.go`, `core/rules/aws.go`, `core/app/checkaws.go`, `coretest/fakeaws/` | neu/update | w1 |
| `internal/adapter/driven/awscli/` | neu | w2 |
| `internal/adapter/driving/cli/` (Befehl `aws`) | update | w2 |

## 6. Trigger

[`welle-03-aws-checks`](../welle-03-aws-checks.md) gestartet (Plan gemerged).

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- Account-IDs und ARNs werden angezeigt (nötig laut Anforderung) und sind keine Secrets; Credential-Werte dürfen nie erscheinen.
- `AWS_ACCESS_KEY_ID` mit `ASIA…`-Präfix wäre ein weiterer Hinweis auf temporäre Keys; der Kern liest keine Werte und stützt sich nur auf `AWS_SESSION_TOKEN`.
- Die Reihenfolge der Credential-Quellen der AWS-CLI (Umgebung vor Profil) wird nicht nachgebildet; der Kern meldet alle erkannten Arten.
- Abgelaufene SSO-Sitzungen erscheinen als `AWS_NO_CREDENTIALS` und nicht als Tool-Fehler.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regeln und Adapter auf dem Kern der
Wellen 1 und 2.
