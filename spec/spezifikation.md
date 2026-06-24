# Spezifikation — Bedrock EU Configuration Checker (`bedrock-eu-check`)

**Status:** Aktiv (Phase 2 — Outline; Detail-`SPEC-*` reifen mit den Slices). **Letzte Änderung:** 2026-06-24.

**Bezug zum Lastenheft:** Diese Spezifikation präzisiert die in
[`lastenheft.md`](lastenheft.md) formulierten Anforderungen (`LH-*`).
**Bei Konflikt gewinnt das Lastenheft.** ID-Schema `SPEC-*` siehe
[`../harness/conventions.md`](../harness/conventions.md) [`MR-002`](../harness/conventions.md#mr-002--id-schema-kanon-präfix-lh--mit-bereichscodes).

---

## 1. Algorithmen und Datenflüsse

### SPEC-001 — Severity- und Aggregationsmodell

Jede Regel liefert ein **Finding** mit Schweregrad aus `{PASS, WARN, FAIL}`.
Aggregation über einen Lauf:

1. Findings aller aktivierten Regeln einsammeln.
2. Höchster Schweregrad bestimmt den Exit-Code (`SPEC-030`).
3. Reihenfolge der Ausgabe = Modus-Reihenfolge, innerhalb eines Modus
   stabil nach Regel-ID.

Präzisiert [`LH-ZIEL-005`](lastenheft.md#lh-ziel-005--entwicklerfreundliche-ausgabe), [`LH-FA-REP-001`](lastenheft.md#lh-fa-rep-001--konsolenausgabe).

### SPEC-002 — Regel-Modell

Eine Regel ist ein deklaratives Objekt:

```text
Rule {
  id:        "LH-FA-LENV-001"   # Lastenheft-Anker (Pflicht)
  area:      LENV | CLAUDE | DEVCON | DOCKER | AWS | BEDROCK | IAM | TF | AUTH | REP
  mode:      local | aws | terraform          # in welchem Modus aktiv
  evaluate:  (ctx, config) -> Finding[]        # rein, ohne Seiteneffekt auf Ressourcen
  offline:   bool                              # true = ohne Netz/AWS lauffähig
}
```

Erweiterbarkeit ([`LH-NF-006`](lastenheft.md#lh-nf-006--erweiterbarkeit)): neue Regeln werden registriert, ohne
bestehende zu ändern (Registry-Pattern, siehe
[`architecture.md`](architecture.md) [`ARC-004`](architecture.md#2-schichten-und-constraints)).

### SPEC-010 — Klassifikation von Modell-/Profil-IDs

Eingabe: eine Modell- oder Inference-Profil-ID. Ausgabe: `allowed` /
`blocked` / `unclassified`. Schritte:

1. Trifft ein Muster aus `blockedModelPatterns` → **blocked** (`FAIL`).
2. Sonst trifft ein Muster aus `allowedModelPatterns` → **allowed** (`PASS`).
3. Sonst → **unclassified** (`WARN`).

Block hat Vorrang vor Allow. Default-Muster siehe `SPEC-020`.
Präzisiert [`LH-FA-LENV-006`](lastenheft.md#lh-fa-lenv-006--prüfung-von-claude-default-modellvariablen), [`LH-FA-BEDROCK-002`](lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen)/[`LH-FA-BEDROCK-003`](lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen),
[`LH-FA-TF-004`](lastenheft.md#lh-fa-tf-004--nicht-erlaubte-modell-ids-finden).

### SPEC-011 — Regionsklassifikation

Eine Region ist erlaubt, gdw. sie in `allowedRegions` enthalten ist.
Fehlende Region → `FAIL` (lokaler Env-Check) bzw. `WARN` (dynamische
Terraform-Region, nicht auflösbar). Präzisiert [`LH-FA-LENV-001`](lastenheft.md#lh-fa-lenv-001--prüfung-von-aws_region)/[`LH-FA-LENV-002`](lastenheft.md#lh-fa-lenv-002--prüfung-von-aws_default_region),
[`LH-FA-TF-002`](lastenheft.md#lh-fa-tf-002--aws-provider-region-prüfen).

## 2. Datenstrukturen und Schemas

### SPEC-020 — Konfigurationsschema `bedrock-eu-check.yml`

```yaml
allowedRegions:        [eu-central-1]              # Default; LH-OFF-001
allowedAccounts:       []                          # leer = nicht geprüft
allowedModelPatterns:  ['^eu\.', '^anthropic\.claude-sonnet-4-6$', '^anthropic\.claude-haiku-4-5$']
blockedModelPatterns:  ['^us\.', '^global\.', 'us-east-1', 'us-west-2']
allowedProfiles:       ['eu.*', 'eu-central-1.*']
mantle:                warn                         # warn | allow | block; LH-OFF-003
```

Alle Felder optional; fehlende Felder fallen auf die Defaults zurück
([`LH-NF-007`](lastenheft.md#lh-nf-007--konfigurierbarkeit)). Block-Muster haben Vorrang (`SPEC-010`).

### SPEC-021 — Finding-Schema (JSON, `--format json`)

```json
{
  "status":   "PASS | WARN | FAIL",
  "id":       "LH-FA-LENV-001",
  "message":  "AWS_REGION=eu-central-1",
  "severity": "info | warning | error",
  "file":     "path or null",
  "line":     0,
  "recommendation": "string or null"
}
```

Präzisiert [`LH-FA-REP-002`](lastenheft.md#lh-fa-rep-002--json-ausgabe). **Secrets erscheinen nie als Wert**
(`message` trägt höchstens „set"/„not set", [`LH-NF-004`](lastenheft.md#lh-nf-004--keine-secret-ausgabe)).

### SPEC-022 — SARIF-Abbildung (`--format sarif`)

`PASS`→ kein Result oder `level: none`; `WARN`→ `level: warning`;
`FAIL`→ `level: error`. `ruleId` = `LH-*`-ID. Optional
([`LH-PRI-KANN-001`](lastenheft.md#lh-pri-kann-001)).

## 3. Defaults und Konstanten

| Name | Wert | Begründung |
|---|---|---|
| `DEFAULT_ALLOWED_REGION` | `eu-central-1` | Vorschlag [`LH-OFF-001`](lastenheft.md#lh-off-001--exakte-erlaubte-aws-regionen) |
| `DEFAULT_SONNET` | `anthropic.claude-sonnet-4-6` | [`LH-OFF-002`](lastenheft.md#lh-off-002--erlaubte-claude-modelle) |
| `DEFAULT_HAIKU` | `anthropic.claude-haiku-4-5` | [`LH-OFF-002`](lastenheft.md#lh-off-002--erlaubte-claude-modelle) |
| `MANTLE_POLICY` | `warn` | [`LH-OFF-003`](lastenheft.md#lh-off-003--mantle-strategie), [`LH-RISK-002`](lastenheft.md#lh-risk-002--mantle-verfügbarkeit-nicht-eindeutig-prüfbar) |
| `CONFIG_FILENAME` | `bedrock-eu-check.yml` | [`LH-TECH-003`](lastenheft.md#lh-tech-003--konfigurationsdatei) |

> Werte, die die ADR-Schärfung betreffen (z. B. die endgültige
> Regionsliste), werden über das `**Schärft:**`-Feld der jeweiligen ADR
> *aufwärts* deklariert — kein ADR-Rückzeiger in dieser Datei.

## 4. Fehler-Codes und Exit-Codes

### SPEC-030 — Exit-Codes

| Exit | Bedingung | Lastenheft |
|---|---|---|
| `0` | kein `FAIL`-Finding | [`LH-FA-EXIT-001`](lastenheft.md#lh-fa-exit-001--erfolgreiche-prüfung) |
| `1` | ≥ 1 `FAIL`-Finding | [`LH-FA-EXIT-002`](lastenheft.md#lh-fa-exit-002--fehlgeschlagene-prüfung) |
| `2` | Tool-Fehler (Config unlesbar, interner Abbruch) | [`LH-FA-EXIT-003`](lastenheft.md#lh-fa-exit-003--tool-fehler) |

Fail-closed: Kann ein offline-Check nicht ausgeführt werden, ist das
Exit `2`, nicht stilles `0`.

## 5. Metriken und Tracing-Felder

Für ein CLI-Werkzeug ohne Service-Laufzeit nicht anwendbar in v1.
`--ci` erzeugt eine kompakte, parsebare Zeilen-Ausgabe statt OTel-Spans
([`LH-MOD-005`](lastenheft.md#lh-mod-005--ci-modus)). Reserviert für spätere Versionen.

## 6. Externe Verträge

| System | Version | Vertrag |
|---|---|---|
| AWS CLI | v2 (lokal konfiguriert) | `sts get-caller-identity`, `bedrock list-inference-profiles` (read-only, [`LH-SEC-003`](lastenheft.md#lh-sec-003--read-only-aws-zugriff)) |
| Terraform | `.tf`-HCL, optional `plan -json` | statischer Scan; dynamische Ausdrücke → `WARN` ([`LH-RISK-003`](lastenheft.md#lh-risk-003--false-positives-bei-terraform)) |
| Claude Code | `~/.claude/settings.json`, `.claude/settings*.json` | Env-Block-Auswertung (`LH-FA-CLAUDE-*`) |

## 7. Historie

| Datum | Änderung | ADR |
|---|---|---|
| 2026-06-24 | Initiale Outline (Phase 2) beim Harness-Bootstrap | — |
