# ADR-Index — bedrock-eu-guard

**Derivativ:** Quelle der Wahrheit sind die ADR-Dateien; dieser Index ist
eine Bequemlichkeits-Sicht — bei jedem neuen/akzeptierten ADR mitziehen.

| ID | Titel | Status | Bezug |
|---|---|---|---|
| [0001](0001-dokumentation-als-source-of-truth.md) | Dokumentation als Source of Truth (Greenfield-Harness-Adoption) | Accepted | — (Prozess) |
| [0002](0002-implementierungssprache.md) | Implementierungssprache (Go empfohlen) | Accepted | [`LH-TECH-001`](../../../spec/lastenheft.md#lh-tech-001--programmiersprache) |
| [0003](0003-hexagonale-architektur.md) | Hexagonale Architektur (Ports & Adapters) | Accepted | [`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)/[`LH-NF-006`](../../../spec/lastenheft.md#lh-nf-006--erweiterbarkeit), [`LH-QA-002`](../../../spec/lastenheft.md#lh-qa-002--integration-tests) |
| [0004](0004-abhaengigkeiten-und-offline-gates.md) | Abhängigkeiten (YAML, HCL) und Offline-Gates | Accepted | [`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit), [`LH-QA-004`](../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build) |
| [0005](0005-coverage-gate.md) | Coverage-Gate — Messweise, Ausschlüsse und Mindestwert | Accepted | [`LH-QA-001`](../../../spec/lastenheft.md#lh-qa-001--unit-tests) |

## Konventionen

- ADRs sind nach `Accepted` **immutable**. Schärfungen entstehen als neue
  ADR mit `Supersedes ADR-NNNN`.
- Bei `Accepted`: diesen Index aktualisieren (Status, Datum).
- Jede ADR deklariert im `**Schärft:**`-Feld *aufwärts*, welche Spec-Stelle
  sie verbindlich macht. Prozess-ADRs ohne Spec-Stratum tragen `—`.
- Vorlage für neue ADRs: [`NNNN-titel.template.md`](NNNN-titel.template.md).
