# Einbindung in einen Devcontainer

`bedrock-eu-check` kann in einem Devcontainer als Feature oder als
vorinstalliertes Tool bereitgestellt werden
([`LH-TECH-005`](../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)).

**Stand:** Das Feature ist als `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check`
(Tags `1`, `1.0`, `1.0.1`, `1.0.0`, `latest`; aktuelle Feature-Version 1.0.1) veröffentlicht und wird in der CI gegen ein lokal
gebautes Release getestet. Zusätzlich ist es mit der Devcontainer-CLI (Version 0.72.0, Workflow
`feature-cli`) erprobt: Das Feature aus dem Arbeitsstand wie auch das veröffentlichte Feature
von `ghcr.io` lassen sich in `mcr.microsoft.com/devcontainers/base:ubuntu` installieren, und
`bedrock-eu-check version` meldet `bedrock-eu-check 1.0.0 (5f1b35cca776)`. Nicht erprobt ist
`devcontainer up` (es blieb auf dem GitHub-Runner nach dem Start des Containers hängen, die
Ursache ist nicht ermittelt); erprobt ist `devcontainer build` mit anschließendem `docker run`. Die
Standard-`baseUrl` zeigt auf das [GitHub-Release v1.0.0](https://github.com/pt9912/bedrock-eu-guard/releases/tag/v1.0.0)
(siehe [`releasing.md`](releasing.md)).

## Variante A: Feature

Der Ordner `devcontainer-feature/src/bedrock-eu-check` enthält
`devcontainer-feature.json` und `install.sh`. Als lokales Feature wird er in den
Ordner `.devcontainer/` des Projekts kopiert:

```text
.devcontainer/
├── devcontainer.json
└── bedrock-eu-check/            # Kopie von devcontainer-feature/src/bedrock-eu-check
    ├── devcontainer-feature.json
    └── install.sh
```

```jsonc
// .devcontainer/devcontainer.json
{
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu",
  "features": {
    "./bedrock-eu-check": { "version": "1.0.0" }
  }
}
```

Oder direkt aus der Registry, ohne Kopie (mit der Devcontainer-CLI erprobt):

```jsonc
"features": {
  "ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check:1": { "version": "1.0.0" }
}
```

| Option | Default | Bedeutung |
|---|---|---|
| `version` | `1.0.0` | Version des Release-Binarys (ohne führendes `v`) |
| `baseUrl` | GitHub-Release-Downloads | Basis-URL der Artefakte, z. B. für einen Spiegel |

`install.sh` lädt `<baseUrl>/v<version>/bedrock-eu-check_<version>_linux_<arch>`
(`amd64` oder `arm64`) und `<baseUrl>/v<version>/SHA256SUMS`, vergleicht die
SHA256-Prüfsumme und installiert das Binary nach `/usr/local/bin`. Fehlt die
Prüfsumme oder weicht sie ab, bricht die Installation ab und es wird nichts
installiert. Voraussetzung sind `curl` und `sha256sum` (bei Debian/Ubuntu-Basen
vorhanden).

**Sicherheit:** `baseUrl` muss mit `https://` oder `file://` beginnen. `SHA256SUMS`
stammt vom selben Ursprung wie das Binary: Die Prüfung schützt vor Korruption und
Teil-Downloads, **nicht** vor einem kompromittierten Ursprung. Signierte Releases
(z. B. Sigstore) sind ein Folgepunkt. Der Default `1.0.0` zeigt auf das veröffentlichte Release v1.0.0; das Werkzeug 1.1.0 wählst du mit `"version": "1.1.0"`.

Zum Testen des Ablaufs ohne Release: `make feature-test` (baut ein lokales
„Release“, installiert es per `file://` in einen Container ohne Netz und prüft
auch die Fehlerfälle).

## Variante B: vorinstalliert über das Image

Statt eines Features kann das Binary aus dem `slim`-Image in ein eigenes
Devcontainer-Image übernommen werden:

```dockerfile
FROM bedrock-eu-check:slim AS tool
FROM mcr.microsoft.com/devcontainers/base:ubuntu
COPY --from=tool /usr/local/bin/bedrock-eu-check /usr/local/bin/bedrock-eu-check
```

Das Image muss dazu lokal gebaut oder aus einer Registry verfügbar sein; siehe
[`docker.md`](docker.md). Diese Variante ist nicht Teil der automatischen Tests.

## Nutzung im Devcontainer

```bash
bedrock-eu-check version
bedrock-eu-check local          # prüft Umgebung, Claude-Settings, Devcontainer, Compose, Dockerfile
bedrock-eu-check terraform
```

Die Benutzer-Settings (`~/.claude/settings.json`) liegen im Devcontainer im Home-Verzeichnis
des Container-Nutzers; `local` prüft sie dort. Der Modus `aws` braucht zusätzlich die
AWS-CLI und Zugangsdaten ([`aws-permissions.md`](aws-permissions.md)).
