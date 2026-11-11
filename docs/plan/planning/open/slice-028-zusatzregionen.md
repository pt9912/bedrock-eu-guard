# Slice slice-028: Zusätzliche Regionen für die Sichtbarkeit nicht erlaubter Profile

**Status:** open → next → in-progress → done (Datei wird durch die
Verzeichnisse bewegt).

**Welle:** [`welle-07-ci-komfort`](../welle-07-ci-komfort.md).

**Bezug:** [`LH-FA-BEDROCK-001`](../../../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten), [`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen), [`LH-NF-007`](../../../../spec/lastenheft.md#lh-nf-007--konfigurierbarkeit), [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml), [`LH-SEC-003`](../../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff), [`ADR-0003`](../../adr/0003-hexagonale-architektur.md).

**Autor:** Projektteam bedrock-eu-guard. **Datum:** 2026-11-11.

---

## 1. Ziel

Der Modus `aws` fragt bisher nur die erlaubten Regionen ab. Nicht erlaubte Profile (`us.*`, `global.*`, …) sind
aber auch **in anderen Regionen** sichtbar, etwa in `us-east-1`. Wer das wissen will, kann **zusätzliche Regionen** konfigurieren;
in ihnen wird ausschließlich die Sichtbarkeit nicht erlaubter Profile geprüft (Folgepunkt aus slice-009).

## 2. Verträge

- **Konfiguration (w1):** neues optionales Feld `visibilityRegions` in `bedrock-eu-check.yml` (Liste von Regionsnamen, Default leer = unverändertes Verhalten). `PartialConfig`/`Resolve`/`Config` und [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) werden ergänzt. Validierung: keine leeren Namen; eine Region darf nicht zugleich in `allowedRegions` stehen (sonst Konfigurationsfehler, Exit `2`, ohne Wert im Fehlertext).
- **Kern (w1):** `rules.QueryRegions` liefert weiterhin die erlaubten Regionen; neu `rules.VisibilityRegions(cfg)`. `app.CheckAWS` liest für die Zusatzregionen **nur** `ListInferenceProfiles` (keine Modell-Liste); `RegionData` bekommt ein Kennzeichen `VisibilityOnly`.
- **Regeln (w1):**
  - [`LH-FA-BEDROCK-001`](../../../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten) (Profile je Region) betrifft nur erlaubte Regionen, unverändert.
  - [`LH-FA-BEDROCK-002`](../../../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen): sichtbare, nach [`SPEC-010`](../../../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids) blockierte Profile in Zusatzregionen führen zu einem `WARN` (`PROFILES_VISIBLE_NOT_ALLOWED`) mit Region in der Meldung; sie sind nie ein `FAIL` (das bleibt „verwendet“). Ein Fehler beim Lesen einer Zusatzregion (zum Beispiel Zugriff verweigert) ist ein `WARN` mit neuem Reason `VISIBILITY_UNKNOWN`, kein `FAIL`.
- **Adapter (w2):** `awscli.Reader` nutzt bereits eine Region je Aufruf; keine Vertragsänderung nötig. `configyaml` mappt `visibilityRegions` 1:1 auf `PartialConfig` (Feld abwesend = `nil`, `visibilityRegions: []` = leer; strikt: ein Eintrag, der kein String ist, ergibt `FILE_INVALID` ohne Wertzitat). Der Composition Root übergibt die Konfiguration unverändert. CLI-Test mit Fake-`aws` für eine Zusatzregion.
- **Doku:** `aws-permissions.md` (w2) nennt `bedrock:ListInferenceProfiles` in den Zusatzregionen als nötiges Recht; eine IAM-Policy mit Region-Condition nur für EU führt dort zu `VISIBILITY_UNKNOWN` (`WARN`); die Beispiel-Policy bekommt nur einen Hinweis. Regelkatalog und Referenz (`rules.md`, `reference.md`) mit den Reason-Codes `PROFILES_VISIBLE_NOT_ALLOWED` und `VISIBILITY_UNKNOWN` sowie dem Konfigurationsfeld pflegt w1 (Doku-Drift-Test).

## 3. Verhalten

| Situation | Ergebnis |
|---|---|
| `visibilityRegions` leer (Default) | wie bisher; keine zusätzlichen Aufrufe |
| Zusatzregion zeigt `us.*`-Profil | `WARN` `PROFILES_VISIBLE_NOT_ALLOWED` (Region genannt) |
| Zusatzregion zeigt nur erlaubte Profile | kein zusätzliches Finding |
| Zusatzregion nicht lesbar (verweigert/nicht erreichbar) | `WARN` `VISIBILITY_UNKNOWN` |
| Region in `allowedRegions` **und** `visibilityRegions` | Konfigurationsfehler, Exit `2` |

## 4. Definition of Done

- [ ] Kern (w1): Konfigurationsfeld mit Validierung, `CheckAWS` liest Zusatzregionen, Regeländerung, Reason-Code, Unit-Tests je Zeile der Tabelle; [`SPEC-020`](../../../../spec/spezifikation.md#spec-020--konfigurationsschema-bedrock-eu-checkyml) und Regelkatalog (Doku-Drift-Test) ergänzt.
- [ ] Adapter/CLI (w2): `configyaml` mappt das Feld strikt (Typprüfung), CLI-Test mit Fake-`aws`, Referenz und Handbuch ergänzt; `docs/user/aws-permissions.md` nennt die zusätzlichen Aufrufe je Zusatzregion.
- [ ] Standardverhalten unverändert (Regressionstest).
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| `internal/hexagon/core/model/{model,config}.go`, `core/rules/bedrock.go`, `core/app/checkaws.go`, `coretest/fakeaws`, `spec/spezifikation.md`, `docs/user/{rules,reference}.md` | update | w1 |
| `internal/adapter/driven/configyaml/`, `internal/adapter/driving/cli/`, `docs/user/{aws-permissions,handbuch}.md` | update | w2 |

## 6. Trigger

Plan der Welle 07 gemerged; Slice-Plan gemerged.

## 7. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 8. Risiken und offene Punkte

- **Mehr AWS-Aufrufe und Rechte:** Je Zusatzregion ein weiterer Aufruf; die nötigen Rechte (`bedrock:ListInferenceProfiles` in dieser Region) müssen vorhanden sein, sonst `WARN`. Bleibt lesend ([`LH-SEC-003`](../../../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)).
- **Rauschen:** In Regionen wie `us-east-1` sind viele System-Profile sichtbar; die Meldung fasst sie zusammen (Anzahl und erstes Beispiel), wie bisher bei „sichtbar, nicht verwendet“.
- **Kein Ersatz für Richtlinien:** Sichtbarkeit zeigt, was angeboten wird, nicht was ein Principal nutzen darf; durchsetzen lässt sich das nur per IAM/SCP.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: optionale Erweiterung bestehender Regeln und der Konfiguration.
