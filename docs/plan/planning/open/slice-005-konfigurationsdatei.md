# Slice slice-005: Konfigurationsdatei `bedrock-eu-check.yml`

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-02-config-scanner.

**Bezug:** [`LH-TECH-003`](../../../../spec/lastenheft.md#lh-tech-003--konfigurationsdatei), [`LH-NF-007`](../../../../spec/lastenheft.md#lh-nf-007--konfigurierbarkeit), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md), [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Eine optionale Datei `bedrock-eu-check.yml` überschreibt die Defaults
(Regionen, Konten, Modell-Muster, Profile, `mantle`); fehlende Felder fallen
auf die Defaults zurück, eine ungültige eigene Konfiguration ist ein
Tool-Fehler (Exit `2`). Das `mantle`-Feld steuert [`LH-FA-LENV-005`](../../../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle).

## 2. Verträge (mit w1 abgestimmt)

- `model.Config` += `AllowedAccounts`, `AllowedProfiles`, `Mantle` (`warn|allow|block`, Default `warn`).
- `model.PartialConfig` (gelesener Inhalt): nil-Slice/leerer String = nicht angegeben (Default gilt); nicht-nil leerer Slice = explizit leer (z. B. `allowedAccounts: []`).
- `model.Resolve(PartialConfig) (Config, error)`: Defaults ergänzen, validieren (ungültige Regex, unbekannter Mantle-Wert, leere `allowedRegions`, Konto nicht 12 Ziffern) → `*model.ConfigError{Field, Index, Reason}`, nie mit Feldwert.
- Port `driven.ConfigLoader`: `Load() (*model.PartialConfig, error)`; Datei fehlt → `(nil, nil)` = Defaults; kaputt oder unbekannte Felder → `*model.FileError` (`FILE_INVALID`, Zeile).
- `app.LoadConfig(loader) (model.Config, error)`.
- Fehlerpolitik: Die **eigene** Konfiguration ungültig → Exit `2` ([`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes)); das unterscheidet sie von gescannten Dateien (dort Finding mit `FAIL`).
- Mantle ([`LH-FA-LENV-005`](../../../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle)): `warn` wie bisher; `allow` → bei aktivem Mantle `PASS` (`MANTLE_ALLOWED`) ohne WARN; `block` → `FAIL` (`MANTLE_BLOCKED`). Blockierte Modell-IDs bleiben in jedem Modus `FAIL` über LENV-006.

## 3. Definition of Done

- [ ] Kern (w1): Verträge aus §2, `lenv005.go` Mantle-Modi, Unit-Tests je Feld und je Fehlerfall (`coretest/fakeconfig`).
- [ ] `yaml.v3` gevendort ([`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md)): exakte Version, Lizenz dokumentiert; `make deps-verify` und `make gates` grün.
- [ ] Adapter `configyaml` (w2): YAML 1:1 auf `PartialConfig`; Abwesenheit (nil) und `[]` (leer) unterscheiden; unbekannte Felder und Syntaxfehler → `FileError` mit Zeile, ohne Dateiinhalt; Datei fehlt → `(nil, nil)`.
- [ ] Composition Root (w2): Suche `<root>/bedrock-eu-check.yml` (Default aktuelles Verzeichnis) oder `--config PATH`; `LoadConfig`; Fehler → stderr ohne Wert, Exit `2`.
- [ ] Testdaten im Repo (gültig, teilweise, ungültig); Akzeptanztests: `mantle: block` → FAIL, ungültige Regex → Exit 2.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/config.go`, `core/app/config.go` | neu | w1 |
| `internal/hexagon/port/driven/` (`ConfigLoader`), `coretest/fakeconfig/` | neu | w1 |
| `internal/hexagon/core/rules/lenv005.go` | update (Mantle-Modi) | w1 |
| `vendor/`, `go.mod`, `go.sum` | neu (`yaml.v3`) | w2 |
| `internal/adapter/driven/configyaml/` | neu | w2 |
| `internal/adapter/driving/cli/` | update (`--config`, `LoadConfig`, Exit 2) | w2 |

## 5. Trigger

[`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md) accepted und Vendoring-Infrastruktur gemergt (erfüllt); Kernvertrag von w1 gemergt, bevor der Adapter integriert wird.

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- Der `vendor/`-PR wird groß; er enthält ausschließlich `yaml.v3` samt Lizenzdatei und wird nicht von Hand editiert.
- `yaml.v3` akzeptiert standardmäßig unbekannte Felder; strikte Prüfung über `KnownFields(true)`.
- Secret-Regel ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)): Fehlermeldungen nennen nie Feldwerte.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
