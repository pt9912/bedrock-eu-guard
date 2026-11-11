# Reviewer-Skill — bedrock-eu-guard

* Status: Accepted
* Bezug: [`AGENTS.md`](../../AGENTS.md) §3 (Harte Regeln), [`ADR-0003`](../../docs/plan/adr/0003-hexagonale-architektur.md), [`ADR-0004`](../../docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md), [`docs/user/releasing.md`](../../docs/user/releasing.md), Kurs Modul 10 (Review Harness)
* Gilt für: das Gegenlesen eines Plans, eines Entwurfs oder eines Pull Requests durch den jeweils **anderen** Agenten (w1 oder w2). Es gibt kein Make-Target dafür.
* Quelle: Vorlage `lab/templates/.harness/skills/reviewer.template.md` im Kurs-Repository `pt9912/ai-harness-course`, Commit `4b597d84af9c` (Branch `main`). Unser adoptierter Kurs-Tag ist v1.4.0 und enthält die Vorlage nicht, wohl aber dasselbe Worked Example in Modul 10.
* Bewusst weggelassen: Die Vorlage enthält drei weitere HIGH-Anker („Norm nur im Template-Kommentar“, „Kommentar trägt keine der Kommentar-Klassen“, „Zustandsfeld trägt Chronik“). Sie verweisen auf Baseline-Dokumente des neueren Kurs-Standes, die in unserem Tag und im Repository nicht vorliegen; sie werden mit einer Anhebung des Kurs-Tags neu bewertet (Entscheidung der Nutzer).

Report-Gerüst pro Lauf: [`docs/reviews/review-report.template.md`](../../docs/reviews/review-report.template.md), ein Report pro Lauf unter `docs/reviews/`.

## Kontext-Eingang (Pflicht)

Was der Reviewer *immer* mitbringt, bevor er den Diff liest:

- Diff des Pull Requests (bei einem Plan-Review: der Plan selbst)
- der Slice-Plan (Ziel, Definition of Done, Verträge) und, falls berührt, [`spec/spezifikation.md`](../../spec/spezifikation.md)
- [`spec/lastenheft.md`](../../spec/lastenheft.md) für die referenzierten `LH-*`-IDs
- die ADRs, deren ID im Pull Request oder in einer Commit-Message vorkommt
- [`AGENTS.md`](../../AGENTS.md) §3 „Harte Regeln“
- die letzten Reports in [`docs/reviews/`](../../docs/reviews/) (etwa fünf)

Ohne diesen Block sieht der Reviewer den Code, aber nicht *die Verträge, gegen die er prüft*.

## Klassifikation

Jeder Anker HIGH, MEDIUM und LOW hat eine *konkrete* Liste. INFO ist bewusst kurz (Ergänzungs-Kanal, nicht Hauptkanal).

**HIGH** (blockiert den Merge) — eines der folgenden:

- **ADR-Verstoß**, zum Beispiel eine Layer-Grenze nach [`ADR-0003`](../../docs/plan/adr/0003-hexagonale-architektur.md) (Importregeln, geprüft von `internal/arch`) oder eine Abhängigkeit am Vendoring vorbei ([`ADR-0004`](../../docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md))
- **Sicherheits-Anti-Pattern: Secret-Wert in Ausgabe, Fehlertext, Log oder Testdaten** (Werte von AWS-Zugangsdaten, Anthropic-Schlüsseln, Token; Testdaten nur mit dem Marker `GEHEIM`). Anker: [`AGENTS.md`](../../AGENTS.md) §3.7 und [`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)
- **Korrektheitsfehler im kritischen Pfad:** Ein Schweregrad oder Exit-Code weicht von [`SPEC-030`](../../spec/spezifikation.md#spec-030--exit-codes) und dem Regelkatalog ab, zum Beispiel eine nicht erlaubte Region oder der direkte Anthropic-Zugang ergibt `PASS` oder `WARN` statt `FAIL`, oder ein Tool-Fehler ergibt Exit `0`
- **Inline-Suppression von Lint-Findings** (`//nolint` und Verwandtes) statt einer begründeten Ausnahme in der zentralen Konfiguration. Anker: [`AGENTS.md`](../../AGENTS.md) §3.2
- **Absenkung einer Schwelle oder Strenge eines Gates ohne ADR** (Coverage, Linter-Strenge, Architekturregel). Anker: [`AGENTS.md`](../../AGENTS.md) §3.6
- **Repo-spezifisch #1 — Verschiebung und Inhaltsänderung in einem Commit:** `git mv` und Inhaltsänderung derselben Datei in einem Commit, auch bei den Moves des Slice-Lebenszyklus (`open → next → in-progress → done`). Anker: [`AGENTS.md`](../../AGENTS.md) §3.3
- **Repo-spezifisch #2 — Änderung an einem Accepted-ADR in der Sache.** Anker: [`AGENTS.md`](../../AGENTS.md) §3.5
- **Repo-spezifisch #3 — schreibender AWS-Zugriff oder Remediation** (das Werkzeug liest nur). Anker: [`AGENTS.md`](../../AGENTS.md) §3.8 und [`LH-SEC-003`](../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)
- **Repo-spezifisch #4 — hartcodierte Allow- oder Blocklisten** statt konfigurierbarer Listen. Anker: [`AGENTS.md`](../../AGENTS.md) §3.9
- **Repo-spezifisch #5 — Veröffentlichungsschritt** (Tag, Release, Image-Push, Feature-Publish) **oder neue Veröffentlichungsfähigkeit ohne dokumentierte Freigabe beider Nutzer.** Anker: [`docs/user/releasing.md`](../../docs/user/releasing.md) (Quelle im Rang 6 der Source Precedence, keine Harte Regel, aber verbindlicher Ablauf)

**MEDIUM** (sollte vor dem Merge geklärt werden) — eines der folgenden:

- unklare Fehlerbehandlung am Rand des Spec-Bereichs
- fehlende Negativtests bei neuem öffentlichem Vertrag (neue Option, neuer Reason-Code, neues Konfigurationsfeld)
- Wiederholung eines Musters, das schon zweimal `LOW` war
- Slice-Lebenszyklus nicht eingehalten (Plan nicht in `next/` oder `in-progress/`, obwohl gearbeitet wird). Anker: [`docs/plan/planning/README.md`](../../docs/plan/planning/README.md)

**LOW** — *mit Konventions-Anker* (ADR, Harte Regel, Linter-Regel, Eintrag in diesem Skill): stilistisch unschön ohne semantische Auswirkung, einmalige Tippfehler, unbenutzte Imports.

**INFO** — Hinweis ohne erwartete Aktion (zum Beispiel „diese Stelle hat ein passendes Testpendant, das du nicht kennst“).

## Was dieser Skill NICHT macht

- Keine Lösungsvorschläge („schreib das so“) — der Reviewer kategorisiert, die Implementation entscheidet.
- Kein Refactoring-Vorschlag, der über den Diff hinausgeht.
- Keine Verifikation gegen die Definition of Done — das ist Aufgabe des Verifiers (Modul 11); der Report ersetzt sie nicht.
- Keine Validation gegen reale Bedürfnisse — das ist Aufgabe des Validators.
- **Kein Stil-Polizist:** Formatierung oder Benennung ohne Konventions-Anker ist kein Finding.
- **Kein HIGH- oder MEDIUM-Finding ohne Failure-Szenario:** Was sich nicht als konkretes Versagen erzählen lässt, wird nicht als HIGH oder MEDIUM gemeldet.

Wenn etwas auffällt, das in diese Kategorien gehört: ein INFO-Finding mit Verweis auf die zuständige Rolle.

## Output-Schema

Jedes Finding:

- `kategorie`: HIGH | MEDIUM | LOW | INFO
- `quelle`: ADR-ID, `LH-*`-ID, Hard-Rule-Name oder „Maintainability“
- `pfad`: Datei · wörtliches, in der Datei eindeutig auffindbares Kurzzitat der Stelle als Anker; die Zeile darf als Lesehilfe dazu, ist aber nicht der Anker
- `befund`: 1–2 Sätze, beobachtbar, ohne Lösungsvorschlag
- `verifizierbar`: ja/nein — gibt es einen Gate-Lauf, der es bestätigen würde?
- `klasse`: stabile Kurz-Bezeichnung des Fehlermusters, zum Beispiel „Tie-Break in sortierender Operation nicht dokumentiert“ — speist den Steering-Loop-Zähler (siehe Pflege)

Zusätzlich am Ende: eine Zeile „geprüft, ohne Befund“ pro betrachtetem Verzeichnis (Negativbefund-Zeile — sonst ist „keine Findings“ nicht von „nicht geprüft“ unterscheidbar). Report-Gerüst für den ganzen Lauf: [`docs/reviews/review-report.template.md`](../../docs/reviews/review-report.template.md), ein Report pro Lauf, Folgeläufe als neue Datei statt Überschreibung.

Reports bis einschließlich des Nachlaufs zu PR #107 (`pr90`, `pr96`, `pr106`, `pr107` und der Nachlauf) tragen das ältere Format (`pfad` als `Datei:Zeile`, kein Feld `klasse`) und werden nicht umgeschrieben; sie sind Artefakte eines Laufs.

## Pflege (Steering-Loop)

Bei dreimaligem Auftreten desselben Findings:

- Ist die Kategorie noch richtig? → Klassifikation schärfen.
- Gibt es einen ADR oder `AGENTS.md`-Eintrag, der das verhindert hätte? → Folge-ADR oder `AGENTS.md`-Update.
- Gibt es eine Fitness Function, die das prüfen würde? → Gate hinzufügen (Modul 13).

Diese Skill-Datei wird **nicht** überschrieben, sondern versioniert (ADR-Hard-Rule, Modul 4).
