# Review-Report: PR #107 (slice-024, Code) — 2026-11-11

**Review-Art:** Code — geprüft werden Workflow `feature-cli`, die Änderung an `publish.yml` (Eingabe `ref`), `devcontainer-feature.json` und die Doku-Aussagen gegen den freigegebenen Plan (Plan-Review: `2026-11-11-pr106-slice-024-plan.md`) und gegen die tatsächlichen Läufe auf GitHub. Es wurde kein Workflow von mir gestartet; die Läufe wurden nur gelesen.

**Gegenstand:** Pull Request #107 (`feat/slice-024-feature-cli`, Autor w2), Stand bei Beginn des Reviews `abdf5c2`; Diff gegen `main` (`git diff origin/main...`): `.github/workflows/feature-cli.yml` (neu), `.github/workflows/publish.yml`, `devcontainer-feature/src/bedrock-eu-check/devcontainer-feature.json`, `README.md`, `docs/user/devcontainer.md`.

**Skill:** keiner (im Repository existiert keine Reviewer-Skill-Datei).
**Modell:** claude-sonnet-5-5 (Agent w1) · **Datum:** 2026-11-11

**Eingangs-Kontext:**

- Slice-Plan `slice-024` (`docs/plan/planning/in-progress/slice-024-devcontainer-feature-erproben.md`), einschließlich der Veröffentlichungsgrenze (Ja beider Nutzer)
- [`LH-TECH-005`](../../spec/lastenheft.md#lh-tech-005--devcontainer-integration)
- Läufe auf GitHub: Run 42986608727 (Fall A und Fall B, `workflow_dispatch`, Commit `9cad7af`, success) und Run 42983578810 (`devcontainer up`, cancelled)
- [`AGENTS.md`](../../AGENTS.md) (Hard Rules)

---

## Findings

### F-1 — `devcontainer up` hängt, Ursache ungeklärt

- `kategorie`: MEDIUM
- `quelle`: Plan slice-024 (Ziel: Feature mit der echten CLI erproben)
- `pfad`: `.github/workflows/feature-cli.yml` (Kommentar vor `devcontainer build`), `docs/user/devcontainer.md` (Abschnitt „Stand“)
- `befund`: Der Lauf mit `devcontainer up` blieb nach „Container started“ bis zum Job-Timeout hängen (Run 42983578810, `cancelled`), obwohl das Feature im Build installiert war. Der Workflow weicht auf `devcontainer build` plus `docker run` aus; ob das Verhalten vom Feature, von der CLI-Version 0.72.0 oder vom Runner stammt, ist nicht ermittelt. Das Feature ist damit nur bis zum Bau des Images erprobt, nicht im von der CLI hochgefahrenen Container.
- `verifizierbar`: nein — es gibt keinen Lauf, der Feature und Runner als Ursache trennt.

### F-2 — Neuer `ref`-Pfad in `publish.yml` wird von keinem Lauf ausgeführt

- `kategorie`: MEDIUM
- `quelle`: Plan slice-024 (Veröffentlichungsweg, Variante b)
- `pfad`: `.github/workflows/publish.yml` (Eingabe `ref`, Checkout und Schritt „Version und Commit gegen den Tag prüfen“)
- `befund`: `publish` hat keinen Trockenlauf; der erste Lauf, der den `ref`-Pfad ausführt, ist die tatsächliche Veröffentlichung (nicht rückholbar, verschiebt die Tags `1`, `1.0`, `latest`). Die Änderung wird weder durch den Pfadfilter von `feature-cli` noch durch ein anderes Gate erfasst. Die Shell-Logik ist beim Lesen plausibel (leeres `ref` → `v<version>`; `ref` muss `feature-v<version>` entsprechen; Commit gegen den Tag; Feature-Version gleich `version`), aber nicht ausgeführt.
- `verifizierbar`: nein — ohne Trockenlauf gibt es keinen Gate-Lauf, der es bestätigt.

### F-3 — `ref` wirkt nur im Feature-Job

- `kategorie`: LOW
- `quelle`: Maintainability
- `pfad`: `.github/workflows/publish.yml` (Job `images`, Checkout `refs/tags/v${{ inputs.version }}`)
- `befund`: Wird `ref` zusammen mit `images: true` angegeben, checkt der Job `images` weiterhin `v<version>` aus, wobei `version` dann die Feature-Version ist; die Eingabebeschreibung sagt „Nur für das Feature“, der Workflow weist eine Kombination mit `images` aber nicht ab.
- `verifizierbar`: nein — Verhalten nur im Lauf sichtbar.

### F-4 — Fall A setzt `version` ausdrücklich und prüft den Default nicht

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `.github/workflows/feature-cli.yml` (Fall A und B, `"version": "1.0.0"`)
- `befund`: Beide Fälle übergeben `version` explizit; der Default der Option in `devcontainer-feature.json` (dessen Beschreibung in diesem PR geändert wurde) wird dadurch nicht über die CLI geprüft. `feature-test` deckt den Default lokal ab, soweit nach der Änderung beurteilbar.
- `verifizierbar`: ja — ein Fall ohne `version` würde es bestätigen.

### F-5 — Erwartete Ausgabe fest an `v1.0.0` gebunden

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `.github/workflows/feature-cli.yml` (`EXPECTED`)
- `befund`: `EXPECTED` enthält Version und Commit des Release v1.0.0. Die Prüfung bleibt gültig, solange der Default des Features auf 1.0.0 zeigt (Plan: entkoppelt von der Feature-Version); ändert sich der Default, schlägt sie fehl.
- `verifizierbar`: ja — Workflow-Lauf.

## Negativbefunde

- geprüft, ohne Befund: Rechte — Workflow-Ebene `permissions: {}`, beide Jobs nur `contents: read`; kein `packages: write`, kein Push, kein Tag. Der Workflow veröffentlicht nichts.
- geprüft, ohne Befund: Claims im PR gegen GitHub — Run 42986608727 ist `success` mit Fall A und Fall B, beide Ausgaben `bedrock-eu-check 1.0.0 (5f1b35cca776)`; Run 42983578810 ist `cancelled`. Beides stimmt mit der PR-Beschreibung und der Doku überein.
- geprüft, ohne Befund: Plan-Review-Einarbeitung — Fall B nur bei `workflow_dispatch` (F-2 des Plan-Reviews), Fall A auch bei Änderungen am Feature, Ausgabeformat mit Präfix `bedrock-eu-check` (F-4 des Plan-Reviews).
- geprüft, ohne Befund: Einschleusung — `inputs.ref`, `inputs.version` und `inputs.commit` gelangen nur über `env` in die Shell (`REF_INPUT`, `VERSION`, `COMMIT_INPUT`), nicht per `${{ }}` in `run:`; der Checkout nutzt `ref` ausschließlich als Tag-Ref.
- geprüft, ohne Befund: Ehrlichkeit der Doku — `devcontainer.md` und `README.md` nennen `devcontainer up` ausdrücklich als nicht erprobt und die Ursache als ungeklärt; die Aussage „nicht mit der Devcontainer-CLI erprobt“ wurde entsprechend ersetzt.
- geprüft, ohne Befund: Pinning — Actions per SHA, CLI `@devcontainers/cli@0.72.0` wie in `publish.yml`.
- geprüft, ohne Befund: Veröffentlichungsgrenze — die Feature-Version in `devcontainer-feature.json` ist 1.0.1 (noch nicht veröffentlicht); der PR legt keinen Tag an und startet kein `publish`.
- nicht geprüft: eigene Ausführung von `feature-cli`, `publish` oder `release-create` (target `feature`); das Verhalten von `devcontainer up` auf anderen Runnern oder lokal; Wirkung auf VS Code Dev Containers (nutzt eigene Implementierung); Ergebnis von `make verify-clean` und CI des PR zum Zeitpunkt dieses Reports (CI lief noch).

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 1 |
| INFO | 2 |

## Verdikt

**Merge-blockierend:** nein — kein HIGH. Die beiden MEDIUM-Findings blockieren den Merge dieses PR nicht, weil der Merge selbst nichts veröffentlicht, F-1 in Doku und Workflow offen benannt ist und F-2 erst bei der Veröffentlichung wirkt. Diese Abweichung von „MEDIUM vor Merge klären“ ist hier bewusst begründet: Beide Punkte müssen **vor der Entscheidungsvorlage und dem Ja der Nutzer zur Feature-Version 1.0.1** geklärt oder ausdrücklich als bekannte Grenze in die Vorlage übernommen werden.

**Übergabe:** Findings gehen an den Autor. Der Report ersetzt keine Verifikation — DoD und Spezifikations-Konformität prüft der Verifier separat.
