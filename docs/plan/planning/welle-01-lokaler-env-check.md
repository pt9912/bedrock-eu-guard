# Welle welle-01-lokaler-env-check: Lokaler Env- und Region-Check

**Status:** open

**Zielmeilenstein:** M1 — erstes lauffähiges Tool

**Verantwortlich:** Projektteam bedrock-eu-guard. **Datum:** 2026-06-24.

---

## 1. Welle-Ziel

Eine erste lauffähige Version: `bedrock-eu-check local` prüft die lokalen
Umgebungsvariablen und Modell-IDs, liefert `PASS`/`WARN`/`FAIL` auf der
Konsole und setzt die Exit-Codes `0`/`1`/`2`. Damit ist das Hexagon-Gerüst
(Kern, Ports, CLI-Composition-Root, ein driven-Adapter je Quelle) etabliert
und die `MUSS`-Anforderungen `LH-PRI-MUSS-001..005` sind im Kern
erfüllt (`RM-001` / v0.1).

## 2. Trigger (Welle startet)

- ADR-0002 (Implementierungssprache) **accepted** — Sprachwahl bestätigt.
- ADR-0003 (Hexagonale Architektur) accepted ✅.

## 3. Closure-Trigger (Welle schließt)

- Alle Slices der Welle done.
- `make gates` grün — inkl. der neu entstehenden `make lint`/`make test`.
- `bedrock-eu-check local` erfüllt die Abnahmekriterien `LH-ABN-001`
  (kritische Fehlkonfiguration → ≥ 2 FAIL, Exit 1, kein Secret-Wert) und
  `LH-ABN-002` (EU-Minimalumgebung → PASS, Exit 0).
- Closure-Notiz in `done/welle-01-results.md` mit Steering-Loop-Eintrag.

## 4. Slices in dieser Welle

| Slice | Titel | Status | Bezug |
|---|---|---|---|
| slice-001 | Hexagon-Gerüst + Env-/Region-Check + Exit-Codes + Konsolen-Reporter | open | `LH-FA-LENV-001/002`, `EXIT-*`, `REP-001` |
| slice-002 | Modell-ID-Klassifikation (allow/block) | geplant | `LH-FA-LENV-006`, `SPEC-010` |
| slice-003 | Anthropic-Key- und Bedrock-Flag-Regeln | geplant | `LH-FA-LENV-003/004`, `PRI-MUSS-002` |

> Nur slice-001 ist als Datei in `open/` angelegt; slice-002/003 entstehen
> aus dem Closure von slice-001 (Plan-vor-Code, Modul 5).

## 5. Abhängigkeiten

- Blockiert: welle-02 (Config-Scanner bauen auf Kern + Reporter auf).
- Wird blockiert von: ADR-0002-Sign-off.

## 6. Out-of-Scope für diese Welle

- Jeder AWS-Live-Call (Welle 3) — `local` bleibt strikt offline.
- JSON-/SARIF-Ausgabe (Welle 4).
- Terraform-/Container-Scanner (Welle 2).

## 7. Closure-Notiz

<!-- Erst nach Welle-Abschluss füllen. Verweis auf welle-01-results.md. -->
