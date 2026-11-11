# Slice slice-029: Bewertung der d-check-Prüfziele für `make gates`

**Status:** open.

**Welle:** [`welle-07-ci-komfort`](../welle-07-ci-komfort.md).

**Bezug:** `AGENTS.md` §4 (Quality Gates), [`welle-05-results.md`](../done/welle-05-results.md) (Folgeaufgaben), PR #82.

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Die mit d-check v0.91.0 neu verfügbaren Ziele (`doc-planning`, `doc-targets`, `doc-structure`, `doc-tracked`, …) werden auf diesem Repository bewertet, und es wird entschieden, welche in `make gates` aufgenommen werden. „Nicht aufnehmen“ ist ein zulässiges Ergebnis.

## 2. Definition of Done

- [ ] Jedes Ziel einmal auf `main` ausgeführt; Befunde, Laufzeit und Netzbedarf je Ziel dokumentiert (hermetisch `--network none`, ohne Range).
- [ ] Entscheidung je Ziel (aufnehmen / nicht aufnehmen / später) mit Begründung in der Closure-Notiz.
- [ ] Aufgenommene Ziele laufen in `make gates` und `make verify-clean`; Befunde sind bereinigt, ohne Inhalte abzuschwächen.
- [ ] `AGENTS.md` und `harness/README.md` (Sensors-Tabelle) spiegeln die aufgenommenen Ziele.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `Makefile`, `AGENTS.md`, `harness/README.md` | update (je nach Entscheidung) | Gate-Aufnahme und Sensors-Tabelle |
| Planungsdateien | update bei Befund | Befunde der Ziele bereinigen |

## 4. Trigger

Welle 07 gestartet; unabhängig von den anderen Slices.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Ziele mit `.git`-Zugriff (`doc-tracked`) brauchen den Git-Stand im Mount und passen schlecht zu `verify-clean` im frischen Worktree; ggf. nicht aufnehmen.
- Ziele mit `RANGE` (`doc-immutable`, `doc-commits`) sind PR-spezifisch und kein Teil des lokalen Gates.
- Mehr Gates verlängern die Laufzeit der CI; Nutzen gegen Kosten abwägen.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
