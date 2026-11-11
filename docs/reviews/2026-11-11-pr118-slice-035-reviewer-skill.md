# Review-Report: PR #118 (slice-035, Reviewer-Skill) — 2026-11-11

**Review-Art:** Code — geprüft werden die neue Skill-Datei `.harness/skills/reviewer.md`, die Anpassung der Report-Vorlage und der Lifecycle-Teil des PR gegen den Slice-Plan, `AGENTS.md` §3 und die zitierten Dokumente. Es ist ein reiner Doku-Diff; es wurde nichts ausgeführt außer `make gates` (grün) und dem Lesen der Dateien.

**Gegenstand:** Pull Request #118 (`feat/slice-035-reviewer`, Autor w2), Commits `797262c` (Move `next/` → `in-progress/`) und `378a349` (Inhalt), Stand bei Beginn des Reviews.

**Skill:** [`.harness/skills/reviewer.md`](../../.harness/skills/reviewer.md) @ `378a349` (dieser PR; der Lauf prüft den Skill also mit dem Schema, das er selbst einführt — das ist eine Selbstanwendung und kein unabhängiger Test).
**Modell:** claude-sonnet-5-5 (Agent w1) · **Datum:** 2026-11-11

**Eingangs-Kontext:**

- Slice-Plan `slice-035-reviewer-skill` (`docs/plan/planning/in-progress/`)
- [`AGENTS.md`](../../AGENTS.md) §3.2 bis §3.9 (zitierte Anker), §2 (Source Precedence)
- [`docs/user/releasing.md`](../../docs/user/releasing.md) (Anker der Veröffentlichungs-Regel)
- [`docs/reviews/review-report.template.md`](../../docs/reviews/review-report.template.md)
- bestehende Reports unter `docs/reviews/`

---

## Findings

### F-1 — Datumsangabe zu den „älteren“ Reports ist falsch

- `kategorie`: LOW
- `quelle`: Maintainability
- `pfad`: `.harness/skills/reviewer.md` · „Reports vor dem 2026-11-11 (`pr90`, `pr96`, `pr106`, `pr107` und der Nachlauf)“
- `befund`: Alle genannten Reports tragen das Datum 2026-11-11 im Dateinamen; „vor dem 2026-11-11“ trifft auf keinen zu. Gemeint sind die Reports bis einschließlich des Nachlaufs zu PR #107.
- `verifizierbar`: ja — `ls docs/reviews/`.
- `klasse`: Zeitangabe widerspricht den Belegen

### F-2 — Anker für „Suppression oder Lockerung“ vermischt zwei Regeln

- `kategorie`: LOW
- `quelle`: Maintainability
- `pfad`: `.harness/skills/reviewer.md` · „Suppression oder Lockerung eines Gates ohne ADR“
- `befund`: `AGENTS.md` §3.2 verbietet Inline-Suppression und lässt Ausnahmen mit Begründung in einer zentralen Konfiguration zu; §3.6 verlangt eine ADR nur für das Senken von Schwellen. Die Zeile verlangt für beides „ohne ADR“ als HIGH-Kriterium, was §3.2 so nicht sagt.
- `verifizierbar`: nein — Lesen von `AGENTS.md` §3.2 und §3.6.
- `klasse`: Regelzitat fasst zwei Regeln ungenau zusammen

### F-3 — Veröffentlichungs-Anker stützt sich auf `releasing.md`, nicht auf eine Harte Regel

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `.harness/skills/reviewer.md` · „Repo-spezifisch #5 — Veröffentlichungsschritt“
- `befund`: Der Anker ist `docs/user/releasing.md` (Rang 6 der Source Precedence); in `AGENTS.md` §3 steht keine Regel zur Freigabe beider Nutzer. Das stimmt mit dem Dokument überein, ist aber schwächer verankert als die übrigen HIGH-Punkte.
- `verifizierbar`: nein — Lesen von `AGENTS.md` §3 und `releasing.md`.
- `klasse`: Anker schwächer als die Schwere

## Negativbefunde

- geprüft, ohne Befund: Zitate der Harten Regeln — §3.7 (Secret-Werte, `LH-NF-004`), §3.3 (`git mv` plus Inhalt in zwei Commits), §3.5 (Accepted-ADR unveränderlich), §3.8 (nur lesend, `LH-SEC-003`) und §3.9 (konfigurierbare Listen) stimmen im Wortlaut mit `AGENTS.md` überein; die Verweise auf `ADR-0003` und `ADR-0004` zeigen auf vorhandene Dateien.
- geprüft, ohne Befund: Schema — `kategorie`, `quelle`, `pfad`, `befund`, `verifizierbar`, `klasse` und die Negativbefund-Zeile entsprechen dem Worked Example des Kurses und der angepassten Report-Vorlage; „ohne Lösungsvorschlag“ und „kein HIGH/MEDIUM ohne Failure-Szenario“ stehen im Abschnitt „Was dieser Skill NICHT macht“.
- geprüft, ohne Befund: Report-Vorlage — `pfad` folgt dem Skill (Datei · Zitat), das Feld `klasse` ist ergänzt, der `d-check:ignore`-Hinweis ist entfernt, weil die Datei nun existiert; alte Reports bleiben unverändert und der Skill nennt das.
- geprüft, ohne Befund: Lifecycle — Commit `797262c` ist ein reiner Move (separat vom Inhaltscommit `378a349`), die Roadmap-Änderung entfernt den Ruhe-Marker im selben PR.
- geprüft, ohne Befund: Herkunft — der Kopf nennt Quelle, Commit `4b597d84af9c`, den abweichenden Kurs-Tag und die drei bewusst weggelassenen Anker mit Begründung; nichts davon wurde von mir gegen das Kurs-Repository geprüft.
- geprüft, ohne Befund: Gates — `make gates` im Branch grün; CI zum Zeitpunkt des Reports noch laufend.
- nicht geprüft: Inhalt der Kurs-Vorlage `4b597d84af9c` und des Worked Example in Modul 10 (kein Zugriff in diesem Lauf); ob die HIGH-Liste für künftige PRs vollständig ist (zum Beispiel eine Regel, dass das Lastenheft nur mit Zustimmung der Nutzer geändert wird — sie steht in `AGENTS.md` nicht als Harte Regel).

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 2 |
| INFO | 1 |

## Verdikt

**Merge-blockierend:** nein — kein HIGH, kein MEDIUM. F-1 und F-2 sind Textkorrekturen und können vor dem Merge im selben PR geschehen.

**Übergabe:** Findings gehen an den Autor. Der Report ersetzt keine Verifikation — DoD und Spezifikations-Konformität prüft der Verifier separat. Dieser Lauf wendet den Skill auf seine eigene Einführung an; ein späterer Lauf eines anderen Agenten ist der erste unabhängige Test des Skills.
