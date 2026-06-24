# ADR-0001: Dokumentation als Source of Truth (Greenfield-Harness-Adoption)

**Status:** Accepted

**Datum:** 2026-06-24

**Autor:** Projektteam bedrock-eu-guard

**Bezug:** — (Prozess-ADR; bindet keine einzelne `LH-*`-Anforderung, sondern den Umgang mit allen)

**Schärft:** — (Prozess-ADR ohne Spec-Stratum)

---

## Kontext

`bedrock-eu-guard` startet als leeres Repo mit einem bereits
ausformulierten Lastenheft (Projektidee). Bevor Code entsteht, muss
festgelegt sein, *wer die Wahrheit trägt*, wenn Dokumentation und Code
auseinanderlaufen — sonst driftet die Spec gegen die Implementierung, und
ein Compliance-Prüfwerkzeug, dessen eigene Konfiguration nicht
nachvollziehbar ist, untergräbt seinen Zweck.

Annahme: Das Projekt wird überwiegend von KI-Agenten implementiert. Für
die gilt „was nicht im Kontext steht, existiert nicht" — die kanonischen
Quellen müssen explizit, rang-geordnet und maschinell prüfbar sein.

## Entscheidung

Wir adoptieren den **AI-Harness-Kurs (`pt9912/ai-harness-course`, v1.4.0)**
als Baseline und betreiben das Repo im **Greenfield-Modus: Dokumentation
führt, Code folgt.** Die Source Precedence (Lastenheft › Spezifikation ›
Architektur › ADRs › Roadmap › Operations › README › AGENTS ›
harness/README) ist in [`../../../harness/README.md`](../../../harness/README.md)
verankert; bei Konflikt gewinnt die höher rangierte Quelle.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — Code-first, Doku nachziehen | schneller erster Prototyp | erzeugt sofort Brownfield-Drift; für ein Compliance-Tool widersprüchlich |
| B — Ad-hoc-Doku ohne Harness | kein Setup-Aufwand | keine Source Precedence, keine Traceability, halluzinierte Gates wahrscheinlich |
| **C — AI-Harness-Kurs adoptieren (gewählt)** | etablierte Konventionen, Templates, echter Doku-Gate (`make doc-check`), Traceability-Constraint | initialer Bootstrap-Aufwand; Disziplin nötig |

## Konsequenzen

- Positiv: Jede Änderung hängt an einer `LH-*`/`ADR-*`-ID; die
  Artefaktkette ist auditierbar (passt zum Compliance-Charakter des Tools).
- Positiv: `make doc-check` greift ab Tag 1 als realer Sensor.
- Negativ: Mehr Schreibarbeit vor dem ersten Code; Agenten müssen die
  Source Precedence einhalten.
- Folgepflicht: ID-Schema und Modus pro Sub-Area in
  [`../../../harness/conventions.md`](../../../harness/conventions.md)
  gepflegt halten; Sprachwahl in **ADR-0002** entscheiden.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| d-check | Doku-Referenzen (Links/Anker) lösen auf | `make doc-check` |

## Re-Evaluierungs-Trigger

Wenn die Baseline-Version des Kurses wechselt (> v1.4.0) oder das Repo die
Repo-Klasse ändert (z. B. von Tooling/Referenz zu Policy/Compliance).

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-06-24 | Proposed | Harness-Bootstrap |
| 2026-06-24 | Accepted | Harness-Bootstrap (Adoptionsentscheidung mit dem Repo-Setup enacted) |
| 2026-06-24 | Redaktion | Gate-Target `docs-check` → `doc-check` (d-check `--print-mk`-Fragment, v0.29.0). Mechanische Korrektur im selben uncommitteten Bootstrap; die Entscheidung bleibt unverändert. |
