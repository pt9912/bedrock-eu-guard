# Roadmap

**Status:** Aktiv. **Letzte Änderung:** 2026-06-24.

**Format-Regel:** Die Roadmap ist eine Reihenfolge von **Wellen**, keine
Reihenfolge von Terminen. Die Wellen bilden die Versions-Stufen
`LH-RM-*` des Lastenhefts ab.

---

## Aktuelle Welle

Keine aktive Welle. welle-01-lokaler-env-check ist abgeschlossen
(2026-11-10, siehe [`../done/welle-01-results.md`](../done/welle-01-results.md));
als Nächstes steht [`welle-02-config-scanner`](../welle-02-config-scanner.md) an (Plan-Entwurf liegt vor, Start nach Abhängigkeits-ADR bzw. mit slice-004).

## Nächste Wellen

| Welle | Version (LH-RM) | Trigger | Wichtigste Slices | Aufwand |
|---|---|---|---|---|
| welle-02-config-scanner | v0.2 (`RM-002`) | Welle 1 done | Claude-Code-, Devcontainer-, Docker-, Terraform-Scanner (`CLAUDE/DEVCON/DOCKER/TF-*`) | L |
| welle-03-aws-checks | v0.3 (`RM-003`) | Welle 2 done | AWS-STS-, Bedrock-, IAM-Prüfung (`AWS/BEDROCK/IAM-*`, `AUTH-001`); `awscli`-Adapter + Mock | L |
| welle-04-output-ci | v0.4 (`RM-004`) | Welle 3 done | JSON-/SARIF-Ausgabe, CI-Modus (`REP-002/003`, `MOD-005`) | M |
| welle-05-release | v1.0 (`RM-005`) | Welle 4 done | Doku, Docker-Image, Devcontainer-Feature, Release-Artefakte (`QA-004`, `TECH-004/005`) | M |

## Meilensteine

| Meilenstein | Welle(n) | Trigger | Status |
|---|---|---|---|
| M0 — Harness bereit | — | Greenfield-Bootstrap abgeschlossen | erreicht (2026-06-24) |
| M1 — erstes lauffähiges Tool | welle-01 | `bedrock-eu-check local` liefert PASS/WARN/FAIL + Exit-Codes | erreicht (2026-11-10) |
| M2 — vollständiger statischer Scan | welle-02 | Config/Container/Terraform abgedeckt | offen |
| M3 — AWS-Live-Checks | welle-03 | STS/Bedrock/IAM read-only | offen |
| M4 — CI-fähig | welle-04 | SARIF + `--ci` + Exit-Code 1 bei FAIL | offen |
| M5 — Release 1.0 | welle-05 | Docker-Image + Doku + Tests grün | offen |

## Abhängigkeitsgraph

```mermaid
flowchart LR
    W1[welle-01<br/>local env] --> W2[welle-02<br/>config scanner]
    W2 --> W3[welle-03<br/>aws checks]
    W3 --> W4[welle-04<br/>output + ci]
    W4 --> W5[welle-05<br/>release 1.0]
```

## Abgeschlossene Wellen

- welle-01-lokaler-env-check, abgeschlossen 2026-11-10, siehe [`done/welle-01-results.md`](../done/welle-01-results.md).

## Historische Trigger-Verschiebungen

(noch keine)
