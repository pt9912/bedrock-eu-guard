# Slice slice-003: Mantle-Prüfung (CLAUDE_CODE_USE_MANTLE)

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-01-lokaler-env-check.

**Bezug:** [`LH-FA-LENV-005`](../../../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle), [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check local` erkennt `CLAUDE_CODE_USE_MANTLE=1` und meldet
WARN, solange keine Modell-/Region-Prüfung möglich ist, bzw. PASS, wenn
alle zusätzlich konfigurierten Modell-IDs erlaubt sind
([`LH-FA-LENV-005`](../../../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle)).

> **Umdefiniert (2026-11-10):** Der ursprünglich geplante Inhalt
> (Anthropic-Key- und Bedrock-Flag-Regeln, [`LH-FA-LENV-003`](../../../../spec/lastenheft.md#lh-fa-lenv-003--prüfung-von-claude_code_use_bedrock)/[`-004`](../../../../spec/lastenheft.md#lh-fa-lenv-004--prüfung-von-anthropic_api_key)) wurde
> bereits in slice-001 geliefert. Dieser Slice deckt stattdessen
> [`LH-FA-LENV-005`](../../../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle) ab.

## 2. Definition of Done

- [ ] Regel [`LH-FA-LENV-005`](../../../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle): Mantle nicht gesetzt → kein Befund; aktiv ohne Modellvariablen → WARN; aktiv und alle gesetzten Modell-IDs `allowed` → PASS; blockierte ID → FAIL über LENV-006.
- [ ] Nutzt `Classify` aus slice-002 (keine eigene Klassifikation).
- [ ] Unit-Tests je Zweig; Registrierung in `LocalRules()`.
- [ ] `make gates` grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `internal/hexagon/core/rules/lenv005.go` | neu | Mantle-Regel |
| `internal/hexagon/core/rules/lenv.go` | update | Einzeiler in `LocalRules()` |

## 4. Trigger

slice-002 gepusht (liefert `Classify`).

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Konfigurierbares Verhalten (`mantle: warn|allow|block`, [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml)) erst mit der Config-Datei in Welle 2; hier Default `warn`.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
