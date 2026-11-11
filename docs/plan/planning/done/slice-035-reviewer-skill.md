# Slice slice-035: Reviewer-Skill `.harness/skills/reviewer.md` anlegen

**Status:** done. **Verantwortlich:** w2 (Gegenlesen: w1). **Freigabe:** Nutzer 2026-11-11 („ja“).

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

- [x] `.harness/skills/reviewer.md` mit den Abschnitten der Vorlage: Kontext-Eingang, Klassifikation (HIGH, MEDIUM, LOW, INFO), „Was dieser Skill NICHT macht“, Output-Schema, Pflege (Steering-Loop).
- [x] Die HIGH-Liste nennt mindestens zwei repo-spezifische Regeln, die ein generischer Skill nicht abdeckt, jeweils mit Anker in `AGENTS.md`, Lastenheft oder `releasing.md` (nichts Erfundenes).
- [x] Kopf der Datei nennt Quelle (Repository, Pfad, Commit), den adoptierten Kurs-Tag und die bewusst weggelassenen Anker.
- [x] Das Output-Schema ist mit der Report-Vorlage vereinbar (`pfad`: Datei, Zitat, Zeile als Lesehilfe); Abweichungen zwischen Skill und Report-Vorlage sind benannt.
- [x] Die Report-Vorlage verweist weiter auf die Datei; neue Reports tragen als Skill `.harness/skills/reviewer.md`.
- [x] `.harness/` ist nicht in `.gitignore` (nur `.harness/cache/` ist gitignored); `make verify-clean` und CI grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

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

**Abgeschlossen:** 2026-11-11 (PRs #117 Plan und #118 Skill, w2; Gegenlesen mit Report [`2026-11-11-pr118-slice-035-reviewer-skill.md`](../../../reviews/2026-11-11-pr118-slice-035-reviewer-skill.md) durch w1, vor dem Merge).

- **Funktioniert:** `.harness/skills/reviewer.md` liegt im Repository. Sie folgt der Struktur der Kurs-Vorlage (Kontext-Eingang, Klassifikation, „Was dieser Skill NICHT macht“, Output-Schema, Pflege) und nennt Quelle (Kurs-Repository, Commit `4b597d84af9c`), den adoptierten Kurs-Tag v1.4.0 und die bewusst weggelassenen drei HIGH-Anker. Die HIGH-Liste hat vier Punkte aus dem Kurs (ADR-Verstoß, Secret-Wert in der Ausgabe, Korrektheitsfehler gegen [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes), Suppression) und je einen Anker in `AGENTS.md` §3.2, §3.3, §3.5, §3.6, §3.8, §3.9, dem Lastenheft und `releasing.md`. Die Report-Vorlage folgt dem Skill: `pfad` als Datei und Zitat, neues Feld `klasse`, Skill-Verweis als Link. Ältere Reports (bis einschließlich Nachlauf zu PR #107) tragen das alte Format und blieben unverändert; der Skill vermerkt das.
- **Anders als geplant:** nichts Wesentliches. Der Review von w1 trennte die vermischte Zeile zu Suppression (§3.2) und Gate-Lockerung (§3.6) in zwei Punkte, korrigierte ein falsches Datum und stufte den Anker `releasing.md` als schwächer ein (Rang 6 der Source Precedence, keine Harte Regel); beides ist eingearbeitet.
- **Offen:** Die drei weggelassenen Anker der Vorlage bleiben ausgelassen, bis die Nutzer den Kurs-Tag anheben. Der Skill wurde noch auf keinem Diff „scharf“ geprüft (Übung aus Modul 10: Lauf auf einem realen Diff, ob ein Repo-HIGH greift). `docs/reviews/review-report.template.md` und der Skill können in Einzelheiten wieder auseinanderlaufen; der Skill ist die verbindliche Quelle.
- **Steering-Loop:** Eine Vorlage aus einem neueren Kurs-Stand nicht kopieren, sondern gegen den adoptierten Stand prüfen und Auslassungen im Kopf der Datei nennen; Anker in der HIGH-Liste wörtlich gegen `AGENTS.md` lesen, weil verwandte Regeln (§3.2 und §3.6) leicht vermischt werden.
- **Folge-Slices:** keine.

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Harness-Datei ohne Code.
