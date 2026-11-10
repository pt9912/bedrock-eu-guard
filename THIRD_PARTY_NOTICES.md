# Third-Party Notices

Abhängigkeiten liegen unter `vendor/` (siehe
[ADR-0004](docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md)) und werden
nie von Hand editiert. Lizenztexte stehen in der jeweiligen Lizenzdatei unter
`vendor/`.

| Modul | Version | Lizenz | Verwendung |
|---|---|---|---|
| `gopkg.in/yaml.v3` | v3.0.1 | MIT und Apache-2.0 (Details in `vendor/gopkg.in/yaml.v3/LICENSE` und `NOTICE`) | `bedrock-eu-check.yml` einlesen (Adapter `configyaml`) |
