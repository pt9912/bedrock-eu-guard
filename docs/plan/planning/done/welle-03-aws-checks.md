# Welle welle-03-aws-checks: AWS-Checks (STS, Bedrock, IAM)

**Status:** done

**Zielmeilenstein:** M3 — AWS-Live-Checks

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Welle-Ziel

`bedrock-eu-check aws` prüft read-only gegen AWS: Identität und Account
([`LH-FA-AWS-001`](../../../spec/lastenheft.md#lh-fa-aws-001--sts-identität-prüfen)–[`LH-FA-AWS-003`](../../../spec/lastenheft.md#lh-fa-aws-003--erwartete-account-id-prüfen), [`LH-FA-AUTH-001`](../../../spec/lastenheft.md#lh-fa-auth-001--authentifizierungsart-klassifizieren)), Bedrock-Inference-Profile und
Modellverfügbarkeit ([`LH-FA-BEDROCK-001`](../../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten)–[`LH-FA-BEDROCK-004`](../../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen)) sowie die IAM-Rechte des
aktuellen Principals ([`LH-FA-IAM-001`](../../../spec/lastenheft.md#lh-fa-iam-001--bedrock-rechte-prüfen)–[`LH-FA-IAM-003`](../../../spec/lastenheft.md#lh-fa-iam-003--region-condition-prüfen)). Der Zugriff läuft ausschließlich über
den Adapter `awscli` (einzige AWS-Tür, siehe [`spec/architecture.md`](../../../spec/architecture.md) §2) und ist lesend
([`LH-SEC-003`](../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)). Dazu kommt die Verdrahtung `aws` und `all` ([`LH-MOD-002`](../../../spec/lastenheft.md#lh-mod-002--aws-modus), [`LH-MOD-004`](../../../spec/lastenheft.md#lh-mod-004--gesamtprüfung)).

## 2. Trigger (Welle startet)

- welle-02-config-scanner done (M2), insbesondere slice-007.
- CLI-Struktur entschieden (Agenten untereinander, 2026-11-10): `local` = Env, Claude-Config,
  Devcontainer/Compose/Dockerfile; `terraform` = nur Terraform-Regeln (eigener Befehl,
  [`LH-MOD-003`](../../../spec/lastenheft.md#lh-mod-003--terraform-modus)); `aws` = AWS-Regeln; `all` = `local` + `terraform` + `aws`
  ([`LH-MOD-004`](../../../spec/lastenheft.md#lh-mod-004--gesamtprüfung)). Der Umbau (Terraform aus `local`) ist ein Vorab-PR von w2; der Kern macht dafür alle
  Felder von `app.Scan` optional.

## 3. Closure-Trigger (Welle schließt)

- Alle Slices done; `make verify-clean` und CI grün (ohne Netz und ohne AWS-Zugang).
- `bedrock-eu-check aws` läuft gegen einen Fake-`aws`-Prozess mit je einem
  Positiv- und Negativbeispiel je Regel; ohne Credentials meldet der Modus `FAIL`
  und die lokalen Regeln laufen weiter ([`spec/architecture.md`](../../../spec/architecture.md) §4).
- Kein Credential-Wert in Ausgabe, Logs oder Testdaten ([`LH-SEC-001`](../../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials), [`LH-NF-004`](../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
- Die benötigten IAM-Rechte des Tools sind dokumentiert ([`LH-SEC-004`](../../../spec/lastenheft.md#lh-sec-004--least-privilege)).
- Closure-Notiz in `done/welle-03-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle (Entwurf, Reihenfolge)

| Slice | Titel | Bezug | Dependency |
|---|---|---|---|
| slice-008 | AWS-Port, `awscli`-Adapter (STS), Profil/Account, Auth-Klassifikation, Befehl `aws` | [`LH-FA-AWS-001`](../../../spec/lastenheft.md#lh-fa-aws-001--sts-identität-prüfen), [`LH-FA-AWS-002`](../../../spec/lastenheft.md#lh-fa-aws-002--aws-profil-anzeigen), [`LH-FA-AWS-003`](../../../spec/lastenheft.md#lh-fa-aws-003--erwartete-account-id-prüfen), [`LH-FA-AUTH-001`](../../../spec/lastenheft.md#lh-fa-auth-001--authentifizierungsart-klassifizieren) | nur Stdlib |
| slice-009 | Bedrock-Inference-Profile, Allowlist, Modellverfügbarkeit | [`LH-FA-BEDROCK-001`](../../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten), [`LH-FA-BEDROCK-002`](../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen), [`LH-FA-BEDROCK-003`](../../../spec/lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen), [`LH-FA-BEDROCK-004`](../../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen) | slice-008 |
| slice-010 | IAM: Bedrock-Rechte, zu breite Policies, Region-Condition | [`LH-FA-IAM-001`](../../../spec/lastenheft.md#lh-fa-iam-001--bedrock-rechte-prüfen), [`LH-FA-IAM-002`](../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen), [`LH-FA-IAM-003`](../../../spec/lastenheft.md#lh-fa-iam-003--region-condition-prüfen) | slice-008 |

> Nur der Zuschnitt steht hier; die Slice-Dateien entstehen vor dem
> jeweiligen Start (Plan-vor-Code, Modul 5).

## 5. Abhängigkeiten

- Blockiert: welle-04 (JSON-/SARIF-Ausgabe und CI-Modus nutzen das fertige Findings-Modell).
- Wird blockiert von: welle-02 (slice-007 für die Wiederverwendung der IAM-Bewertung).

## 5a. Architektur-Entscheidungen (Entwurf, in slice-008 zu bestätigen)

1. **Port `driven.AWSReader`** (Vertrag, Kern liefert; geparste Typen statt Roh-JSON):
   `Identity() (*model.AWSIdentity, error)` (Account, Principal-ARN, UserID),
   `ListInferenceProfiles(region string) ([]model.InferenceProfile, error)`,
   `ListFoundationModels(region string) ([]model.FoundationModel, error)`,
   `SimulateBedrock(principalARN string, actions []string) ([]model.ActionDecision, error)` und
   `PrincipalPolicies(principalARN string) ([]model.IAMStatement, error)` (nur Statements, keine Dokumentreste).
   Fehler sind typisiert (`*model.AWSError{Op, Reason}`, Reasons u. a. `AWS_NO_CREDENTIALS`, `AWS_UNREACHABLE`, `AWS_ACCESS_DENIED`) und enthalten nie CLI-Ausgabe oder Credential-Werte.
2. **Wiederverwendung:** Die Bewertung zu breiter Policies ([`LH-FA-IAM-002`](../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen), [`LH-FA-IAM-003`](../../../spec/lastenheft.md#lh-fa-iam-003--region-condition-prüfen)) nutzt dieselbe Logik wie Terraform ([`LH-FA-TF-003`](../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen)) über `model.IAMStatement`; sie wird dafür aus `rules/terraform.go` in einen gemeinsamen Baustein gezogen (Refactoring im Kern, kein neues Verhalten).
3. **Rechteprüfung ([`LH-FA-IAM-001`](../../../spec/lastenheft.md#lh-fa-iam-001--bedrock-rechte-prüfen)):** `iam simulate-principal-policy` ist lesend und liefert pro Aktion allowed/denied. Fehlt dem Tool das Recht zur Simulation, ist das Ergebnis `WARN` („nicht ermittelbar“), nie ein stilles `PASS`.
4. **Fehlerpolitik:** Keine Credentials oder AWS nicht erreichbar → `FAIL` für die AWS-Regeln mit klarer Meldung; die offline-Regeln laufen weiter, Exit `1` (nicht `2`, [`SPEC-030`](../../../spec/spezifikation.md#spec-030--exit-codes)). `aws`-Binary fehlt → Tool-Fehler Exit `2` mit Hinweis.
5. **Adapter `awscli`:** startet das installierte `aws`-Binary ohne Shell (feste Argumentlisten, `--output json`, `--region` aus der Konfiguration), parst die JSON-Ausgabe, reicht keine Umgebung mit Secrets in Meldungen weiter und speichert nichts. Tests laufen offline gegen ein Fake-`aws` (Testskript im `PATH`), nie gegen echtes AWS.
6. **Authentifizierungsart ([`LH-FA-AUTH-001`](../../../spec/lastenheft.md#lh-fa-auth-001--authentifizierungsart-klassifizieren)):** reine Kernregel auf dem `Environment`-Port (`AWS_PROFILE`, `AWS_ACCESS_KEY_ID`/`AWS_SESSION_TOKEN`, `AWS_BEARER_TOKEN_BEDROCK`, `ANTHROPIC_API_KEY`) ohne AWS-Zugriff; Werte werden nie gelesen oder ausgegeben, nur Existenz.
7. **Aufteilung:** w1 = Kern (`model`, Port, Regeln, `coretest/fakeaws`); w2 = Adapter `awscli`, CLI-Befehle `aws`/`all`, Makefile/CI, Doku der IAM-Rechte.

## 6. Out-of-Scope für diese Welle

- Schreibende AWS-Operationen und Remediation ([`LH-SEC-005`](../../../spec/lastenheft.md#lh-sec-005--keine-automatischen-remediations)).
- JSON-/SARIF-Ausgabe und `--ci` (Welle 4).
- AWS-SDK statt CLI (nur falls ein ADR es später ersetzt).
- Mehrere Accounts/Organisationen in einem Lauf.

## 7. Risiken und offene Punkte

- **Offline-Pflicht:** Nur `aws` braucht Netz; `local` bleibt offline ([`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)). Das Docker-Image des Tools braucht das `aws`-Binary erst für den Betrieb, nicht für die Gates.
- **Rechte des Tools:** `sts:GetCallerIdentity`, `bedrock:ListInferenceProfiles`, `bedrock:ListFoundationModels`, `iam:SimulatePrincipalPolicy`, `iam:List*`/`iam:Get*` auf den eigenen Principal. Least-Privilege-Policy wird mit slice-010 dokumentiert ([`LH-SEC-004`](../../../spec/lastenheft.md#lh-sec-004--least-privilege)).
- **„Verwendet“ vs. „sichtbar“** ([`LH-FA-BEDROCK-002`](../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen)): verwendet = in Env/Settings/Konfiguration referenziert (Ergebnis des `local`-Laufs); sichtbar = nur in der Profilliste. Die Zusammenführung beider Modi (`all`) legt slice-009 fest.
- **SSO/temporäre Credentials** laufen ab: eine abgelaufene Sitzung ist `AWS_NO_CREDENTIALS`, kein Tool-Fehler.
- **Regionen der Abfragen:** `--region` kommt aus `allowedRegions` (Default `eu-central-1`). Ob `ListInferenceProfiles` zusätzlich eine feste Liste typischer Nicht-EU-Regionen abfragt, um „sichtbare, nicht erlaubte“ Profile ([`LH-FA-BEDROCK-002`](../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen)) zu finden, legt der Plan von slice-009 fest.

## 8. Closure-Notiz

Abgeschlossen am 2026-11-10, siehe [`done/welle-03-results.md`](done/welle-03-results.md).
