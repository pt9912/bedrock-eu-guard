# Slice slice-002: Modell-ID-Klassifikation (allow/block) + Claude-Default-Modellvariablen

**Status:** done.

**Welle:** welle-01-lokaler-env-check.

**Bezug:** [`LH-FA-LENV-006`](../../../../spec/lastenheft.md#lh-fa-lenv-006--prüfung-von-claude-default-modellvariablen), [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Modell-/Profil-IDs werden nach [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids) in `allowed` / `blocked` /
`unclassified` eingeteilt (Block vor Allow), und `bedrock-eu-check local`
wendet das auf `ANTHROPIC_DEFAULT_SONNET_MODEL`, `..._HAIKU_MODEL` und
`..._OPUS_MODEL` an ([`LH-FA-LENV-006`](../../../../spec/lastenheft.md#lh-fa-lenv-006--prüfung-von-claude-default-modellvariablen)).

## 2. Definition of Done

- [x] `model`: `Config` um `AllowedModelPatterns`/`BlockedModelPatterns` mit den Defaults aus [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml); Klassifikationstyp `allowed|blocked|unclassified`.
- [x] `rules`: `Classify(id, cfg)` — Block hat Vorrang vor Allow, sonst `unclassified` (WARN).
- [~] Regel [`LH-FA-LENV-006`](../../../../spec/lastenheft.md#lh-fa-lenv-006--prüfung-von-claude-default-modellvariablen): FAIL bei `us.`/`global.`-Präfix oder nicht erlaubter Region in der ID; ungesetzte Variable erzeugt kein FAIL. **Teilweise erfüllt:** Die Default-Blockmuster decken `us.`, `global.`, `us-east-1` und `us-west-2` ab ([`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml)). Eine ID mit einer anderen nicht erlaubten Region (z. B. `ap-southeast-1`) wird nur als `unclassified` (WARN) gemeldet; die allgemeine Regionserkennung aus der ID gegen `allowedRegions` ist offen (Folgeaufgabe).
- [x] Unit-Tests je Zweig (blocked, allowed, unclassified, Block-vor-Allow, ungesetzt).
- [x] `make gates` grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `internal/hexagon/core/model/` | update | Muster-Config, Klassifikationstyp, Reason-Codes |
| `internal/hexagon/core/rules/modelid.go` | neu | `Classify` ([`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids)) |
| `internal/hexagon/core/rules/lenv006.go` | neu | LENV-006-Regel, in `LocalRules()` registrieren |

## 4. Trigger

slice-001 gemerged (oder Branch ab dessen Kern).

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Muster sind Regex ([`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml)); ungültige Muster aus einer späteren Config-Datei (Welle 2) brauchen eigene Fehlerbehandlung — hier nur Defaults.
- Aufteilung: w1 (Kern), siehe slice-003 für die Abhängigkeit.

## 7. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-10 (PR #5, w1).

- **Funktioniert:** `Classify` (Block vor Allow, Default-Muster aus der Spec) und LENV-006 mit Tests; ungültige Regex werden übersprungen.
- **Teilweise erfüllt:** LENV-006 „andere nicht erlaubte Regionen“ nur über die Default-Muster; allgemeine Region-Extraktion aus der ID gegen `allowedRegions` ist Folgeaufgabe.
- **Offen/Nicht-Ziel:** Muster aus einer Config-Datei (Welle 2) brauchen eigene Fehlerbehandlung für ungültige Regex.
- **Steering-Loop:** siehe slice-001.

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: reiner Kern ohne Bestandscode außerhalb
von slice-001.
