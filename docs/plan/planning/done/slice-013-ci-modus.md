# Slice slice-013: CI-Modus `--ci` und Beispiel-Workflow

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-04-output-ci.

**Bezug:** [`LH-MOD-005`](../../../../spec/lastenheft.md#lh-mod-005--ci-modus), [`LH-FA-EXIT-001`](../../../../spec/lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung), [`LH-FA-EXIT-002`](../../../../spec/lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung), [`LH-FA-EXIT-003`](../../../../spec/lastenheft.md#lh-fa-exit-003--tool-fehler), [`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`--ci` sorgt für eine für Pipelines taugliche, deterministische Ausgabe mit
den definierten Exit-Codes, und ein dokumentierter Beispiel-Workflow zeigt
den Einsatz mit Code-Scanning.

## 2. Verträge

- **`--ci`** ist eine Option aller Modi. Wirkung: stdout enthält ausschließlich das Ergebnis im gewählten Format (Default `console`), nie Hinweise oder Fortschritt; Hinweise und Fehler gehen auf stderr; keine Farben und keine interaktiven Elemente; stabile Sortierung (wie sonst auch). Die **Exit-Codes bleiben unverändert**: `0` ohne FAIL, `1` mit mindestens einem FAIL, `2` bei Tool-Fehler ([`SPEC-030`](../../../../spec/spezifikation.md#spec-030--exit-codes)).
- **Kombination mit `--format`:** `--ci --format sarif|json` ist der Normalfall; ohne `--format` bleibt das Konsolenformat.
- **`--fail-on warn`** und Baseline-Dateien sind Folgepunkte (außerhalb dieser Welle).
- **Beispiel-Workflow** in `docs/user/ci.md`: GitHub-Actions-Job, der das Werkzeug in einem Schritt ausführt (`all --ci --format sarif`, Ausgabe in eine Datei umgeleitet) und das SARIF zum Code-Scanning hochlädt. Hinweis, dass der Modus `aws` ein installiertes `aws`-Binary und Zugangsdaten braucht (sonst Exit `2` bzw. FAIL) und dass für Pull Requests aus Forks meist nur `local` und `terraform` sinnvoll sind. Actions werden im Beispiel mit Versionen gepinnt und der Hinweis ergänzt, sie per Commit-SHA zu pinnen.

## 3. Definition of Done

- [ ] CLI: `--ci` in allen Modi; Tests: Exit-Codes unverändert, stdout parsebar (mit `--format json|sarif`), stderr enthält Fehler, kein Secret in stdout/stderr.
- [ ] Tool-Fehler im CI-Modus: bereits ermittelte Findings erscheinen im gewählten Format auf stdout, die Fehlermeldung auf stderr, Exit `2`.
- [ ] `docs/user/ci.md` mit Beispiel-Workflow (nur Dokumentation, kein Hochladen durch das Werkzeug), Verweis auf die Exit-Codes und die Rechte für `aws` ([`aws-permissions.md`](../../../user/aws-permissions.md)).
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; Welle-4-Abschluss (M4).

## 4. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/adapter/driving/cli/` (`--ci`) | update | w2 |
| `docs/user/ci.md` | neu | w2 |

## 5. Trigger

[`slice-011`](slice-011-json-reporter.md) gemerged; für das Beispiel mit SARIF zusätzlich [`slice-012`](slice-012-sarif-reporter.md).

## 6. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 7. Risiken und offene Punkte

- Der Beispiel-Workflow lässt sich offline nicht ausführen; er wird sorgfältig gegen die dokumentierten Aufrufe geprüft, aber nicht gegen GitHub getestet.
- Ein Selbsttest des Werkzeugs gegen das eigene Repository ([`LH-QA-005`](../../../../spec/lastenheft.md#lh-qa-005--ci-prüfung) Self-Check) ist ein Folgepunkt: Die Testdaten mit absichtlich falscher Konfiguration würden einen Lauf über das ganze Repository färben, daher braucht es zuerst eine konfigurierbare Ausschlussliste.
- `--output FILE` (statt Umleitung) ist ein Folgepunkt.

## 8. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 9. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**.
