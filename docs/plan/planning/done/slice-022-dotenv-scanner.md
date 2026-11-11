# Slice slice-022: `.env`-Scanner

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-05-release (Vorlauf für das Release 1.0, Entscheidung der Auftraggeber: Option 1b).

**Bezug:** [`LH-FA-LENV-007`](../../../../spec/lastenheft.md#lh-fa-lenv-007--prüfung-von-env-dateien) (neu), [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten), [`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-11.

---

## 1. Ziel

`bedrock-eu-check local` prüft `.env`-Dateien rekursiv auf direkte
Anthropic-Zugänge und nicht erlaubte Regionen. Das schließt die Lücke bei
[`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002): ein `ANTHROPIC_API_KEY` in einer `.env` wird jetzt erkannt. Werte erscheinen nie in der Ausgabe.

## 2. Entscheidung und Lastenheft

Die Auftraggeber haben **Option 1b** gewählt (siehe `release-1.0-entscheidung-qa003.md` in `in-progress/`) und
erlaubt, den Wortlaut im Lastenheft einzutragen. Eingetragen sind:

- **Neu** [`LH-FA-LENV-007`](../../../../spec/lastenheft.md#lh-fa-lenv-007--prüfung-von-env-dateien): Akzeptanzkriterien siehe Lastenheft.
- **Präzisiert** [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten): `iam-policy.json` ist erst gefordert, wenn eine Anforderung zur Prüfung eigenständiger Policy-Dateien besteht (Soll-Anforderung nach 1.0).

**Abweichung von der Entscheidungsvorlage:** Das Dateimuster ist bewusst breiter als in der Vorlage vorgeschlagen (`.env`, `.env.local`, `.env.*.local`): geprüft werden `.env` und `.env.*`, ausgenommen `*.example`, `*.sample`, `*.template`. Grund: `.env.production` und `.env.development` sind üblich, und das Muster entspricht der `.gitignore` des Repositorys. Moduszuordnung `local`, Suche rekursiv (Festlegung 4 der Vorlage).

## 3. Verträge

- `model.EnvDotenv` (`dotenv`) als weitere `EnvSourceKind`; `EnvFile` und `EnvEntry` unverändert.
- Neue kleine Schnittstelle `driven.EnvFileFinder{ Find(kind) ([]string, error) }` (rekursiv, sortiert, relative `/`-Pfade, Ausschlüsse `.git`, `.terraform`, `vendor`, `node_modules`, `.tmp`, keine Symlinks); der bestehende `EnvSourceReader` bleibt unverändert und liest `.env`-Dateien über `ReadEnvSource(EnvDotenv, rel)`.
- `app.Scan` bekommt das optionale Feld `EnvFinder`; ohne `EnvFinder` (oder ohne `EnvSources`) wird keine `.env` geprüft. `exclude` wird wie bei Terraform im Kern angewendet. Ein Fehler der Suche selbst ist ein Finding (`FILE_UNREADABLE`), kein Tool-Fehler.
- Regel [`LH-FA-LENV-007`](../../../../spec/lastenheft.md#lh-fa-lenv-007--prüfung-von-env-dateien) (`CheckDotenv`) mit der Auswertung der bestehenden Env-Quellen: `FAIL` bei `ANTHROPIC_API_KEY`/`ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL` auf `api.anthropic.com` oder nicht erlaubter Region; `PASS`, wenn mindestens eine Region gesetzt ist und alle erlaubt sind; sonst kein Finding; kaputte Datei `FAIL` mit `FILE_INVALID`/`FILE_UNREADABLE`.

**Adapter-Pflichten (w2):** `envsource.Reader` implementiert zusätzlich `EnvFileFinder` und liest `dotenv`: Kommentare, `export`-Präfix, Anführungszeichen, Zeilenfortsetzungen, Zeilennummern; Secret-Schwärzung wie bei den übrigen Quellen (`ANTHROPIC_API_KEY` usw. mit `Secret=true`, `Value=""`); Werte mit Interpolation (`${…}`) werden für Regionen nicht geliefert; Testdaten `good`/`bad`.

## 4. Definition of Done

- [x] Lastenheft: [`LH-FA-LENV-007`](../../../../spec/lastenheft.md#lh-fa-lenv-007--prüfung-von-env-dateien) und die Präzisierung von [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) eingetragen (in diesem Slice umgesetzt).
- [x] Kern (w1): `EnvDotenv`, `EnvFileFinder`, `CheckDotenv`, `app.Scan` mit `EnvFinder`, Double `fakeenvsource.Finder`, Unit-Tests; Regelkatalog und Referenz ergänzt (Doku-Drift-Test grün).
- [x] Adapter (w2): Parser und Finder, Testdaten `good`/`bad` unter `testdata/` mit gezielter `.gitignore`-Negation; CLI übergibt `EnvFinder`; Test ohne Wert in der Ausgabe.
- [x] Doku (w2): Handbuch und Fehlersuche; die Aussage „`.env` wird nicht geprüft“ entfällt.
- [ ] Matrix-Test zu [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) um `.env` erweitert; `iam-policy.json` bleibt als bekannte Lücke benannt.
- [x] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `spec/lastenheft.md`, `internal/hexagon/core/model/envsource.go`, `port/driven/files.go`, `core/rules/{envsource,claude}.go`, `core/app/scan.go`, `coretest/fakeenvsource/`, `docs/user/{rules,reference}.md` | update | w1 |
| `internal/adapter/driven/envsource/`, `internal/adapter/driving/cli/`, Testdaten, `.gitignore`, Handbuch/Fehlersuche | update/neu | w2 |

## 6. Trigger

Entscheidung der Auftraggeber zu Option 1b (erteilt); das Release 1.0 folgt auf dem Stand nach diesem Slice.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- **Testdaten und `.gitignore`:** `.env` und `.env.*` sind ignoriert (Schutz für [`LH-SEC-001`](../../../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials) und [`LH-SEC-002`](../../../../spec/lastenheft.md#lh-sec-002--keine-speicherung-von-claude-anthropic-secrets)); Testdaten brauchen eine gezielte Negation oder `git add -f` und nur offensichtlich synthetische Werte (Marker `GEHEIM-…`). Nach dem Anlegen mit `git status` prüfen, dass sie tatsächlich verfolgt werden.
- **Dialekte:** `.env` ist kein einheitliches Format; der Parser deckt die gängige Syntax ab, alles andere wird als `FILE_INVALID` gemeldet.
- **Fehlalarme:** Beispieldateien sind ausgenommen; echte `.env`-Dateien mit absichtlich gesetzter Anthropic-Variable sind ein berechtigter `FAIL`.
- **Release:** Der Commit-SHA für die Freigabe steht erst nach diesem Slice fest.

## 9. Closure-Notiz (nach `done/`)

**Adapter-Seite (w2):**

- **Funktioniert:** Parser (`dotenv.go`) für Kommentare, `export`, Anführungszeichen (auch mehrzeilig), Inline-Kommentare und BOM; Secret-Werte werden im Adapter verworfen (nur gesetzt/nicht gesetzt); defekte Dateien ergeben `FILE_INVALID` nur mit Zeilennummer, ohne Inhalt. Rekursiver Finder (`finder.go`) überspringt `.git`, `.terraform`, `vendor`, `node_modules`, `.tmp` und Symlinks und lässt Beispieldateien (`.example`, `.sample`, `.template`) aus. Gilt nur im Modus `local`. CLI-Tests über alle Formate ohne Wert in der Ausgabe; Handbuch, README und Quickstart angepasst.
- **Anders als geplant:** Die Suche ist rekursiv (nicht nur im Root), damit `.env`-Dateien in Unterprojekten nicht übersehen werden; `exclude` greift auch dort. Variablenreferenzen (`${X}`) als Region-Wert werden verworfen statt als Fehlalarm gemeldet. Die `.gitignore` bekam eine Ausnahme nur für `internal/**/testdata/**`; eine `.env` im Repository-Root bleibt ignoriert (mit `git check-ignore` geprüft).
- **Offen:** Matrix-Test zu [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) um `.env` (w1); `iam-policy.json` bleibt als eigenständige Policy-Datei ungeprüft.
- **Steering-Loop:** Testdaten, die von Ignore-Regeln geschützt sind, nach dem Anlegen mit `git ls-files` bzw. `git check-ignore` prüfen, bevor man dem grünen lokalen Lauf traut; ein Verhalten, das nur durch eine lokale Ausnahme läuft, fällt sonst erst in CI auf.


## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regel und Reader-Erweiterung auf dem
bestehenden Kern.
