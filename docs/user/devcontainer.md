# Einbindung in einen Devcontainer

`bedrock-eu-check` kann in einem Devcontainer als Feature oder als
vorinstalliertes Tool bereitgestellt werden
([`LH-TECH-005`](../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)).

**Stand:** Das Feature ist im Repository enthalten und wird in der CI gegen ein lokal
gebautes Release getestet. Es ist **noch nicht veröffentlicht** (kein OCI-Feature, kein
Release), und es wurde **nicht mit der Devcontainer-CLI erprobt**. Mit der
Standard-`baseUrl` lässt es sich erst installieren, wenn ein Release mit den
erwarteten Artefakten existiert (siehe [`releasing.md`](releasing.md)).

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
(z. B. Sigstore) sind ein Folgepunkt. Der Default `1.0.0` zeigt auf ein Release, das bis
zur Veröffentlichung nicht existiert; die Installation schlägt dann mit
„Download … fehlgeschlagen“ fehl.

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
