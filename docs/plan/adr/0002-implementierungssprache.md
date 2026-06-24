# ADR-0002: Implementierungssprache

**Status:** Proposed

**Datum:** 2026-06-24

**Autor:** Projektteam bedrock-eu-guard

**Bezug:** [`LH-TECH-001`](../../../spec/lastenheft.md#lh-tech-001--programmiersprache) (Sprache Python/Go), [`LH-NF-002`](../../../spec/lastenheft.md#lh-nf-002--containerfähigkeit) (Containerfähigkeit), [`LH-QA-004`](../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build) (reproduzierbarer Build)

**Schärft:** — (Toolchain-/Distributions-Entscheidung; die Schichtung selbst trägt ADR-0003)

---

## Kontext

`bedrock-eu-check` ist ein **CLI-Prüfwerkzeug** für lokale Nutzung, CI und
Devcontainer (`LH-MOD-*`, [`LH-TECH-005`](../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)), das als **Docker-Image**
([`LH-TECH-004`](../../../spec/lastenheft.md#lh-tech-004--docker-image)) und idealerweise als **einzelnes Binary** verteilt
werden soll ([`LH-TECH-001`](../../../spec/lastenheft.md#lh-tech-001--programmiersprache)). Das Lastenheft lässt **Python oder Go**
offen und empfiehlt: *Python für schnelle Entwicklung, Go für einfache
Verteilung als einzelnes Binary.*

Die in **ADR-0003** gewählte Architektur orientiert sich an **d-check**
(Go) und nutzt dessen `internal/`-Ordnerkonvention (driving/driven,
Paket-Sichtbarkeit). Das verschiebt die Abwägung Richtung Go.

## Entscheidung

Wir wählen **Go** *(Status: Proposed — Sign-off durch den Auftraggeber
ausstehend)*.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| **A — Go (vorgeschlagen)** | ein statisches Binary ([`LH-TECH-001`](../../../spec/lastenheft.md#lh-tech-001--programmiersprache)); reproduzierbarer, schlanker Docker-Build ([`LH-QA-004`](../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)); `internal/`-Paketkonvention deckt ADR-0003 nativ; starke Stdlib für Dateiscan; identischer Stack wie das Vorbild d-check | etwas mehr Code für Parser; HCL-Parsing über Drittbibliothek |
| B — Python | sehr schnelle Entwicklung; reiches Ökosystem (boto3, python-hcl2); flexible Scanner | Verteilung als Binary aufwändig (PyInstaller/uv); kein natives `internal/`; Laufzeit-/Dependency-Footprint im Image größer |
| C — Rust | schnellstes Binary, starke Typen | steilste Lernkurve, langsamste Entwicklung; für ein Config-Scan-Tool überdimensioniert |

## Konsequenzen

- Positiv: Eine Toolchain (Go) für Build, Test, Lint, Single-Binary und
  Docker; deckt [`LH-TECH-001`](../../../spec/lastenheft.md#lh-tech-001--programmiersprache)/[`LH-TECH-004`](../../../spec/lastenheft.md#lh-tech-004--docker-image) und [`LH-NF-002`](../../../spec/lastenheft.md#lh-nf-002--containerfähigkeit) direkt.
- Positiv: ADR-0003-Ordnerkonvention und die `coretest`-Doubles
  ([`LH-QA-002`](../../../spec/lastenheft.md#lh-qa-002--integration-tests)) sind Go-idiomatisch.
- Negativ: HCL-/YAML-Parsing braucht ausgewählte Bibliotheken (z. B.
  `hashicorp/hcl`, `gopkg.in/yaml.v3`) statt Stdlib-only.
- Folgepflicht nach *Accepted*: Go-Toolchain im `Makefile` verdrahten
  (`make lint`/`test`/`build`), Suppression-Marker für AGENTS.md §3.2
  festlegen (`//nolint`), Devcontainer/Docker-Build aufsetzen.

## Re-Evaluierungs-Trigger

Wenn ein Pflicht-Feature nur mit einer Python-spezifischen Bibliothek
sinnvoll umsetzbar ist, oder wenn die Verteilung als Binary doch nicht
gefordert wird.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-06-24 | Proposed | Bootstrap; Empfehlung Go nach Wahl der Hexagonal-/d-check-Architektur (ADR-0003) |
