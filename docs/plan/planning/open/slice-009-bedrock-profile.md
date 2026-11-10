# Slice slice-009: Bedrock-Inference-Profile, Allowlist und Modellverfügbarkeit

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** welle-03-aws-checks.

**Bezug:** [`LH-FA-BEDROCK-001`](../../../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten), [`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen), [`LH-FA-BEDROCK-003`](../../../../spec/lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen), [`LH-FA-BEDROCK-004`](../../../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen), [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), [`LH-SEC-003`](../../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-10.

---

## 1. Ziel

`bedrock-eu-check aws` liest Inference-Profile und Basismodelle der erlaubten
Regionen (lesend, über `awscli`), erkennt nicht erlaubte Profile, prüft die
konfigurierten Profile gegen die Allowlist und meldet, ob die konfigurierten
Claude-Modelle in der Primärregion verfügbar sind.

## 2. Verträge

- Zusätzlicher Port `driven.BedrockReader` (eigene Schnittstelle neben `AWSReader`, damit bestehende Adapter nicht brechen): `ListInferenceProfiles(region) ([]model.InferenceProfile, error)` und `ListFoundationModels(region) ([]model.FoundationModel, error)`; Fehler sind `*model.AWSError` (wie bei `Identity`), nie CLI-Ausgabe.
- `model.InferenceProfile{ID, Name, Type}`, `model.FoundationModel{ID}`.
- `app.CheckAWS` bekommt das optionale Feld `Bedrock`; `nil` = keine Bedrock-Regeln. Der Kern bestimmt die Regionen: alle `allowedRegions` (stabil, ohne Duplikate); der Adapter bekommt je Aufruf eine Region.
- Primärregion: `AWS_REGION`, wenn erlaubt, sonst die erste erlaubte Region.
- Ist die Identität schon als `AWS_NO_CREDENTIALS`/`AWS_UNREACHABLE` bekannt, ruft der Kern Bedrock nicht erneut auf, sondern meldet denselben Fehler je Region.
- Double `coretest/fakeaws.Bedrock`.

**Adapter-Pflichten (w2):** `awscli.Reader` implementiert zusätzlich `BedrockReader` (`aws bedrock list-inference-profiles --region R --output json`, `aws bedrock list-foundation-models …`); gleiche Fehlerklassifikation wie `Identity`; nur lesende Aufrufe; Paginierung (`nextToken`/`--max-results`) vollständig abarbeiten; Fake-`aws` in den Tests.

## 3. Verhalten je Regel

„Verwendet“ heißt: per Umgebung konfiguriert (`ANTHROPIC_MODEL`, `ANTHROPIC_DEFAULT_SONNET|HAIKU|OPUS_MODEL`, `ANTHROPIC_SMALL_FAST_MODEL`); „sichtbar“ heißt: in der Profilliste einer abgefragten Region.

| Regel | Verhalten |
|---|---|
| [`LH-FA-BEDROCK-001`](../../../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten) | Je Region: `PASS` bei gelesenen Profilen, `WARN` (`PROFILES_NONE`) ohne sichtbare Profile, `FAIL` bei AWS-Fehler (Reason der `AWSError`, Region in der Meldung). |
| [`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen) | `FAIL`: eine verwendete ID ist laut [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids) blockiert (je Variable). `WARN` (`PROFILES_VISIBLE_NOT_ALLOWED`): blockierte Profile nur sichtbar, nicht verwendet (ein Finding, mit Anzahl und erstem Beispiel). `PASS`, wenn nichts davon zutrifft. |
| [`LH-FA-BEDROCK-003`](../../../../spec/lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen) | Nur verwendete IDs. Abgedeckt = vollständiger Treffer auf ein Muster aus `allowedProfiles` (Muster werden mit `^(?:…)$` verankert, damit `eu.*` nicht jede ID mit „eu“ im Namen trifft) **oder** laut [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids) erlaubt. `FAIL` je nicht abgedeckter ID, sonst `PASS`; ohne verwendete IDs kein Finding. |
| [`LH-FA-BEDROCK-004`](../../../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen) | Je verwendeter ID in der Primärregion: `PASS`, wenn das Profil gelistet ist oder das Basismodell (ID ohne Geo-Präfix `eu.`/`us.`/`global.`/…) in `ListFoundationModels` steht. `FAIL` (`MODEL_UNAVAILABLE`): Listen lesbar, ID nicht enthalten. `WARN` (`MODEL_AVAILABILITY_UNKNOWN`): Listenfehler, ARN-ID oder Primärregion ohne Daten. Ohne verwendete IDs kein Finding. |

## 4. Definition of Done

- [ ] Kern (w1): `model`-Typen, `driven.BedrockReader`, `rules/bedrock.go`, Erweiterung `app.CheckAWS`, Double `fakeaws.Bedrock`, Unit-Tests je Zweig und Fehlerart.
- [ ] Adapter (w2): `awscli` implementiert `BedrockReader` inkl. Paginierung, Tests mit Fake-`aws`.
- [ ] CLI (w2): Befehl `aws` übergibt `Bedrock`; Akzeptanztest ohne Credential-Werte in der Ausgabe.
- [ ] `make verify-clean` und CI grün; Sensors-Doku aktuell.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/aws.go`, `port/driven/files.go`, `core/rules/bedrock.go`, `core/app/checkaws.go`, `coretest/fakeaws/` | update/neu | w1 |
| `internal/adapter/driven/awscli/` | update | w2 |
| `internal/adapter/driving/cli/` | update | w2 |

## 6. Trigger

slice-008 gemerged; Plan gemerged.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- **Sichtbarkeit nur in erlaubten Regionen:** Es werden bewusst nur `allowedRegions` abgefragt. System-Profile wie `global.*` erscheinen dort ebenfalls. Eine zusätzliche Abfrage fester Nicht-EU-Regionen wäre ein Folgepunkt (mehr Aufrufe, mehr nötige Rechte).
- **„Verwendet“ nur aus der Umgebung:** Modelle aus Claude-Settings oder Konfigurationsdateien fließen nicht ein; die Zusammenführung mit dem Ergebnis des `local`-Laufs (`all`) ist ein Folgepunkt.
- `allowedProfiles` wird als verankerter regulärer Ausdruck gelesen ([`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) zeigt glob-ähnliche Beispiele wie `eu.*`; sie sind als Regex gleichbedeutend).
- Die Verfügbarkeitsprüfung kennt keine Account-/Zugriffsfreigabe pro Modell (Model Access); `ListFoundationModels` zeigt Angebot, nicht Freischaltung.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: neue Regeln und Adapter-Erweiterung auf dem
Kern von slice-008.
