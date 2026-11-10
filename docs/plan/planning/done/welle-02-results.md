# Welle welle-02-config-scanner: Ergebnisse

**Abgeschlossen:** 2026-11-10. **Meilenstein:** M2 — vollständiger statischer Scan.

## Closure-Trigger

- Alle Slices done: slice-004 bis slice-007.
- `make verify-clean`, `make deps-verify` und CI grün auf `main`.
- Je Quelle ein Positiv- und ein Negativbeispiel im Repo (Claude-Settings, Devcontainer, Compose, Dockerfile, Terraform, Konfigurationsdatei); Findings tragen Datei und Zeile.
- Kein Secret-Wert in der Ausgabe ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)), per Akzeptanztest je Quelle geprüft.

## Lieferumfang

Befehle `local` (Env, Claude-Settings, Devcontainer, Compose, Dockerfile), `terraform` und `all` ohne AWS; Konfigurationsdatei `bedrock-eu-check.yml` mit `--config`; Reporter mit `Datei:Zeile`; Vendoring nach [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md) (`yaml.v3`, `hcl/v2`), CI mit `gates` und `deps-verify`.

## Steering-Loop

1. Verträge zuerst, dann parallel: Kern (w1) und Adapter (w2) auf einem im Plan festgehaltenen Vertrag.
2. Gates im sauberen Checkout (`make verify-clean`) und in CI, nicht im Arbeitsbaum; bei Workflows mit mehreren Jobs auf alle Ergebnisse warten.
3. Modus- und Befehlszuordnung vor dem Code im Lastenheft prüfen ([`LH-MOD-001`](../../../../spec/lastenheft.md#lh-mod-001--lokaler-modus), [`LH-MOD-003`](../../../../spec/lastenheft.md#lh-mod-003--terraform-modus)): Eine falsche Zuordnung (`scan`, Terraform in `local`) kostete zwei Umbauten.
4. Gemeinsame Logik früh teilen (`secretkey`); Closure-PRs nicht liegen lassen.

## Folgeaufgaben

- Compose-Interpolation, `env_file`/`.env`/`ARG`, rekursive Suche für Dockerfile- und Compose-Dateien.
- `apiKeyHelper` (Spec-Präzisierung zu CLAUDE-003), externe Policy-Dokumente und `.tfvars` bei Terraform.
- Optionales Refactoring `CheckLocal` mit optionalem Settings-Reader.
