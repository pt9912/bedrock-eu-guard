# Slice slice-007: Terraform-Scanner

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-02-config-scanner.

**Bezug:** [`LH-FA-TF-001`](../../../../spec/lastenheft.md#lh-fa-tf-001--terraform-dateien-finden), [`LH-FA-TF-002`](../../../../spec/lastenheft.md#lh-fa-tf-002--aws-provider-region-prüfen), [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen), [`LH-FA-TF-004`](../../../../spec/lastenheft.md#lh-fa-tf-004--nicht-erlaubte-modell-ids-finden), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), [`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md), [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check local` findet `.tf`-Dateien rekursiv im Scan-Root, liest
AWS-Provider-Regionen, Bedrock-IAM-Statements und Modell-/Profil-IDs und
meldet Findings mit Datei und Zeile (reiner Lesezugriff, offline). Der Kern
sieht nur geparste Fakten, keine Variablenwerte und keine übrigen Literale.

## 2. Verträge

- `model.TerraformFile{Path, Providers []AWSProvider, Statements []IAMStatement, ModelRefs []ModelRef}`:
  - `AWSProvider{Alias, Region, RegionDynamic, Line}`: `Region` nur bei String-Literal; `RegionDynamic=true` bei nicht auflösbarem Ausdruck (`var.x`, `local.x`, …); beides leer = Attribut fehlt.
  - `IAMStatement{Effect, Actions, Resources, HasRegionCondition, Dynamic, Line}`: nur Literale; `Dynamic=true`, wenn Einträge nicht auflösbar waren; `HasRegionCondition` bei Condition auf `aws:RequestedRegion` (auch `aws:RequestedRegion` in `StringEquals`/`StringLike`/`ForAnyValue:*`).
  - `ModelRef{Value, Line}`: nur das **gefundene Token**, nie das ganze Literal.
- Port `driven.TerraformReader`: `Find() ([]string, error)` (rekursive Suche nach `*.tf`, sortiert, relative Pfade; schließt `.git`, `.terraform`, `vendor`, `node_modules` aus) und `Read(rel) (*model.TerraformFile, error)` (fehlt → `(nil, nil)`; kaputt → `*model.FileError`).
- `app.Scan` bekommt das optionale Feld `Terraform`; ohne Reader unverändertes Verhalten.

**Adapter-Pflichten (w2):** HCL über `hashicorp/hcl/v2` (`hclsyntax`, gevendort, [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md)); Zeilen aus den Ranges.
Provider: `provider "aws" { region = … }` (Literal, sonst `RegionDynamic`; mehrere Provider mit `alias`).
Policies: `data "aws_iam_policy_document"` (`statement { effect, actions, resources, condition { test, variable, values } }`), `aws_iam_policy`/`aws_iam_role_policy` mit `jsonencode({...})` und JSON-Heredoc; je Statement ein `IAMStatement`.
Modell-IDs: aus allen String-Literalen nur Tokens extrahieren, die auf `(us|eu|apac|global)?\.?(anthropic|amazon|meta|mistral|cohere|ai21|deepseek|…)\.…` bzw. `arn:aws:bedrock:…` passen. **Keine** `.tfvars`, `variable`-Defaults, `locals`-Werte oder sonstige Literale in den Kern geben ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).

## 3. Verhalten je Regel

Datei vorhanden, aber kaputt → `FAIL` `FILE_INVALID`/`FILE_UNREADABLE` unter `LH-FA-TF-001`, Scan läuft weiter. Pro Datei und Regel höchstens ein Finding (Zeile des ersten Belegs), außer `LH-FA-TF-003` (ein Finding je riskantem Statement).

| Regel | Verhalten |
|---|---|
| [`LH-FA-TF-001`](../../../../spec/lastenheft.md#lh-fa-tf-001--terraform-dateien-finden) | `PASS`: mindestens eine Datei gefunden und gelesen. `WARN` `NO_TERRAFORM_FILES`: keine `.tf`-Datei gefunden. Ohne Reader: kein Finding. |
| [`LH-FA-TF-002`](../../../../spec/lastenheft.md#lh-fa-tf-002--aws-provider-region-prüfen) | `FAIL`: Region nicht in `allowedRegions`. `WARN`: Region dynamisch (`REGION_DYNAMIC`) oder Attribut fehlt (`REGION_MISSING`). `PASS`: nur erlaubte. `FAIL` schlägt `WARN`. Dateien ohne Provider: kein Finding. |
| [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) | Nur `Allow`-Statements mit Bedrock-Aktionen. Riskant = Aktion `bedrock:*`, `*` oder `bedrock:InvokeModel*` **und** Resource `*`: `FAIL` ohne, `WARN` mit Region-Condition. Sonst `PASS` je Datei (`WARN` `IAM_DYNAMIC`, wenn nicht auflösbare Einträge vorliegen). `Deny` und Nicht-Bedrock-Statements werden ignoriert. |
| [`LH-FA-TF-004`](../../../../spec/lastenheft.md#lh-fa-tf-004--nicht-erlaubte-modell-ids-finden) | Klassifikation mit `Classify` ([`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), inkl. Region in der ID): `FAIL` bei blockierten (`us.`, `global.`, fremde Region), `WARN` bei unklassifizierten, `PASS` bei nur erlaubten; ohne Funde kein Finding. |

## 4. Definition of Done

- [ ] Kern (w1): `model/terraform.go`, `TerraformReader`, `rules/terraform.go`, Erweiterung `app.Scan`, Double `coretest/faketerraform`, Unit-Tests je Zweig.
- [ ] Adapter (w2): `adapter/driven/terraform` (rekursive Suche, HCL-Parser, Policy-/Modell-Extraktion, Zeilen), HCL-Abhängigkeit gevendort, `THIRD_PARTY_NOTICES.md` ergänzt, Tests inkl. Secret-/`tfvars`-Fälle.
- [ ] CLI (w2): `local` übergibt `Terraform`; Testdaten good/bad; Akzeptanztest ohne Secret-Wert in der Ausgabe.
- [ ] `make verify-clean`, `deps-verify` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; Welle-2-Abschluss (M2).

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/terraform.go`, `port/driven/files.go`, `core/rules/terraform.go`, `core/app/scan.go`, `coretest/faketerraform/` | neu/update | w1 |
| `internal/adapter/driven/terraform/`, `vendor/` (HCL + transitive), `THIRD_PARTY_NOTICES.md` | neu | w2 |
| `internal/adapter/driving/cli/` | update | w2 |

## 6. Trigger

slice-006 gemerged; Plan gemerged.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- `hashicorp/hcl/v2` ist MPL-2.0 und zieht transitive Abhängigkeiten (u. a. `go-cty`) nach `vendor/`; die Repo-Größe wächst spürbar (siehe [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md), Re-Evaluierungs-Trigger). `vendor/` wird nie von Hand editiert.
- Ausdrücke werden nicht ausgewertet: Variablen, `locals`, Module und `for_each` sind nicht auflösbar → `WARN` statt `PASS`/`FAIL`. Auflösung über `.tfvars` ist ein Folgepunkt (und secret-sensibel).
- Rekursive Suche kann große Repos durchlaufen; Ausschlüsse sind fest (`.git`, `.terraform`, `vendor`, `node_modules`), Konfiguration folgt bei Bedarf.
- Region-Condition wird nur für `aws:RequestedRegion` erkannt; andere Einschränkungen (z. B. ARN-Region in `Resource`) zählen als „eng“ nur, wenn `Resource` nicht `*` ist.
- Policy-Dokumente als externe `.json`-Dateien (`file("policy.json")`) sind ein Folgepunkt.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regeln und Adapter auf dem Kern aus
Welle 1 und slice-004 bis -006.
