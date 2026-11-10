# Regelkatalog — bedrock-eu-check

**Status:** Stand v1.0-Kandidat. Dieses Dokument trägt keine eigenen Anforderungen
(die stehen im Lastenheft, Rang 1 der Source Precedence); es beschreibt, was jede
Regel prüft, welche Ergebnisse sie liefert und welche Reason-Codes im JSON-/SARIF-Ausgabeformat
auftauchen (`reason`).

Schweregrade: `PASS` (in Ordnung), `WARN` (manuell prüfen), `FAIL` (nicht
EU-konform bzw. nicht belegbar). Der Exit-Code ist `1`, sobald mindestens ein
`FAIL` vorliegt ([`SPEC-030`](../../spec/spezifikation.md#spec-030--exit-codes)). Regionen gelten als erlaubt, wenn sie in `allowedRegions`
stehen (Default `eu-central-1`, [`SPEC-011`](../../spec/spezifikation.md#spec-011--regionsklassifikation)). Modell-/Profil-IDs werden nach
[`SPEC-010`](../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids) klassifiziert. Secret-Werte erscheinen nie in
der Ausgabe, nur „gesetzt“/„nicht gesetzt“.


## Modus `local` (offline)

| Regel | Prüfgegenstand | Ergebnis | Reason-Codes |
|---|---|---|---|
| [`LH-FA-LENV-001`](../../spec/lastenheft.md#lh-fa-lenv-001--prüfung-von-aws_region) | `AWS_REGION` | PASS: gesetzt und erlaubt. FAIL: fehlt (auch ohne `AWS_DEFAULT_REGION`) oder nicht erlaubt. | `OK`, `REGION_MISSING`, `REGION_NOT_ALLOWED` |
| [`LH-FA-LENV-002`](../../spec/lastenheft.md#lh-fa-lenv-002--prüfung-von-aws_default_region) | `AWS_DEFAULT_REGION` (nur wenn `AWS_REGION` fehlt) | WARN: erlaubt (Fallback). FAIL: nicht erlaubt. Beide fehlend meldet LENV-001. | `REGION_FALLBACK_DEFAULT`, `REGION_NOT_ALLOWED` |
| [`LH-FA-LENV-003`](../../spec/lastenheft.md#lh-fa-lenv-003--prüfung-von-claude_code_use_bedrock) | `CLAUDE_CODE_USE_BEDROCK` | PASS: Wert `1`. FAIL: fehlt oder anderer Wert. | `OK`, `BEDROCK_NOT_ENABLED` |
| [`LH-FA-LENV-004`](../../spec/lastenheft.md#lh-fa-lenv-004--prüfung-von-anthropic_api_key) | `ANTHROPIC_API_KEY` | PASS: nicht gesetzt. FAIL: gesetzt (nur „set“/„not set“, nie der Wert). | `ANTHROPIC_API_KEY_NOT_SET`, `ANTHROPIC_API_KEY_SET` |
| [`LH-FA-LENV-005`](../../spec/lastenheft.md#lh-fa-lenv-005--prüfung-von-claude_code_use_mantle) | `CLAUDE_CODE_USE_MANTLE` (nur bei Wert `1`) | Konfig `mantle: warn` (Default): PASS, wenn mindestens eine Modellvariable gesetzt und alle erlaubt; sonst WARN. `allow`: PASS. `block`: FAIL. | `MANTLE_VERIFIED`, `MANTLE_UNVERIFIED`, `MANTLE_ALLOWED`, `MANTLE_BLOCKED` |
| [`LH-FA-LENV-006`](../../spec/lastenheft.md#lh-fa-lenv-006--prüfung-von-claude-default-modellvariablen) | `ANTHROPIC_MODEL`, `ANTHROPIC_DEFAULT_{SONNET,HAIKU,OPUS}_MODEL` | Je gesetzter Variable: PASS erlaubt, WARN nicht klassifizierbar, FAIL blockiert (Klassifikation nach [`SPEC-010`](../../spec/spezifikation.md#spec-010--klassifikation-von-modell-profil-ids), siehe oben). Ungesetzte Variablen erzeugen kein Finding. | `MODEL_ALLOWED`, `MODEL_UNCLASSIFIED`, `MODEL_BLOCKED` |
| [`LH-FA-CLAUDE-001`](../../spec/lastenheft.md#lh-fa-claude-001--benutzerweite-claude-code-settings-prüfen) | `~/.claude/settings.json` | WARN: Datei fehlt (`FILE_MISSING`) oder Bedrock nicht aktiviert. FAIL: direkte Anthropic-Konfiguration oder Datei ungültig/nicht lesbar. PASS: Bedrock aktiviert, keine direkte Konfiguration. | `FILE_MISSING`, `BEDROCK_NOT_ENABLED`, `DIRECT_ANTHROPIC`, `FILE_INVALID`, `FILE_UNREADABLE`, `OK` |
| [`LH-FA-CLAUDE-002`](../../spec/lastenheft.md#lh-fa-claude-002--projektweite-claude-code-settings-prüfen) | `.claude/settings.json`, `.claude/settings.local.json` (je Datei) | FAIL: nicht erlaubtes Modell, nicht erlaubte `AWS_REGION` oder `ANTHROPIC_API_KEY`. WARN: nicht klassifizierbares Modell. PASS: EU-safe. Fehlende Dateien: kein Finding. | `SETTINGS_UNSAFE`, `SETTINGS_REVIEW`, `OK`, `FILE_INVALID`, `FILE_UNREADABLE` |
| [`LH-FA-CLAUDE-003`](../../spec/lastenheft.md#lh-fa-claude-003--erkennung-direkter-anthropic-endpunkte) | alle vorhandenen Settings-Dateien | Rangfolge je Datei laut [`SPEC-012`](../../spec/spezifikation.md#spec-012--direkte-anthropic-konfiguration-in-claude-settings): FAIL (API-Key/Token, `api.anthropic.com`) > WARN `apiKeyHelper` > WARN fremder Host > WARN Text-Fund > PASS. | `DIRECT_ANTHROPIC`, `APIKEY_HELPER`, `ENDPOINT_FOREIGN`, `ENDPOINT_TEXT_ONLY`, `OK` |
| [`LH-FA-CLAUDE-004`](../../spec/lastenheft.md#lh-fa-claude-004--claude-code-wizard-konfiguration-prüfen) | Wizard-Konfiguration in `~/.claude/settings.json` (nur bei `CLAUDE_CODE_USE_BEDROCK=1`) | PASS: erlaubte `AWS_REGION`. FAIL: Region fehlt oder nicht erlaubt, oder parallel direkte Anthropic-Konfiguration. | `OK`, `REGION_MISSING`, `REGION_NOT_ALLOWED`, `DIRECT_ANTHROPIC` |
| [`LH-FA-DEVCON-001`](../../spec/lastenheft.md#lh-fa-devcon-001--prüfung-von-devcontainerdevcontainerjson) | `.devcontainer/devcontainer.json`, `.devcontainer.json` | FAIL: Anthropic-Zugang oder nicht erlaubte Region. PASS: Bedrock-Flag `1` und erlaubte Region. WARN: keine Claude-/Bedrock-Konfiguration oder unvollständig. | `DIRECT_ANTHROPIC`, `REGION_NOT_ALLOWED`, `OK`, `NO_CLAUDE_CONFIG`, `ENV_INCOMPLETE` |
| [`LH-FA-DOCKER-001`](../../spec/lastenheft.md#lh-fa-docker-001--prüfung-von-docker-composeyml) | `docker-compose*.yml|yaml`, `compose*.yml|yaml` | FAIL: Anthropic-Zugang oder nicht erlaubte Region. PASS: mindestens eine Region gesetzt, alle erlaubt. Ohne relevante Variablen: kein Finding. | `DIRECT_ANTHROPIC`, `REGION_NOT_ALLOWED`, `OK` |
| [`LH-FA-DOCKER-002`](../../spec/lastenheft.md#lh-fa-docker-002--prüfung-von-dockerfile) | `Dockerfile*` | FAIL: Anthropic-Zugang oder nicht erlaubte Region. WARN: keine Region per `ENV`. PASS: Region erlaubt. | `DIRECT_ANTHROPIC`, `REGION_NOT_ALLOWED`, `REGION_NOT_SET`, `OK` |

## Modus `terraform` (offline)

| Regel | Prüfgegenstand | Ergebnis | Reason-Codes |
|---|---|---|---|
| [`LH-FA-TF-001`](../../spec/lastenheft.md#lh-fa-tf-001--terraform-dateien-finden) | `*.tf` (rekursiv) | PASS: Dateien gefunden und gelesen. WARN: keine gefunden. FAIL: Datei ungültig/nicht lesbar. | `TERRAFORM_FOUND`, `NO_TERRAFORM_FILES`, `FILE_INVALID`, `FILE_UNREADABLE` |
| [`LH-FA-TF-002`](../../spec/lastenheft.md#lh-fa-tf-002--aws-provider-region-prüfen) | AWS-Provider-Region je Datei | FAIL: nicht erlaubte Region. WARN: Region dynamisch oder fehlend. PASS: nur erlaubte. | `REGION_NOT_ALLOWED`, `REGION_DYNAMIC`, `REGION_MISSING`, `OK` |
| [`LH-FA-TF-003`](../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) | Bedrock-IAM-Statements | FAIL: `bedrock:*`/`InvokeModel*`/`*` auf `*` ohne Region-Condition. WARN: mit Region-Condition oder nicht auflösbar. PASS: eng eingeschränkt. | `IAM_WILDCARD`, `IAM_WILDCARD_REGION_CONDITION`, `IAM_DYNAMIC`, `OK` |
| [`LH-FA-TF-004`](../../spec/lastenheft.md#lh-fa-tf-004--nicht-erlaubte-modell-ids-finden) | Modell-/Profil-IDs in Literalen | FAIL: blockiert (`us.`, `global.`, fremde Region). WARN: nicht klassifizierbar. PASS: nur erlaubte. | `MODEL_BLOCKED`, `MODEL_UNCLASSIFIED`, `OK` |

## Modus `aws` (lesend gegen AWS)

| Regel | Prüfgegenstand | Ergebnis | Reason-Codes |
|---|---|---|---|
| [`LH-FA-AWS-001`](../../spec/lastenheft.md#lh-fa-aws-001--sts-identität-prüfen) | STS-Identität | PASS: Account und Principal ermittelt. FAIL: keine Credentials, Zugriff verweigert oder AWS nicht erreichbar. | `OK`, `AWS_NO_CREDENTIALS`, `AWS_ACCESS_DENIED`, `AWS_UNREACHABLE` |
| [`LH-FA-AWS-002`](../../spec/lastenheft.md#lh-fa-aws-002--aws-profil-anzeigen) | `AWS_PROFILE` | PASS: gesetzt und verwendbar. WARN: nicht gesetzt oder nicht verwendbar. | `AWS_PROFILE_SET`, `AWS_PROFILE_NOT_SET`, `AWS_PROFILE_UNUSABLE` |
| [`LH-FA-AWS-003`](../../spec/lastenheft.md#lh-fa-aws-003--erwartete-account-id-prüfen) | Account-ID gegen `allowedAccounts` / `--allowed-account` | Ohne Liste: kein Finding. PASS: erlaubt. FAIL: nicht erlaubt. WARN: nicht prüfbar. | `ACCOUNT_ALLOWED`, `ACCOUNT_NOT_ALLOWED`, `ACCOUNT_UNCHECKED` |
| [`LH-FA-AUTH-001`](../../spec/lastenheft.md#lh-fa-auth-001--authentifizierungsart-klassifizieren) | Authentifizierungsart aus Umgebungsvariablen (Existenz, nie Werte) | Ein Finding je Art: FAIL `ANTHROPIC_API_KEY`; WARN Bedrock API Key und statische Access Keys; PASS Profil und temporäre Credentials; WARN, wenn nichts erkannt. | `AUTH_ANTHROPIC_API_KEY`, `AUTH_BEDROCK_API_KEY`, `AUTH_STATIC_KEYS`, `AUTH_PROFILE`, `AUTH_TEMPORARY_CREDENTIALS`, `AUTH_NONE_DETECTED` |
| [`LH-FA-BEDROCK-001`](../../spec/lastenheft.md#lh-fa-bedrock-001--bedrock-inference-profile-auflisten) | Inference-Profile je erlaubter Region | PASS: Profile gelesen. WARN: keine sichtbar. FAIL: AWS-Fehler. | `PROFILES_LISTED`, `PROFILES_NONE`, `AWS_*` |
| [`LH-FA-BEDROCK-002`](../../spec/lastenheft.md#lh-fa-bedrock-002--nicht-erlaubte-profile-erkennen) | verwendete und sichtbare Profile | FAIL: verwendete ID blockiert. WARN: blockierte Profile nur sichtbar. PASS sonst. | `MODEL_BLOCKED`, `PROFILES_VISIBLE_NOT_ALLOWED`, `OK` |
| [`LH-FA-BEDROCK-003`](../../spec/lastenheft.md#lh-fa-bedrock-003--erlaubte-eu-profile-prüfen) | verwendete IDs gegen `allowedProfiles` | FAIL: nicht abgedeckt. PASS: alle abgedeckt. Ohne verwendete IDs: kein Finding. | `PROFILE_NOT_ALLOWLISTED`, `OK` |
| [`LH-FA-BEDROCK-004`](../../spec/lastenheft.md#lh-fa-bedrock-004--modellverfügbarkeit-prüfen) | Verfügbarkeit in der Primärregion | PASS: verfügbar. FAIL: nicht verfügbar. WARN: nicht eindeutig ermittelbar. | `MODEL_AVAILABLE`, `MODEL_UNAVAILABLE`, `MODEL_AVAILABILITY_UNKNOWN` |
| [`LH-FA-IAM-001`](../../spec/lastenheft.md#lh-fa-iam-001--bedrock-rechte-prüfen) | Bedrock-Rechte des Principals (Simulation) | FAIL: Aufruf-Rechte fehlen. WARN: nur Erkennungs-Rechte fehlen oder Simulation nicht erlaubt. PASS: alles erlaubt. | `IAM_INVOKE_DENIED`, `IAM_DISCOVERY_DENIED`, `IAM_UNKNOWN`, `OK` |
| [`LH-FA-IAM-002`](../../spec/lastenheft.md#lh-fa-iam-002--zu-breite-bedrock-policies-erkennen) | zu breite Policies des Principals | Wie [`LH-FA-TF-003`](../../spec/lastenheft.md#lh-fa-tf-003--bedrock-iam-policies-prüfen) auf den Policies des Principals. | `IAM_WILDCARD`, `IAM_WILDCARD_REGION_CONDITION`, `IAM_DYNAMIC`, `OK` |
| [`LH-FA-IAM-003`](../../spec/lastenheft.md#lh-fa-iam-003--region-condition-prüfen) | Region-Condition | FAIL: Bedrock-Allow ohne `aws:RequestedRegion` oder mit nicht erlaubter Region. WARN: Regionen nicht auswertbar. PASS: eingeschränkt. | `IAM_NO_REGION_CONDITION`, `IAM_REGION_CONDITION_NOT_ALLOWED`, `IAM_REGION_CONDITION_UNKNOWN`, `OK` |

## Datei-Reason-Codes

Bei vorhandenen, aber nicht nutzbaren Dateien melden die Datei-Regeln ein `FAIL`
mit `FILE_UNREADABLE` (nicht lesbar) oder `FILE_INVALID` (ungültig, mit Zeile,
ohne Dateiinhalt); der Lauf geht weiter. Fehlende optionale Dateien erzeugen
kein Finding (Ausnahme: Benutzer-Settings, `FILE_MISSING` als `WARN`).

## Hinweise

- `all` führt `local`, `terraform` und `aws` aus; `aws` benötigt das
  `aws`-Binary und Zugangsdaten (siehe `aws-permissions.md`).
- Pfade, die in `exclude` stehen, werden nicht gelesen (siehe `reference.md`).
- Jedes Finding trägt Regel-ID, Meldung, Empfehlung und – falls bekannt – Datei
  und Zeile; die Ausgabeformate stehen in `ci.md` und `reference.md`.
