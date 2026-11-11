# Slice slice-034: devcontainer-CLI aktualisieren und `devcontainer up` erproben

**Status:** done. **Verantwortlich:** w2. **Freigabe:** Nutzer 2026-11-11 („ja“).

**Welle:** ohne Welle (Wartung, Folge von slice-024).

**Bezug:** [`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration), [`slice-024`](../done/slice-024-devcontainer-feature-erproben.md) (Befund: `devcontainer up` hängt).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Die Workflows `feature-cli` und `publish` nutzen die aktuelle `devcontainer`-CLI statt der veralteten Version 0.72.0. `feature-cli` erprobt wieder den Weg, den Nutzer nehmen (`devcontainer up` und `devcontainer exec`), und nicht nur `build` mit `docker run`.

## 2. Ausgangslage (belegt)

- Die Diagnose zu `devcontainer up` (Läufe 42999455137 und 42999946761, Draft-PR #111, geschlossen) zeigt: Mit CLI 0.72.0 (veröffentlicht 2024-11-01) hängt `up` auf dem GitHub-Runner mit Docker 29.4.2; mit 0.86.1 und 0.90.0 endet es regulär. Die Varianten Shell-Probe, stdin und Lebenszyklus-Befehle ändern nichts.
- Aktuell sind CLI 0.90.0 (2026-10-28) und Docker 29.9.3; die 0.72.0 war beim Pinnen nicht gegen das aktuelle Release geprüft worden.

## 3. Definition of Done

- [x] `feature-cli`: CLI-Version 0.90.0 (gepinnt); Fall A und Fall B nutzen `devcontainer up` und `devcontainer exec` mit Schritt-Timeout; Fall A läuft grün auf dem PR, Fall B per Dispatch (Run-IDs in der Closure-Notiz).
- [x] `publish.yml`: CLI-Version 0.90.0 im Job `feature`; Beleg durch einen Trockenlauf (`dry_run=true`, `ref=feature-v1.0.1`, `version=1.0.1`, Commit des Tags): `features package` liefert dieselbe Struktur wie zuvor. Es wird **nichts veröffentlicht**.
- [x] Die CLI-Version steht an einer Stelle je Workflow (Umgebungsvariable), mit Kommentar zum Grund des Pins.
- [x] `devcontainer.md` und README: `devcontainer up` ist erprobt, der Hinweis auf das Hängen und die Ursache stehen (mit Läufen) in der Referenz, nicht als offene Grenze.
- [x] `make verify-clean` und CI grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.github/workflows/feature-cli.yml` | update | CLI 0.90.0, `up` und `exec` statt `build` und `docker run` |
| `.github/workflows/publish.yml` | update | CLI 0.90.0 im Job `feature`; kein Einfluss auf Images |
| `docs/user/devcontainer.md`, `README.md` | update | erprobter Weg, Ursache des früheren Hängens |

## 5. Trigger

Diagnose abgeschlossen (siehe Ausgangslage); Freigabe des Nutzers („ja“) am 2026-11-11.

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- **Veröffentlichungswerkzeug:** Die CLI in `publish.yml` zu ändern berührt den Weg, über den Features veröffentlicht werden. Deshalb nur mit Trockenlauf-Beleg; eine Veröffentlichung gehört nicht zu diesem Slice und braucht das Ja beider Nutzer.
- Eine neuere CLI kann das Paketformat oder die Ausgabe verändern; der Trockenlauf vergleicht die gepackte Struktur (Datei- und Sammlungs-Metadaten) mit der bisherigen.
- Das Docker auf dem Runner wechselt mit dem Runner-Image; ein erneutes Hängen wäre ein neuer Befund.

## 8. Closure-Notiz (nach `done/`)

**Abgeschlossen:** 2026-11-11 (PR #115, w2). Der PR wurde vor dem Gegenlesen durch w1 gemergt; w1 hat den Diff danach gelesen und keinen Befund gemeldet. Für Änderungen an Workflows gilt künftig: Gegenlesen vor dem Merge.

- **Funktioniert:** `feature-cli` und der Job `feature` in `publish.yml` nutzen die `devcontainer`-CLI 0.90.0 (Variable `CLI_VERSION`, Kommentar zum Grund des Pins). `feature-cli` erprobt wieder den Weg der Nutzer: `devcontainer up` und `devcontainer exec`. Fall A lief grün auf dem PR (21 s), Fall B und der Trockenlauf von `publish` per Dispatch auf dem Branch (Läufe 43002260929 und 43002265534); beide Fälle melden `bedrock-eu-check 1.0.0 (5f1b35cca776)`. Der Trockenlauf mit `ref=feature-v1.0.1` packt dieselbe Struktur wie mit 0.72.0 (Lauf 42996795891): `devcontainer-collection.json` 1160 Byte und `devcontainer-feature-bedrock-eu-check.tgz` 6144 Byte. Es wurde nichts veröffentlicht, die Tags blieben `v1.0.0`, `v1.1.0` und `feature-v1.0.1`. `devcontainer.md` und README nennen die erprobte Version und den Hinweis zur alten CLI.
- **Anders als geplant:** nichts Wesentliches. Der Plan hatte `devcontainer up` mit Docker 29.4.2 auf dem Runner als Ausgangslage; Das Hängen tritt bei der Kombination aus CLI 0.72.0 und Docker 29.4.2 auf; mit neueren CLI-Versionen läuft es auf demselben Runner durch (Diagnose, Läufe 42999455137 und 42999946761). Eine Ursachenanalyse im Code der CLI gab es nicht. Der Slice entstand aus der Nutzerfrage, ob die Versionen aktuell sind.
- **Offen:** Das veröffentlichte Feature 1.0.1 wurde mit CLI 0.72.0 veröffentlicht (Lauf 42996881739); der Inhalt ist davon unabhängig (gleiche Struktur im Trockenlauf), die nächste Feature-Veröffentlichung nutzt die CLI 0.90.0 und braucht das Ja beider Nutzer. Die gepinnten Versionen (CLI, Actions) werden nicht automatisch auf Aktualität geprüft.
- **Steering-Loop:** Beim Pinnen einer Werkzeugversion das Alter gegen das aktuelle Release prüfen und den Grund des Pins im Workflow nennen; ein Pin, der „einmal funktionierte“, ist eine versteckte Altlast. Ein Hänger auf dem Runner wird mit einer Diagnose-Matrix (Version, Eingaben) eingegrenzt, statt umgangen.
- **Folge-Slices:** keine; ein regelmäßiger Aktualitätsprüfer für gepinnte Versionen ist eine mögliche spätere Entscheidung der Nutzer.

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Anhebung einer gepinnten Werkzeugversion in bestehenden Workflows.
