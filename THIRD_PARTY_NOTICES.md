# Third-Party Notices

Abhängigkeiten liegen unter `vendor/` (siehe
[ADR-0004](docs/plan/adr/0004-abhaengigkeiten-und-offline-gates.md)) und werden
nie von Hand editiert. Lizenztexte stehen in der jeweiligen Lizenzdatei unter
`vendor/`.

| Modul | Version | Lizenz | Verwendung |
|---|---|---|---|
| `gopkg.in/yaml.v3` | v3.0.1 | MIT und Apache-2.0 (Details in `vendor/gopkg.in/yaml.v3/LICENSE` und `NOTICE`) | `bedrock-eu-check.yml` und Compose-Dateien einlesen (Adapter `configyaml`, `envsource`) |
| `github.com/hashicorp/hcl/v2` | v2.25.0 | MPL-2.0 (dateibasiertes Copyleft: unverändert eingebunden, Änderungen an `vendor/` sind verboten) | Terraform-Dateien parsen (Adapter `terraform`) |
| `github.com/zclconf/go-cty` | v1.20.0 | MIT | Werttypen von HCL, JSON-Umwandlung für `jsonencode(...)` |
| `github.com/agext/levenshtein` | v1.2.1 | Apache-2.0 | transitiv (HCL) |
| `github.com/apparentlymart/go-textseg/v15`, `/v17` | v15.0.0, v17.0.1 | MIT | transitiv (HCL, go-cty) |
| `github.com/mitchellh/go-wordwrap` | v1.0.1 | MIT | transitiv (HCL) |
| `golang.org/x/mod`, `golang.org/x/sync`, `golang.org/x/text`, `golang.org/x/tools` | v0.29.0, v0.18.0, v0.31.0, v0.38.0 | BSD-3-Clause | transitiv (HCL, go-cty) |
