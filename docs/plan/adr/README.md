# ADR-Index — bedrock-eu-guard

**Derivativ:** Quelle der Wahrheit sind die ADR-Dateien; dieser Index ist
eine Bequemlichkeits-Sicht — bei jedem neuen/akzeptierten ADR mitziehen.

| ID | Titel | Status | Bezug |
|---|---|---|---|
| [0001](0001-dokumentation-als-source-of-truth.md) | Dokumentation als Source of Truth (Greenfield-Harness-Adoption) | Accepted | — (Prozess) |
| [0002](0002-implementierungssprache.md) | Implementierungssprache (Go empfohlen) | Proposed | [`LH-TECH-001`](../../../spec/lastenheft.md) |
| [0003](0003-hexagonale-architektur.md) | Hexagonale Architektur (Ports & Adapters) | Accepted | [`LH-NF-005`](../../../spec/lastenheft.md)/[`LH-NF-006`](../../../spec/lastenheft.md), [`LH-QA-002`](../../../spec/lastenheft.md) |

## Konventionen

- ADRs sind nach `Accepted` **immutable**. Schärfungen entstehen als neue
  ADR mit `Supersedes ADR-NNNN`.
- Bei `Accepted`: diesen Index aktualisieren (Status, Datum).
- Jede ADR deklariert im `**Schärft:**`-Feld *aufwärts*, welche Spec-Stelle
  sie verbindlich macht. Prozess-ADRs ohne Spec-Stratum tragen `—`.
- Vorlage für neue ADRs: [`NNNN-titel.template.md`](NNNN-titel.template.md).
