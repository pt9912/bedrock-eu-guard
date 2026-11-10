# Slice slice-017: Devcontainer-Feature

**Status:** done.

**Welle:** welle-05-release.

**Bezug:** [`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration), [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check` lässt sich in einem Devcontainer als Feature installieren
([`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration) verlangt „Feature oder vorinstalliertes Tool“). Die Installation lädt
ein Release-Binary und prüft es gegen eine Prüfsummendatei; ohne passende
Prüfsumme wird nichts installiert.

## 2. Verträge

- **Layout** nach der Devcontainer-Feature-Konvention: `devcontainer-feature/src/bedrock-eu-check/devcontainer-feature.json` und `install.sh`.
- **Optionen:** `version` (Default `1.0.0`, ohne führendes `v`) und `baseUrl` (Default: GitHub-Release-Downloads; für Spiegel oder Tests, `file://` möglich). Die Optionen kommen als `VERSION` und `BASEURL` in die Umgebung von `install.sh`.
- **Artefakte, die das Feature erwartet** (der Release-Workflow aus slice-018 erzeugt sie): `<baseUrl>/v<version>/bedrock-eu-check_<version>_linux_<arch>` für `amd64` und `arm64` sowie `<baseUrl>/v<version>/SHA256SUMS` im Format `<hash>  <datei>`.
- **Ablauf:** Architektur erkennen, Binary und `SHA256SUMS` laden, Prüfsumme vergleichen, nach `/usr/local/bin/bedrock-eu-check` installieren, `version` ausführen. `baseUrl` darf nur mit `https://` oder `file://` beginnen. Fehlerfälle (Download, fehlende oder falsche Prüfsumme, unbekannte Architektur, ungültige Version, unzulässige `baseUrl`) enden mit Exit ≠ 0 und installieren nichts.
- **Alternative „vorinstalliert“:** Das Binary lässt sich per Multi-Stage-`COPY --from=<image>` aus dem `slim`-Image in ein eigenes Devcontainer-Image übernehmen (siehe `docs/user/devcontainer.md`).
- **Außenwirkung:** Das Feature wird **nicht** als OCI-Feature veröffentlicht; das ist ein eigener, freigegebener Schritt.

## 3. Definition of Done

- [x] `devcontainer-feature.json` (gültiges JSON, `id` = Ordnername) und `install.sh` (POSIX-Shell, `set -eu`).
- [x] `tools/feature-test.sh` / `make feature-test`: Struktur, Syntax, **Installation gegen ein lokal gebautes „Release“** (per `file://`, Container ohne Netz), Aufruf von `version`; Negativfälle (manipulierte Prüfsumme, ungültige Version, unzulässige `baseUrl`).
- [x] CI: Schritt im Workflow `image` (Docker nötig), Pfadfilter um `devcontainer-feature/**` erweitert.
- [x] `docs/user/devcontainer.md` mit Nutzung und Hinweis auf den Stand (nicht mit der Devcontainer-CLI erprobt, noch nicht veröffentlicht).
- [x] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `devcontainer-feature/src/bedrock-eu-check/` | neu | w2 |
| `tools/feature-test.sh`, `Makefile` (`feature-test`) | neu/update | w2 |
| `.github/workflows/image.yml` | update | w2 |
| `docs/user/devcontainer.md` | neu | w2 |

## 5. Trigger

[`slice-016`](slice-016-docker-image.md) gemerged (Image, Test-Infrastruktur).

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- **Kein echtes Release:** Bis der Release-Workflow Artefakte veröffentlicht, ist das Feature mit Standard-`baseUrl` nicht installierbar; der Test deckt den Ablauf mit einem lokalen Release ab.
- **Nicht mit der Devcontainer-CLI erprobt:** Die Feature-Metadaten folgen der Konvention, ein Lauf mit der Devcontainer-CLI (`devcontainer features test`) braucht Netz und Container-Laufzeit und ist ein Folgepunkt.
- **Vertrauensmodell:** `SHA256SUMS` stammt vom selben Ursprung wie das Binary und schützt nur vor Korruption, nicht vor einem kompromittierten Ursprung; Signaturen (Sigstore) sind ein Folgepunkt.
- **Plattformen:** nur Linux (`amd64`, `arm64`); `install.sh` setzt `curl` und `sha256sum` voraus (Debian/Ubuntu-Basen) und installiert `curl` nur per `apt-get`.
- **Veröffentlichung als OCI-Feature** (Registry, Versionierung) ist nicht Teil des Slices.

## 8. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-10 (PR #64, w2).

- **Funktioniert:** `devcontainer-feature/src/bedrock-eu-check` (`devcontainer-feature.json` mit `version`, `baseUrl` und `install.sh`) installiert das Release-Binary: Architektur erkennen, Binary und `SHA256SUMS` laden, Prüfsumme vergleichen, erst dann installieren. `baseUrl` darf nur mit `https://` oder `file://` beginnen. `make feature-test` installiert ein lokal gebautes „Release“ per `file://` in einen Container ohne Netz, prüft `version` und die Negativfälle (manipulierte Prüfsumme, ungültige Version, unzulässige `baseUrl`); der Test läuft im Workflow `image` grün. Nutzung und Variante „vorinstalliert“ (Multi-Stage-`COPY`) stehen in `docs/user/devcontainer.md`.
- **Anders als geplant:** Auf Review wurde das Vertrauensmodell ehrlich benannt: `SHA256SUMS` stammt vom selben Ursprung wie das Binary und schützt vor Korruption, nicht vor einem kompromittierten Ursprung.
- **Offen:** Nicht veröffentlicht (kein OCI-Feature) und nicht mit der Devcontainer-CLI erprobt; mit der Standard-`baseUrl` erst nach einem Release mit den erwarteten Artefaktnamen installierbar. Signaturen (Sigstore) sind ein Folgepunkt; nur Linux (`amd64`, `arm64`).
- **Steering-Loop:** Ein Installationsskript gegen ein lokal gebautes „Release“ testen (inklusive Negativfälle) macht es ohne Veröffentlichung überprüfbar. Sicherheitsaussagen (was die Prüfsumme leistet) gehören ausdrücklich in die Doku.
- **Folge-Slices:** slice-018.

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
