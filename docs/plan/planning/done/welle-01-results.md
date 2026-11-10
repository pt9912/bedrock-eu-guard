# Welle welle-01-lokaler-env-check: Ergebnisse

**Abgeschlossen:** 2026-11-10. **Meilenstein:** M1 — erstes lauffähiges Tool.

## Closure-Trigger

- Alle Slices done: slice-001, slice-002, slice-003.
- `make gates` (`doc-check`, `lint`, `test`) grün auf `main`, im sauberen Checkout geprüft.
- `bedrock-eu-check local` erfüllt [`LH-ABN-001`](../../../../spec/lastenheft.md#lh-abn-001--kritische-lokale-fehlkonfiguration) (≥ 2 FAIL, Exit 1, kein Secret-Wert) und [`LH-ABN-002`](../../../../spec/lastenheft.md#lh-abn-002--eu-konforme-minimalumgebung) (Exit 0) als Test in `internal/adapter/driving/cli`.

## Lieferumfang

Hexagon-Kern, Regeln LENV-001 bis -006, Konsolen-Reporter, CLI mit Exit-Codes, `make lint/test/build`, containerisierte Go-Toolchain (Image per Digest).

## Steering-Loop

1. Gates vor dem Handoff im sauberen Checkout oder in CI, nicht im Arbeitsbaum. Auslöser war ein roter `main` (PR #4, fehlende Dateien, behoben in #6).
2. Parallelarbeit funktioniert mit einem früh festgelegten Schnittstellenvertrag und getrennten Verzeichnissen je Agent.
3. Spezifikationswiderspruch bei LENV-002 (PASS vs. WARN) klären und in der Spec präzisieren.

## Folgeaufgaben

- `make verify-clean` oder CI-Workflow für `make gates`.
- Spec-Präzisierung LENV-002.
- Region-Extraktion aus Modell-/Profil-IDs gegen `allowedRegions` ([`LH-FA-LENV-006`](../../../../spec/lastenheft.md#lh-fa-lenv-006--prüfung-von-claude-default-modellvariablen), derzeit nur Default-Blockmuster).
- Welle 2: Config-Datei und Adapter, `mantle: warn|allow|block`, Fehlerbehandlung ungültiger Regex-Muster.
