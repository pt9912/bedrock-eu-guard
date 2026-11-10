# Slice slice-016: Docker-Image und reproduzierbarer Build

**Status:** done.

**Welle:** welle-05-release.

**Bezug:** [`LH-TECH-004`](../../../../spec/lastenheft.md#lh-tech-004--docker-image), [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`LH-NF-005`](../../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit), [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

Das Werkzeug ist als Docker-Image verfügbar
(`docker run --rm -v "$PWD:/work" bedrock-eu-check all`), baut reproduzierbar
aus gepinnten Basen und läuft ohne Root.

## 2. Verträge

- **Zwei Ziele** in einem mehrstufigen `Dockerfile`:
  - `slim` (Default): `FROM scratch` mit dem statischen Binary. Modi `local` und `terraform`. Ohne Shell.
  - `full`: Basis `amazon/aws-cli` (Version 2.39.7, per Tag **und** Digest gepinnt) plus das Binary. Alle Modi, auch `aws` und `all`.
- **Build-Stufe:** Go-Image per Digest (dasselbe wie in den Gates), `vendor/` und `-mod=vendor`, `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `VERSION`/`COMMIT` als Build-Argumente. Der Build selbst braucht kein Netz; nur das Ziehen der Basis-Images.
- **Laufzeit:** Nutzer `65532:65532` (kein Root), Arbeitsverzeichnis `/work`, `ENTRYPOINT` ist das Binary; die Images laufen mit `--read-only` und `--network none` (außer dem Modus `aws`, der Netz braucht).
- **`.dockerignore`:** nur `go.mod`, `go.sum`, `vendor`, `cmd`, `internal` gelangen in den Kontext; Testdaten und `.git` nicht.
- **Make-Targets:** `make image` baut beide Ziele, `make image-test` ruft zusätzlich `tools/image-smoke.sh` auf. Beide sind **nicht** Teil von `make gates` (Netz nötig).
- **CI:** eigener Workflow `image` (mit Netz), nur bei Änderungen, die das Image betreffen.
- **Außenwirkung:** Das Werkzeug **veröffentlicht nichts**. Push in eine Registry, Tags und Release sind ein eigener Schritt mit ausdrücklicher Freigabe beider Nutzer.

## 3. Definition of Done

- [x] `Dockerfile`, `.dockerignore`, `tools/image-smoke.sh`, Make-Targets, Workflow `image`.
- [x] Smoke-Test je Image: läuft nicht als Root; `version`; `local` auf Gut- und Schlecht-Beispiel (Exit `0` bzw. `1`); `terraform --format json`; `aws`: `slim` Exit `2` (kein `aws`-Binary), `full` Exit `1` (keine Credentials); `full` meldet die gepinnte AWS-CLI-Version.
- [x] Reproduzierbarkeit: Das Binary im Image, aus `make build` und aus einem Neubau ohne Cache hat dieselbe SHA256; Ergebnis in der Closure-Notiz.
- [x] Doku: welche Modi je Image funktionieren (siehe unten) in `docs/user/` bzw. im Release-Slice.
- [x] `make verify-clean` und CI grün (`gates`, `image`).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

| Image | Modi |
|---|---|
| `slim` | `local`, `terraform`, `version` (`aws`/`all` enden mit Tool-Fehler Exit 2) |
| `full` | alle |

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `Dockerfile`, `.dockerignore` | neu | w2 |
| `tools/image-smoke.sh`, `Makefile` (`image`, `image-test`) | neu/update | w2 |
| `.github/workflows/image.yml` | neu | w2 |

## 5. Trigger

[`slice-014`](slice-014-version-befehl.md) gemerged (Version/Build-Argumente).

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- **Image-Reproduzierbarkeit:** Das Binary ist bitgleich; die Image-ID kann wegen Zeitstempeln der Ebenen abweichen. Bitgleiche Images (z. B. über `SOURCE_DATE_EPOCH` und OCI-Export) sind ein Folgepunkt.
- **Größe von `full`:** rund 660 MB durch die AWS-CLI-Basis; Updates der AWS-CLI erfordern ein neues Pin (Tag und Digest).
- **Digest-Pin und Sicherheitsupdates:** Gepinnte Basen bekommen keine automatischen Updates; ein Aktualisierungsprozess gehört in die Release-Doku.
- **Windows/macOS-Hosts:** nur Linux-Container getestet.
- **Veröffentlichung:** nicht Teil dieses Slices (siehe Außenwirkung).

## 8. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-10 (PR #53, w2).

- **Funktioniert:** Mehrstufiges `Dockerfile` mit zwei Zielen: `slim` (`FROM scratch`, rund 7 MB, Modi `local`/`terraform`) und `full` (`amazon/aws-cli` 2.39.7 per Tag und Digest, rund 660 MB, alle Modi). Basis-Images per Digest gepinnt, Build ohne Netz (`vendor/`), Nutzer `65532`, `/work`, Binary als `ENTRYPOINT`. `tools/image-smoke.sh` prüft je Image: kein Root, `version`, `local` auf Gut-/Schlecht-Beispiel (Exit 0 bzw. 1), `terraform --format json`, `aws` (`slim` Exit 2, `full` Exit 1 ohne Credentials), AWS-CLI-Version; alles mit `--network none` und `--read-only`. Das Binary ist bitgleich im Image, aus `make build` und aus einem Neubau ohne Cache. Der Workflow `image` lief auf GitHub grün.
- **Anders als geplant:** Der Test deckte nicht auf, dass ein für uid 65532 nicht lesbares Mount zu Exit 2 führt (schon die Konfiguration ist nicht lesbar), sondern erst die Dokumentation am echten Image; das steht jetzt in `docs/user/docker.md`.
- **Offen:** Die Image-ID kann wegen Ebenen-Zeitstempeln abweichen (bitgleiche Images über OCI-Export mit festem Zeitstempel sind ein Folgepunkt). Registry und Namen für die Veröffentlichung sind nicht festgelegt; nichts wurde veröffentlicht. Aktualisierung der gepinnten Basen bleibt ein bewusster Schritt.
- **Steering-Loop:** Aussagen in der Nutzerdoku am echten Artefakt prüfen, nicht aus dem Code ableiten. Tests, die Netz brauchen (Basis-Images ziehen), laufen als eigener CI-Job und nicht in den netzlosen Gates.
- **Folge-Slices:** slice-017, slice-018.

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
