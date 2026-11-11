# Slice slice-023: Workflow `release-artifacts` auf GitHub erproben

**Status:** done.

**Welle:** [`welle-06-release-haertung`](welle-06-release-haertung.md).

**Bezug:** [`LH-QA-004`](../../../../spec/lastenheft.md#lh-qa-004--reproduzierbarer-build), [`welle-05-results.md`](../done/welle-05-results.md) (Herkunft der Release-Dateien).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Der Workflow `release-artifacts` läuft einmal auf GitHub (`workflow_dispatch` auf dem Tag-Ref `v1.0.0`, Version 1.0.0). Seine Binaries und `SHA256SUMS` werden mit dem veröffentlichten Release verglichen. Der Workflow baut nur und veröffentlicht nichts.

## 2. Definition of Done

- [x] Lauf: `gh workflow run release.yml --ref v1.0.0 -f version=1.0.0`; Ergebnis (Run-ID, Dauer) in der Closure-Notiz.
- [x] Artefakte des Laufs herunterladen und gegen das Release vergleichen: `SHA256SUMS` identisch oder Abweichungen erklärt.
- [x] Prüfung, dass der Lauf nichts veröffentlicht hat (kein neuer Tag, Release, Paket); Berechtigungen des Workflows sind nur `contents: read`.
- [x] Befund und mögliche Korrekturen am Workflow (z. B. `COMMIT`-Übergabe) als PR; ohne Befund nur Dokumentation in `releasing.md` (Status „erprobt“) und README.
- [x] `make verify-clean` und CI grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.github/workflows/release.yml` | update falls Befund | `make dist` im Workflow nutzt `git rev-parse` als Commit; auf dem Tag-Ref gleich dem Release-Commit |
| `docs/user/releasing.md`, `README.md` | update | „nicht erprobt“ entfällt nach erfolgreichem Lauf |

## 4. Trigger

Welle 06 gestartet (Pläne gemerged). Erster Slice der Welle; keine Abhängigkeit.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Der Lauf auf dem Tag-Ref nutzt die Workflow-Datei des Tags; Abweichungen zwischen Tag und `main` sind erwartbar und im Vergleich zu erklären.
- Ein Lauf auf `main` ergäbe einen anderen Commit und damit andere Prüfsummen; er ist nicht der Vergleichsfall.
- Der Workflow-Artefakt-Download braucht `gh` mit Leserechten auf Actions; fehlende Rechte nicht umgehen, sondern den Nutzer fragen.

## 7. Closure-Notiz (nach `done/`)

- **Funktioniert:** Der Workflow lief am 2026-11-11 auf GitHub per `workflow_dispatch` auf dem Tag-Ref `v1.0.0` (Run 42970275950, Commit `5f1b35cca776`, Ergebnis `success`). Seine `SHA256SUMS` sind **bitgleich** mit denen des veröffentlichten Releases (amd64 `32928970…`, arm64 `8f91f8e7…`); damit sind die lokal gebauten Release-Dateien auch aus CI reproduzierbar.
- **Nichts veröffentlicht:** Der Workflow hat nur `contents: read` und lädt nur ein Workflow-Artefakt hoch; danach gab es weiterhin nur den Tag `v1.0.0` und ein Release, kein Paket wurde angefasst.
- **Anders als geplant:** Es war keine Korrektur am Workflow nötig; geändert wurden nur README und `releasing.md` („nicht erprobt“ entfällt).
- **Hinweis:** Auf `main` statt auf dem Tag-Ref ergäbe der Lauf einen anderen Commit und damit andere Prüfsummen; der Tag-Ref ist der Vergleichsfall. GitHub meldet die Abkündigung von Node 20 für `actions/checkout@v4` und `actions/upload-artifact@v4`; die Pinning- und Versionsfrage ist ein Folgepunkt für `release.yml`.
- **Steering-Loop:** Eine Erprobung gegen ein veröffentlichtes Artefakt braucht den Tag-Ref als Eingabe, weil Version und Commit im Binary stecken; den Vergleichsfall vorher festlegen, statt das Ergebnis zu deuten.

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
