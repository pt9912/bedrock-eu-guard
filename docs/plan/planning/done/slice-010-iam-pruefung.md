# Slice slice-010: IAM-Prüfung (Bedrock-Rechte, zu breite Policies, Region-Condition)

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-03-aws-checks.

**Bezug:** [`LH-FA-IAM-001`](../../../../spec/lastenheft.md#lh-fa-iam-001--bedrock-rechte-prüfen), [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen), [`LH-FA-IAM-003`](../../../../spec/lastenheft.md#lh-fa-iam-003--region-condition-prüfen), [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen), [`LH-SEC-003`](../../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff), [`LH-SEC-004`](../../../../spec/lastenheft.md#lh-sec-004--least-privilege), [`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check aws` prüft die IAM-Seite des aktuellen Principals (lesend, über
`awscli`): darf er Bedrock aufrufen, sind seine Policies zu breit, und sind
Aufrufe auf erlaubte Regionen eingeschränkt. Dazu wird die dokumentierte
Mindestberechtigung des Prüfwerkzeugs selbst festgehalten ([`LH-SEC-004`](../../../../spec/lastenheft.md#lh-sec-004--least-privilege)).

## 2. Verträge

- Zusätzlicher Port `driven.IAMReader` (eigene Schnittstelle): `SimulateBedrock(principalARN, actions) ([]model.ActionDecision, error)` und `PrincipalPolicies(principalARN) ([]model.IAMStatement, error)`. `principalARN` ist der von STS gemeldete ARN; der Adapter löst ihn bei Bedarf auf (`assumed-role` → IAM-Rolle, inklusive Pfad über `get-role`). Fehler sind `*model.AWSError`.
- `model.ActionDecision{Action, Allowed}`; `model.IAMStatement` bekommt additiv `ConditionRegions []string` (Werte der `aws:RequestedRegion`-Condition, soweit auswertbar; leer bei `HasRegionCondition` = Regionen unbekannt).
- `app.CheckAWS` bekommt das optionale Feld `IAM`; `nil` = keine IAM-Regeln. Ohne ermittelte Identität meldet der Kern denselben Fehler (Reason der Identität) für alle drei IAM-Regeln.
- Gemeinsame Bewertung: die Logik von [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) liegt jetzt in `assessBedrockStatements` (Refactoring ohne Verhaltensänderung) und dient auch [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen).
- Double `coretest/fakeaws.IAM`.

**Adapter-Pflichten (w2):** `awscli.Reader` implementiert `IAMReader`: `iam simulate-principal-policy` (Aktionen als eigene Argv-Elemente, ARN als einzelnes Element, Pagination), Policies über `list-attached-role-policies`/`get-policy`/`get-policy-version` und Inline-Policies (`list-role-policies`/`get-role-policy`), analog für Benutzer und Gruppen; Policy-Dokumente werden URL-decodiert und zu `IAMStatement`s verflacht (`Action`/`NotAction`, `Resource`, `Effect`, `aws:RequestedRegion`-Werte aus `Condition`); `NotAction` und `NotResource` gelten als nicht auswertbar (`Dynamic`). Nur lesende Aufrufe; Fake-`aws` in den Tests.

## 3. Verhalten je Regel

| Regel | Verhalten |
|---|---|
| [`LH-FA-IAM-001`](../../../../spec/lastenheft.md#lh-fa-iam-001--bedrock-rechte-prüfen) | Simuliert `bedrock:InvokeModel`, `…WithResponseStream`, `Converse`, `ConverseStream` (Aufruf) und `ListInferenceProfiles`, `GetInferenceProfile` (Erkennung). `FAIL` (`IAM_INVOKE_DENIED`): ein Aufruf-Recht fehlt. `WARN` (`IAM_DISCOVERY_DENIED`): nur Erkennungs-Rechte fehlen. `PASS`: alles erlaubt. Simulation nicht erlaubt (`AWS_ACCESS_DENIED`): `WARN` `IAM_UNKNOWN`; keine Credentials/nicht erreichbar: `FAIL`. |
| [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen) | Wie [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) auf den Statements des Principals: `bedrock:*`/`InvokeModel*`/`*` auf `*` ist `FAIL` ohne und `WARN` mit Region-Condition; sonst `PASS`. Statements lesen nicht erlaubt: `WARN` `IAM_UNKNOWN`. |
| [`LH-FA-IAM-003`](../../../../spec/lastenheft.md#lh-fa-iam-003--region-condition-prüfen) | Je Bedrock-`Allow`-Statement muss eine `aws:RequestedRegion`-Condition mit erlaubten Regionen gelten ([`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation)) oder jede Resource ein Bedrock-ARN einer erlaubten Region sein. `FAIL`: fehlt (`IAM_NO_REGION_CONDITION`) oder enthält eine nicht erlaubte Region (`IAM_REGION_CONDITION_NOT_ALLOWED`). `WARN`: Condition ohne auswertbare Regionen. `PASS`: alle Statements eingeschränkt; ohne Bedrock-Allow kein Finding. |

## 4. Definition of Done

- [ ] Kern (w1): `model`-Typen, `driven.IAMReader`, `rules/iam.go` (inkl. Refactoring `assessBedrockStatements`), Erweiterung `app.CheckAWS`, Double `fakeaws.IAM`, Unit-Tests je Zweig und Fehlerart; bestehende Terraform-Tests unverändert grün.
- [ ] Adapter (w2): `awscli` implementiert `IAMReader` (Simulation, Policy-Verflachung, URL-Decoding, Pagination), Tests mit Fake-`aws`.
- [ ] CLI (w2): Befehl `aws` übergibt `IAM`; Akzeptanztest ohne Credential-Werte in der Ausgabe.
- [ ] Doku der Mindestrechte des Prüfwerkzeugs ([`LH-SEC-004`](../../../../spec/lastenheft.md#lh-sec-004--least-privilege)): `docs/user/` (w2), inkl. Beispiel-Policy mit `sts:GetCallerIdentity`, `bedrock:ListInferenceProfiles`, `bedrock:ListFoundationModels`, `iam:SimulatePrincipalPolicy`, `iam:GetRole`, `iam:List*`/`iam:Get*` auf den eigenen Principal.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; Welle-3-Abschluss (M3).

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/{aws,terraform}.go`, `port/driven/files.go`, `core/rules/{iam,terraform}.go`, `core/app/checkaws.go`, `coretest/fakeaws/` | neu/update | w1 |
| `internal/adapter/driven/awscli/` | update | w2 |
| `internal/adapter/driving/cli/`, `docs/user/` | update | w2 |

## 6. Trigger

slice-009 gemerged; Plan gemerged.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- **Auflösung des Principals:** Rollen mit Pfad (`/service-role/…`) und SSO-Rollen (`AWSReservedSSO_…`) brauchen eine Auflösung über `get-role`; der Adapter meldet `AWS_ACCESS_DENIED`/`WARN`, wenn das nicht erlaubt ist.
- **Simulation kennt keine Organisations-SCPs und Permission Boundaries vollständig;** das Ergebnis ist eine Näherung und wird als solche in der Meldung nicht überhöht (nur `Allowed`/`Denied` je Aktion).
- **Policy-Verflachung:** Variablen in Policies (`${aws:username}`), `NotAction`, `NotResource` und Condition-Operatoren außer `aws:RequestedRegion` werden nicht ausgewertet (`Dynamic` bzw. unbekannte Regionen → `WARN`).
- **Vorgehen nur für den eigenen Principal;** andere Principals, Gruppen-Vererbung über Organisationen und Ressourcen-Policies sind Folgepunkte.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regeln und Adapter-Erweiterung auf dem
Kern von slice-008 und slice-009; das Refactoring von `TF-003` ist durch die
bestehenden Tests abgesichert.
