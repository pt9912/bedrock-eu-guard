# Welle welle-06-release-haertung: Ergebnisse

**Abgeschlossen:** 2026-11-11. **Status:** done; Feature 1.0.1 **veröffentlicht**.
**Meilenstein:** M6 — Release-Prozess erprobt **erreicht** (2026-11-11).

## Closure-Trigger

| Kriterium | Stand |
|---|---|
| Alle Slices done (023, 024, 025) | erfüllt |
| `make verify-clean` und CI grün | erfüllt (Stand `5274f3dce55f`) |
| `release-artifacts` auf GitHub gelaufen, Prüfsummen == Release | erfüllt (slice-023, Run 42970275950, bitgleich zum Release 1.0.0) |
| Feature mit der echten `devcontainer`-CLI installiert, Mängel behoben | erfüllt für `devcontainer build` mit `docker run` (slice-024); `devcontainer up` mit 0.72.0 hängt (Ursache belegt, siehe „Offen“) |
| `release-create` im Trockenlauf erprobt, Tag und Release nur auf Eingabe | erfüllt (slice-025) und seit 2026-11-11 auch **scharf** durchlaufen (siehe unten) |
| Jede Veröffentlichung nur nach dem Ja beider Nutzer mit Version, Commit und Zielen | erfüllt (Wortlaut siehe Welle 7, Entscheidung B) |

## Veröffentlicht

Feature 1.0.1 am 2026-11-11 auf Commit `5274f3dce55f4423e2937978d3c4d05c888a13cb` (Entscheidung B der gemeinsamen Vorlage):

| Schritt | Beleg |
|---|---|
| Tag `feature-v1.0.1` (annotiert, kein GitHub-Release) | `release-create` mit `target=feature`, scharf: Run 42993689182 |
| `publish` mit `ref=feature-v1.0.1`, Trockenlauf | Run 42996795891 (erster Lauf des `ref`-Pfads) |
| `publish`, scharf | Run 42996881739; Tags `1`, `1.0`, `1.0.1`, `latest` auf `sha256:34aa4f35d7704d07c21e4955ffc343091f679df267da91e6472139190e47f37a`; `1.0.0` unverändert (`sha256:cd61b5bc…`) |
| Nachprüfung | `feature-cli` Fall A und B, Run 42996952937: Ausgabe `bedrock-eu-check 1.0.0 (5f1b35cca776)` (Standard der Option `version` bleibt 1.0.0) |

w1 hat die Tags, die Digests, die OCI-Metadaten (Feature-Version 1.0.1, korrigierte Option-Beschreibung) und die Läufe unabhängig gegen GitHub und `ghcr.io` geprüft.

## Was die Welle gebracht hat

- Der Workflow `release-artifacts` ist erprobt; seine Binaries sind bitgleich zum Release.
- Der Workflow `release-create` legt Tag und Release an, mit Trockenlauf als Standard; er ist für `v1.1.0` und `feature-v1.0.1` scharf gelaufen.
- Der Workflow `publish` hat einen Trockenlauf (`dry_run`, Standard) und eine Eingabe `ref` für Feature-Tags.
- Der Workflow `feature-cli` installiert das Feature mit der echten `devcontainer`-CLI (Version 0.72.0).
- Die veraltete Option-Beschreibung des Features ist korrigiert (Feature 1.0.1).

## Offen

1. **`devcontainer up` hängt mit der CLI 0.72.0** auf dem GitHub-Runner nach „Container started“ (Run 42983578810, Job-Timeout nach 20 Minuten). Die Diagnose von w2 (Läufe 42999455137 und 42999946761, Docker 29.4.2 auf dem Runner) belegt: Mit 0.72.0 läuft `up` in den Timeout, mit 0.86.1 und 0.90.0 läuft es durch; das Feature ist nicht die Ursache. Die Welle hat deshalb bei `devcontainer build` mit `docker run` erprobt. Die Umstellung auf CLI 0.90.0 mit `up` und `exec` ist als Wartungsslice (slice-034, ohne Welle) geplant.
2. **Umgebung `release`:** Sie ist mit dem Prüfer `pt9912` (nur Branch `main`) angelegt. Weil alle Beteiligten dasselbe GitHub-Konto nutzen, erzwingt sie das „Ja beider Nutzer“ nicht und erlaubt Selbstfreigabe; das Ja bleibt eine Wortlaut-Freigabe im Dialog. Das Token der Agenten darf Deployments nicht freigeben (HTTP 403); die Freigabe erteilte ein Mensch in der Weboberfläche.
3. Der Beispiel-Workflow für GitHub Code Scanning und die Beispiel-IAM-Policy sind weiter nicht auf externen Plattformen erprobt (nicht Teil dieser Welle).

## Fehlläufe und Abweichungen (ehrlich)

- `release-create`: Der erste Trockenlauf (Run 42972041118) scheiterte, weil `actions/checkout` keine verkürzte SHA als `ref` akzeptiert; behoben in PR #93.
- `devcontainer up` hing mit der CLI 0.72.0 (siehe „Offen“); der Workflow wich auf `devcontainer build` aus, statt das Hängen zu verschweigen. Die Ursache wurde nach der Veröffentlichung mit einer Diagnose-Matrix (PR #111, geschlossen) eingegrenzt.
- Der Trockenlauf von `publish` für die Feature-Version konnte erst laufen, nachdem der Tag existierte; der Beleg ist deshalb erst bei der Veröffentlichung entstanden (Review-Finding F-2 zu PR #107).
- Der Reviewer-Skill `.harness/skills/reviewer.md` existiert im Repository nicht; die Review-Reports tragen „Skill: keiner“. Die Anlage ist als eigener Doku-Slice abgestimmt, aber noch nicht umgesetzt.

## Steering-Loop

Ein außenwirksamer Ablauf wird erst im Trockenlauf auf der echten Plattform erprobt, dann scharf. Wo der Trockenlauf einen Pfad nicht erreichen kann (hier `ref` ohne Tag), steht die Grenze vorher in Plan und Entscheidungsvorlage, nicht erst im Nachhinein. Und eine Schutzeinstellung (hier die Umgebung `release`) ersetzt die Freigabe nur, wenn sie zwischen den Beteiligten unterscheiden kann; sonst bleibt die dokumentierte Wortlaut-Freigabe die tragende Kontrolle.
