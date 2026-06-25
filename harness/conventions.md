# Harness-Konventionen — bedrock-eu-guard

## Purpose

Diese Datei deklariert die *repo-lokalen* Strukturregeln von
**bedrock-eu-guard** gegenüber der adoptierten Harnesskonvention
(Baseline: AI-Harness-Kurs). Sie ist der Default-Ort für:

- **Adaptionen** ggü. der Baseline (mit Begründung und Auflösungs-Trigger).
- **ID-Schema-Deklaration** — welches Präfix-Schema dieses Repo nutzt.
- **Zusatzklassen-Deklarationen** für repo-spezifische Sensors-Bindungen.
- **Modus-Deklarationen** pro Sub-Area (Greenfield / Brownfield / Hybrid).

Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt die
kanonische Quelle (Source Precedence, siehe [`README.md`](README.md)).
Diese Datei ist konformitätsbringend für *Form*-Fragen, nicht autoritativ
über Inhalt.

**Repo-Klasse:** *Tooling/Referenz*. `bedrock-eu-check` ist ein lokal und
in CI laufendes Prüfwerkzeug ohne produktiven Laufzeitpfad. Daraus folgt
die Default-Source-Precedence (keine Regulatorik-Voranstellung wie bei
einem Policy/Compliance-Repo) und eine moderate Hard-Rule-Schärfe — die
domänenkritischen Negativ-Regeln (kein Secret-Klartext, Read-only-AWS)
stehen trotzdem als Safety-Boundaries in [`README.md`](README.md).

## Baseline

- **Konvention:** AI-Harness-Kurs (`pt9912/ai-harness-course`).
- **Stand:** v1.4.0 (Regelwerk + Templates als Release-Assets).
- **Datum der Adoption:** 2026-06-24.

## Adoptierte Konventions-Quellen

- **Extern (Regelwerk, Modul-Bundle):**
  <https://github.com/pt9912/ai-harness-course/releases/download/v1.4.0/lab-regelwerk.zip>
  — self-navigierbares ZIP, interne Verweise auf den Tag gepinnt;
  adoptierter Stand **v1.4.0**. Für harte Reproduzierbarkeit den
  Release/Tag pinnen (nicht `main`).
- **In-Repo (verkörperte Form):** die co-located `*.template.md` der
  **wiederkehrenden** Artefakte (ADR, Slice, Welle, Carveout, Review-Report)
  — beim Anlegen kopieren; **autoritative** Autorenquelle. Singleton-Skelette
  (AGENTS, conventions, README, spec, roadmap, Index-READMEs) sind bewusst
  **nicht** co-located ([`MR-003`](#mr-003--template-verkörperung-nur-für-wiederkehrende-skelette)).
  Dazu die Gate-Baseline (`d-check.mk`, `.d-check.yml`, `Makefile`);
  `d-check.mk` ist mit `d-check --print-mk` (v0.29.0) erzeugt, Templates aus
  `lab-templates.zip` v1.4.0.
- **Lokale Lese-/Adoptions-Form (Cache, ephemer):** das **Regelwerk** wird
  zum Lesen nach `.harness/cache/<tag>/regelwerk/` entpackt (`<tag>` =
  §Baseline-`**Stand:**`, derzeit `v1.4.0`); `tools/harness/fetch-baseline-cache.sh`
  materialisiert es (Tag aus der §Baseline-Zeile oder als Argument). Dasselbe
  Skript legt `lab-templates.zip` nach `.harness/cache/<tag>/templates/` —
  **nur** als Adoptions-/Drift-Audit-Staging für die Singleton-Skelette
  ([`MR-003`](#mr-003--template-verkörperung-nur-für-wiederkehrende-skelette)),
  nicht als Autorenquelle. Der Cache ist gitignored und aus `make doc-check`
  ausgenommen — `.d-check.yml` `scan.ignore` **und** `ids.scope.ignore` —, da er
  externes Derivat-/Fremd-ID-Material trägt (gleiche Klasse wie die
  `**/*.template.md`- und `.tmp/**`-Ausnahme).

Diese Datei dupliziert keinen Baseline-Text — sie verweist. Bei Konflikt
gilt das Lehrmaterial.

## Adaptions-Block

### MR-000 — Baseline-Aussage

- **Datum:** 2026-06-24
- **Geltungsbereich:** gesamtes Repo
- **Adaption:** *keine inhaltlichen Adaptionen ggü. Baseline-Default* für
  Verzeichniskonvention, Slice-Lifecycle (`open → next → in-progress →
  done`), Carveout-Disziplin und Trigger-Klassen. Repo-Klasse:
  *Tooling/Referenz*. Abweichungen am Präfix-Schema und an der
  Source-Precedence sind als eigene `MR-<NNN>` unten dokumentiert.
- **Begründung:** Initial-Setzung beim Greenfield-Bootstrap.
- **Auflösungs-Trigger:** permanent.

### MR-001 — Source Precedence mit eigener Spezifikations-Schicht

- **Datum:** 2026-06-24
- **Geltungsbereich:** [`README.md`](README.md) §Source precedence
- **Adaption:** Die Source-Precedence-Tabelle führt
  [`spec/spezifikation.md`](../spec/spezifikation.md) als eigenen
  **Rang 2** zwischen Lastenheft (Rang 1) und Architektur (Rang 3).
  Der Kurs-Default kennt zwei Spec-Ränge; dieses Repo nutzt drei.
- **Begründung:** Das Tool hat eine eigenständige technische Schicht
  (Regel-Engine, Klassifikations-Algorithmen, Config-Schema, Finding-/
  SARIF-Schema), die fortschreibbar ist, ohne das Lastenheft zu ändern.
  Damit die ADR-Schärfungs-Regel („ADR darf Spezifikation schärfen, nicht
  Lastenheft") strukturell abgebildet ist, muss die Spezifikation als
  eigener Rang sichtbar sein.
- **Auflösungs-Trigger:** permanent.

### MR-002 — ID-Schema: Kanon-Präfix `LH-` mit Bereichscodes

- **Datum:** 2026-06-24
- **Geltungsbereich:** alle Spec-Straten, ADRs, Commits, PRs
- **Adaption:** Dieses Repo nutzt das **kanonische** Anforderungs-Schema
  `LH-FA-<AREA>-<NNN>` (funktional) und `LH-QA-<NNN>` (qualitativ) — exakt
  die ai-harness/d-check-Form `LH-(FA-[A-Z]+|QA)-\d+`, sodass
  `make doc-trace` Anforderungen erkennt. Über die zwei Kanon-Klassen
  hinaus führt es **dokumentierte Zusatzkategorien** `LH-NF-*`
  (nichtfunktional), `LH-SEC-*` (Sicherheit) und Kontext-Präfixe
  (`LH-AUSG-*`, `LH-ZIEL-*`, …); diese sind linkpflichtig, zählen aber
  **nicht** als `--trace`-Requirements. Spec/Sicht nutzen `SPEC-*` /
  `ARC-*`:

  | Stratum / Artefakt                           | Präfix-Form          | Beispiel                                                                              |
  | -------------------------------------------- | -------------------- | ------------------------------------------------------------------------------------- |
  | Lastenheft — funktional                      | `LH-FA-<AREA>-<NNN>` | [`LH-FA-LENV-001`](../spec/lastenheft.md#lh-fa-lenv-001--prüfung-von-aws_region)      |
  | Lastenheft — nichtfunktional                 | `LH-NF-<NNN>`        | [`LH-NF-004`](../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)                  |
  | Lastenheft — Qualität                        | `LH-QA-<NNN>`        | [`LH-QA-001`](../spec/lastenheft.md#lh-qa-001--unit-tests)                            |
  | Lastenheft — Sicherheit                      | `LH-SEC-<NNN>`       | [`LH-SEC-003`](../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)               |
  | Lastenheft — sonst (Ziele, Modi, Risiken, …) | `LH-<KAT>-<NNN>`     | [`LH-ZIEL-002`](../spec/lastenheft.md#lh-ziel-002--eu-konformitätscheck)              |
  | Spezifikation (Technik)                      | `SPEC-<NNN>`         | [`SPEC-010`](../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids) |
  | Architektur (Sicht)                          | `ARC-<NNN>`          | [`ARC-003`](../spec/architecture.md#2-schichten-und-constraints)                      |
  | ADR                                          | `ADR-<NNNN>`         | [`ADR-0001`](../docs/plan/adr/0001-dokumentation-als-source-of-truth.md)              |
  | Carveout                                     | `CO-<NNN>`           | `CO-001`                                                                              |
  | Slice                                        | `slice-<NNN>`        | [`slice-001`](../docs/plan/planning/open/slice-001-hexagon-skelett-env-region.md)     |
  | Adaption                                     | `MR-<NNN>`           | `MR-002`                                                                              |

  Die `<AREA>`-Codes des Lastenhefts: `LENV` (lokale Env), `CLAUDE`
  (Claude-Code-Config), `DEVCON`/`DOCKER` (Container), `AWS`, `BEDROCK`,
  `IAM`, `TF` (Terraform), `REP` (Reporting), `EXIT` (Exit-Codes),
  `AUTH` (Authentifizierung).
- **Begründung:** Konformität zum Kanon, damit die Harness-Tooling
  (`--trace`, `--suggest-config`) greift, **ohne** die d-check-Config
  aufzuweichen. Die `<AREA>`-Codes (LENV, CLAUDE, …) halten die große
  funktionale Fläche (8 Prüf-Modi, ~30 Regeln) navigierbar — genau wie
  d-checks eigene `DC-FA-CLI-*`. **Migration 2026-06-24:** löst das
  ursprünglich vorgeschlagene bespoke `LH-BECC-*`-Schema ab; die IDs
  wurden repo-weit auf `LH-FA-*`/`LH-QA-*` gezogen, damit `--trace`
  Requirements zählt.
- **Auflösungs-Trigger:** permanent.

### MR-003 — Template-Verkörperung nur für wiederkehrende Skelette

- **Datum:** 2026-06-25
- **Geltungsbereich:** co-located `*.template.md`; [`AGENTS.md`](../AGENTS.md) §1; §Adoptierte Konventions-Quellen
- **Adaption:** bedrock verkörpert **nur die wiederkehrenden** Skelette
  co-located — ADR, Slice, Welle, Carveout, Review-Report (pro Artefakt
  kopiert). Die **Singleton**-Skelette (AGENTS, conventions, harness-/
  project-README, die spec-Straten, roadmap, Index-READMEs) werden **nicht**
  co-located: sie werden beim Bootstrap einmalig gefüllt, danach ist das
  *gefüllte* File die Verkörperung; das Roh-Skelett bleibt im Cache
  (`.harness/cache/<tag>/templates/` via `tools/harness/fetch-baseline-cache.sh`)
  für Adoption/Drift-Audit.
- **Begründung:** Der Baseline-Bootstrap (Regelwerk Modul 2) schreibt »alle
  Skelette kopieren« — eine **defekte Kanon-Regel**: Sie kennt nur den
  Adopter/Consumer und **verschweigt den Producer-/Self-Hosting-Fall**
  (verifiziert 2026-06-25: 0 Treffer für Producer/Self-Hosting über alle
  20 Module + 3 Grundlagen; die Repo-Klassen-Taxonomie achst nur nach Domäne,
  nicht nach Baseline-Beziehung). Eine Ebene tiefer verschmilzt »alle Skelette«
  die producer-autorierten **Singletons** (einmal gefüllt) mit den
  consumer-**wiederkehrenden** Artefakten (pro Anlage kopiert). bedrock folgt
  dem schärferen Prinzip (Modul 0: per-Lauf-Relevantes gehört verkörpert)
  statt der Pauschale — wiederkehrend → co-located, Singleton → Cache. Erklärt
  zugleich, warum d-check (Producer/Self-Hoster) korrekt **keine** Templates
  co-located.
- **Auflösungs-Trigger:** entfällt, sobald der Kurs eine Producer/Self-Hosting-
  Klasse bzw. die Singleton-/Wiederkehr-Unterscheidung kanonisiert; bis dahin
  permanent.

## Zusatzklassen-Deklaration für Sensors-Bindung

Es werden **keine** Zusatzklassen über die vier kanonischen hinaus (ADR,
Carveout, Schwelle, Reproduzierbarkeit) verwendet.

— keine —

> Wird später z. B. ein Gate eingeführt, das genau eine Lastenheft-Regel
> prüft (etwa ein „No-Secret-Output"-Test gebunden an [`LH-NF-004`](../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)),
> wird die **LH-Bindung** hier als Klasse `LH-<…>` deklariert, bevor
> sie in der Sensors-Tabelle auftaucht.

## Modus-Deklaration pro Sub-Area

Das Repo ist ein **Greenfield**-Projekt: Dokumentation führt, Code folgt.
Es existiert noch kein Bestandscode, daher gibt es keine Brownfield-
Inventur und keinen Reconciliation-Backlog.

| Sub-Area (Pfad / Modul)                 | Modus      | Begründung                                                                                              | Graduation-Bedingung |
| --------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------- | -------------------- |
| `*` (Default für gesamtes Repo)         | Greenfield | Leeres Repo, Spec und ADRs werden vor Code geschrieben.                                                 | n/a (GF)             |
| Spezifikation & Architektur (`spec/`)   | Greenfield | Verträge entstehen aus dem Lastenheft, nicht aus Code.                                                  | n/a (GF)             |
| Regel-Engine & Scanner (`src/` geplant) | Greenfield | Regeln folgen den `LH-FA-*`-Anforderungen.                                                              | n/a (GF)             |
| Test-Infrastruktur                      | Greenfield | Tests werden gegen `LH-*`-IDs geschrieben ([`LH-QA-001`](../spec/lastenheft.md#lh-qa-001--unit-tests)). | n/a (GF)             |

> Sub-Area-Disziplin: Die obigen Zeilen sind beim Bootstrap teils
> *Sub-Area-Aspirantinnen* (Struktur vor Substanz). Sobald eine eigene
> `MR-NNN`-Adaption oder eine eigenständige Inventur-Linie entsteht,
> qualifizieren sie nach der Drei-Achsen-Schwelle (≥ 2) zur vollen
> Sub-Area.

## Glossar (optional)

| Begriff               | Bedeutung in diesem Repo                                                                                                                                                                                                              |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| EU-safe Konfiguration | Konfiguration, die ausschließlich zugelassene EU-Regionen, freigegebene Modell-/Inference-Profile und AWS Bedrock statt direkter Anthropic-API nutzt (siehe [`LH-BEG-005`](../spec/lastenheft.md#lh-beg-005--eu-safe-konfiguration)). |
| Finding               | Einzelnes Prüfergebnis mit Schweregrad `PASS` / `WARN` / `FAIL`, Kennung, Nachricht und Empfehlung.                                                                                                                                   |
| Mantle                | Bedrock-Endpunkt für Claude Code in nativer Anthropic-API-Form, aber mit AWS-IAM (siehe [`LH-BEG-003`](../spec/lastenheft.md#lh-beg-003--mantle)).                                                                                    |
