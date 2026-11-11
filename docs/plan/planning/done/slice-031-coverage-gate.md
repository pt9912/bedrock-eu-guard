# Slice slice-031: `coverage-gate` mit begründeter Schwelle und ADR

**Status:** next. **Verantwortlich:** w1 (Plan-Review: w2). **Freigabe:** Nutzer 2026-11-11 („031 coverage-gate + ADR“; beide Nutzer wählten 031).

**Welle:** [`welle-08-erweiterungen`](../welle-08-erweiterungen.md).

**Bezug:** [`AGENTS.md`](../../../../AGENTS.md) §3.6 (Gates nur per ADR lockern) und §4 (Eintrag erst, wenn das Target läuft), [`LH-QA-001`](../../../../spec/lastenheft.md#lh-qa-001--unit-tests), [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md) (Gates offline und per Digest gepinnt).

**Autor:** w1 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

`make coverage-gate` misst die Testabdeckung offline im gepinnten Go-Image und bricht ab, wenn sie unter einer **begründeten Schwelle** liegt. Die Schwelle und ihre Messweise stehen in einer neuen ADR, nicht in einem PR-Kommentar; das Target ist Teil von `make gates`.

## 2. Ist-Stand (Messung am 2026-11-11, `main` bei `4413da6`)

Gemessen mit `go test` im gepinnten Image, ohne Netz:

| Messweise | Ergebnis |
|---|---|
| `go test -coverprofile ./...` (jedes Paket nur durch die eigenen Tests) | **91,8 %** gesamt |
| `-coverpkg` über alle Pakete außer `coretest/*` und `cmd/` (Paket-übergreifend gemessen) | **95,4 %** gesamt |

Pakete bei der einfachen Messung: `rules` 97,2 %, `app` 97,5 %, `report` 97,0 %, `cli` 94,9 %, `awscli` 94,4 %, `iampolicy` 94,4 %, `configyaml` 93,1 %, `terraform` 92,1 %, `envsource` 90,4 %, `fs` 90,0 %, `claudesettings` 88,3 %, `model` **77,7 %** (der Kern wird zum Teil erst über `rules` und `app` geprüft, das zählt bei dieser Messweise nicht); `envos` und `secretkey` 100 %. Ohne Tests und mit 0 % stehen `cmd/bedrock-eu-check` (nur `main`) und die Test-Doubles unter `coretest/*`. Die einzige Funktion mit 0 % im Kern ist `Severity.String` in `model.go`.

## 3. Definition of Done

- [ ] **Neue ADR** `docs/plan/adr/0005-coverage-gate.md` legt fest: Messweise, Ausschlüsse mit Begründung, den Mindestwert mit Begründung und die Regel für spätere Änderungen. Sie wird als `Proposed` eingereicht und erst nach dem Sign-off der Auftraggeber (beide Nutzer, wie bei [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md)) `Accepted`; der Sign-off liegt vor dem Merge des Implementierungs-PR. Der ADR-Index ist aktualisiert.
- [ ] `make coverage-gate` läuft offline im gepinnten Image, schreibt das Profil nach `.tmp/` (nicht ins Repository), gibt die Gesamt- und die Paketwerte aus und endet mit einem Fehler unterhalb der Schwelle.
- [ ] Negativprobe: eine absichtlich zu hohe Schwelle lässt das Target scheitern (Beleg in der Closure-Notiz); danach der echte Wert grün.
- [ ] `coverage-gate` ist in `make gates` aufgenommen; `AGENTS.md` §4 trägt das Target in die Tabelle ein und streicht den Satz „entsteht in einem späteren Slice“.
- [ ] Der CI-Job `gates` ruft es über `make gates` auf; die Laufzeit steigt nicht merklich (Messung in der Closure-Notiz).
- [ ] `make verify-clean` und CI grün; `docs/user` bleibt unberührt (kein Vertrag nach außen).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

**Entscheidungen, die die ADR festlegt** (Vorschlag, der Review darf ihn ändern):

1. **Messweise:** Paket-übergreifend (`-coverpkg`) über die Produktionspakete; die einfache Messung unterschätzt den Kern (`model` 77,7 % gegen die Wahrheit, dass `rules` und `app` ihn voll nutzen).
2. **Ausschlüsse (jeder ist eine Lockerung und steht mit Begründung in der ADR):** `internal/hexagon/coretest/*` (Test-Doubles, selbst Testcode) und `cmd/bedrock-eu-check` (nur `main`, deckt die CLI-Verdrahtung über den Test des Pakets `cli` ab).
3. **Mindestwert:** Gesamt **90 %** (Ist-Stand 95,4 % mit Abstand von rund fünf Punkten, damit normale Änderungen nicht an der Schwelle scheitern, ein spürbarer Einbruch aber auffällt). Die ADR nennt diesen Mindestwert; der **wirksame Wert** steht im `Makefile` (Variable `COVERAGE_MIN`, Anfangswert 90). Eine zusätzliche Paket-Schwelle wird nur eingeführt, wenn die ADR einen Grund nennt; Vorschlag: keine, weil kleine Pakete ein Prozentziel verzerren.
4. **Änderung der Schwelle:** Anheben von `COVERAGE_MIN` ist ein gewöhnlicher PR (strenger ist keine Lockerung nach `AGENTS.md` §3.6); eine Accepted-ADR wird dabei nicht verändert (§3.5). Senken unter den Mindestwert der ADR oder ein neuer Ausschluss verlangt eine neue ADR mit `Supersedes`, nie einen PR-Kommentar.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `docs/plan/adr/0005-coverage-gate.md`, `docs/plan/adr/README.md` | neu / update | Entscheidung zur Messweise, zu den Ausschlüssen und zur Schwelle |
| `Makefile` (`coverage-gate`, `gates`), optional `tools/coverage-gate.sh` | neu / update | Messung offline im gepinnten Image, Schwellenvergleich ohne weitere Abhängigkeit (nur Go-Werkzeuge und POSIX-Shell) |
| `AGENTS.md` §4 | update | Target in die Tabelle, Satz „späterer Slice“ entfällt |
| `.gitignore` | prüfen | `.tmp/` ist bereits ausgeschlossen; sonst ergänzen |

Die Änderung am `Makefile` liegt im Bereich von w2 (Makefile, CI); w2 liest dort gegen, bevor gemergt wird.

## 5. Trigger

Nutzerfreigabe liegt vor (siehe Kopf). Plan-Review durch w2 liegt vor ([`2026-11-11-pr122-slice-031-plan.md`](../../../reviews/2026-11-11-pr122-slice-031-plan.md)); `open → next` ist erfolgt, `next → in-progress` mit dem Start der Umsetzung.

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- Eine zu hohe Schwelle erzeugt Tests, die nur Prozente füllen; eine zu niedrige gibt keine Sicherheit. Deshalb steht die Begründung in der ADR und der Abstand zum Ist-Stand ist benannt.
- `-coverpkg` führt die Tests ein zweites Mal aus: gemessen im gepinnten Image `go test ./...` rund 3 s, mit `-coverpkg` und Profil rund 6 s (Plan-Review); die Messung nach der Umsetzung steht in der Closure-Notiz (`make gates` bleibt offline und ohne Zusatzabhängigkeiten, [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md)).
- Die Gesamtzahl hängt vom Messumfang ab; wer ein Paket ausschließt, senkt die Latte. Jeder Ausschluss braucht den Eintrag in der ADR.
- Kein Einfluss auf Veröffentlichung oder Werkzeug-Version: das Gate betrifft nur Entwicklung und CI.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neues Gate und neue ADR im bestehenden Make-/CI-Ablauf, keine neue Sub-Area.
