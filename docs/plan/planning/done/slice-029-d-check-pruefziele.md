# Slice slice-029: Bewertung der d-check-Prüfziele für `make gates`

**Status:** done.

**Welle:** [`welle-07-ci-komfort`](welle-07-ci-komfort.md).

**Bezug:** `AGENTS.md` §4 (Quality Gates), [`welle-05-results.md`](../done/welle-05-results.md) (Folgeaufgaben), PR #82.

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Die mit d-check v0.91.0 neu verfügbaren Ziele (`doc-planning`, `doc-targets`, `doc-structure`, `doc-tracked`, …) werden auf diesem Repository bewertet, und es wird entschieden, welche in `make gates` aufgenommen werden. „Nicht aufnehmen“ ist ein zulässiges Ergebnis.

## 2. Definition of Done

- [x] Jedes Ziel einmal auf `main` ausgeführt; Befunde, Laufzeit und Netzbedarf je Ziel dokumentiert (hermetisch `--network none`, ohne Range).
- [x] Entscheidung je Ziel (aufnehmen / nicht aufnehmen / später) mit Begründung in der Closure-Notiz.
- [x] Aufgenommene Ziele laufen in `make gates` und `make verify-clean`; Befunde sind bereinigt, ohne Inhalte abzuschwächen.
- [x] `AGENTS.md` und `harness/README.md` (Sensors-Tabelle) spiegeln die aufgenommenen Ziele.
- [x] `make verify-clean` und CI grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

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

**Abgeschlossen:** 2026-11-11.

**Bewertung** (d-check v0.91.0, jedes Ziel auf `main` ausgeführt; alle laufen hermetisch mit `--network none`, je Lauf unter einer Sekunde):

| Ziel | Ergebnis | Entscheidung |
|---|---|---|
| `doc-targets` | Ohne Konfiguration inert. Mit Konfiguration (Makefiles, Tabellen in `AGENTS.md` und `harness/README.md`, Autorität `AGENTS.md`) 14 Befunde `gate-undocumented`: `help` und die von `d-check --print-mk` erzeugten `doc-*`-Regeln. Umgekehrte Richtung (dokumentiert, aber nicht definiert) 0 Befunde; die Negativprobe mit einem erfundenen Target schlägt mit `gate-phantom` an. | **aufgenommen** in `make gates` und damit `verify-clean`; die erzeugten Regeln und `help` stehen unter `exempt-targets` in `.d-check.yml` |
| `doc-planning` | Ohne Konfiguration inert. Mit Konfiguration (Roadmap, Closure-Verzeichnis `done/`) 1 echter Befund: `roadmap.md` führt bei „kein Slice in `in-progress/`“ den Ruhe-Marker „Keine aktive Welle“ nicht; zugleich sind Welle 06 und 07 gestartet, die Slice-Dateien wandern aber von `open/` direkt nach `done/`. Die Closure-Notizen aller abgeschlossenen Slices bestehen die Strukturprüfung. | **noch nicht aufgenommen, Entscheidung der Nutzer offen:** Das Gate setzt den Slice-Lifecycle `open → next → in-progress → done` aus `docs/plan/planning/README.md` voraus; unsere Slices gehen bisher direkt von `open/` nach `done/`. Aufnehmen heißt, den Lifecycle einzuhalten; sonst wäre das Gate dauerhaft rot |
| `doc-structure` | Ohne Konfiguration inert. | nicht aufgenommen (kein Konfigurationsbedarf erkennbar) |
| `doc-tracked` | Braucht `.git` im Mount; im frischen Worktree von `verify-clean` ist `.git` eine Datei. | nicht aufgenommen |
| `doc-immutable`, `doc-commits` | Brauchen eine `RANGE` (PR-Bezug). | nicht Teil von `make gates`; mögliche CI-Schritte für Pull Requests |
| `doc-usage`, `doc-manual` | Hilfsziele ohne Prüfung. | nicht aufgenommen (Hilfsziele) |

- **Funktioniert:** `doc-targets` hält Doku und Makefile gleich; die Tabellen in `AGENTS.md` und `harness/README.md` nennen das neue Ziel selbst.
- **Anders als geplant:** Die Ziele sind ohne Konfiguration inert; „0 Befunde“ allein sagt deshalb nichts. Erst mit einem Probelauf auf einem temporären Profil (`--config`) zeigte sich, welche etwas prüfen.
- **Offen:** Die Nutzer entscheiden, ob der Lifecycle aus `docs/plan/planning/README.md` künftig eingehalten wird (Regelwerk) und `planning` damit als Gate aufgenommen werden kann (Roadmap mit Ruhe-Marker bzw. Slices unter `in-progress/` während einer Welle, Closure-Prüfung inklusive), oder ob der Ablauf `open → done` bewusst bleibt und die Konvention angepasst wird. `doc-immutable` und `doc-commits` als PR-Schritte in der CI sind ein weiterer Kandidat.
- **Steering-Loop:** Ein neues Gate erst mit einer Negativprobe belegen (hier ein erfundenes Target), bevor man dem grünen Lauf traut. Opt-in-Module ohne Konfiguration melden „0 Befunde“, ohne etwas geprüft zu haben.

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
