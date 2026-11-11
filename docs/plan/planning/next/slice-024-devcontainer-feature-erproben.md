# Slice slice-024: Devcontainer-Feature mit der devcontainer-CLI erproben und korrigieren

**Status:** next. **Verantwortlich:** w2 (Plan-Review: w1). **Freigabe zum Erproben:** Nutzer 2026-11-11 („slice 024 freigegeben“).

**Welle:** [`welle-06-release-haertung`](../welle-06-release-haertung.md).

**Bezug:** [`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration), [`welle-05-results.md`](../done/welle-05-results.md) (bekannter Mangel im veröffentlichten Feature).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Das Devcontainer-Feature wird mit der echten `devcontainer`-CLI installiert und geprüft. Gefundene Mängel werden behoben, darunter die veraltete Option-Beschreibung. Das Ergebnis erscheint als neue **Feature-Version** (Vorschlag 1.0.1); das Werkzeug bleibt 1.0.0.

## 2. Definition of Done

- [ ] CI-Job (Netz, nicht Teil von `make gates`), der mit der `devcontainer`-CLI (Version gepinnt) zwei Beispiel-Container baut und `bedrock-eu-check version` ausführt: Fall A mit dem Feature aus dem Arbeitsstand, Fall B mit dem veröffentlichten Feature von `ghcr.io`.
- [ ] Mängel aus der Erprobung behoben; die Option-Beschreibung „zeigt auf ein Release, das noch nicht existiert“ ist korrigiert.
- [ ] `feature-test` (lokal, `file://`) bleibt grün; `devcontainer.md` nennt den erprobten Weg und entfernt „nicht erprobt“.
- [ ] Neue Feature-Version in `devcontainer-feature.json`; Veröffentlichung über den Workflow `publish` (Schalter `feature`) **nur nach dem Ja beider Nutzer** zu Feature-Version, Commit und Ziel.
- [ ] Veröffentlichungsweg festgelegt (Variante b): eigener Tag `feature-v<x.y.z>` auf dem Fix-Commit; der Workflow `publish` nimmt dafür eine Eingabe `ref` (Standard `v<version>`) und prüft Commit gegen diesen Tag. Grund: Der Tool-Tag `v1.0.0` zeigt auf einen Stand mit der alten `devcontainer-feature.json`; ein `publish` mit `feature=true` am Tool-Tag würde die alte Datei erneut veröffentlichen. Variante a (Tool-Patch 1.0.1 mit neuem Tool-Tag, neuen Binaries und Images) nur, falls sich auch Tool-Code ändert.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

**Ablauf der Erprobung** (veröffentlicht nichts):

1. Workflow `feature-cli` (Rechte `contents: read`, kein `packages: write`). **Fall A** läuft bei `workflow_dispatch` und bei Änderungen unter `devcontainer-feature/**` und `.github/workflows/feature-cli.yml`; **Fall B** (veröffentlichtes Artefakt) läuft nur bei `workflow_dispatch`, weil sein Ergebnis nicht von einem PR abhängt und unbeteiligte PRs nicht blockieren soll. Der Workflow: installiert die `devcontainer`-CLI in gepinnter Version (`@devcontainers/cli@0.72.0`, dieselbe wie in `publish.yml`) und baut zwei Beispiel-Container.
2. **Fall A, Feature aus dem Arbeitsstand:** Die Konfiguration verweist auf das Verzeichnis `devcontainer-feature/src` (lokaler Pfad). `devcontainer build` installiert das Feature mit der Standard-`baseUrl` aus dem GitHub-Release `v1.0.0`; im Container prüft `bedrock-eu-check version` die Ausgabe `bedrock-eu-check 1.0.0 (5f1b35cca776)` (Tool-Version 1.0.0 aus dem Release, unabhängig von der Feature-Version).
3. **Fall B, veröffentlichtes Feature:** Die Konfiguration verweist auf `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check:1`; gleiche Prüfung. Das belegt, was Nutzer heute bekommen.
4. Beide Fälle laufen mit einer Basis aus `mcr.microsoft.com/devcontainers/base:ubuntu` (die Vorlage in `devcontainer.md`); Fehler der CLI (zum Beispiel Rechte, `curl`, Architektur) gelten als Mangel und werden behoben, nicht umgangen.
5. Befunde, Run-ID und Ergebnis kommen in die Closure-Notiz; erst dann wird entschieden, ob eine neue Feature-Version nötig ist.

**Veröffentlichung:** Eine neue Feature-Version erscheint nur über `publish` mit Eingabe `ref` und dem Tag `feature-v<x.y.z>` (angelegt über `release-create`, Ziel `feature`) und nur nach dem Ja **beider** Nutzer zu Feature-Version, Commit und Ziel. Dieser Slice bereitet sie vor, führt sie aber nicht aus.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.github/workflows/feature-cli.yml` | neu | Erprobung mit der echten CLI, eigener Job mit Netz |
| `devcontainer-feature/src/bedrock-eu-check/devcontainer-feature.json` | update | Option-Beschreibung, Feature-Version |
| `.github/workflows/publish.yml` | update | Eingabe `ref` für `feature-v<x.y.z>` (nur Feature), gleiche Commit-gegen-Tag-Prüfung. Bedeutung von `version`: Bei `ref=feature-v<x.y.z>` ist `version` die **Feature-Version** `x.y.z` und muss gleich der Version in `devcontainer-feature.json` und dem Suffix des Tags sein (die bestehende Prüfung bleibt); die Eingabebeschreibung nennt das ausdrücklich. Ohne `ref` gilt wie bisher der Tool-Tag `v<version>` |
| `docs/user/devcontainer.md`, `README.md` | update | erprobter Weg |

## 4. Trigger

slice-025 done (nutzt `publish` und den gehärteten Ablauf). Mängel werden gebündelt, damit das Feature nur einmal neu veröffentlicht wird.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Die Neuveröffentlichung des Features ist nicht rückholbar (Tags `1`, `1.0`, `latest` wandern); nur mit Ja beider Nutzer.
- Die `devcontainer`-CLI braucht Container-Laufzeit und Netz; die Version der CLI wird gepinnt (Runner-Abhängigkeit).
- Fall B setzt voraus, dass das Paket `ghcr.io/pt9912/bedrock-eu-guard/bedrock-eu-check` anonym ladbar ist; geprüft am 2026-11-11 (Manifest `1` anonym abrufbar, HTTP 200). Wird es privat, schlägt Fall B fehl.
- Feature-Version und Tool-Version sind entkoppelt; der Default `version` des Features bleibt 1.0.0.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
