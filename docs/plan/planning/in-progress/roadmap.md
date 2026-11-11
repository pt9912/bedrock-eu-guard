# Roadmap

**Status:** Aktiv. **Letzte Änderung:** 2026-11-11.

**Format-Regel:** Die Roadmap ist eine Reihenfolge von **Wellen**, keine
Reihenfolge von Terminen. Die Wellen bilden die Versions-Stufen
`LH-RM-*` des Lastenhefts ab.

---

## Aktuelle Welle

(keine; alle geplanten Wellen sind abgeschlossen)

welle-01 bis welle-05 sind abgeschlossen. Release 1.0.0 ist am 2026-11-11 veröffentlicht
(Tag `v1.0.0` auf Commit `5f1b35cca776`, GitHub-Release, Images auf `ghcr.io`, Devcontainer-Feature),
siehe [`../done/welle-01-results.md`](../done/welle-01-results.md),
[`../done/welle-02-results.md`](../done/welle-02-results.md),
[`../done/welle-03-results.md`](../done/welle-03-results.md),
[`../done/welle-04-results.md`](../done/welle-04-results.md) und
[`../done/welle-05-results.md`](../done/welle-05-results.md). Die Pläne liegen unter
[`../done/`](../done/), zuletzt [`../done/welle-05-release.md`](../done/welle-05-release.md).

## Nächste Wellen

Alle drei sind **geplant, noch nicht gestartet**. Versionsnummern nach 1.0 sind projektintern
(das Lastenheft kennt `LH-RM-*` nur bis Version 1.0); jede Veröffentlichung braucht das Ja beider Nutzer.

| Welle | Ziel | Trigger | Wichtigste Slices | Aufwand |
|---|---|---|---|---|
| welle-06-release-haertung | Release-Prozess erproben und härten (Feature 1.0.1, Werkzeug bleibt 1.0.0) | Release 1.0.0 veröffentlicht | 023 `release-artifacts` erproben, 025 `release-create`, 024 Feature mit `devcontainer`-CLI | S |
| welle-07-ci-komfort | Werkzeug 1.1.0: CI-Komfort | Release 1.0.0 veröffentlicht | 026 `--fail-on warn`, 027 `--output FILE`, 028 Zusatzregionen, 029 `d-check`-Prüfziele | M |
| welle-08-erweiterungen | Erweiterungen und Härtung (nach Entscheidung) | je Slice: Entscheidung der Auftraggeber bzw. ADR | 030 `iam-policy.json`, 031 `coverage-gate`, 032 Sigstore/SBOM, 033 macOS/Windows | M bis L |

Pläne: [`../welle-06-release-haertung.md`](../welle-06-release-haertung.md), [`../welle-07-ci-komfort.md`](../welle-07-ci-komfort.md), [`../welle-08-erweiterungen.md`](../welle-08-erweiterungen.md).

## Meilensteine

| Meilenstein | Welle(n) | Trigger | Status |
|---|---|---|---|
| M0 — Harness bereit | — | Greenfield-Bootstrap abgeschlossen | erreicht (2026-06-24) |
| M1 — erstes lauffähiges Tool | welle-01 | `bedrock-eu-check local` liefert PASS/WARN/FAIL + Exit-Codes | erreicht (2026-11-10) |
| M2 — vollständiger statischer Scan | welle-02 | Config/Container/Terraform abgedeckt | erreicht (2026-11-10) |
| M3 — AWS-Live-Checks | welle-03 | STS/Bedrock/IAM read-only | erreicht (2026-11-10) |
| M4 — CI-fähig | welle-04 | SARIF + `--ci` + Exit-Code 1 bei FAIL | erreicht (2026-11-10) |
| M5 — Release 1.0 | welle-05 | Docker-Image + Doku + Tests grün | erreicht (2026-11-11, veröffentlicht) |
| M6 — Release-Prozess erprobt | welle-06 | `release-artifacts` und Feature erprobt, `release-create` vorhanden | offen (geplant) |
| M7 — CI-Komfort (1.1.0) | welle-07 | `--fail-on warn`, `--output FILE`, Zusatzregionen, Gate-Bewertung | offen (geplant) |
| M8 — Erweiterungen | welle-08 | je Slice: Entscheidung der Auftraggeber bzw. ADR | offen (nicht freigegeben) |

## Abhängigkeitsgraph

```mermaid
flowchart LR
    W1[welle-01<br/>local env] --> W2[welle-02<br/>config scanner]
    W2 --> W3[welle-03<br/>aws checks]
    W3 --> W4[welle-04<br/>output + ci]
    W4 --> W5[welle-05<br/>release 1.0]
    W5 --> W6[welle-06<br/>release-haertung]
    W5 --> W7[welle-07<br/>ci-komfort]
    W5 --> W8[welle-08<br/>erweiterungen]
    W6 -.-> W8
```

## Abgeschlossene Wellen

- welle-01-lokaler-env-check, abgeschlossen 2026-11-10, siehe [`done/welle-01-results.md`](../done/welle-01-results.md).
- welle-02-config-scanner, abgeschlossen 2026-11-10, siehe [`done/welle-02-results.md`](../done/welle-02-results.md).
- welle-03-aws-checks, abgeschlossen 2026-11-10, siehe [`done/welle-03-results.md`](../done/welle-03-results.md).
- welle-04-output-ci, abgeschlossen 2026-11-10, siehe [`done/welle-04-results.md`](../done/welle-04-results.md).
- welle-05-release, abgeschlossen 2026-11-11 (Release 1.0.0 veröffentlicht), siehe [`done/welle-05-results.md`](../done/welle-05-results.md).

## Historische Trigger-Verschiebungen

(noch keine)
