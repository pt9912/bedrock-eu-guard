# Slice slice-006: Devcontainer-, Compose- und Dockerfile-Scanner

**Status:** done.

**Welle:** welle-02-config-scanner.

**Bezug:** [`LH-FA-DEVCON-001`](../../../../spec/lastenheft.md#lh-fa-devcon-001--prüfung-von-devcontainerdevcontainerjson), [`LH-FA-DOCKER-001`](../../../../spec/lastenheft.md#lh-fa-docker-001--prüfung-von-docker-composeyml), [`LH-FA-DOCKER-002`](../../../../spec/lastenheft.md#lh-fa-docker-002--prüfung-von-dockerfile), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md), [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check local` prüft zusätzlich `devcontainer.json`, Compose-Dateien
und Dockerfiles des Projekts auf problematische Env-Definitionen (reiner
Lesezugriff, offline). Secret-Werte erreichen nie den Kern.

## 2. Verträge

- **Gemeinsamer Typ** `model.EnvFile{Path, Kind, Env []EnvEntry}` mit `EnvSourceKind` (`devcontainer|compose|dockerfile`). `EnvEntry` bekommt additiv `Scope` (Compose-Service bzw. Dockerfile-Stage). Secret-Schwärzung wie bei `ClaudeSettings` im Adapter (`Secret=true`, `Value=""`, zählt als gesetzt).
- **Port** `driven.EnvSourceReader.ReadEnvSource(kind, rel) (*model.EnvFile, error)`: Datei fehlt → `(nil, nil)`; kaputt → `*model.FileError`.
- **Dateisuche** im Kern über `driven.Filesystem.Glob` (optional in `app.Scan`: `EnvSources` + `Files`; ohne beides unverändertes Verhalten). Muster: `.devcontainer/devcontainer.json`, `.devcontainer.json`; `docker-compose*.yml|yaml`, `compose*.yml|yaml`; `Dockerfile*`.
- **Adapter-Pflichten (w2):** `devcontainer.json` ist JSONC (Kommentare, Trailing Commas, Zeilen erhalten); Env aus `containerEnv`, `remoteEnv` (und `build.args` nur als Hinweis); Compose: `environment` als Map und als Liste `KEY=VAL`, `env_file` nur als Fund (Inhalt nicht lesen), Service als `Scope`; Dockerfile: `ENV KEY=VAL` und alte Form `ENV KEY VAL`, mehrere Paare je Zeile, Zeilenfortsetzung mit `\`, Stage als `Scope`, `ARG` nur als Hinweis.

## 3. Verhalten je Regel

Gemeinsam: Anthropic-Zugang (`ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN` gesetzt oder `ANTHROPIC_BASE_URL` auf `api.anthropic.com`) → `FAIL` (`DIRECT_ANTHROPIC`); `AWS_REGION`/`AWS_DEFAULT_REGION` nicht erlaubt ([`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation)) → `FAIL` (`REGION_NOT_ALLOWED`). Datei fehlt → kein Finding; vorhanden, aber kaputt → `FAIL` mit `FILE_UNREADABLE`/`FILE_INVALID`, Scan läuft weiter.

| Regel | Zusätzlich |
|---|---|
| [`LH-FA-DEVCON-001`](../../../../spec/lastenheft.md#lh-fa-devcon-001--prüfung-von-devcontainerdevcontainerjson) | `PASS`: `CLAUDE_CODE_USE_BEDROCK=1` **und** erlaubte Region. `WARN` `NO_CLAUDE_CONFIG`: weder Bedrock-Flag noch Region noch Anthropic-Variable. `WARN` `ENV_INCOMPLETE`: nur eines von beiden bzw. Flag ≠ `1`. |
| [`LH-FA-DOCKER-001`](../../../../spec/lastenheft.md#lh-fa-docker-001--prüfung-von-docker-composeyml) | `PASS`: mindestens eine Region gesetzt und alle erlaubt. Keine relevante Variable → kein Finding. |
| [`LH-FA-DOCKER-002`](../../../../spec/lastenheft.md#lh-fa-docker-002--prüfung-von-dockerfile) | `PASS`: Region erlaubt. `WARN` `REGION_NOT_SET`: keine Region per `ENV`. |

Pro Datei und Regel ein Finding (Zeile des ersten Belegs); mehrere Dateien einer Quelle werden einzeln bewertet.

## 4. Definition of Done

- [x] Kern (w1): `model.EnvFile`, `EnvSourceReader`, Regeln `rules/envsource.go`, Erweiterung `app.Scan` (Glob-Discovery), Double `coretest/fakeenvsource`, Unit-Tests je Zweig und Dateizustand.
- [x] Adapter (w2): Reader für alle drei Quellen (JSONC, Compose-Map/Liste, Dockerfile-Parser), Secret-Schwärzung, Zeilen, Tests inkl. Secret-Fälle.
- [x] CLI (w2): `local` übergibt `EnvSources` und `Files`; Testdaten good/bad je Quelle; Akzeptanztest ohne Secret-Wert in der Ausgabe.
- [x] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/`, `port/driven/`, `core/rules/envsource.go`, `core/app/scan.go`, `coretest/fakeenvsource/` | neu/update | w1 |
| `internal/adapter/driven/envsource/` (Reader, JSONC-/Compose-/Dockerfile-Parser) | neu | w2 |
| `internal/adapter/driving/cli/` | update | w2 |

## 6. Trigger

slice-005 gemerged; Plan gemerged.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- Compose-Interpolation (`${VAR}`, `${VAR:-default}`) wird nicht aufgelöst; der Wert gilt wörtlich (Hinweis, kein Fehlalarm bei `${AWS_REGION}`-Referenzen: Adapter markiert sie nicht als Region-Wert).
- `env_file`, `.env`-Dateien und Dockerfile-`ARG` sind ein Folgepunkt.
- Kandidatenmuster `Dockerfile*` trifft auch `Dockerfile.dev`; Treffer in `vendor/`-ähnlichen Verzeichnissen entstehen nicht, da nur das Scan-Root geprüft wird (keine Rekursion).

## 9. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-10 (PRs #23 Kern/w1, #25 Adapter/CLI/w2).

- **Funktioniert:** Devcontainer (JSONC), Compose (Map und Liste, Anker, Merge-Keys, Service als Scope) und Dockerfile (`ENV` alte und neue Form, Fortsetzungen, Stages) werden gelesen und gegen DEVCON-001, DOCKER-001 und DOCKER-002 geprüft; Findings tragen Datei und Zeile. Secrets werden im Adapter geschwärzt, Akzeptanztests prüfen, dass kein Wert in stdout/stderr erscheint. Variablenreferenzen (`$`) in Regionen erzeugen keinen Fehlalarm.
- **Anders als geplant:** Die Secret-Erkennung liegt jetzt in einem gemeinsamen Paket (`secretkey`), das auch `claudesettings` nutzt (Duplikat vermieden).
- **Offen:**
  - Compose-Interpolation (`${VAR:-default}`) wird nicht aufgelöst; `env_file`, `.env`-Dateien und Dockerfile-`ARG` sind Folgepunkte.
  - Die Dateisuche ist nicht rekursiv (nur Scan-Root bzw. `.devcontainer/`).
  - Die Secret-Schlüsselliste bleibt eine großzügige Heuristik.
- **Steering-Loop:** Gemeinsame Logik früh in ein geteiltes Paket ziehen, sobald ein zweiter Adapter sie braucht. Kern und Adapter parallel auf einem festen Vertrag funktionieren gut, solange der Vertrag im Plan steht und beide Seiten ihn vor dem Code bestätigen. Beim Mergen von Stapel-PRs vor jedem Schritt prüfen, ob die Basis bereits auf `main` ist.
- **Folge-Slices:** slice-007 (Terraform-Scanner, HCL-Bibliothek per Vendoring), danach Abschluss von Welle 2.

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regeln und Adapter auf dem Kern aus
Welle 1 und slice-004.
