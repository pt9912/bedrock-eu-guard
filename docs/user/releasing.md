# Releasing & Betrieb — bedrock-eu-check

**Status:** Outline (Phase 2). Konkretisiert sich mit Welle 4/5
([`LH-RM-004`](../../spec/lastenheft.md#lh-rm-004--version-04)/[`LH-RM-005`](../../spec/lastenheft.md#lh-rm-005--version-10)). **Letzte Änderung:** 2026-06-24.

Dieses Dokument (Rang 6 der Source Precedence) beschreibt, *wie*
`bedrock-eu-check` versioniert, gebaut, verteilt und in CI betrieben wird.
Es trägt **keine** eigenen Anforderungen — die stehen im Lastenheft.

## 1. Versionierung

- **SemVer** (`Major.Minor.Patch`). Die Minor-Stufen folgen den
  Release-Stufen des Lastenhefts:

  | Version | Welle | Inhalt | LH |
  |---|---|---|---|
  | v0.1 | welle-01 | lokaler Env-/Modell-ID-Check, Exit-Codes | `RM-001` |
  | v0.2 | welle-02 | Claude-Code-/Devcontainer-/Docker-/Terraform-Scanner | `RM-002` |
  | v0.3 | welle-03 | AWS-STS-/Bedrock-/IAM-Prüfung | `RM-003` |
  | v0.4 | welle-04 | JSON-/SARIF-Ausgabe, CI-Modus | `RM-004` |
  | v1.0 | welle-05 | stabile Team-Version, Doku, Tests, Release-Artefakte | `RM-005` |

## 2. Build & Distribution

- **Einzelnes Binary** als primäres Artefakt ([`LH-TECH-001`](../../spec/lastenheft.md#lh-tech-001--programmiersprache),
  Sprachwahl [ADR-0002](../plan/adr/0002-implementierungssprache.md)).
- **Docker-Image** ([`LH-TECH-004`](../../spec/lastenheft.md#lh-tech-004--docker-image)):
  `docker run --rm -v "$PWD:/work" bedrock-eu-check all`.
- **Devcontainer-Feature / vorinstalliertes Tool** ([`LH-TECH-005`](../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)).
- **Reproduzierbarer Build** über `make` ([`LH-QA-004`](../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)); Toolchain
  per Image-Digest gepinnt.

## 3. Exit-Code-Vertrag (für CI-Konsumenten)

| Exit | Bedeutung |
|---|---|
| `0` | kein `FAIL`-Finding ([`LH-FA-EXIT-001`](../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung)) |
| `1` | mindestens ein `FAIL` ([`LH-FA-EXIT-002`](../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung)) |
| `2` | Tool-Fehler ([`LH-FA-EXIT-003`](../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler)) |

CI bricht bei `1` ab ([`LH-ZIEL-004`](../../spec/lastenheft.md#lh-ziel-004--cicd-fähigkeit), [`LH-PRI-MUSS-005`](../../spec/lastenheft.md#lh-pri-muss-005)).

## 4. Quality Gates in CI (`LH-QA-005`)

- **Doku-Referenz-Gate:** `make doc-check` (via d-check, Version/Digest im
  Makefile gepinnt — die *Reproduzierbarkeits-Bindung* der Sensors-Tabelle
  in [`../../harness/README.md`](../../harness/README.md)).
- **Code-Gates** (entstehen ab Welle 1): `make lint`, `make test`,
  `make self-check` (`bedrock-eu-check` gegen das eigene Repo).
- Nur real existierende Targets werden in den Sensors-Tabellen geführt
  (keine halluzinierten Gates).

## 5. Sicherheits-Hinweise für den Betrieb

- Das Tool benötigt für AWS-Checks nur **lesende** IAM-Rechte
  ([`LH-SEC-003`](../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)/[`LH-SEC-004`](../../spec/lastenheft.md#lh-sec-004--least-privilege)); minimale Policy wird mit Welle 3 dokumentiert.
- Es gibt **keine** Secrets aus ([`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)) und speichert keine
  Credentials ([`LH-SEC-001`](../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials)/[`LH-SEC-002`](../../spec/lastenheft.md#lh-sec-002--keine-speicherung-von-claude-anthropic-secrets)).
