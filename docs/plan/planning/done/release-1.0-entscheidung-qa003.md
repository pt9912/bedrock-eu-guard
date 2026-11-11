# Entscheidungsvorlage: Lücke bei LH-QA-003 (`.env`, `iam-policy.json`)

**Status:** entschieden und umgesetzt (Option 1b: `.env`-Scanner, siehe `slice-022-dotenv-scanner.md`; `iam-policy.json` bleibt Soll-Anforderung nach 1.0). **Datum:** 2026-11-11.
**Verfasser:** w1. **Teil der gemeinsamen Vorlage zu Release 1.0** (Release-Checkliste und die übrigen Entscheidungen: `release-1.0-entscheidungsvorlage.md`, w2).

Dieses Dokument trägt keine Anforderungen; es bereitet eine Entscheidung zum
vertraglich bindenden Lastenheft vor. Änderungen am Lastenheft nehmen nur die
Auftraggeber vor.

## 1. Ausgangslage

[`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) verlangt Positiv- und Negativbeispiele für sieben Dateiarten. Für fünf
(`devcontainer.json`, `docker-compose.yml`, `Dockerfile`, `main.tf`,
`.claude/settings.json`) gibt es Scanner, Regeln und Testdaten (`good`/`bad`,
Matrix-Test). Für **`.env`** und **`iam-policy.json`** beschreibt das Lastenheft
**keine funktionale Anforderung**: keine Regel prüft diese Dateien, deshalb gibt es
weder Scanner noch sinnvolle Testdaten. Der Matrix-Test loggt die Lücke bei jedem Lauf.

Bezug zu Muss-Anforderungen: [`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002) („Direkte Anthropic-API-Nutzung muss erkannt
werden“) lässt sich an der häufigsten Stelle, einer `.env`-Datei mit
`ANTHROPIC_API_KEY`, derzeit nicht erkennen, wenn die Variable nicht zusätzlich in
Umgebung, Devcontainer, Compose oder Settings steht.

## 2. Optionen

### Option 1 — Scanner ergänzen (Empfehlung, Mindestvariante 1b)

Das Lastenheft wird um zwei Anforderungen erweitert; die Scanner entstehen auf
vorhandenen Bausteinen.

**Vorschlag für neue Anforderungen** (Nummern und Wortlaut legen die Auftraggeber fest; sie folgen dem Schema der bestehenden Bereiche):

| Vorschlags-ID | Titel | Akzeptanzkriterium (Vorschlag) |
|---|---|---|
| LENV-007 (neu) | Prüfung von `.env`-Dateien | `FAIL`, wenn `ANTHROPIC_API_KEY` (oder `ANTHROPIC_AUTH_TOKEN`) gesetzt ist. `FAIL`, wenn `AWS_REGION`/`AWS_DEFAULT_REGION` auf eine nicht erlaubte Region zeigt. `PASS`, wenn nur erlaubte Regionen gesetzt sind. Ohne relevante Variablen kein Befund. Werte werden nie ausgegeben. |
| IAM-004 (neu) | Prüfung von IAM-Policy-Dokumenten (JSON) | Wie [`LH-FA-IAM-002`](../../../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen) und [`LH-FA-TF-003`](../../../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) für eigenständige Policy-Dateien: `FAIL` bei `bedrock:*`/`InvokeModel*`/`*` auf `*` ohne Region-Condition, `WARN` mit Region-Condition, `PASS` bei eng eingeschränkten Policies. |

**Varianten**

- **1a (vollständig):** beide Anforderungen.
- **1b (Mindestvariante):** nur `.env`. Das schließt die Lücke bei [`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002); für `iam-policy.json` wird [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) auf „Beispiele, soweit ein Scanner existiert“ präzisiert oder die Datei als Soll-Anforderung nach 1.0 verschoben.

**Festlegungen, die die Auftraggeber treffen müssen**

1. **Dateimuster für `.env`:** nur `.env`, oder auch `.env.local` und `.env.*`? Vorschlag: `.env`, `.env.local` und `.env.*.local`; Beispieldateien (`*.example`, `*.sample`, `*.template`) bleiben ausgenommen (Fehlalarme).
2. **Moduszuordnung:** `.env` gehört zu `local`. `iam-policy.json` passt zu `terraform` ([`LH-MOD-003`](../../../../spec/lastenheft.md#lh-mod-003--terraform-modus): IAM-Konfigurationen in Infrastrukturdateien), nicht zu `local`. Vorschlag: `terraform`.
4. **Suchtiefe für `.env`:** Die Datei-Discovery der Env-Quellen (Devcontainer, Compose, Dockerfile) findet heute nur Dateien **im Scan-Root** (nicht rekursiv); Terraform sucht rekursiv. Für `.env` ist festzulegen: nur im Scan-Root (kleiner Aufwand, deckt `apps/api/.env` nicht ab) oder rekursiv wie Terraform (größerer Aufwand: rekursive Suche, Ausschlüsse über `exclude`). Vorschlag: rekursiv, weil `.env`-Dateien in Monorepos typischerweise in Unterverzeichnissen liegen.
3. **Suchmuster für Policy-Dateien:** `iam-policy.json` allein ist eng; Vorschlag: `**/iam-policy*.json` und `**/*-policy.json` (rekursiv wie Terraform, Ausschlüsse über `exclude`).

**Aufwand** (Schätzung, je ein kleiner Slice nach dem bekannten Muster Plan, Kern, Adapter, Closure):

| Teil | Kern (w1) | Adapter/CLI/Doku (w2) |
|---|---|---|
| `.env` | `EnvFile` mit neuer Art `dotenv`, Regel nach dem Muster von [`LH-FA-DOCKER-001`](../../../../spec/lastenheft.md#lh-fa-docker-001--prüfung-von-docker-composeyml) (vorhandene Auswertung wiederverwendbar), Unit-Tests | `.env`-Parser (Kommentare, `export`, Anführungszeichen, Zeilen), Secret-Schwärzung wie bei den übrigen Quellen, Testdaten `good`/`bad`, CLI-Test ohne Wert in der Ausgabe |
| `iam-policy.json` | Regel über die vorhandene Bewertung `assessBedrockStatements`, Unit-Tests | JSON-Reader auf dem vorhandenen Paket `iampolicy`, rekursive Suche, Testdaten `good`/`bad` |
| Gemeinsam | Plan-Dateien, Regelkatalog (`rules.md`), Doku-Drift-Test (greift automatisch), `reference.md` | Handbuch, Fehlersuche |

Hinweis zum Aufwand der Discovery: Der `.env`-Parser ist klein, die Suche nicht, wenn sie rekursiv sein soll (Festlegung 4). Optional und kein Bestandteil dieser Entscheidung: Compose-`env_file:` zeigt oft auf genau diese Datei; der Compose-Reader könnte referenzierte Dateien als Fund melden (Folgepunkt).

Größenordnung: je Teil ein Arbeitsblock vergleichbar mit slice-006 (`.env`) bzw. einer Erweiterung von slice-010 (`iam-policy.json`); kein neuer Port außer einer weiteren `EnvSourceKind`-Art und einem Policy-Reader.

**Folgen**

- **Lastenheft** (bindend): zwei neue Abschnitte (Variante 1a) bzw. einer (1b), Anpassung von [`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) falls nötig, Eintrag der Regel-IDs im Regelkatalog.
- **Release 1.0:** verschiebt sich um die Slices; Alternative: Release 1.0 ohne die Scanner mit dokumentiertem Gap, Nachlieferung in 1.1 (dann bleibt die Lücke zu [`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002) für `.env` bis dahin offen).
- **Risiken:** Fehlalarme bei Beispieldateien (siehe Festlegung 1); **Testdaten werden von `.gitignore` verschluckt:** das Repository ignoriert `.env` und `.env.*` bewusst (Schutz für [`LH-SEC-001`](../../../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials) und [`LH-SEC-002`](../../../../spec/lastenheft.md#lh-sec-002--keine-speicherung-von-claude-anthropic-secrets)); Testdaten `good`/`bad` für den Scanner brauchen eine gezielte Negation (z. B. `!internal/**/testdata/**/.env`) oder `git add -f` und dürfen ausschließlich offensichtlich synthetische Werte enthalten (Marker wie `GEHEIM-…`), damit Secret-Scanner und Reviews sie nicht für echte Schlüssel halten (dieselbe Falle wie beim früheren, durch `.git/info/exclude` verschluckten Quellcode: Testdaten nach dem Anlegen mit `git status` prüfen); `.env`-Dateien enthalten reale Secrets, deshalb darf kein Wert den Adapter verlassen (Muster und Tests wie bei Compose und Dockerfile vorhanden); `.env` ist kein einheitlich spezifiziertes Format (Dialekte), der Parser deckt die gängige Syntax ab und meldet unlesbare Dateien als `FILE_INVALID`.

### Option 2 — Anforderung streichen oder präzisieren

[`LH-QA-003`](../../../../spec/lastenheft.md#lh-qa-003--testdaten) wird auf die tatsächlich geprüften Dateiarten beschränkt (oder `.env`/`iam-policy.json` ausdrücklich als „Beispiele für künftige Scanner“ gekennzeichnet).

- **Aufwand:** nur Textänderung im Lastenheft; Matrix-Test und Katalog entfallen bzw. verlieren die Log-Zeile.
- **Folgen:** Release 1.0 ohne Verzug. **Preis:** `.env` mit `ANTHROPIC_API_KEY` bleibt unentdeckt (Lücke zu [`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002)); das müssen die Auftraggeber bewusst akzeptieren und im Handbuch als bekannte Grenze benennen.

## 3. Empfehlung

**Option 1b als Minimum, 1a bevorzugt.** Begründung: Die Lücke betrifft eine Muss-Anforderung
([`LH-PRI-MUSS-002`](../../../../spec/lastenheft.md#lh-pri-muss-002)) an einer realistisch häufigen Stelle; der Aufwand ist klein, weil Auswertung, Secret-Regeln,
Testdatenmuster und Policy-Bewertung bereits existieren. `iam-policy.json` ist weniger kritisch und kann als Soll-Anforderung in 1.1 folgen, wenn der Zeitplan das verlangt.
Option 2 ist nur vertretbar, wenn Release 1.0 ausdrücklich ohne `.env`-Prüfung erscheinen soll und diese Grenze dokumentiert ist.

## 4. Benötigte Entscheidungen

1. Option 1a, 1b oder 2?
2. Bei Option 1: Dateimuster (Festlegung 1 und 3), Moduszuordnung (Festlegung 2) und Suchtiefe (Festlegung 4) bestätigen oder ändern; Wortlaut und Nummern der neuen Lastenheft-Anforderungen freigeben.
3. Bei Option 1: Release 1.0 nach den Slices (Zeitplan) oder Release 1.0 mit dokumentiertem Gap und Nachlieferung in 1.1?

Nichts davon wird ohne Entscheidung umgesetzt; das Lastenheft bleibt unverändert, bis die Auftraggeber zugestimmt haben.
