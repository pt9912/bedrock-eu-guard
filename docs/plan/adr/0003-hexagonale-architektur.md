# ADR-0003: Hexagonale Architektur (Ports & Adapters)

**Status:** Accepted

**Datum:** 2026-06-24

**Autor:** Projektteam bedrock-eu-guard

**Bezug:** [`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit) (Offline-Fähigkeit), [`LH-NF-006`](../../../spec/lastenheft.md#lh-nf-006--erweiterbarkeit) (Erweiterbarkeit), [`LH-QA-002`](../../../spec/lastenheft.md#lh-qa-002--integration-tests) (mockbare AWS-Adapter)

**Schärft:** [`spec/architecture.md` §2 Schichten und Constraints](../../../spec/architecture.md) — diese ADR macht die dort beschriebene Schichtung und die Constraints (einzige AWS-Tür, importfreier Kern) verbindlich.

---

## Kontext

`bedrock-eu-check` hat einen **offline-fähigen Prüfkern** (lokale +
Terraform-Regeln ohne Netz, [`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)), eine **eng begrenzte
AWS-Berührung** (nur lesend, [`LH-SEC-003`](../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)), **mehrere Eingabequellen**
(Env, Dateien, AWS) und **mehrere Ausgabeformate** (console/json/sarif).
Diese Asymmetrie — reiner Kern, wenige Seiteneffekt-Ränder — verlangt eine
Struktur, die den Kern testbar ([`LH-QA-002`](../../../spec/lastenheft.md#lh-qa-002--integration-tests)) und seiteneffektfrei
hält und neue Regeln ohne Architekturbruch zulässt ([`LH-NF-006`](../../../spec/lastenheft.md#lh-nf-006--erweiterbarkeit)).

Als Vorbild dient **d-check** (`pt9912/d-check`, Go) — dasselbe Werkzeug,
das unseren `make doc-check`-Gate stellt. Sein `internal/`-Layout ist ein
ausgereiftes Ports-&-Adapters-Hexagon mit klarer driving/driven-Trennung,
importfreiem `model`-Ring und genau *einer* Netzwerk-Tür.

## Entscheidung

Wir wählen **Hexagonale Architektur (Ports & Adapters)** und übernehmen
die **Ordnerkonvention von d-check** (driving/driven). Der Kern definiert
Ports als Interfaces; Adapter implementieren sie; die CLI ist die einzige
**Composition Root**.

```text
internal/
├── hexagon/
│   ├── core/
│   │   ├── model/      # importfreier Ring: Finding, Severity, Reason, Config, Rule
│   │   ├── rules/      # Regel-Module je Bereich + run.go (Orchestrator) + Registry
│   │   ├── app/        # Use-Cases: checkLocal, checkAws, checkTerraform, checkAll
│   │   └── coretest/   # In-Memory-Doubles: memfs, fakeaws, fakeenv  (LH-QA-002)
│   └── port/
│       └── driven/     # Interfaces: Filesystem, Environment, AWSReader
└── adapter/
    ├── driving/
    │   └── cli/        # Arg-Parsing, Composition Root, Modus-Dispatch, Exit-Code
    └── driven/
        ├── configyaml/ # bedrock-eu-check.yml laden/validieren
        ├── fs/         # Dateisystem (.env, .tf, settings.json, devcontainer, compose, Dockerfile)
        ├── envos/      # Umgebungsvariablen lesen
        ├── awscli/     # read-only AWS: STS / Bedrock / IAM   ← einzige AWS-Tür
        └── report/     # console | json | sarif
```

**Tragende Constraints (als Fitness Function durchsetzbar, sobald Code
existiert):**

1. `core/model` importiert **nichts** Internes (innerster Ring).
2. `core/rules` und `core/app` importieren nur `model` und `port`, nie
   einen Adapter.
3. **Genau ein Adapter (`awscli`) berührt AWS**; jeder Netz-/AWS-Zugriff
   außerhalb davon ist ein Architekturverstoß (spiegelt d-checks
   „einzige Netzwerk-Tür", erzwingt [`LH-NF-005`](../../../spec/lastenheft.md#lh-nf-005--offline-fähigkeit)/`SEC-003`).
4. Der **Reporter erhält nur `Finding`-Objekte** — Secret-Maskierung
   passiert vorher im Kern ([`LH-NF-004`](../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — Pragmatisch geschichtet (CLI→services→adapters) | wenig Zeremonie, schnell | schwache Grenzen; offline/online-Disziplin nur per Konvention, nicht erzwingbar |
| B — Pipes & Filters (Scan-Pipeline) | passt zur linearen Scan-Natur | AWS-Seiteneffekte schlechter isoliert; Mockbarkeit unklarer |
| **C — Hexagonal / Ports & Adapters (gewählt)** | Kern rein & testbar; AWS hinter einem Port; Regeln als Plugins; bewährtes Vorbild (d-check) | mehr Struktur/Boilerplate im Bootstrap |

## Konsequenzen

- Positiv: Der Offline-Kern ist strukturell vom AWS-Rand getrennt — die
  wichtigste Compliance-Eigenschaft des Tools ist *architektonisch*
  abgesichert, nicht nur per Test.
- Positiv: Regeln wachsen in `core/rules/` ohne Adapter-Änderung
  ([`LH-NF-006`](../../../spec/lastenheft.md#lh-nf-006--erweiterbarkeit)); jede Quelle ist über ihren Port mockbar
  ([`LH-QA-002`](../../../spec/lastenheft.md#lh-qa-002--integration-tests)).
- Negativ: Mehr Pakete/Interfaces als ein flaches Skript; Einstieg etwas
  steiler.
- Folgepflicht: Import-Constraints als Fitness Function (`make arch-check`)
  binden, sobald Code existiert; in der Sensors-Tabelle erst dann
  eintragen. Sprachwahl in **ADR-0002** (die Ordnerkonvention setzt eine
  Sprache mit `internal/`-/Paket-Sichtbarkeit voraus).

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| Import-Linter (depguard/import-linter, je Sprache) | `core/model` ohne Internal-Import; nur `awscli` darf AWS-SDK importieren | `make arch-check` (geplant, Welle 1) |

## Re-Evaluierungs-Trigger

Wenn das Tool über das CLI-Prüfwerkzeug hinaus zu einem dauerhaft
laufenden Service erweitert wird (dann zusätzliche driving-Adapter), oder
wenn die in ADR-0002 gewählte Sprache die Ordnerkonvention nicht trägt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-06-24 | Proposed | Architektur-Entscheidung im Bootstrap |
| 2026-06-24 | Accepted | Vom Auftraggeber gewählt, Vorbild [`pt9912/d-check`](https://github.com/pt9912/d-check/tree/main/internal) |
