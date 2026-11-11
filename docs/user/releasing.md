# Releasing & Betrieb — bedrock-eu-check

**Status:** Version 1.0.0 veröffentlicht am 2026-11-11 auf Commit `5f1b35cca776` (Tag, GitHub-Release, Images, Feature). Zuletzt konkretisiert mit
Welle 5 ([`LH-RM-005`](../../spec/lastenheft.md#lh-rm-005--version-10)).

Dieses Dokument (Rang 6 der Source Precedence) beschreibt, *wie*
`bedrock-eu-check` versioniert, gebaut, verteilt und in CI betrieben wird.
Es trägt **keine** eigenen Anforderungen — die stehen im Lastenheft.

> **Außenwirkung:** Tag, GitHub-Release, Image-Push und das Veröffentlichen des
> Devcontainer-Features sind nicht rückholbar. Sie geschehen **nur nach
> ausdrücklicher Freigabe beider Nutzer** und nie automatisch (siehe Abschnitt 5).

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

- Die Version steckt zur Buildzeit im Binary (`bedrock-eu-check version`, zugleich die
  Tool-Version in SARIF). Der Git-Tag heißt `v<version>`, z. B. `v1.0.0`.

## 2. Build & Distribution

- **Binary** (primäres Artefakt, [`LH-TECH-001`](../../spec/lastenheft.md#lh-tech-001--programmiersprache),
  Sprachwahl [ADR-0002](../plan/adr/0002-implementierungssprache.md)): `make dist VERSION=x.y.z` baut
  statisch gelinkte Binaries für Linux (`amd64`, `arm64`) nach `dist/v<version>/` und
  erzeugt `SHA256SUMS`; die Prüfsummen werden direkt verifiziert.

  ```text
  dist/v1.0.0/bedrock-eu-check_1.0.0_linux_amd64
  dist/v1.0.0/bedrock-eu-check_1.0.0_linux_arm64
  dist/v1.0.0/SHA256SUMS
  ```

  Andere Plattformen (macOS, Windows) werden nicht gebaut und nicht getestet.
- **Docker-Image** ([`LH-TECH-004`](../../spec/lastenheft.md#lh-tech-004--docker-image)): `slim` und `full`,
  siehe [`docker.md`](docker.md). Nicht veröffentlicht.
- **Devcontainer-Feature** ([`LH-TECH-005`](../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)):
  siehe [`devcontainer.md`](devcontainer.md). Es erwartet genau die Artefaktnamen oben unter
  `<baseUrl>/v<version>/`. Nicht veröffentlicht.
- **Reproduzierbarer Build** über `make` ([`LH-QA-004`](../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build)):
  Toolchain per Image-Digest gepinnt, Abhängigkeiten unter `vendor/`
  ([ADR-0004](../plan/adr/0004-abhaengigkeiten-und-offline-gates.md)), `-trimpath`,
  `CGO_ENABLED=0`, `-buildvcs=false`. Geprüft: zwei Builds mit gleichen Eingaben liefern
  dieselben Prüfsummen. Bei Docker-Images kann die Image-ID wegen Zeitstempeln abweichen;
  das Binary darin ist bitgleich.

## 3. Exit-Code-Vertrag (für CI-Konsumenten)

| Exit | Bedeutung |
|---|---|
| `0` | kein `FAIL`-Finding ([`LH-FA-EXIT-001`](../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung)) |
| `1` | mindestens ein `FAIL` ([`LH-FA-EXIT-002`](../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung)) |
| `2` | Tool-Fehler ([`LH-FA-EXIT-003`](../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler)) |

CI bricht bei `1` ab ([`LH-ZIEL-004`](../../spec/lastenheft.md#lh-ziel-004--cicd-fähigkeit), [`LH-PRI-MUSS-005`](../../spec/lastenheft.md#lh-pri-muss-005)).
Einbindung: [`ci.md`](ci.md).

## 4. Quality Gates in CI (`LH-QA-005`)

| Gate | Wo | Netz |
|---|---|---|
| `make gates` (doc-check, lint, test inkl. Architektur- und Doku-Abgleichstests) | Workflow `gates` | nein |
| `make self-check` (`bedrock-eu-check` gegen das eigene Repo) | Workflow `gates` | nein |
| `make image-test`, `make feature-test` | Workflow `image` (bei betroffenen Pfaden) | Basis-Images ziehen |
| `make deps-verify` (`vendor/` gegen Upstream) | Workflow `deps-verify` (bei Änderungen an `vendor/`, `go.mod`, `go.sum`) | ja |

Lokal prüft `make verify-clean` den committeten Stand in einem frischen Worktree.
Nur real existierende Targets stehen in den Sensors-Tabellen
([`../../harness/README.md`](../../harness/README.md)).

## 5. Release-Ablauf

Voraussetzung ist die **ausdrückliche Freigabe beider Nutzer** für diese Version. Der
Workflow `release-artifacts` baut nur und veröffentlicht nichts.

1. **Stand prüfen:** `main` ist in CI grün; lokal `make verify-clean`, `make self-check`,
   `make deps-verify`, `make image-test`, `make feature-test`. Alle Slices der Welle sind
   abgeschlossen, die Roadmap nennt den Meilenstein.
2. **Freigabe einholen** und festhalten (z. B. als Kommentar im Release-PR).
3. **Artefakte bauen:** Workflow `release-artifacts` manuell starten (Eingabe: Version,
   z. B. `1.0.0`) oder lokal `make dist VERSION=1.0.0`. Das Ergebnis liegt als
   Workflow-Artefakt bzw. unter `dist/v1.0.0/`; `sha256sum -c SHA256SUMS` muss „OK“ melden.
   Der Workflow ist erprobt: Auf dem Tag-Ref `v1.0.0` (`gh workflow run release.yml --ref v1.0.0 -f version=1.0.0`)
   liefert er `SHA256SUMS` bitgleich zum veröffentlichten Release. Beim Vergleich den Tag-Ref
   verwenden: Version und Commit stecken im Binary, auf `main` ergäben sich andere Prüfsummen.
4. **Veröffentlichen (manuell, außenwirksam):**
   - Tag setzen und pushen: `git tag -a v1.0.0 -m "bedrock-eu-check 1.0.0"`,
     `git push origin v1.0.0`.
   - GitHub-Release mit den Artefakten und Release-Notizen anlegen
     (`gh release create v1.0.0 dist/v1.0.0/* --notes-file <datei>`).
   - Docker-Images und Devcontainer-Feature nach `ghcr.io` über den Workflow `publish`
     (nur manuell, `workflow_dispatch`; Eingaben: `version`, `commit`, Schalter `images`
     und `feature`). Er baut aus dem Tag `v<version>`, prüft Commit und Labels, führt die
     Smoke-Tests vor dem Push aus und nutzt das `GITHUB_TOKEN` des Laufs (keine eigenen
     Zugangsdaten). Zuerst nur `images`, nach der Nachprüfung erst `feature`. Images:
     `ghcr.io/pt9912/bedrock-eu-check:<version>-slim|-full`; Feature:
     `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check`. Neue Pakete sind zunächst
     privat; die Sichtbarkeit wird danach in den Paketeinstellungen auf GitHub bewusst
     geändert.
5. **Nachprüfen:** Artefakt frisch herunterladen und `sha256sum -c` ausführen; das Feature mit
   der Standard-`baseUrl` in einem frischen Devcontainer installieren.

Nicht Teil von 1.0: signierte Releases (Sigstore/GPG), SBOM, Installer für macOS/Windows.

## 6. Gepinnte Abhängigkeiten aktualisieren

Gepinnte Basen erhalten keine automatischen Updates. Aktualisiert wird bewusst per PR,
danach laufen die Gates:

| Pin | Ort | Vorgehen |
|---|---|---|
| Go-Image (Digest) | `Makefile` (`GO_IMAGE`), `Dockerfile` | beide Stellen gleichzeitig ändern |
| AWS-CLI-Image (Tag und Digest) | `Dockerfile` (Ziel `full`), Test in `tools/image-smoke.sh` | Version in beiden anpassen |
| d-check (Digest) | `d-check.mk` | wie dort beschrieben |
| Go-Bibliotheken | `vendor/`, `go.mod`, `go.sum` | `make deps`, dann `make deps-verify`; `vendor/` nie von Hand ändern |

## 7. Sicherheits-Hinweise für den Betrieb

- Das Tool benötigt für AWS-Checks nur **lesende** IAM-Rechte
  ([`LH-SEC-003`](../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)/[`LH-SEC-004`](../../spec/lastenheft.md#lh-sec-004--least-privilege));
  die minimale Policy steht in [`aws-permissions.md`](aws-permissions.md).
- Es gibt **keine** Secrets aus ([`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)) und speichert keine
  Credentials ([`LH-SEC-001`](../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials)/[`LH-SEC-002`](../../spec/lastenheft.md#lh-sec-002--keine-speicherung-von-claude-anthropic-secrets)).
- Prüfsummen schützen vor Korruption, nicht vor einem kompromittierten Ursprung; signierte
  Releases sind ein Folgepunkt.
