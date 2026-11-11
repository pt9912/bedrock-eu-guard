# Review-Report: PR #107 (slice-024, Code) — Nachlauf — 2026-11-11

**Review-Art:** Code — Folgelauf zu [`2026-11-11-pr107-slice-024-feature-cli.md`](2026-11-11-pr107-slice-024-feature-cli.md). Geprüft wird der Commit `9628e9f` (`publish` mit Trockenlauf, `ref` nur für das Feature) gegen die Findings F-2 und F-3 des ersten Laufs.

**Gegenstand:** Pull Request #107, Commit `9628e9f` (Diff gegen `fe8ea99`: `.github/workflows/publish.yml`, `docs/user/releasing.md`).

**Skill:** keiner (im Repository existiert keine Reviewer-Skill-Datei).
**Modell:** claude-sonnet-5-5 (Agent w1) · **Datum:** 2026-11-11

**Eingangs-Kontext:**

- erster Report zu PR #107 (F-1 bis F-5)
- Lauf 42987262764 auf GitHub (`publish`, `workflow_dispatch`, Commit `9628e9f`, `dry_run=true`, Ergebnis success)
- Slice-Plan `slice-024`, Veröffentlichungsgrenze (Ja beider Nutzer)
- [`AGENTS.md`](../../AGENTS.md) (Hard Rules)

---

## Findings

### F-1 — Der Trockenlauf-Beleg prüft den neuen `ref`-Pfad nicht

- `kategorie`: LOW
- `quelle`: Plan slice-024 (Veröffentlichungsweg)
- `pfad`: `.github/workflows/publish.yml` (Job `feature`); Lauf 42987262764
- `befund`: Der Lauf verwendete `version` 1.0.0 ohne `ref`, checkte also den Tag `v1.0.0` mit der alten `devcontainer-feature.json` aus und packte diesen Stand; der Zweig mit `ref=feature-v<version>` und die Feature-Version 1.0.1 wurden nicht ausgeführt. Der Autor nennt diese Grenze selbst (Tag entsteht erst über `release-create` mit `target=feature`).
- `verifizierbar`: ja — ein Trockenlauf mit existierendem Tag `feature-v1.0.1` würde es bestätigen.

## Negativbefunde

- geprüft, ohne Befund: Lauf 42987262764 — Ergebnis success, Event `workflow_dispatch`, Commit `9628e9f`; übersprungen waren genau „Bei ghcr.io anmelden“, „Nach ghcr.io pushen“ (Job `images`) und „Feature veröffentlichen“ (Job `feature`); `git ls-remote --tags` zeigt weiter nur `v1.0.0`.
- geprüft, ohne Befund: Der Standard ist `dry_run: true`; alle schreibenden Schritte (Anmeldung, Push, Veröffentlichung) tragen `if: ${{ !inputs.dry_run }}`, der Trockenlauf-Schritt `if: ${{ inputs.dry_run }}`.
- geprüft, ohne Befund: Die Prüfung der OCI-Labels läuft jetzt auch im Trockenlauf (vor dem Push), ihre Logik ist unverändert.
- geprüft, ohne Befund: Der frühere F-3 ist behoben — `ref` zusammen mit dem Job `images` bricht mit der Meldung „ref gilt nur für das Feature“ ab.
- geprüft, ohne Befund: Rechte unverändert (`permissions: {}` oben, `packages: write` nur in den Jobs); keine neuen Eingaben fließen per `${{ }}` in `run:`.
- geprüft, ohne Befund: `releasing.md` nennt `dry_run` (Standard) und `ref`. Beobachtung ohne Kategorie: Der Satz „Er baut aus dem Tag `v<version>` …“ steht jetzt hinter der `ref`-Erläuterung und liest sich dadurch holprig.
- geprüft, ohne Befund: CI des PR auf `9628e9f` — `gates`, `image`, Fall A grün (Fall B übersprungen, nur per Dispatch).
- nicht geprüft: Lauf mit `dry_run=false` (bewusst nie ohne Ja beider Nutzer); Verhalten mit existierendem Tag `feature-v1.0.1`.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 1 |
| INFO | 0 |

## Verdikt

**Merge-blockierend:** nein — kein HIGH, kein MEDIUM. F-2 und F-3 des ersten Laufs sind behoben (F-3) bzw. durch den Trockenlauf mit dokumentierter Restgrenze abgedeckt (F-2, hier als LOW F-1 geführt). F-1, F-4 und F-5 des ersten Laufs bleiben als bekannte Grenzen für die Closure und die Entscheidungsvorlage bestehen.

**Übergabe:** Findings gehen an den Autor. Der Report ersetzt keine Verifikation — DoD und Spezifikations-Konformität prüft der Verifier separat.
