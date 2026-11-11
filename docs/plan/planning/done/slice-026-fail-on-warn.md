# Slice slice-026: Schwelle `--fail-on warn`

**Status:** done.

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
- **CLI (w2):** Option `--fail-on fail|warn` für die Prüfbefehle `local`, `terraform`, `aws` und `all`; beim Befehl `version` ist sie ein Aufruffehler (Exit `2`, `usage`), wie andere unpassende Optionen; ungültiger Wert oder fehlendes Argument ist ein Aufruffehler (Exit `2`, `usage`). Die Zusammenfassungszeile im CI-Modus zeigt weiterhin den tatsächlichen Exit-Code. Reporter, JSON und SARIF bleiben unverändert (der Schweregrad der Befunde ändert sich nicht, nur der Exit-Code).

## 3. Verhalten

| Befunde | Default | `--fail-on warn` |
|---|---|---|
| nur `PASS` | `0` | `0` |
| `WARN`, kein `FAIL` | `0` | `1` |
| mindestens ein `FAIL` | `1` | `1` |
| Tool-Fehler | `2` | `2` |

Bei `all` mit fehlendem `aws`-Binary bleibt es bei Exit `2`; die Schwelle wird nur auf ein vollständiges Ergebnis angewendet (der Tool-Fehler hat Vorrang).

## 4. Definition of Done

- [x] Kern (w1): `ExitCodeAt` mit Tabellentest über alle Kombinationen; Spec fortgeschrieben.
- [x] CLI (w2): Option, Validierung (auch `version` mit der Option = Exit `2`), Tests je Modus (Default unverändert, `warn` greift, ungültiger Wert Exit `2`); Referenz (Option und Exit-Tabelle; Doku-Drift-Test beachten), Handbuch (Aufgabe „In CI einbinden“ und Fehlersuche: „`WARN` ändert den Exit-Code nur mit `--fail-on warn`“), `ci.md` (Beispiel „strenge Pipeline“), README-Optionsliste aktualisiert; der `jq`-Workaround verweist auf die Option.
- [x] Regressionstest: ohne Option unverändertes Verhalten in allen Modi.
- [x] `make verify-clean` und CI grün; Doku-Drift-Test grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

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

**Abgeschlossen:** 2026-11-11 (PRs #89 Kern/w1, #91 CLI/Doku/w2).

- **Funktioniert:** `--fail-on fail|warn` für `local`, `terraform`, `aws` und `all`. Der Kern liefert `app.ExitCodeAt(findings, err, failOn)` (Tool-Fehler hat Vorrang vor der Schwelle; nur `WARN` und `FAIL` sind zulässige Schwellen, alles andere gilt als `FAIL`) und `model.ParseFailOn` als einzige Quelle der zulässigen Werte; `app.ExitCode` ist unverändert die Schwelle `FAIL`. Die CLI nimmt die Option an, weist ungültige, leere oder fehlende Werte sowie die Option beim Befehl `version` als Aufruffehler ab (Exit `2`, `usage`) und zeigt in der Zusammenfassung des CI-Modus den tatsächlichen Exit-Code. [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes) ist fortgeschrieben (Erweiterung, keine Umgebungsvariable, kein Konfigurationsfeld); Referenz, `ci.md`, Handbuch, Fehlersuche, Schnellstart und README sind aktualisiert. Tests: Tabellentest über alle Kombinationen im Kern; in der CLI Default unverändert, `warn` greift, Ausgabe in Konsole/JSON/SARIF unverändert, Tool-Fehler bleibt Exit `2`, ungültige Aufrufe.
- **Anders als geplant:** nichts Wesentliches. Die Ergänzung des `jq`-Workarounds in der Doku verweist jetzt auf die Option.
- **Offen:** Dauerhafte, erwartbare `WARN`-Befunde (zum Beispiel fehlende Benutzer-Settings in Containern) machen `--fail-on warn` nur mit `exclude` oder angepasster Konfiguration brauchbar; das steht in der Doku. Die Schwelle ist bewusst nur auf der Kommandozeile einstellbar.
- **Steering-Loop:** Die Menge der zulässigen Werte liegt einmal im Kern (`ParseFailOn`) und wird von der CLI nur verwendet; so können Kern und Option nicht auseinanderlaufen. Ein Standardverhalten, das sich nicht ändern darf, braucht einen eigenen Regressionstest, nicht nur Tests der neuen Option.
- **Folge-Slices:** keine.

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: kleine Erweiterung des Exit-Code-Mappings auf bestehendem Code.
