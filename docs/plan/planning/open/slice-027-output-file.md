# Slice slice-027: `--output FILE` für alle Ausgabeformate

**Status:** open.

**Welle:** [`welle-07-ci-komfort`](../welle-07-ci-komfort.md).

**Bezug:** [`LH-MOD-005`](../../../../spec/lastenheft.md#lh-mod-005--ci-modus), [`LH-FA-REP-002`](../../../../spec/lastenheft.md#lh-fa-rep-002--json-ausgabe), [`LH-FA-REP-003`](../../../../spec/lastenheft.md#lh-fa-rep-003--sarif-ausgabe), [`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Die CLI schreibt das Ergebnis mit `--output FILE` in eine Datei statt nach stdout, für Konsole, JSON und SARIF. Ohne `--output` bleibt das Verhalten unverändert.

## 2. Definition of Done

- [ ] Flag `--output FILE` in allen Modi; Pfad relativ zum aktuellen Verzeichnis; die Datei wird atomar geschrieben (temporär, dann umbenannt).
- [ ] Bei `--output` bleibt stdout leer; die `--ci`-Zusammenfassung geht weiter nach stderr; Exit-Codes unverändert.
- [ ] Fehler beim Schreiben (Verzeichnis fehlt, keine Rechte) ergeben Exit 2 mit klarer Meldung, ohne Dateiinhalt.
- [ ] Tests für alle drei Formate, Fehlerfall und unveränderte Ausgabe ohne Flag; `reference.md`, `ci.md` und README aktualisiert (Beispiel ohne Shell-Umleitung).
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `internal/adapter/driving/cli/cli.go` | update | Flag, Ausgabeziel, Fehlerpfad |
| `internal/adapter/driving/cli/cli_test.go` | update | Tests (in getrenntem Abschnitt, Konflikte vermeiden) |
| `docs/user/reference.md`, `docs/user/ci.md`, `README.md` | update | Doku-Drift-Test beachten |

## 4. Trigger

Welle 07 gestartet; unabhängig von den anderen Slices der Welle.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Eine Ausgabedatei im Scan-Root kann bei späteren Läufen als Eingabe gelten (z. B. SARIF unter `testdata`); in der Doku erwähnen.
- Überschreiben bestehender Dateien ist gewollt; kein Anhängen.
- Der Doku-Drift-Test (`reference.md`) muss das neue Flag kennen, sonst schlägt er an.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
