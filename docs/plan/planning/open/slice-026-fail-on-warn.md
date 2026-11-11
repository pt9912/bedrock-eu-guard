# Slice slice-026: Schwelle `--fail-on warn`

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** [`welle-07-ci-komfort`](../welle-07-ci-komfort.md).

**Bezug:** [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes), [`LH-FA-EXIT-001`](../../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung), [`LH-FA-EXIT-002`](../../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung), [`LH-FA-EXIT-003`](../../../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler), [`LH-MOD-005`](../../../../spec/lastenheft.md#lh-mod-005--ci-modus), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-11.

---

## 1. Ziel

Pipelines können Warnungen als Fehler behandeln: mit `--fail-on warn` endet der Lauf mit Exit `1`, sobald mindestens
ein `WARN` oder `FAIL` vorliegt. Ohne die Option bleibt das Verhalten **unverändert** (Exit `1` nur bei `FAIL`).

## 2. Verträge

- **Kern (w1):** `app.ExitCodeAt(findings, err, failOn model.Severity) int` mit der Semantik: Fehler → `2`; höchster Schweregrad `>= failOn` → `1`; sonst `0`. Die bestehende Funktion `app.ExitCode` bleibt und ist gleichbedeutend mit `failOn = FAIL`. Zulässige Schwellen: `FAIL` (Default) und `WARN`; `PASS` ist keine zulässige Schwelle.
- **Spec (w1):** [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes) wird fortgeschrieben: Tabelle um die Bedingung „mit `--fail-on warn`: mindestens ein `WARN` oder `FAIL`“ ergänzt; Hinweis, dass die Option eine Erweiterung ist und [`LH-FA-EXIT-001`](../../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung)/[`LH-FA-EXIT-002`](../../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung) im Default erfüllt bleiben.
- **CLI (w2):** Option `--fail-on fail|warn` für alle Prüfbefehle; ungültiger Wert oder fehlendes Argument ist ein Aufruffehler (Exit `2`, `usage`). Die Zusammenfassungszeile im CI-Modus zeigt weiterhin den tatsächlichen Exit-Code. Reporter, JSON und SARIF bleiben unverändert (der Schweregrad der Befunde ändert sich nicht, nur der Exit-Code).

## 3. Verhalten

| Befunde | Default | `--fail-on warn` |
|---|---|---|
| nur `PASS` | `0` | `0` |
| `WARN`, kein `FAIL` | `0` | `1` |
| mindestens ein `FAIL` | `1` | `1` |
| Tool-Fehler | `2` | `2` |

## 4. Definition of Done

- [ ] Kern (w1): `ExitCodeAt` mit Tabellentest über alle Kombinationen; Spec fortgeschrieben.
- [ ] CLI (w2): Option, Validierung, Tests je Modus (Default unverändert, `warn` greift, ungültiger Wert Exit `2`); Referenz, Handbuch (Aufgabe „In CI einbinden“), `ci.md` und Fehlersuche aktualisiert; der `jq`-Workaround verweist auf die Option.
- [ ] Regressionstest: ohne Option unverändertes Verhalten in allen Modi.
- [ ] `make verify-clean` und CI grün; Doku-Drift-Test grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/app/checklocal.go` (oder neue Datei), `spec/spezifikation.md` | update | w1 |
| `internal/adapter/driving/cli/`, `docs/user/{reference,handbuch,ci,troubleshooting}.md`, README | update | w2 |

## 6. Trigger

Plan der Welle 07 gemerged; Slice-Plan gemerged.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- **Exit-Code-Vertrag:** Konsumenten verlassen sich auf `0`/`1`/`2`. Die Schwelle darf nur auf ausdrückliche Anforderung greifen; es gibt keine Umgebungsvariable und keinen Eintrag in `bedrock-eu-check.yml` (kein verstecktes Umschalten).
- **Konfigurierbarkeit:** Ob die Schwelle zusätzlich in der Konfigurationsdatei stehen soll, ist eine bewusste Nicht-Entscheidung dieses Slices (nur Kommandozeile); eine Aufnahme wäre eine Änderung von [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml).
- **Erwartete Warnungen:** Viele Umgebungen haben dauerhafte, erwartbare `WARN`-Befunde (zum Beispiel fehlende Benutzer-Settings); `--fail-on warn` kann dort nur mit `exclude` oder angepasster Konfiguration sinnvoll genutzt werden. Das gehört in die Doku.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: kleine Erweiterung des Exit-Code-Mappings auf bestehendem Code.
