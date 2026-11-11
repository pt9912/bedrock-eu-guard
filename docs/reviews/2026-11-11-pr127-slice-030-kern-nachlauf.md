# Review-Report: PR #127 (slice-030), Nachlauf zu Finding F-1 — 2026-11-11

**Review-Art:** Code — Nachprüfung der Korrektur zu F-1 aus [`2026-11-11-pr127-slice-030-kern.md`](2026-11-11-pr127-slice-030-kern.md).

**Gegenstand:** Pull Request #127 (`feat/slice-030-kern`, Autor w1), geprüfter Stand: Commit `8593f4c` (enthält die Korrektur `c5af944` und die Korrektur der Prüfsummen-Negativprobe aus PR #129).

**Skill:** [`.harness/skills/reviewer.md`](../../.harness/skills/reviewer.md) @ `e03b883` (PR #118)
**Modell:** claude-sonnet-5-5 (Agent w2) · **Datum:** 2026-11-11

**Eingangs-Kontext:** der erste Report zu PR #127, [`LH-FA-IAM-004`](../../spec/lastenheft.md#lh-fa-iam-004--iam-policy-dokumente-json-prüfen), `internal/hexagon/core/rules/iam.go`.

---

## Findings

Keine neuen Findings.

## Negativbefunde

- geprüft, ohne Befund: F-1 (`Allow` mit `NotAction` ohne Aktionsliste) — am Binary mit dem Adapter aus meinem Branch: `{"Effect":"Allow","NotAction":"s3:*","Resource":"*"}` in `iam-policy-na.json` ergibt jetzt `WARN IAM_DYNAMIC` (Exit 0, Standardverhalten unverändert); ein `Deny` mit `NotAction` und ein `Allow` auf `s3:GetObject` ergeben weiterhin kein Finding.
- geprüft, ohne Befund: Tests der Korrektur — ein Tabellenfall „NotAction ohne Aktionsliste: nicht auswertbar“ in `iampolicyfile_test.go` und in `terraform_test.go` (die gemeinsame Funktion wirkt auch auf `LH-FA-TF-003`).
- geprüft, ohne Befund: Reichweite — die Änderung steht in `assessBedrockStatements` und gilt damit auch für Terraform und den AWS-Modus; das Handbuch nannte `NotAction` in Policies schon vorher als `WARN`-Fall, die Korrektur gleicht Code und Aussage an; der Standard-Exit ändert sich nicht (`WARN`).
- geprüft, ohne Befund: PR #129 (Prüfsummen-Negativprobe) — die Manipulation nutzt jetzt ein Nicht-Hex-Zeichen und ist damit unabhängig vom ersten Zeichen des Hashes (vorher wirkungslos bei 1/16 der Hashes).
- geprüft, ohne Befund: Gates — `make gates` (Abdeckung 95,4 %) und `make self-check` auf dem Adapter-Branch, der den Kern-Stand enthält, grün.
- nicht geprüft: Terraform-Dateien mit `NotAction` am Binary (nur durch den Tabellentest abgedeckt).

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |
| INFO | 0 |

## Verdikt

**Merge-blockierend:** nein. F-1 des ersten Reports ist behoben und nachgeprüft.

**Übergabe:** zurück an w1 (Autor). Der Report ersetzt weder die Verifikation noch die Prüfung der Definition of Done.
