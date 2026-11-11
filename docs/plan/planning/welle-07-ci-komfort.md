# Welle welle-07-ci-komfort: Komfort für CI und Auswertung

**Status:** in-progress (gestartet; slice-026, -028, -029 done, slice-027 umgesetzt, Closure und Entscheidungsvorlage 1.1.0 offen)

**Zielmeilenstein:** M7 — CI-Komfort (Werkzeug-Version 1.1.0)

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-11.

---

## 1. Welle-Ziel

Das Werkzeug lässt sich in Pipelines komfortabler und strenger einsetzen: Warnungen können den Lauf scheitern lassen
(`--fail-on warn`), das Ergebnis kann in eine Datei geschrieben werden (`--output FILE`), sichtbare Bedrock-Profile
werden auch in weiteren Regionen gefunden, und die Dokumentations-Prüfziele von `d-check` werden bewertet und, wo sinnvoll,
in `make gates` aufgenommen. Das ist eine **Minor-Version 1.1.0** (neue Optionen, neues Verhalten auf Anforderung); das
Standardverhalten bleibt unverändert (Exit `1` nur bei `FAIL`).

## 2. Trigger (Welle startet)

- Release 1.0.0 veröffentlicht (erfüllt, M5).
- Die Pläne der Wellen 06 bis 08 und die Roadmap sind gemerged. Welle 06 muss nicht abgeschlossen sein; die Veröffentlichung von 1.1.0 nutzt aber den dort gehärteten Ablauf.

## 3. Closure-Trigger (Welle schließt)

- Alle Slices done; `make verify-clean` und CI grün.
- Standardverhalten unverändert (Regressionstests); die neuen Optionen sind dokumentiert (Referenz, Handbuch, CI-Doku, Regelkatalog, Doku-Drift-Test grün).
- Die Veröffentlichung 1.1.0 (Tag, Release, Images, Feature) erfolgte nur mit dem Ja beider Nutzer zu Version, Commit und Zielen; sonst gilt die Welle als „Release-Kandidat“ wie bei 1.0.0.
- Closure-Notiz in `done/welle-07-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle

| Slice | Titel | Zuständig | Bezug / Dependency |
|---|---|---|---|
| slice-026 | `--fail-on warn` (Exit-Code-Schwelle) | Kern w1 (inkl. Spec), Flag und Doku w2 | [`SPEC-030`](../../../spec/spezifikation.md#spec-030--exit-codes), [`LH-FA-EXIT-001`](../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung)/[`LH-FA-EXIT-002`](../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung) |
| slice-027 | `--output FILE` für alle Ausgabeformate | w2 | — |
| slice-028 | Zusätzliche Regionen für „sichtbare“ Bedrock-Profile | Kern w1, Adapter w2 | [`LH-FA-BEDROCK-002`](../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen) |
| slice-029 | Bewertung der `d-check`-Prüfziele (`doc-planning`, `doc-targets`, …) und gegebenenfalls Aufnahme in `make gates` | w2 | — |

> Die Slice-Dateien entstehen vor dem jeweiligen Start (Plan-vor-Code) unter `open/`.

## 5. Abhängigkeiten

- Blockiert: nichts.
- Wird blockiert von: Release 1.0.0 (erfüllt); Veröffentlichung von 1.1.0 profitiert von Welle 06.

## 5a. Entscheidungen (Entwurf, in den Slices zu bestätigen)

1. **`--fail-on warn`:** Schwelle als Wert (`fail` Default, `warn`); bei `warn` endet der Lauf mit Exit `1`, sobald mindestens ein `WARN` oder `FAIL` vorliegt. Der Default bleibt unverändert und erfüllt weiter [`LH-FA-EXIT-001`](../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung)/[`LH-FA-EXIT-002`](../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung); die Option ist eine Erweiterung, kein Widerspruch. Auswirkung auf die Zusammenfassungszeile im CI-Modus und auf SARIF/JSON (unverändert) prüfen. [`SPEC-030`](../../../spec/spezifikation.md#spec-030--exit-codes) wird fortgeschrieben.
2. **`--output FILE`:** Schreibt das gewählte Format in die Datei statt nach stdout; Fehlerfälle (Pfad nicht schreibbar) sind Tool-Fehler (Exit `2`); stdout bleibt dann leer, Meldungen gehen weiter auf stderr; keine Überschreibung außerhalb des angegebenen Pfads, Datei wird atomar ersetzt oder vollständig geschrieben.
3. **Zusatzregionen:** Konfigurierbare Liste weiterer Regionen, in denen nur die Sichtbarkeit nicht erlaubter Profile geprüft wird (WARN laut [`LH-FA-BEDROCK-002`](../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen)); mehr Aufrufe und mehr nötige IAM-Rechte werden dokumentiert (`aws-permissions.md`).
4. **`d-check`-Ziele:** Zuerst bewerten (hermetisch, netzlos, Befunde auf unserem Repository), dann entscheiden; „nicht aufnehmen“ ist ein zulässiges Ergebnis.

## 6. Out-of-Scope für diese Welle

- Neue Dateiarten (`iam-policy.json`), Signaturen, weitere Plattformen (Welle 8).
- Änderungen am Standard-Exit-Code-Verhalten.

## 7. Risiken

- **Exit-Code-Vertrag:** Konsumenten verlassen sich auf `0`/`1`/`2`; die neue Schwelle darf nur auf Anforderung greifen.
- **Mehr AWS-Aufrufe** bei Zusatzregionen: Rechte, Laufzeit und Fehlerbehandlung je Region.
- **`d-check`-Ziele** können Befunde auf bestehender Dokumentation erzeugen, die erst bereinigt werden müssen.

## 8. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-07-results.md. -->
