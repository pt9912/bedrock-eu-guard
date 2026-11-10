# Slice slice-004: Filesystem-Port, Scan-Gerüst und Claude-Code-Settings

**Status:** done.

**Welle:** welle-02-config-scanner.

**Bezug:** [`LH-FA-CLAUDE-001`](../../../../spec/lastenheft.md#lh-fa-claude-001--benutzerweite-claude-code-settings-prüfen), [`LH-FA-CLAUDE-002`](../../../../spec/lastenheft.md#lh-fa-claude-002--projektweite-claude-code-settings-prüfen), [`LH-FA-CLAUDE-003`](../../../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte), [`LH-FA-CLAUDE-004`](../../../../spec/lastenheft.md#lh-fa-claude-004--claude-code-wizard-konfiguration-prüfen), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`SPEC-021`](../../../../spec/spezifikation.md#spec-021--finding-schema-json---format-json), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check local` liest `~/.claude/settings.json` sowie
`.claude/settings.json` und `.claude/settings.local.json` des Projekts
(reiner Lesezugriff, offline), prüft sie gegen die vier CLAUDE-Anforderungen
und meldet Findings mit Datei und Zeile. Secret-Werte erreichen nie den Kern.

## 2. Verträge (mit w1 abgestimmt)

- `model.Finding` += `File string`, `Line int`, `Recommendation string` (additiv; Nullwerte = bisheriges Verhalten).
- Reason-Konstanten `FILE_MISSING`, `FILE_UNREADABLE`, `FILE_INVALID`; `model.FileError{Path, Reason, Line}` ohne Dateiinhalt in der Message.
- `model.ClaudeSettings{Path, Model, Env []EnvEntry, Hits []Hit}`; `EnvEntry{Key, Value, Secret, Line}`; `Hit{Pattern, Line}` mit den Mustern `api.anthropic.com`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_API_KEY`.
- **Secret-Schwärzung im Adapter:** Für Secret-Schlüssel ist `Value=""` und `Secret=true`; das zählt als gesetzt. Liste (dokumentiert, getestet): `ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `apiKeyHelper` sowie `*_KEY`, `*_TOKEN`, `*_SECRET`, `*_PASSWORD`, `*_CREDENTIALS`. Hits tragen nur Muster und Zeile, nie Textausschnitte.
- Ports: `driven.Filesystem` (`ReadFile`, `Exists`, `Glob`; Pfade relativ zum Scan-Root) und `driven.ClaudeSettingsReader` (`ReadUser`, `ReadProject(rel)`). Datei fehlt → `(nil, nil)`; kaputt oder nicht lesbar → `*model.FileError`. Die Regeln machen daraus Findings, nie Exit `2`.

## 3. Verhalten je Regel (Entwurf, vor dem Test zu bestätigen)

| Regel | Datei fehlt | Datei vorhanden |
|---|---|---|
| [`LH-FA-CLAUDE-001`](../../../../spec/lastenheft.md#lh-fa-claude-001--benutzerweite-claude-code-settings-prüfen) (User-Settings) | `WARN` (`FILE_MISSING`) | `PASS`, wenn `CLAUDE_CODE_USE_BEDROCK=1` und keine direkte Anthropic-Konfiguration; `FAIL` bei direkter Konfiguration; sonst `WARN` (Bedrock nicht aktiviert) |
| [`LH-FA-CLAUDE-002`](../../../../spec/lastenheft.md#lh-fa-claude-002--projektweite-claude-code-settings-prüfen) (Projekt: `settings.json`, `settings.local.json`) | kein Finding (optional) | `PASS`, wenn EU-safe; `FAIL` bei nicht erlaubtem Modell (Klassifikation aus [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids)), nicht erlaubter Region oder API-Key |
| [`LH-FA-CLAUDE-003`](../../../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte) (Endpunkte) | kein Finding | `FAIL` bei gesetztem `ANTHROPIC_API_KEY`/`ANTHROPIC_AUTH_TOKEN` (Eintrag in `env`), bei Treffer `api.anthropic.com` oder bei `ANTHROPIC_BASE_URL` mit Host `api.anthropic.com`; reiner Textfund von `ANTHROPIC_API_KEY`/`ANTHROPIC_BASE_URL` außerhalb von `env` (z. B. in `permissions`, Kommentarstrings, Hooks) sowie `ANTHROPIC_BASE_URL` auf anderen Host: `WARN` |
| [`LH-FA-CLAUDE-004`](../../../../spec/lastenheft.md#lh-fa-claude-004--claude-code-wizard-konfiguration-prüfen) (Wizard in User-Settings) | kein Finding (deckt 001 ab) | `PASS` bei Bedrock `1` und erlaubter `AWS_REGION`; `FAIL` bei fehlender oder nicht erlaubter Region oder parallelem direktem Anthropic-Zugang |

Präzisierungen aus dem Review (w1):

- **Aktiv konfiguriert** ([`LH-FA-CLAUDE-003`](../../../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte) verlangt `FAIL` nur bei aktiver Konfiguration): Textfunde außerhalb von `env` sind höchstens `WARN`.
- **CLAUDE-002:** Das Modell stammt aus `model` bzw. aus `ANTHROPIC_*_MODEL` in `env` und wird mit `Classify` bewertet (`blocked` = `FAIL`, `unclassified` = `WARN`, `allowed` = `PASS`). Eine nicht erlaubte `AWS_REGION` ist `FAIL` ([`SPEC-011`](../../../../spec/spezifikation.md#spec-011--regionsklassifikation)). `settings.json` und `settings.local.json` werden einzeln bewertet, je Datei ein Finding pro Regel mit `File` und `Line`.
- **Doppelmeldungen sind gewollt:** Dieselbe Ursache (z. B. API-Key in den User-Settings) erzeugt je Regel-ID ein Finding (001, 003 und 004). Nicht deduplizieren.

Datei vorhanden, aber nicht lesbar oder ungültig: `FAIL` mit
`FILE_UNREADABLE`/`FILE_INVALID`, Scan läuft weiter
(siehe [`welle-02-config-scanner.md`](../welle-02-config-scanner.md) §5a).

## 4. Definition of Done

- [x] Verträge aus §2 umgesetzt, additiv und rückwärtskompatibel (`bedrock-eu-check local` unverändert).
- [x] Kern (w1): Regeln `rules/claude*.go`, Use-Case `app.Scan`, Doubles `coretest/memfs` und `coretest/fakesettings`, Unit-Tests je Zweig und je Dateizustand.
- [x] Adapter (w2): `adapter/driven/fs`, `adapter/driven/claudesettings` (JSON-Parser mit Zeilen aus Decoder-Offsets, Schwärzung, Hits, Home-Injektion), Tests inkl. Secret-Fälle.
- [x] CLI (w2): `bedrock-eu-check local [--root DIR]` führt Env- und Datei-Regeln aus (Scan-Root Default aktuelles Verzeichnis); Reporter gibt `Datei:Zeile` aus. Ein eigener Befehl `scan` entfällt (Entscheidung des Auftraggebers: [`SPEC-002`](../../../../spec/spezifikation.md#spec-002--regel-modell) kennt nur die Modi `local`/`aws`/`terraform`).
- [x] Testdaten im Repo: je ein positives und ein negatives Beispiel je Quelle; Akzeptanztest, dass kein Secret-Wert (`sk-…`) in stdout/stderr erscheint.
- [x] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/` | update (Finding, Settings-Typen, Reasons) | w1 |
| `internal/hexagon/port/driven/` | neu (`Filesystem`, `ClaudeSettingsReader`) | w1 |
| `internal/hexagon/core/rules/claude*.go`, `core/app/scan.go` | neu | w1 |
| `internal/hexagon/coretest/{memfs,fakesettings}/` | neu | w1 |
| `internal/adapter/driven/fs/` | neu | w2 |
| `internal/adapter/driven/claudesettings/` | neu | w2 |
| `internal/adapter/driven/report/` | update (`Datei:Zeile`) | w2 |
| `internal/adapter/driving/cli/` | update (`scan`) | w2 |

## 6. Trigger

[`welle-02-config-scanner`](../welle-02-config-scanner.md) gestartet (Plan gemerged) und Freigabe der Nutzer; Kernverträge von w1 gepusht, bevor die Adapter integriert werden.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- Zeilen aus `encoding/json`: der Decoder liefert nur Byte-Offsets; Umrechnung auf Zeilen im Adapter, mit Tests für CRLF.
- Die Schlüsselliste für Secrets ist eine Heuristik; Fehlalarme (zu viel geschwärzt) sind akzeptabel, Lücken nicht.
- `~/.claude/settings.json` liegt außerhalb des Scan-Roots; Home wird injiziert, Zugriff nur lesend ([`LH-SEC-003`](../../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff) gilt analog).

## 9. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-10 (PRs #15 Kern/w1, #16 Adapter/CLI/w2).

- **Funktioniert:** Claude-Code-Settings (User und Projekt) werden gelesen, geparst und gegen CLAUDE-001 bis -004 geprüft; Findings tragen Datei, Zeile und Empfehlung. Secrets werden im Adapter geschwärzt, der Kern sieht nie einen Wert; Akzeptanztests prüfen, dass kein Secret in stdout/stderr erscheint. Abstimmung über einen früh festgelegten Vertrag, Kern und Adapter parallel.
- **Anders als geplant:** Es gibt keinen Befehl `scan`. `local` führt Env- und Datei-Regeln aus ([`SPEC-002`](../../../../spec/spezifikation.md#spec-002--regel-modell) kennt nur die Modi `local`/`aws`/`terraform`). Das Verhalten je Regel (§3) wurde im Review von w1 präzisiert (aktiv konfiguriert, Doppelmeldungen gewollt).
- **Offen:**
  - `apiKeyHelper` (Top-Level-Schlüssel) wird nicht als direkte Anthropic-Konfiguration gewertet; das Lastenheft nennt ihn nicht. Dafür ist eine Spec-Präzisierung zu CLAUDE-003 nötig, nicht stilles FAIL.
  - Optionales Refactoring: `app.CheckLocal` mit optionalem Settings-Reader, damit der Composition Root nur noch einen Aufruf macht (separater kleiner PR, nur auf Wunsch).
  - Die Secret-Schlüsselliste ist eine großzügige Heuristik; Fehlalarme sind akzeptiert.
- **Steering-Loop:** Verträge zwischen den Agenten zuerst festlegen und im Plan festhalten, bevor Kern und Adapter parallel entstehen. Die Namensfrage für Befehle gehört vor den Code (Spec-Anker prüfen); hier kostete sie einen Umbau von #16. CI startet verzögert; vor dem Merge auf den Lauf für den Ende-Stand (Commit-SHA) warten.
- **Folge-Slices:** Plan für slice-005 (Konfigurationsdatei, YAML) nach Vendoring-Umbau gemäß [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md).

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Adapter und Regeln auf dem Kern aus
Welle 1, kein Bestandscode außerhalb davon.
