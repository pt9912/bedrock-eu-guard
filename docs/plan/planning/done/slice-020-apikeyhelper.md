# Slice slice-020: apiKeyHelper als Hinweis auf direkte Anthropic-Konfiguration

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** Qualitätsrunde nach welle-04-output-ci (Folgepunkt aus slice-004).

**Bezug:** [`LH-FA-CLAUDE-003`](../../../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Der Top-Level-Schlüssel `apiKeyHelper` in Claude-Settings wird als möglicher
Weg zu einer direkten Anthropic-Verbindung gemeldet (`WARN`), ohne den Befehl
je zu lesen oder auszugeben.

## 2. Verträge

- `model.ClaudeSettings.APIKeyHelperLine int` (0 = nicht gesetzt). Der Adapter `claudesettings` (w2) setzt die Zeile des Schlüssels `apiKeyHelper`; der Wert wird nie gelesen.
- Reason `APIKEY_HELPER`. Regelverhalten und Rangfolge stehen in [`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings) (neu).
- [`LH-FA-CLAUDE-001`](../../../../spec/lastenheft.md#lh-fa-claude-001--benutzerweite-claude-code-settings-prüfen) bleibt unverändert: ein Helper allein belegt keine direkte Konfiguration.

## 3. Definition of Done

- [ ] Kern (w1): Feld, Reason, Regel in `CheckClaudeEndpoints`, Tests inkl. Rangfolge, [`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings).
- [ ] Adapter (w2): `claudesettings` liefert `APIKeyHelperLine`; Testdaten mit Helper, Test „Befehl erscheint nirgends in der Ausgabe“.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/claude.go`, `core/rules/claude.go`, `spec/spezifikation.md` | update | w1 |
| `internal/adapter/driven/claudesettings/` | update | w2 |

## 5. Risiken und offene Punkte

- `WARN` statt `FAIL` ist eine Auslegung (ein Helper ist nicht zwingend direkte Nutzung); die Auftraggeber können auf `FAIL` bei fehlendem Bedrock-Flag verschärfen. Das wäre eine Änderung von [`SPEC-012`](../../../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings).
- Die Wirkung von `apiKeyHelper` zusammen mit `CLAUDE_CODE_USE_BEDROCK=1` ist nicht verlässlich belegt; die Meldung bleibt deshalb vorsichtig („manuell prüfen“).

## 6. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->
