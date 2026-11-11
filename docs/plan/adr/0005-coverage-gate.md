# ADR-0005: Coverage-Gate — Messweise, Ausschlüsse und Mindestwert

**Status:** Proposed

**Datum:** 2026-11-11

**Autor:** w1 (Projektteam bedrock-eu-guard)

**Bezug:** [`LH-QA-001`](../../../spec/lastenheft.md#lh-qa-001--unit-tests) (Unit-Tests), [`AGENTS.md`](../../../AGENTS.md) §3.6 (Gates nur per ADR lockern), [ADR-0004](0004-abhaengigkeiten-und-offline-gates.md) (Gates offline, Image per Digest gepinnt)

**Schärft:** — (Prozess-ADR ohne Spec-Stratum)

---

## Kontext

`AGENTS.md` §4 kündigt ein `coverage-gate` an; §3.6 verlangt, dass jede Schwelle (Coverage, Linter-Strenge, Architekturregel) eine ADR ist und kein PR-Kommentar. Bisher misst kein Gate die Abdeckung.

Gemessener Ist-Stand (2026-11-11, gepinntes Image, ohne Netz, [`slice-031`](../planning/in-progress/slice-031-coverage-gate.md) §2):

| Messweise | Gesamt |
|---|---|
| `go test -coverprofile ./...`, jedes Paket nur durch die eigenen Tests | 91,8 % |
| `-coverpkg` über die Produktionspakete (ohne Test-Doubles und `cmd/`) | 95,4 % |

Die einfache Messung unterschätzt den Kern: `internal/hexagon/core/model` steht dort bei 77,7 %, weil der Kern erst durch die Tests von `rules` und `app` voll genutzt wird; paketübergreifend gemessen sind es 95,1 %.

## Entscheidung

Wir wählen **paketübergreifende Messung mit festem Mindestwert und benannten Ausschlüssen**:

1. **Messweise:** `go test -coverpkg=<Produktionspakete> -coverprofile ./...`; der Gesamtwert ist der Anteil abgedeckter Anweisungen (`go tool cover -func`, Zeile `total`). Das Target gibt zusätzlich den Wert je Paket aus, damit ein Einbruch lokalisierbar ist.
2. **Ausschlüsse (jeder ist eine Lockerung und steht hier mit Grund):**
   - `internal/hexagon/coretest/*` — Test-Doubles, selbst Testcode; ihre Abdeckung sagt nichts über das Produkt.
   - `cmd/bedrock-eu-check` — enthält nur `main`; die Verdrahtung deckt der Test des Pakets `cli` ab.
   Jeder weitere Ausschluss braucht eine neue ADR.
3. **Mindestwert:** **90 %** gesamt. Er liegt rund fünf Punkte unter dem Ist-Stand: normale Änderungen scheitern nicht an der Schwelle, ein spürbarer Einbruch fällt auf. Es gibt keine zusätzliche Paket-Schwelle, weil kleine Pakete ein Prozentziel verzerren.
4. **Verbindliche Schwelle und wirksamer Wert:** Verbindlich ist der Mindestwert **90 %** dieser ADR. Der wirksame Wert steht im `Makefile` (`COVERAGE_MIN`, Anfangswert 90, nicht aus der Umgebung überschreibbar); `tools/coverage-gate.sh` führt die Untergrenze 90 als Konstante, lehnt ungültige Werte ab (Exit 2) und lässt keinen Wert unter 90 zu. Werte ab 90 setzt ein gewöhnlicher PR; ein Wert über 90 ist freiwilliger Abstand und keine eigene Schwelle, sein Senken bis auf 90 ist deshalb keine Lockerung im Sinn von `AGENTS.md` §3.6. **Senken unter 90 %** und jeder neue Ausschluss ist eine neue ADR mit `Supersedes`; die Konstante im Skript ändert sich nur zusammen mit ihr.
5. **Einordnung:** `make coverage-gate` läuft offline im gepinnten Go-Image und ist Teil von `make gates` (und damit der CI).

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — kein Gate | keine Zusatzkosten | Abdeckung kann unbemerkt sinken; §4 in `AGENTS.md` bleibt ein Versprechen |
| B — einfache Messung je Paket (`go test -cover`) | trivial | unterschätzt den Kern (77,7 %), erzwingt Tests an der falschen Stelle |
| C — Schwelle je Paket | lokalisiert Lücken | kleine Pakete verzerren; viele Werte zu pflegen |
| **D — paketübergreifend, ein Gesamtwert, Mindestwert 90 %** | misst, was der Kern wirklich nutzt; ein Wert; Paketwerte als Diagnose | Gesamtwert kann Lücken in einem Paket verdecken |

## Konsequenzen

- Positiv: Sinkende Abdeckung fällt in der CI auf; Schwelle, Ausschlüsse und Änderungsweg sind nachvollziehbar festgehalten.
- Negativ: Die Tests laufen im Gate ein zweites Mal (gemessen rund 3 s gegen rund 6 s); ein Gesamtwert kann Lücken in einzelnen Paketen verdecken. Die Untergrenze steht an zwei Stellen (ADR und Konstante im Skript); beide ändern sich nur gemeinsam.
- Folgepflicht: `AGENTS.md` §4 trägt das Target ein; `.tmp/coverage.out` bleibt außerhalb des Repositorys (`.tmp/` ist ignoriert).

## Fitness Function

| Tooling | Regel | Make-Target |
|---|---|---|
| `go test -coverpkg`, `tools/coverage-gate.sh` | Gesamtabdeckung der Produktionspakete mindestens `COVERAGE_MIN` | `make coverage-gate` |

## Re-Evaluierungs-Trigger

- Wenn die Abdeckung dauerhaft deutlich über dem Mindestwert liegt (Anheben prüfen).
- Wenn ein Paket Logik ohne Tests aufnimmt und der Gesamtwert das verdeckt (Paket-Schwelle prüfen).
- Wenn die Testlaufzeit des Gates spürbar steigt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-11 | Proposed | [`slice-031`](../planning/in-progress/slice-031-coverage-gate.md) |
