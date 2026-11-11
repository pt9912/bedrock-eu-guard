# Slice slice-027: `--output FILE` für alle Ausgabeformate

**Status:** done.

**Welle:** [`welle-07-ci-komfort`](../welle-07-ci-komfort.md).

**Bezug:** [`LH-MOD-005`](../../../../spec/lastenheft.md#lh-mod-005--ci-modus), [`LH-FA-REP-002`](../../../../spec/lastenheft.md#lh-fa-rep-002--json-ausgabe), [`LH-FA-REP-003`](../../../../spec/lastenheft.md#lh-fa-rep-003--sarif-ausgabe), [`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Die CLI schreibt das Ergebnis mit `--output FILE` in eine Datei statt nach stdout, für Konsole, JSON und SARIF. Ohne `--output` bleibt das Verhalten unverändert.

## 2. Definition of Done

- [x] Flag `--output FILE` in allen Modi; Pfad relativ zum aktuellen Verzeichnis; die Datei wird atomar geschrieben (temporär, dann umbenannt).
- [x] Bei `--output` bleibt stdout leer; die `--ci`-Zusammenfassung geht weiter nach stderr; Exit-Codes unverändert.
- [x] Fehler beim Schreiben (Verzeichnis fehlt, keine Rechte) ergeben Exit 2 mit klarer Meldung, ohne Dateiinhalt. Endet der Lauf vor dem Ergebnis mit einem Tool-Fehler (Exit 2), wird die Ausgabedatei gar nicht angelegt (keine halbe Datei); eine bestehende Datei wird nur bei vorhandenem Ergebnis atomar ersetzt.
- [x] Tests für alle drei Formate, Fehlerfall und unveränderte Ausgabe ohne Flag; `reference.md`, `ci.md` und README aktualisiert (Beispiel ohne Shell-Umleitung).
- [x] `make verify-clean` und CI grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

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

**Abgeschlossen:** 2026-11-11 (PRs #96 und #103, beide w2; Spezifikation [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes), Abschnitt „Ausgabeziel `--output FILE`“, PR #97 von w1; Review-Report [`2026-11-11-pr96-slice-027-output-file.md`](../../../reviews/2026-11-11-pr96-slice-027-output-file.md), PR #102 von w1).

- **Funktioniert:** `--output FILE` für alle Prüfbefehle und Formate. Das Ergebnis geht in die Datei, stdout bleibt leer, die `--ci`-Zusammenfassung steht weiter auf stderr, die Exit-Codes (auch mit `--fail-on`) bleiben gleich. Die Datei wird atomar ersetzt (temporäre Datei im Zielverzeichnis, dann Umbenennen, Berechtigung `0644`) und nur geschrieben, wenn ein Ergebnis vorliegt: ein Tool-Fehler vor dem Ergebnis legt keine Datei an und lässt eine bestehende unverändert; ein Teilergebnis bei spätem Tool-Fehler wird wie auf stdout geschrieben. Schreibfehler sind Exit `2`. Tests in `output_test.go` (Gleichheit mit stdout in drei Formaten, `--ci` und `--fail-on`, Ersetzen ohne Temp-Rest, Schreibfehler, früher Tool-Fehler, Teilergebnis, Aufruffehler, Rename-Fehler ohne Temp-Rest, Berechtigung); Doku in `reference.md`, `ci.md` und README.
- **Anders als geplant:** Der Lebenszyklus `open → next → in-progress → done` wurde übersprungen: Die Plan-Datei lag bis zur Closure in `open/` (Versäumnis beider Agenten, in der Roadmap festgehalten; ab slice-024 gilt der Lebenszyklus). Auf Review-Hinweis von w1 kamen Symlink- und `0644`-Hinweis in die Referenz und ein Test auf Temp-Reste. Der Review-Befund F-3 (feste Fehlermeldung ohne Ursache) ist in #103 erledigt: Die Meldung nennt jetzt die Klasse (Zielverzeichnis fehlt, Ziel ist ein Verzeichnis, keine Berechtigung, sonstiger Fehler), weiterhin ohne Pfad und Inhalt.
- **Offen:** Die temporäre Datei `.bedrock-eu-check-*` bleibt bei einem harten Abbruch (zum Beispiel `kill -9`) im Zielverzeichnis zurück (Review-Befund F-1, bekannte Grenze). Ein Konsument, der nur die Datei auswertet, kann ein Teilergebnis bei Exit `2` für vollständig halten (F-2); der Exit-Code ist maßgeblich und in SPEC-030 festgelegt. Nicht geprüft: Windows, Dateisysteme ohne atomares Umbenennen, gleichzeitige Läufe auf dasselbe Ziel, sehr große Ausgaben (der Reporter schreibt zuerst in einen Speicherpuffer).
- **Steering-Loop:** Eine Ausgabe-Option verändert den Vertrag für CI-Konsumenten (wann existiert die Datei, was steht bei Fehlern darin); das gehört in die Spezifikation und nicht nur in die Referenz, und ein unabhängiger Review-Report am gebauten Binary findet Dinge, die der Diff nicht zeigt (hier F-2 und F-3).
- **Folge-Slices:** keine.

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
