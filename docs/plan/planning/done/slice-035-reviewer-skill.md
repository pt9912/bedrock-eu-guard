# Slice slice-035: Reviewer-Skill `.harness/skills/reviewer.md` anlegen

**Status:** in-progress. **Verantwortlich:** w2 (Gegenlesen: w1). **Freigabe:** Nutzer 2026-11-11 („ja“).

**Welle:** ohne Welle (Harness/Doku, kein Code).

**Bezug:** `AGENTS.md` §3 (Harte Regeln), [`docs/reviews/review-report.template.md`](../../../reviews/review-report.template.md) (verweist auf die Skill-Datei), Kurs Modul 10 (Review Harness, Worked Example „eine Reviewer-Skill-Datei schreiben“).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Das Repository bekommt die Reviewer-Skill-Datei `.harness/skills/reviewer.md`, auf die die Vorlage für Review-Reports verweist. Sie legt für jeden Review (Plan, Design, Code) fest, was der Reviewer mitbringt, wie er einstuft und in welchem Schema er berichtet. Bisher tragen unsere Reports „Skill: keiner“.

## 2. Quelle und Abweichung (festgelegt vom Nutzer)

- **Quelle:** die Vorlage `lab/templates/.harness/skills/reviewer.template.md` im Kurs-Repository `pt9912/ai-harness-course`, Stand Commit `4b597d84af9c` (2026-10-03, Branch `main`). Unser adoptierter Kurs-Tag ist v1.4.0; dort gibt es die Vorlage **nicht**, wohl aber das gleiche Worked Example in Modul 10.
- **Nutzerentscheidung:** Die Struktur der Vorlage wird übernommen (2026-11-11, „ja“).
- **Bewusst weggelassen:** die drei neueren HIGH-Anker der Vorlage („Norm nur im Template-Kommentar“, „Kommentar trägt keine der Kommentar-Klassen“, „Zustandsfeld trägt Chronik“). Sie verweisen auf Baseline-Dokumente (`grundlagen-harness-dateien.md`), die in unserem Tag nicht existieren und im Repository nicht vorliegen. Die Datei nennt die Quelle und diese Auslassung im Kopf.

## 3. Definition of Done

- [ ] `.harness/skills/reviewer.md` mit den Abschnitten der Vorlage: Kontext-Eingang, Klassifikation (HIGH, MEDIUM, LOW, INFO), „Was dieser Skill NICHT macht“, Output-Schema, Pflege (Steering-Loop).
- [ ] Die HIGH-Liste nennt mindestens zwei repo-spezifische Regeln, die ein generischer Skill nicht abdeckt, jeweils mit Anker in `AGENTS.md`, Lastenheft oder `releasing.md` (nichts Erfundenes).
- [ ] Kopf der Datei nennt Quelle (Repository, Pfad, Commit), den adoptierten Kurs-Tag und die bewusst weggelassenen Anker.
- [ ] Das Output-Schema ist mit der Report-Vorlage vereinbar (`pfad`: Datei, Zitat, Zeile als Lesehilfe); Abweichungen zwischen Skill und Report-Vorlage sind benannt.
- [ ] Die Report-Vorlage verweist weiter auf die Datei; neue Reports tragen als Skill `.harness/skills/reviewer.md`.
- [ ] `.harness/` ist nicht in `.gitignore` (nur `.harness/cache/` ist gitignored); `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/skills/reviewer.md` | neu | Skill-Datei nach der Vorlage, repo-spezifisch gefüllt |
| `.gitignore` | prüfen, ggf. update | `.harness/skills/` darf nicht ignoriert sein |
| `docs/reviews/review-report.template.md` | prüfen | Verweis und Schema-Feld `pfad` abgleichen (Änderung nur nach Abstimmung mit w1) |

## 5. Trigger

Nutzerentscheidung vom 2026-11-11 („ja“ zur Struktur der Vorlage aus dem aktuellen Kurs-Repository).

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben; Gegenlesen durch w1.

## 7. Risiken und offene Punkte

- Die Vorlage stammt aus einem neueren Kurs-Stand als unser Tag; Teile der Vorlage setzen Dokumente voraus, die wir nicht haben (siehe Abschnitt 2). Eine Anhebung des Kurs-Tags ist eine eigene Entscheidung der Nutzer.
- Das Output-Schema der Vorlage (`pfad` als wörtliches Zitat, Feld `klasse`) weicht von der Report-Vorlage ab (`pfad` als `Datei:Zeile`, kein `klasse`); die Report-Vorlage nennt den Skill als verbindliche Quelle.
- Die Datei wird, wie die Vorlage verlangt, nicht überschrieben, sondern versioniert.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Harness-Datei ohne Code.
