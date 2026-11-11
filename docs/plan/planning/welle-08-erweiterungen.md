# Welle welle-08-erweiterungen: Erweiterungen und Härtung (nach Entscheidung der Auftraggeber)

**Status:** in-progress (teilweise: slice-031 freigegeben und in Arbeit; die übrigen Slices warten auf ihre Trigger)

**Zielmeilenstein:** M8 — Erweiterungen (Werkzeug-Version 1.2.0 oder später, je Slice)

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-11.

---

## 1. Welle-Ziel

Eine Sammelwelle für Vorhaben, die **eine Entscheidung der Auftraggeber, eine Lastenheft-Änderung oder einen ADR** brauchen
oder die sich lokal nicht abschließend erproben lassen. Kein Slice startet ohne seinen Trigger; die Welle schließt nicht
als Ganzes, sondern Slice für Slice (jeder Slice kann auch ersatzlos gestrichen werden).

## 2. Trigger (Welle startet)

Pro Slice (siehe Tabelle); Release 1.0.0 ist veröffentlicht (erfüllt).

## 3. Closure-Trigger

Pro Slice: Definition of Done des Slice-Plans erfüllt, `make verify-clean` und CI grün, Closure-Notiz. Die Welle ist
abgeschlossen, wenn alle Slices entweder done oder ausdrücklich gestrichen sind. Eine Veröffentlichung geschieht nur mit dem
Ja beider Nutzer zu Version, Commit und Zielen.

## 4. Slices in dieser Welle (Entwurf; keine Slice-Dateien vor dem Start)

| Slice | Titel | Zuständig | Trigger |
|---|---|---|---|
| slice-030 | Scanner für eigenständige IAM-Policy-Dateien (`iam-policy.json`) | beide | Entscheidung der Auftraggeber **und** neue Anforderung im Lastenheft (Soll-Anforderung, siehe [`LH-QA-003`](../../../spec/lastenheft.md#lh-qa-003--testdaten)) |
| slice-031 | `coverage-gate` mit begründeter Schwelle | w1 | ADR (jede Schwelle ist ein ADR, nicht ein PR-Kommentar; `AGENTS.md` §3.6) |
| slice-032 | Signierte Releases (Sigstore) und SBOM | w2 | Entscheidung der Auftraggeber zu Signatur- und Vertrauensmodell; Welle 06 (`release-create`) |
| slice-033 | Binaries für macOS und Windows | w2 | Entscheidung der Auftraggeber; **lokal nicht erprobbar** ohne Runner (Cross-Build ja, Test eingeschränkt) |

## 5. Abhängigkeiten

- Blockiert: nichts.
- Wird blockiert von: Entscheidungen der Auftraggeber bzw. ADRs je Slice; slice-032 baut auf dem gehärteten Ablauf aus Welle 06 auf.

## 6. Out-of-Scope für diese Welle

- Alles, was Welle 06 und Welle 07 liefern.
- Neue Prüfregeln außerhalb von slice-030.

## 7. Risiken

- **Lastenheft ist bindend:** slice-030 verändert den Vertragsumfang und braucht den Wortlaut der Auftraggeber.
- **Nicht erprobbar:** slice-033 lässt sich ohne Runner nur eingeschränkt prüfen und ist in den Release-Notizen als solcher zu kennzeichnen.
- **Vertrauensmodell:** Signaturen und SBOM verändern, was Nutzer prüfen können; sie ersetzen nicht die Freigabe-Pflicht.

## 8. Closure-Notiz

<!-- Erst nach Abschluss der Slices füllen. Verweis auf welle-08-results.md. -->
