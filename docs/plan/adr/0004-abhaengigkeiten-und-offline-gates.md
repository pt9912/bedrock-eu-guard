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
- Das `Makefile` setzt heute `GOFLAGS=-mod=mod`; das muss sich bei Vendoring ändern. `go mod vendor` selbst braucht Modul-Cache oder Netz.

## Entscheidung

Wir wählen **Vendoring**: Abhängigkeiten werden mit `go mod vendor` unter
`vendor/` eingecheckt, `go.sum` bleibt im Repo, Gates bauen mit
`-mod=vendor` und weiterhin mit `--network none`.
Neue oder aktualisierte Abhängigkeiten laufen über einen eigenen Schritt
mit Netz (`make deps`) und einen PR, in dem das Diff von `vendor/` und
`go.sum` sichtbar ist. `vendor/` wird **nie von Hand editiert**.

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
- Folgepflicht:
  1. `Makefile`: `GOFLAGS=-mod=vendor` statt `-mod=mod`; `make deps` (mit Netz) für `go mod tidy` und `go mod vendor`.
  2. `make lint` darf `vendor/` nicht scannen (`gofmt -l .` würde anschlagen); `go vet ./...` übergeht `vendor/` von selbst.
  3. d-check (`doc-check`) schließt `vendor/` aus (`.d-check.yml`), sonst erzeugen Markdown-Dateien der Bibliotheken Befunde.
  4. `.gitattributes`: `vendor/** linguist-vendored`, damit GitHub die Diffs einklappt.
  5. Versionen festlegen: exakte Versionen von `gopkg.in/yaml.v3` und `hashicorp/hcl/v2` samt transitiven Abhängigkeiten stehen beim Aufnehmen in `go.mod`/`go.sum` (Go-Mindestversion laut `go.mod`); Aufnahme nur mit dokumentierter Lizenzprüfung (optional `THIRD_PARTY_NOTICES`). HCL ist MPL-2.0 (dateibasiertes Copyleft): Vendoring unverändert ist zulässig, Änderungen an `vendor/` sind verboten.
  6. Integrität gegen Upstream (`go mod verify`, erneutes `go mod vendor` ohne Diff) braucht Netz und ist deshalb **nicht Teil der netzlosen Gates**, sondern ein eigener Schritt `make deps-verify` bzw. CI-Job, der nur läuft, wenn ein PR `vendor/`, `go.mod` oder `go.sum` ändert.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `go build -mod=vendor` | Build bricht ab, wenn `vendor/modules.txt` nicht zu `go.mod` passt (Konsistenz, netzlos) | `make build`, `make gates` |
| `--network none` | Gates bauen und testen ohne Netz | `make gates` |
| `go mod verify` und `go mod vendor` ohne Diff | `vendor/` entspricht Upstream (braucht Netz) | `make deps-verify` (nur bei Änderungen an `vendor/`, `go.mod`, `go.sum`) |

## Re-Evaluierungs-Trigger

- Wenn `vendor/` Review und Repo-Größe spürbar behindert.
- Wenn ein vertrauenswürdiger, signierter Modul-Mirror zur Verfügung steht.
- Wenn eine Abhängigkeit mit nicht passender Lizenz nötig würde.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-10 | Proposed | [`welle-02-config-scanner`](../planning/welle-02-config-scanner.md) §7 Risiken |
