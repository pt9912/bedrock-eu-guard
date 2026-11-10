# Slice slice-014: Befehl `version` und Build-Informationen

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-05-release.

**Bezug:** [`LH-TECH-002`](../../../../spec/lastenheft.md#lh-tech-002--cli-struktur), [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check version` gibt Version und Commit des Builds aus. Beide
kommen zur Buildzeit per `-ldflags` und sind zugleich die Quelle der
SARIF-Tool-Version. Der Build ist dafür vorbereitet, bitgleich
reproduzierbar zu sein.

## 2. Verträge

- **Quelle:** Zwei Paketvariablen im CLI-Rahmen (`version`, `commit`), Defaults `dev` und `unknown`. Sie werden per `-ldflags "-X …/cli.version=… -X …/cli.commit=…"` gesetzt; die bisherige Variable `toolVersion` entfällt.
- **Befehl:** `bedrock-eu-check version` schreibt genau eine Zeile `bedrock-eu-check <version> (<commit>)` auf stdout, ohne Netz, Exit `0`. Zusätzliche Argumente sind ein Aufruffehler (Exit `2`).
- **SARIF:** `tool.driver.version` ist `version` (nicht der Commit).
- **Makefile:** `make build` setzt `VERSION` (Default `dev`) und `COMMIT` (Default: `git rev-parse --short=12 HEAD`, sonst `unknown`) und baut mit `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false` und `-ldflags "-s -w …"`.

## 3. Definition of Done

- [ ] `version`-Befehl mit Tests (Ausgabeformat, Defaults, mit gesetzten Variablen, Argumentfehler).
- [ ] SARIF nutzt dieselbe Version; Test.
- [ ] `make build` bettet Version und Commit ein; ein Aufruf des gebauten Binarys bestätigt es.
- [ ] Reproduzierbarkeit geprüft: zwei Builds mit gleichen Eingaben liefern dieselbe SHA256; Ergebnis in der Closure-Notiz.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/adapter/driving/cli/` (`version`, Variablen) | update | w2 |
| `Makefile` (`build` mit Flags) | update | w2 |

## 5. Trigger

[`welle-05-release`](../welle-05-release.md) gestartet (Plan gemerged); `--ci` und SARIF gemergt (gleiche Datei).

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- Bitgleiche Builds hängen auch vom Go-Image (per Digest gepinnt) und vom Modulinhalt (`vendor/`) ab; Unterschiede zwischen Plattformen sind erwartbar.
- Ein Versionsschema (SemVer) und die Quelle der Version im Release (Tag) legt der Release-Slice fest.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
