# Slice slice-001: Hexagon-Gerüst + Env-/Region-Check + Exit-Codes + Konsolen-Reporter

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-01-lokaler-env-check.

**Bezug:** [`LH-FA-LENV-001`](../../../../spec/lastenheft.md#lh-fa-lenv-001--prüfung-von-aws_region), [`LH-FA-LENV-002`](../../../../spec/lastenheft.md#lh-fa-lenv-002--prüfung-von-aws_default_region), [`LH-FA-LENV-003`](../../../../spec/lastenheft.md#lh-fa-lenv-003--prüfung-von-claude_code_use_bedrock),
[`LH-FA-LENV-004`](../../../../spec/lastenheft.md#lh-fa-lenv-004--prüfung-von-anthropic_api_key), [`LH-FA-EXIT-001`](../../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung), [`LH-FA-EXIT-002`](../../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung),
[`LH-FA-EXIT-003`](../../../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler), [`LH-FA-REP-001`](../../../../spec/lastenheft.md#lh-fa-rep-001--konsolenausgabe), [`SPEC-001`](../../../../spec/spezifikation.md#spec-001--severity--und-aggregationsmodell), [`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation),
[`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes), [`ARC-001`](../../../../spec/architecture.md#2-schichten-und-constraints)–[`ARC-007`](../../../../spec/architecture.md#2-schichten-und-constraints), [`ADR-0002`](../../adr/0002-implementierungssprache.md), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-06-24.

---

## 1. Ziel

Das hexagonale Grundgerüst stehen lassen und den `local`-Modus soweit
ausführen, dass `bedrock-eu-check local` die Region-/Bedrock-/Anthropic-
Env-Variablen prüft, `PASS`/`WARN`/`FAIL` auf der Konsole ausgibt und mit
`0`/`1`/`2` endet.

## 2. Definition of Done

- [ ] Paketstruktur nach [`ADR-0003`](../../adr/0003-hexagonale-architektur.md) angelegt: `internal/hexagon/core/{model,rules,app}`, `internal/hexagon/port/driven`, `internal/adapter/driving/cli`, `internal/adapter/driven/{envos,report}`.
- [ ] `model`: `Finding`, `Severity` (`PASS/WARN/FAIL`), Reason-Codes, `Config` (importfreier Ring) — [`SPEC-001`](../../../../spec/spezifikation.md#spec-001--severity--und-aggregationsmodell).
- [ ] `Environment`-Port + `envos`-Adapter + `coretest/fakeenv`-Double.
- [ ] Regeln [`LH-FA-LENV-001`](../../../../spec/lastenheft.md#lh-fa-lenv-001--prüfung-von-aws_region) (AWS_REGION), `-002` (AWS_DEFAULT_REGION-Fallback), `-003` (CLAUDE_CODE_USE_BEDROCK), `-004` (ANTHROPIC_API_KEY) mit Unit-Tests je Regel ([`LH-QA-001`](../../../../spec/lastenheft.md#lh-qa-001--unit-tests)).
- [ ] Region-Klassifikation gegen `allowedRegions` (Default `eu-central-1`) — [`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation).
- [ ] Konsolen-Reporter ([`LH-FA-REP-001`](../../../../spec/lastenheft.md#lh-fa-rep-001--konsolenausgabe)); **kein Secret-Wert** in der Ausgabe ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
- [ ] Exit-Codes `0/1/2` über `app`-Aggregation — [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes).
- [ ] `make lint`, `make test` im `Makefile` verdrahtet und grün; in `harness/README.md`/`AGENTS.md`-Sensors-Tabelle nachgetragen (dann real).
- [ ] `make gates` grün (inkl. `doc-check`).
- [ ] Akzeptanz [`LH-ABN-001`](../../../../spec/lastenheft.md#lh-abn-001--kritische-lokale-fehlkonfiguration) und [`LH-ABN-002`](../../../../spec/lastenheft.md#lh-abn-002--eu-konforme-minimalumgebung) als Test reproduziert.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `internal/hexagon/core/model/` | neu | Finding/Severity/Config-Ring ([`ARC-003`](../../../../spec/architecture.md#2-schichten-und-constraints)) |
| `internal/hexagon/port/driven/environment.go` | neu | Env-Port ([`ARC-005`](../../../../spec/architecture.md#2-schichten-und-constraints)) |
| `internal/hexagon/core/rules/lenv.go` | neu | LENV-Regeln ([`ARC-004`](../../../../spec/architecture.md#2-schichten-und-constraints)) |
| `internal/hexagon/core/app/checklocal.go` | neu | Use-Case + Aggregation ([`ARC-002`](../../../../spec/architecture.md#2-schichten-und-constraints)) |
| `internal/adapter/driven/{envos,report}/` | neu | Env-Adapter + Konsolen-Reporter ([`ARC-006`](../../../../spec/architecture.md#2-schichten-und-constraints)) |
| `internal/adapter/driving/cli/` | neu | Composition Root, `local`-Befehl, Exit-Code ([`ARC-001`](../../../../spec/architecture.md#2-schichten-und-constraints)) |
| `Makefile` | update | `lint`/`test`/`build` ergänzen |

## 4. Trigger

[`ADR-0002`](../../adr/0002-implementierungssprache.md) accepted (Sprachwahl) → Slice nach `next/`, dann `in-progress/`.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Sprachwahl ([`ADR-0002`](../../adr/0002-implementierungssprache.md)) noch *Proposed* — blockiert Start.
- HCL-/YAML-Parserwahl betrifft erst spätere Slices, hier nicht relevant.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF** (siehe Kurs Modul 5 §Worked Mini-Example):
Es existiert kein Bestandscode; Spec und ADRs führen, der Code folgt. Kein
Brownfield-Inventur- oder Reconciliation-Aufwand.
