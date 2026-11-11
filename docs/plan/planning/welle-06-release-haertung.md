# Welle welle-06-release-haertung: Release-Prozess erproben und härten

**Status:** open (geplant, noch nicht gestartet)

**Zielmeilenstein:** M6 — Release-Prozess erprobt

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-11.

---

## 1. Welle-Ziel

Die Teile von Release 1.0.0, die bei der Veröffentlichung **nicht erprobt** waren
([`welle-05-results.md`](done/welle-05-results.md), Abschnitte „Herkunft der Release-Dateien“ und „Bekannter Mangel im veröffentlichten Feature“),
werden erprobt und korrigiert, und der Veröffentlichungsablauf wird nachvollziehbar und
wiederholbar (Workflows statt lokaler Handarbeit), ohne dass etwas automatisch veröffentlicht wird.

**Versionen:** Das **Werkzeug bleibt 1.0.0**, solange sich sein Code nicht ändert. Neu veröffentlicht wird
höchstens die **Feature-Version** (Vorschlag 1.0.1); die Feature-Version ist von der Tool-Version entkoppelt.
Ändert sich doch Code, entsteht eine Patch-Version des Werkzeugs, über die vorher die Nutzer entscheiden.

## 2. Trigger (Welle startet)

- Release 1.0.0 ist veröffentlicht (erfüllt am 2026-11-11, M5).
- Die Welle-Pläne 06 bis 08 und die Roadmap sind gemerged.

## 3. Closure-Trigger (Welle schließt)

- Alle Slices done; `make verify-clean` und CI grün.
- Der Workflow `release-artifacts` ist auf GitHub gelaufen; seine Binaries und Prüfsummen stimmen mit dem veröffentlichten Release überein (oder Abweichungen sind erklärt).
- Das Devcontainer-Feature wurde mit der echten `devcontainer`-CLI installiert und geprüft; die Mängel sind behoben.
- Der Workflow `release-create` existiert und wurde im Trockenlauf erprobt; Tag und Release legt er nur nach einer Eingabe mit Version und freigegebenem Commit an, nie automatisch.
- **Jede Veröffentlichung** (neue Feature-Version, Tag, Release) geschah nur nach dem ausdrücklichen Ja beider Nutzer mit Version, Commit und Zielen.
- Closure-Notiz in `done/welle-06-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle

Nummern bleiben, die Reihenfolge ist durch die Abhängigkeiten festgelegt: **023 → 025 → 024**.

| Slice | Titel | Zuständig | Dependency |
|---|---|---|---|
| slice-023 | Workflow `release-artifacts` auf GitHub erproben (per Tag-Ref) und mit dem Release vergleichen | w2 | — |
| slice-025 | Workflow `release-create` (nur `workflow_dispatch`; Eingaben Version und Commit; Trockenlauf; Tag und Release nur auf Anforderung) | w2 | slice-023 |
| slice-024 | Devcontainer-Feature mit der `devcontainer`-CLI erproben, Mängel beheben, neue Feature-Version | w2 | slice-025 (nutzt `publish`) |

> Die Slice-Dateien entstehen vor dem jeweiligen Start (Plan-vor-Code) unter `open/`.

## 5. Abhängigkeiten

- Blockiert: nichts zwingend; Welle 7 (1.1.0) kann parallel geplant werden, ihre Veröffentlichung nutzt den hier gehärteten Ablauf.
- Wird blockiert von: Release 1.0.0 (erfüllt).

## 5a. Architektur- und Prozess-Entscheidungen (Entwurf, in den Slices zu bestätigen)

1. **`release-artifacts` erproben:** Aufruf per `gh workflow run release.yml --ref v1.0.0` mit Version 1.0.0, damit der Commit dem des Releases entspricht und ein bitgleicher Vergleich möglich ist (auf dem Kopf von `main` wäre der Commit ein anderer).
2. **`release-create`:** nur `workflow_dispatch` mit den Eingaben Version und Commit; prüft wie `publish` Version, Commit und Tag; Schreibrechte (`contents: write`) nur im Job; ein Eingabeschalter für den Trockenlauf (Standard: Trockenlauf, baut und prüft nur). Ausdrücklich **nie** auf Push, Merge oder Zeitplan. Die neue Veröffentlichungsfähigkeit selbst braucht das Ja beider Nutzer.
3. **Feature:** Erprobung mit der `devcontainer`-CLI (voraussichtlich als CI-Job, weil die CLI lokal nicht vorliegt); die Option-Beschreibung („zeigt auf ein Release, das noch nicht existiert“) wird korrigiert; die Korrektur erscheint erst mit einer neuen Feature-Version (Neuveröffentlichung nötig, deshalb gebündelt mit allen weiteren Mängeln).
4. **Freigabe-Wortlaut unverändert:** „Ja zu Release X auf Commit `<SHA>` für die Ziele …“ von beiden Nutzern, getrennt. Zustand vor und nach jedem Schritt auf GitHub prüfen, nicht nur im Chat.

## 6. Out-of-Scope für diese Welle

- Neue Prüfregeln oder CLI-Optionen (Welle 7).
- Signierte Releases und SBOM, weitere Plattformen (Welle 8).
- Automatisches Veröffentlichen irgendeiner Art.

## 7. Risiken

- **Nicht rückholbar:** Eine neue Feature-Version, ein Tag und ein Release lassen sich praktisch nicht zurücknehmen; sie laufen nur mit dem Ja beider Nutzer.
- **Erprobung ohne Seiteneffekte:** `release-artifacts` und der Trockenlauf von `release-create` dürfen nichts veröffentlichen; das ist beim Entwurf der Workflows zu prüfen (Berechtigungen, keine Push-Schritte).
- **Werkzeug auf der Plattform:** Die `devcontainer`-CLI braucht Container-Laufzeit und Netz; eine Erprobung in den netzlosen Gates ist nicht möglich und bleibt ein eigener CI-Job.

## 8. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-06-results.md. -->
