# ADR-0004: Abhängigkeiten (YAML, HCL) und Offline-Gates

**Status:** Proposed

**Datum:** 2026-11-10

**Autor:** Projektteam bedrock-eu-guard

**Bezug:** [`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit) (Offline-Fähigkeit), [`LH-QA-004`](../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build) (reproduzierbarer Build), [`LH-TECH-003`](../../../spec/lastenheft.md#lh-tech-003--konfigurationsdatei) (Konfigurationsdatei), [ADR-0002](0002-implementierungssprache.md)

**Schärft:** [`spec/architecture.md`](../../../spec/architecture.md) — Parser leben im jeweiligen driven-Adapter; diese ADR legt fest, wie deren Drittbibliotheken beschafft werden.

---

## Kontext

Bisher ist der Code Stdlib-only; die Gates (`make lint/test/build`,
`verify-clean`, CI) laufen in einem Go-Container mit `--network none`
und einem per Digest gepinnten Image. Ab der Konfigurationsdatei und dem Terraform-Scanner in
[welle-02](../planning/welle-02-config-scanner.md) braucht das Tool
Parser, die nicht in der Stdlib liegen ([ADR-0002](0002-implementierungssprache.md) nennt
`gopkg.in/yaml.v3` und `hashicorp/hcl`):

| Bedarf | Welle-2-Baustein | Bibliothek (Kandidat) |
|---|---|---|
| `bedrock-eu-check.yml`, Compose-Dateien | Konfiguration, Compose-Scanner | YAML (`gopkg.in/yaml.v3`) |
| Terraform-Dateien | Terraform-Scanner | HCL (`hashicorp/hcl/v2`, zieht eigene transitive Abhängigkeiten nach sich) |

Randbedingungen:

- Gates müssen netzlos und reproduzierbar laufen ([`LH-QA-004`](../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)); das ist heute durch `--network none` erzwungen und soll nicht aufgeweicht werden.
- Das gebaute Tool muss offline laufen ([`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)). Das ist unabhängig von der Beschaffung, ein Go-Binary linkt statisch.
- Jede Abhängigkeit ist Angriffsfläche und Lizenzfrage (z. B. ist HCL MPL-2.0, YAML Apache-2.0/MIT; Lizenzen vor Aufnahme prüfen).
- Das `Makefile` setzt heute `GOFLAGS=-mod=mod`; das muss sich bei Vendoring ändern.

## Entscheidung

Wir wählen **Vendoring**: Abhängigkeiten werden mit `go mod vendor` unter
`vendor/` eingecheckt, `go.sum` bleibt im Repo, Gates bauen mit
`-mod=vendor` und weiterhin mit `--network none`.
Neue Abhängigkeiten laufen über einen eigenen Schritt (`make deps`, mit
Netz) und einen PR, in dem das Diff von `vendor/` und `go.sum` sichtbar ist.

*(Status: Proposed — Sign-off durch den Auftraggeber ausstehend.)*

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| **A — Vendoring (`vendor/` im Repo)** | Gates bleiben netzlos und reproduzierbar; kein externer Zustand; CI braucht kein Netz für Module; Änderungen an Abhängigkeiten sind im PR-Diff prüfbar | größeres Repo und PR-Diffs; Updates erzeugen Rauschen; Lizenzdateien wandern ins Repo |
| B — Modul-Cache als Volume, vorab mit Netz gefüllt | kleines Repo | Cache ist externer Zustand; lokal und in CI getrennt zu pflegen; weniger reproduzierbar; zusätzlicher Make-Schritt vor jedem Gate |
| C — Netz im Gate erlauben (Proxy oder direkt) | einfachste Einrichtung | hebt `--network none` auf; Supply-Chain-Risiko beim Build; widerspricht der Reproduzierbarkeit |
| D — keine Drittbibliotheken, eigene Parser | keine Abhängigkeit | YAML und vor allem HCL sind aufwendig und fehleranfällig; hoher Wartungs- und Sicherheitsaufwand für Kernfunktionen |

## Konsequenzen

- Positiv: Gates und Release-Build bleiben netzlos; die Drittcode-Oberfläche ist im Repo sichtbar und reviewbar.
- Negativ: `vendor/` bläht das Repo auf (HCL mit transitiven Abhängigkeiten); Dependency-Updates sind laute PRs.
- Folgepflicht: `Makefile` auf `-mod=vendor` umstellen (`GOFLAGS`), `make deps` mit Netz ergänzen, CI-Check, dass `vendor/` zu `go.mod`/`go.sum` passt (`go mod verify` bzw. `go mod vendor` ohne Diff); Lizenzprüfung der aufgenommenen Pakete dokumentieren; `.gitignore` prüfen (kein `vendor/`-Eintrag).

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `go mod vendor` im Container | `vendor/` ist konsistent mit `go.mod`/`go.sum` (kein Diff) | `make deps-check` (neu, in `make gates`) |
| `--network none` | Gates bauen ohne Netz | `make gates` |

## Re-Evaluierungs-Trigger

- Wenn `vendor/` Review und Repo-Größe spürbar behindert.
- Wenn ein vertrauenswürdiger, signierter Modul-Mirror zur Verfügung steht.
- Wenn eine Abhängigkeit mit nicht passender Lizenz nötig würde.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-10 | Proposed | [`welle-02-config-scanner`](../planning/welle-02-config-scanner.md) §7 Risiken |
