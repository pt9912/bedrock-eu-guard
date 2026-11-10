# Welle welle-04-output-ci: Ergebnisse

**Abgeschlossen:** 2026-11-10. **Meilenstein:** M4 — CI-fähig.

## Closure-Trigger

- Alle Slices done: slice-011 bis slice-013.
- `make verify-clean` und CI grün auf `main`.
- `--format json` und `--format sarif` erzeugen deterministische Ausgaben; Tests prüfen Pflichtfelder und Struktur offline.
- `--ci` endet mit Exit `0` ohne FAIL, `1` mit FAIL und `2` bei Tool-Fehler; stdout enthält nur das Ergebnis.
- Kein Secret-Wert in keinem Format ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)), per Test geprüft.
- Beispiel-Workflow in `docs/user/ci.md` dokumentiert.

## Lieferumfang

`--format console|json|sarif` und `--ci` in allen Modi; Reporter als Adapter hinter dem unveränderten Port (`report.Console`, `report.JSON`, `report.SARIF`); `Severity.Level()`/`SARIFLevel()` als einzige Kernänderung; [`SPEC-021`](../../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json) fortgeschrieben.

## Steering-Loop

1. Neue Ausgabeformate sind Adapter; der Port und der Kern bleiben unverändert.
2. Spec und Code im selben PR ändern, wenn das Schema wächst; Versionskonstante im Test fixieren.
3. Ausgaben ohne offiziell verfügbares Schema strukturell testen und die Grenzen offen benennen.
4. Beispiele für externe Systeme (GitHub) nur als „nicht erprobt“ kennzeichnen.

## Folgeaufgaben

- `--fail-on warn`, `--output FILE`, Baseline-Dateien.
- SARIF-Schema offline prüfen, wenn es als Testdatum beschafft wird; Verhalten der Ersatzposition in GitHub Code Scanning erproben.
