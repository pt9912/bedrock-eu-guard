# Slice slice-001: Hexagon-Gerüst + Env-/Region-Check + Exit-Codes + Konsolen-Reporter

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-01-lokaler-env-check.

**Bezug:** `LH-FA-LENV-001`, `LH-FA-LENV-002`, `LH-FA-LENV-003`,
`LH-FA-LENV-004`, `LH-FA-EXIT-001`, `LH-FA-EXIT-002`,
`LH-FA-EXIT-003`, `LH-FA-REP-001`, `SPEC-001`, `SPEC-011`,
`SPEC-030`, `ARC-001`–`ARC-007`, ADR-0002, ADR-0003.

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-06-24.

---

## 1. Ziel

Das hexagonale Grundgerüst stehen lassen und den `local`-Modus soweit
ausführen, dass `bedrock-eu-check local` die Region-/Bedrock-/Anthropic-
Env-Variablen prüft, `PASS`/`WARN`/`FAIL` auf der Konsole ausgibt und mit
`0`/`1`/`2` endet.

## 2. Definition of Done

- [ ] Paketstruktur nach ADR-0003 angelegt: `internal/hexagon/core/{model,rules,app}`, `internal/hexagon/port/driven`, `internal/adapter/driving/cli`, `internal/adapter/driven/{envos,report}`.
- [ ] `model`: `Finding`, `Severity` (`PASS/WARN/FAIL`), Reason-Codes, `Config` (importfreier Ring) — `SPEC-001`.
- [ ] `Environment`-Port + `envos`-Adapter + `coretest/fakeenv`-Double.
- [ ] Regeln `LH-FA-LENV-001` (AWS_REGION), `-002` (AWS_DEFAULT_REGION-Fallback), `-003` (CLAUDE_CODE_USE_BEDROCK), `-004` (ANTHROPIC_API_KEY) mit Unit-Tests je Regel (`LH-QA-001`).
- [ ] Region-Klassifikation gegen `allowedRegions` (Default `eu-central-1`) — `SPEC-011`.
- [ ] Konsolen-Reporter (`LH-FA-REP-001`); **kein Secret-Wert** in der Ausgabe (`LH-NF-004`).
- [ ] Exit-Codes `0/1/2` über `app`-Aggregation — `SPEC-030`.
- [ ] `make lint`, `make test` im `Makefile` verdrahtet und grün; in `harness/README.md`/`AGENTS.md`-Sensors-Tabelle nachgetragen (dann real).
- [ ] `make gates` grün (inkl. `doc-check`).
- [ ] Akzeptanz `LH-ABN-001` und `LH-ABN-002` als Test reproduziert.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `internal/hexagon/core/model/` | neu | Finding/Severity/Config-Ring (`ARC-003`) |
| `internal/hexagon/port/driven/environment.go` | neu | Env-Port (`ARC-005`) |
| `internal/hexagon/core/rules/lenv.go` | neu | LENV-Regeln (`ARC-004`) |
| `internal/hexagon/core/app/checklocal.go` | neu | Use-Case + Aggregation (`ARC-002`) |
| `internal/adapter/driven/{envos,report}/` | neu | Env-Adapter + Konsolen-Reporter (`ARC-006`) |
| `internal/adapter/driving/cli/` | neu | Composition Root, `local`-Befehl, Exit-Code (`ARC-001`) |
| `Makefile` | update | `lint`/`test`/`build` ergänzen |

## 4. Trigger

ADR-0002 accepted (Sprachwahl) → Slice nach `next/`, dann `in-progress/`.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Sprachwahl (ADR-0002) noch *Proposed* — blockiert Start.
- HCL-/YAML-Parserwahl betrifft erst spätere Slices, hier nicht relevant.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF** (siehe Kurs Modul 5 §Worked Mini-Example):
Es existiert kein Bestandscode; Spec und ADRs führen, der Code folgt. Kein
Brownfield-Inventur- oder Reconciliation-Aufwand.
