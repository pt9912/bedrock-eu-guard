# Lastenheft — Bedrock EU Configuration Checker (`bedrock-eu-check`)

**Version:** 0.1.0 · **Status:** Draft · **Autor:** Projektteam bedrock-eu-guard · **Datum:** 2026-06-24

Dieses Lastenheft ist die **vertraglich abnahmebindende** Quelle der
Wahrheit für `bedrock-eu-check` (Rang 1 der Source Precedence). Es prüft
lokale, AWS-, Bedrock-, Claude-Code- und Terraform-Konfigurationen auf
EU-/Frankfurt-Konformität. Das ID-Schema `LH-*` ist in
[`harness/conventions.md`](../harness/conventions.md) ([`MR-002`](../harness/conventions.md#mr-002--id-schema-kanon-präfix-lh--mit-bereichscodes)) deklariert.

## 1. Dokumentinformationen

| Feld             | Inhalt                                                                |
| ---------------- | --------------------------------------------------------------------- |
| Dokument-ID      | LH-001                                                           |
| Titel            | Lastenheft Bedrock EU Configuration Checker                           |
| Version          | 0.1.0                                                                 |
| Status           | Draft (Entwurf)                                                       |
| Zielsystem       | CLI-Tool zur Prüfung von Claude Code / AWS Bedrock EU-Konfigurationen |
| Zielgruppe       | Softwareentwicklung, DevOps, Cloud Engineering, Datenschutz, Security |
| Primäre Umgebung | Linux, VS Code, Devcontainer, Docker, Terraform, AWS                  |

---

## 2. Ausgangssituation

### LH-AUSG-001 — Nutzung von Claude Code über AWS Bedrock

Claude Code soll über Amazon Bedrock betrieben werden, nicht über die direkte Anthropic API.

### LH-AUSG-002 — EU-/Frankfurt-Anforderung

Für den geplanten Betrieb sollen AWS-Bedrock-Konfigurationen bevorzugt auf EU-Regionen, insbesondere `eu-central-1` / Frankfurt, eingeschränkt werden.

### LH-AUSG-003 — Fehlkonfigurationsrisiko

Durch falsche Umgebungsvariablen, AWS-Profile, Claude-Code-Settings, Inference-Profile oder Terraform/IAM-Definitionen kann unbeabsichtigt eine nicht gewünschte Region, ein globales Profil oder ein US-Profil verwendet werden.

### LH-AUSG-004 — Manuelle Prüfung reicht nicht aus

Eine manuelle Kontrolle ist fehleranfällig und skaliert nicht für mehrere Entwickler, Devcontainer, CI/CD-Pipelines und AWS-Accounts.

---

## 3. Zielsetzung

### LH-ZIEL-001 — Automatisierte Konfigurationsprüfung

Es soll ein Tool erstellt werden, das relevante lokale und AWS-seitige Konfigurationen automatisiert prüft.

### LH-ZIEL-002 — EU-Konformitätscheck

Das Tool soll erkennen, ob die Konfiguration auf zugelassene EU-Regionen und freigegebene Bedrock-/Claude-Profile eingeschränkt ist.

### LH-ZIEL-003 — Früherkennung kritischer Risiken

Das Tool soll harte Verstöße wie gesetzte direkte Anthropic-API-Keys, US-Regionen, globale Inference-Profile oder fehlende AWS-Regionen erkennen.

### LH-ZIEL-004 — CI/CD-Fähigkeit

Das Tool soll in CI/CD-Pipelines verwendbar sein und bei kritischen Verstößen mit einem Fehlercode abbrechen.

### LH-ZIEL-005 — Entwicklerfreundliche Ausgabe

Das Tool soll klare, verständliche Ausgaben mit `PASS`, `WARN` und `FAIL` liefern.

---

## 4. Nicht-Ziele

### LH-NZ-001 — Kein rechtlicher Nachweis

Das Tool ersetzt keine rechtliche Prüfung, kein Datenschutzgutachten und keine vertragliche Zusicherung durch AWS oder Anthropic.

### LH-NZ-002 — Keine Garantie interner AWS-Routings

Das Tool kann nicht technisch beweisen, wie AWS intern Daten verarbeitet oder routet. Es prüft nur die sichtbaren und steuerbaren Konfigurationen.

### LH-NZ-003 — Keine automatische Änderung produktiver AWS-Konfigurationen

In der ersten Ausbaustufe soll das Tool nur prüfen und berichten. Es soll keine produktiven AWS-Ressourcen automatisch ändern.

### LH-NZ-004 — Kein vollständiger Cloud-Security-Scanner

Das Tool ist kein Ersatz für CSPM-Lösungen, AWS Security Hub, IAM Access Analyzer oder CloudTrail-Auswertungen.

---

## 5. Begriffe

### LH-BEG-001 — Claude Code

CLI-/Agentenwerkzeug von Anthropic zur Unterstützung bei Softwareentwicklung.

### LH-BEG-002 — AWS Bedrock

AWS-Service zum Aufruf verwalteter Foundation Models, einschließlich Anthropic Claude.

### LH-BEG-003 — Mantle

Amazon-Bedrock-Endpunkt für Claude Code, der Claude-Modelle über eine native Anthropic-API-Form bereitstellt, aber AWS-Anmeldedaten und IAM nutzt.

### LH-BEG-004 — Inference Profile

AWS-Bedrock-Profil zur Modellnutzung, potenziell regional, geo-/in-region oder global/cross-region.

### LH-BEG-005 — EU-safe Konfiguration

Eine Konfiguration, die ausschließlich zugelassene EU-Regionen, zugelassene Modell-/Inference-Profile und AWS Bedrock statt direkter Anthropic API verwendet.

---

## 6. Systemkontext

### LH-KONT-001 — Lokale Entwicklerumgebung

Das Tool soll auf Linux-Systemen und in Devcontainern ausführbar sein.

### LH-KONT-002 — AWS CLI

Das Tool darf die lokal konfigurierte AWS CLI verwenden, um AWS-Identität, Region, Bedrock-Profile und Berechtigungen zu prüfen.

### LH-KONT-003 — Terraform-Repositories

Das Tool soll Terraform-Dateien in Repositories scannen können.

### LH-KONT-004 — Claude-Code-Konfiguration

Das Tool soll Claude-Code-Konfigurationen in Benutzer- und Projektverzeichnissen prüfen können.

---

## 7. Betriebsmodi

### LH-MOD-001 — Lokaler Modus

Das Tool soll lokale Umgebungsvariablen, Claude-Code-Konfigurationen, Devcontainer-Konfigurationen und Docker-Konfigurationen prüfen.

Beispiel:

```bash
bedrock-eu-check local
```

### LH-MOD-002 — AWS-Modus

Das Tool soll AWS-Account, AWS-Region, STS-Identität, Bedrock-Inference-Profile und IAM-Berechtigungen prüfen.

Beispiel:

```bash
bedrock-eu-check aws
```

### LH-MOD-003 — Terraform-Modus

Das Tool soll Terraform-Dateien auf riskante AWS-Bedrock-, IAM- und Region-Konfigurationen prüfen.

Beispiel:

```bash
bedrock-eu-check terraform
```

### LH-MOD-004 — Gesamtprüfung

Das Tool soll alle verfügbaren Prüfungen in einem Lauf ausführen können.

Beispiel:

```bash
bedrock-eu-check all
```

### LH-MOD-005 — CI-Modus

Das Tool soll eine CI-optimierte Ausgabe und definierte Exit-Codes unterstützen.

Beispiel:

```bash
bedrock-eu-check all --ci
```

---

## 8. Funktionale Anforderungen

## 8.1 Lokale Umgebungsprüfung

### LH-FA-LENV-001 — Prüfung von `AWS_REGION`

Das Tool muss prüfen, ob `AWS_REGION` gesetzt ist.

Akzeptanzkriterium:

```text
PASS, wenn AWS_REGION auf eine erlaubte EU-Region gesetzt ist.
FAIL, wenn AWS_REGION fehlt.
FAIL, wenn AWS_REGION auf eine nicht erlaubte Region zeigt.
```

### LH-FA-LENV-002 — Prüfung von `AWS_DEFAULT_REGION`

Das Tool muss `AWS_DEFAULT_REGION` berücksichtigen, falls `AWS_REGION` nicht gesetzt ist.

Akzeptanzkriterium:

```text
PASS, wenn AWS_DEFAULT_REGION auf eine erlaubte EU-Region gesetzt ist.
WARN, wenn AWS_REGION fehlt, aber AWS_DEFAULT_REGION gesetzt ist.
FAIL, wenn beide fehlen.
```

### LH-FA-LENV-003 — Prüfung von `CLAUDE_CODE_USE_BEDROCK`

Das Tool muss prüfen, ob `CLAUDE_CODE_USE_BEDROCK=1` gesetzt ist.

Akzeptanzkriterium:

```text
PASS, wenn CLAUDE_CODE_USE_BEDROCK=1 gesetzt ist.
FAIL, wenn die Variable fehlt oder einen anderen Wert hat.
```

### LH-FA-LENV-004 — Prüfung von `ANTHROPIC_API_KEY`

Das Tool muss prüfen, ob `ANTHROPIC_API_KEY` gesetzt ist.

Akzeptanzkriterium:

```text
PASS, wenn ANTHROPIC_API_KEY nicht gesetzt ist.
FAIL, wenn ANTHROPIC_API_KEY gesetzt ist.
```

### LH-FA-LENV-005 — Prüfung von `CLAUDE_CODE_USE_MANTLE`

Das Tool muss erkennen, ob `CLAUDE_CODE_USE_MANTLE=1` gesetzt ist.

Akzeptanzkriterium:

```text
WARN, wenn Mantle aktiv ist und keine explizite Modell-/Region-Prüfung möglich ist.
PASS, wenn Mantle aktiv ist und alle zusätzlich konfigurierten Modell-IDs erlaubt sind.
```

### LH-FA-LENV-006 — Prüfung von Claude-Default-Modellvariablen

Das Tool muss folgende Variablen prüfen:

```text
ANTHROPIC_DEFAULT_SONNET_MODEL
ANTHROPIC_DEFAULT_HAIKU_MODEL
ANTHROPIC_DEFAULT_OPUS_MODEL
```

Akzeptanzkriterium:

```text
FAIL, wenn eine Modell-ID mit us. oder global. beginnt.
FAIL, wenn eine Modell-ID us-east-1, us-west-2 oder andere nicht erlaubte Regionen enthält.
PASS, wenn die Modell-ID einem erlaubten Muster entspricht.
WARN, wenn die Modell-ID nicht klassifizierbar ist.
```

---

## 8.2 Claude-Code-Konfigurationsprüfung

### LH-FA-CLAUDE-001 — Benutzerweite Claude-Code-Settings prüfen

Das Tool soll `~/.claude/settings.json` prüfen.

Akzeptanzkriterium:

```text
PASS, wenn Bedrock aktiviert und keine direkte Anthropic API konfiguriert ist.
FAIL, wenn direkte Anthropic-Konfigurationen erkannt werden.
WARN, wenn keine Datei vorhanden ist.
```

### LH-FA-CLAUDE-002 — Projektweite Claude-Code-Settings prüfen

Das Tool soll projektlokale Claude-Konfigurationen prüfen, insbesondere:

```text
.claude/settings.json
.claude/settings.local.json
```

Akzeptanzkriterium:

```text
PASS, wenn die Projektkonfiguration EU-safe ist.
FAIL, wenn nicht erlaubte Modelle, Regionen oder API-Keys erkannt werden.
```

### LH-FA-CLAUDE-003 — Erkennung direkter Anthropic-Endpunkte

Das Tool soll direkte Anthropic-Endpunkte erkennen.

Suchmuster:

```text
api.anthropic.com
ANTHROPIC_BASE_URL
ANTHROPIC_API_KEY
```

Akzeptanzkriterium:

```text
FAIL, wenn direkte Anthropic-Endpunkte oder API-Keys aktiv konfiguriert sind.
```

### LH-FA-CLAUDE-004 — Claude-Code-Wizard-Konfiguration prüfen

Das Tool muss erkennen, ob Claude Code durch den Bedrock-Login-Wizard eine Konfiguration in `~/.claude/settings.json` gespeichert hat.

Akzeptanzkriterium:

```text
PASS, wenn env.CLAUDE_CODE_USE_BEDROCK=1 gesetzt ist.
PASS, wenn env.AWS_REGION eine erlaubte EU-Region ist.
FAIL, wenn env.AWS_REGION fehlt oder auf us-east-1/us-west-2 zeigt.
FAIL, wenn direkte Anthropic-API-Konfigurationen parallel gesetzt sind.
```

---

## 8.3 Devcontainer- und Docker-Prüfung

### LH-FA-DEVCON-001 — Prüfung von `.devcontainer/devcontainer.json`

Das Tool soll Devcontainer-Konfigurationen auf relevante Environment-Variablen prüfen.

Akzeptanzkriterium:

```text
PASS, wenn AWS_REGION und CLAUDE_CODE_USE_BEDROCK korrekt gesetzt sind.
FAIL, wenn ANTHROPIC_API_KEY gesetzt wird.
WARN, wenn keine Claude-/Bedrock-Konfiguration vorhanden ist.
```

### LH-FA-DOCKER-001 — Prüfung von `docker-compose.yml`

Das Tool soll Docker-Compose-Dateien nach Bedrock-/Anthropic-Konfigurationen durchsuchen.

Akzeptanzkriterium:

```text
FAIL, wenn direkte Anthropic-API-Keys gesetzt werden.
FAIL, wenn nicht erlaubte AWS-Regionen gesetzt werden.
PASS, wenn nur erlaubte EU-Regionen verwendet werden.
```

### LH-FA-DOCKER-002 — Prüfung von `Dockerfile`

Das Tool soll Dockerfiles nach problematischen `ENV`-Definitionen durchsuchen.

Akzeptanzkriterium:

```text
FAIL, wenn ANTHROPIC_API_KEY per ENV gesetzt wird.
FAIL, wenn AWS_REGION auf eine nicht erlaubte Region gesetzt wird.
WARN, wenn keine Region gesetzt wird.
```

---

## 8.4 AWS-Identitätsprüfung

### LH-FA-AWS-001 — STS-Identität prüfen

Das Tool muss `aws sts get-caller-identity` ausführen können.

Akzeptanzkriterium:

```text
PASS, wenn Account-ID und Principal-ARN ermittelt werden können.
FAIL, wenn keine gültigen AWS-Credentials vorhanden sind.
```

### LH-FA-AWS-002 — AWS-Profil anzeigen

Das Tool soll das aktive AWS-Profil ausgeben.

Akzeptanzkriterium:

```text
PASS, wenn AWS_PROFILE gesetzt und verwendbar ist.
WARN, wenn kein AWS_PROFILE gesetzt ist.
```

### LH-FA-AWS-003 — Erwartete Account-ID prüfen

Das Tool soll optional eine erlaubte Account-ID prüfen.

Beispiel:

```bash
bedrock-eu-check aws --allowed-account 123456789012
```

Akzeptanzkriterium:

```text
PASS, wenn die aktuelle Account-ID erlaubt ist.
FAIL, wenn die aktuelle Account-ID nicht erlaubt ist.
```

### LH-FA-AUTH-001 — Authentifizierungsart klassifizieren

Das Tool soll erkennen, ob Claude Code über `AWS_PROFILE`, Environment-Credentials, Bedrock API Key oder statische Access Keys läuft.

Akzeptanzkriterium:

```text
PASS, wenn AWS_PROFILE/SSO/temporäre Credentials verwendet werden.
WARN, wenn Bedrock API Key verwendet wird.
WARN, wenn statische Access Keys verwendet werden.
FAIL, wenn ANTHROPIC_API_KEY verwendet wird.
```

---

## 8.5 Bedrock-Prüfung

### LH-FA-BEDROCK-001 — Bedrock-Inference-Profile auflisten

Das Tool soll verfügbare Bedrock-Inference-Profile mit der AWS CLI auflisten können.

Akzeptanzkriterium:

```text
PASS, wenn Bedrock-Inference-Profile gelesen werden können.
WARN, wenn keine Profile sichtbar sind.
FAIL, wenn die AWS API nicht erreichbar ist.
```

### LH-FA-BEDROCK-002 — Nicht erlaubte Profile erkennen

Das Tool muss Profile mit folgenden Mustern als kritisch erkennen:

```text
us.*
global.*
*us-east-1*
*us-west-2*
```

Akzeptanzkriterium:

```text
FAIL, wenn ein verwendetes oder explizit konfiguriertes Profil nicht erlaubt ist.
WARN, wenn ein nicht erlaubtes Profil nur sichtbar, aber nicht aktiv verwendet wird.
```

### LH-FA-BEDROCK-003 — Erlaubte EU-Profile prüfen

Das Tool soll erlaubte Profile anhand einer Allowlist prüfen.

Beispiel-Allowlist:

```yaml
allowedProfiles:
  - eu.*
  - eu-central-1.*
  - anthropic.claude-sonnet-4-6
```

Akzeptanzkriterium:

```text
PASS, wenn alle verwendeten Profile durch die Allowlist abgedeckt sind.
FAIL, wenn verwendete Profile nicht durch die Allowlist abgedeckt sind.
```

### LH-FA-BEDROCK-004 — Modellverfügbarkeit prüfen

Das Tool soll prüfen, ob die gewünschten Claude-Modelle in der konfigurierten Region verfügbar sind.

Akzeptanzkriterium:

```text
PASS, wenn das Modell in der Region verfügbar ist.
WARN, wenn die Verfügbarkeit nicht eindeutig ermittelt werden kann.
FAIL, wenn das Modell in der Region nicht verfügbar ist.
```

---

## 8.6 IAM-Prüfung

### LH-FA-IAM-001 — Bedrock-Rechte prüfen

Das Tool soll prüfen, ob der aktuelle Principal Bedrock-Modellaufrufe ausführen darf.

Relevante Actions:

```text
bedrock:InvokeModel
bedrock:InvokeModelWithResponseStream
bedrock:Converse
bedrock:ConverseStream
bedrock:ListInferenceProfiles
bedrock:GetInferenceProfile
```

Akzeptanzkriterium:

```text
PASS, wenn die benötigten Rechte vorhanden sind.
FAIL, wenn Invoke-Rechte fehlen.
WARN, wenn Discovery-Rechte fehlen.
```

### LH-FA-IAM-002 — Zu breite Bedrock-Policies erkennen

Das Tool soll IAM-Policies erkennen, die Bedrock ohne Regionseinschränkung erlauben.

Kritische Muster:

```json
{
  "Action": "bedrock:*",
  "Resource": "*"
}
```

Akzeptanzkriterium:

```text
FAIL, wenn bedrock:* auf Resource="*" ohne aws:RequestedRegion-Condition erlaubt ist.
WARN, wenn Resource="*" verwendet wird, aber eine Region-Condition vorhanden ist.
PASS, wenn Actions und Resources eng eingeschränkt sind.
```

### LH-FA-IAM-003 — Region-Condition prüfen

Das Tool soll prüfen, ob IAM-Policies `aws:RequestedRegion` verwenden.

Akzeptanzkriterium:

```text
PASS, wenn eine erlaubte EU-Region per Condition erzwungen wird.
FAIL, wenn Bedrock-Aufrufe ohne Region-Condition erlaubt sind.
```

---

## 8.7 Terraform-Prüfung

### LH-FA-TF-001 — Terraform-Dateien finden

Das Tool muss `.tf`-Dateien rekursiv im Repository finden.

Akzeptanzkriterium:

```text
PASS, wenn Terraform-Dateien gefunden und gelesen werden.
WARN, wenn keine Terraform-Dateien gefunden werden.
```

### LH-FA-TF-002 — AWS Provider Region prüfen

Das Tool soll AWS-Provider-Regionen in Terraform erkennen.

Akzeptanzkriterium:

```text
PASS, wenn nur erlaubte EU-Regionen verwendet werden.
FAIL, wenn nicht erlaubte Regionen verwendet werden.
WARN, wenn die Region dynamisch ist und nicht aufgelöst werden kann.
```

### LH-FA-TF-003 — Bedrock-IAM-Policies prüfen

Das Tool soll Terraform-IAM-Policy-Dokumente auf riskante Bedrock-Rechte prüfen.

Akzeptanzkriterium:

```text
FAIL, wenn bedrock:* oder InvokeModel auf "*" ohne Region-Condition erlaubt ist.
WARN, wenn "*" verwendet wird, aber eine Region-Condition vorhanden ist.
PASS, wenn die Policy eng eingeschränkt ist.
```

### LH-FA-TF-004 — Nicht erlaubte Modell-IDs finden

Das Tool soll Terraform-Dateien nach nicht erlaubten Modell-/Profil-IDs durchsuchen.

Akzeptanzkriterium:

```text
FAIL, wenn us.* oder global.* Claude-/Bedrock-Profil-IDs gefunden werden.
WARN, wenn unklassifizierbare Modell-IDs gefunden werden.
PASS, wenn nur erlaubte IDs gefunden werden.
```

---

## 8.8 Reporting

### LH-FA-REP-001 — Konsolenausgabe

Das Tool muss eine klare Konsolenausgabe erzeugen.

Format:

```text
PASS  AWS_REGION=eu-central-1
WARN  CLAUDE_CODE_USE_MANTLE=1; verify Mantle availability manually
FAIL  ANTHROPIC_API_KEY is set
```

### LH-FA-REP-002 — JSON-Ausgabe

Das Tool soll eine maschinenlesbare JSON-Ausgabe unterstützen.

Beispiel:

```bash
bedrock-eu-check all --format json
```

Akzeptanzkriterium:

```text
JSON enthält Status, Kennung, Nachricht, Schweregrad, Datei, Zeile und Empfehlung.
```

### LH-FA-REP-003 — SARIF-Ausgabe

Das Tool sollte optional SARIF für GitHub/GitLab-Code-Scanning unterstützen.

Beispiel:

```bash
bedrock-eu-check terraform --format sarif
```

Akzeptanzkriterium:

```text
SARIF-Datei kann von kompatiblen Code-Scanning-Systemen verarbeitet werden.
```

---

## 8.9 Exit-Codes

### LH-FA-EXIT-001 — Erfolgreiche Prüfung

Das Tool muss Exit-Code `0` zurückgeben, wenn keine `FAIL`-Findings existieren.

### LH-FA-EXIT-002 — Fehlgeschlagene Prüfung

Das Tool muss Exit-Code `1` zurückgeben, wenn mindestens ein `FAIL`-Finding existiert.

### LH-FA-EXIT-003 — Tool-Fehler

Das Tool muss Exit-Code `2` zurückgeben, wenn das Tool selbst nicht korrekt ausgeführt werden konnte.

---

## 9. Nichtfunktionale Anforderungen

### LH-NF-001 — Plattform

Das Tool muss unter Linux lauffähig sein.

### LH-NF-002 — Containerfähigkeit

Das Tool muss in Docker- und Devcontainer-Umgebungen ausführbar sein.

### LH-NF-003 — Geschwindigkeit

Eine lokale Prüfung ohne AWS API Calls soll bei typischen Repositories unter 5 Sekunden dauern.

### LH-NF-004 — Keine Secret-Ausgabe

Das Tool darf Secrets niemals im Klartext ausgeben.

Akzeptanzkriterium:

```text
ANTHROPIC_API_KEY wird nur als "set" oder "not set" gemeldet.
Der Wert wird nie ausgegeben.
```

### LH-NF-005 — Offline-Fähigkeit

Lokale und Terraform-Prüfungen müssen ohne Internet und ohne AWS-Zugriff funktionieren.

### LH-NF-006 — Erweiterbarkeit

Neue Regeln sollen ohne größere Architekturänderung ergänzt werden können.

### LH-NF-007 — Konfigurierbarkeit

Erlaubte Regionen, Accounts, Modell-IDs und Profile sollen über eine Projektkonfiguration steuerbar sein.

Beispiel:

```yaml
allowedRegions:
  - eu-central-1

allowedAccounts:
  - "123456789012"

allowedModelPatterns:
  - "^eu\\."
  - "^anthropic\\.claude-haiku-4-5$"
  - "^anthropic\\.claude-sonnet-4-6$"

blockedModelPatterns:
  - "^us\\."
  - "^global\\."
  - "us-east-1"
  - "us-west-2"
```

---

## 10. Sicherheitsanforderungen

### LH-SEC-001 — Keine Speicherung von AWS-Credentials

Das Tool darf keine AWS-Credentials speichern.

### LH-SEC-002 — Keine Speicherung von Claude-/Anthropic-Secrets

Das Tool darf keine Claude-/Anthropic-Secrets speichern.

### LH-SEC-003 — Read-only AWS-Zugriff

Das Tool soll in der ersten Version ausschließlich lesende AWS-Operationen verwenden.

### LH-SEC-004 — Least Privilege

Für den AWS-Prüfmodus sollen minimale IAM-Rechte dokumentiert werden.

### LH-SEC-005 — Keine automatischen Remediations

Automatische Korrekturen sollen in Version 1 nicht enthalten sein.

---

## 11. Qualitätsanforderungen

### LH-QA-001 — Unit Tests

Für jede Regel muss mindestens ein Unit Test existieren.

### LH-QA-002 — Integration Tests

AWS-Prüfungen sollen über mockbare Adapter getestet werden.

### LH-QA-003 — Testdaten

Es sollen Positiv- und Negativbeispiele für folgende Dateien existieren:

```text
.env
devcontainer.json
docker-compose.yml
Dockerfile
main.tf
iam-policy.json
.claude/settings.json
```

### LH-QA-004 — Reproduzierbarer Build

Das Tool soll per Docker reproduzierbar gebaut werden können.

### LH-QA-005 — CI-Prüfung

Pull Requests sollen Unit Tests, Linting und einen Self-Check ausführen.

---

## 12. Vorgeschlagene technische Umsetzung

### LH-TECH-001 — Programmiersprache

Bevorzugte Umsetzung: Python oder Go.

Empfehlung:

```text
Python für schnelle Entwicklung und flexible Scanner.
Go für einfache Verteilung als einzelnes Binary.
```

### LH-TECH-002 — CLI-Struktur

Vorgeschlagene Befehle:

```bash
bedrock-eu-check local
bedrock-eu-check aws
bedrock-eu-check terraform
bedrock-eu-check all
bedrock-eu-check version
```

### LH-TECH-003 — Konfigurationsdatei

Das Tool soll optional eine Datei lesen:

```text
bedrock-eu-check.yml
```

### LH-TECH-004 — Docker-Image

Das Tool soll als Docker-Image bereitgestellt werden können.

Beispiel:

```bash
docker run --rm -v "$PWD:/work" bedrock-eu-check all
```

### LH-TECH-005 — Devcontainer-Integration

Das Tool soll in einem Devcontainer als Feature oder vorinstalliertes Tool nutzbar sein.

---

## 13. Beispielausgabe

### LH-BSP-001 — Erfolgreiche Prüfung

```text
PASS  LH-FA-LENV-001  AWS_REGION=eu-central-1
PASS  LH-FA-LENV-003  CLAUDE_CODE_USE_BEDROCK=1
PASS  LH-FA-LENV-004  ANTHROPIC_API_KEY is not set
PASS  LH-FA-AWS-001   AWS account: 123456789012
PASS  LH-FA-TF-002    Terraform provider region is eu-central-1
```

### LH-BSP-002 — Fehlgeschlagene Prüfung

```text
FAIL  LH-FA-LENV-001  AWS_REGION is not set. Claude Code may fall back to us-east-1.
FAIL  LH-FA-LENV-004  ANTHROPIC_API_KEY is set. This may bypass AWS Bedrock.
FAIL  LH-FA-LENV-006  ANTHROPIC_DEFAULT_SONNET_MODEL uses us.* profile.
WARN  LH-FA-LENV-005  Mantle is enabled. Verify EU availability manually.
```

---

## 14. Priorisierung

## Muss-Anforderungen

### LH-PRI-MUSS-001

Lokale Umgebungsvariablen müssen geprüft werden.

### LH-PRI-MUSS-002

Direkte Anthropic-API-Nutzung muss erkannt werden.

### LH-PRI-MUSS-003

Nicht erlaubte AWS-Regionen müssen erkannt werden.

### LH-PRI-MUSS-004

Nicht erlaubte `us.*` und `global.*` Modell-/Inference-Profile müssen erkannt werden.

### LH-PRI-MUSS-005

Das Tool muss CI-fähige Exit-Codes liefern.

## Soll-Anforderungen

### LH-PRI-SOLL-001

Terraform-Dateien sollen geprüft werden.

### LH-PRI-SOLL-002

AWS-STS-Identität soll geprüft werden.

### LH-PRI-SOLL-003

Bedrock-Inference-Profile sollen geprüft werden.

### LH-PRI-SOLL-004

JSON-Ausgabe soll unterstützt werden.

## Kann-Anforderungen

### LH-PRI-KANN-001

SARIF-Ausgabe kann unterstützt werden.

### LH-PRI-KANN-002

AWS Organizations SCP-Prüfung kann ergänzt werden.

### LH-PRI-KANN-003

Automatische Remediation kann in einer späteren Version ergänzt werden.

---

## 15. Abnahmekriterien

### LH-ABN-001 — Kritische lokale Fehlkonfiguration

Gegeben:

```bash
export ANTHROPIC_API_KEY=secret
unset AWS_REGION
```

Erwartung:

```text
Tool meldet mindestens zwei FAIL-Findings.
Exit-Code ist 1.
Secret-Wert wird nicht ausgegeben.
```

### LH-ABN-002 — EU-konforme Minimalumgebung

Gegeben:

```bash
export AWS_REGION=eu-central-1
export CLAUDE_CODE_USE_BEDROCK=1
unset ANTHROPIC_API_KEY
```

Erwartung:

```text
Tool meldet PASS für Region, Bedrock-Nutzung und fehlenden Anthropic-Key.
Exit-Code ist 0, sofern keine weiteren FAIL-Findings vorliegen.
```

### LH-ABN-003 — Nicht erlaubtes Modellprofil

Gegeben:

```bash
export ANTHROPIC_DEFAULT_SONNET_MODEL=us.anthropic.claude-sonnet-4-6
```

Erwartung:

```text
Tool meldet FAIL wegen nicht erlaubtem us.* Profil.
Exit-Code ist 1.
```

### LH-ABN-004 — Terraform mit zu breiter IAM-Policy

Gegeben eine Terraform-Datei mit:

```hcl
actions   = ["bedrock:*"]
resources = ["*"]
```

ohne Region-Condition.

Erwartung:

```text
Tool meldet FAIL wegen zu breiter Bedrock-Berechtigung.
Exit-Code ist 1.
```

---

## 16. Risiken

### LH-RISK-001 — AWS-Modellverfügbarkeit ändert sich

AWS- und Anthropic-Modellverfügbarkeiten können sich ändern.

Maßnahme:

```text
Allow-/Blocklisten konfigurierbar machen.
Regelwerk regelmäßig aktualisieren.
```

### LH-RISK-002 — Mantle-Verfügbarkeit nicht eindeutig prüfbar

Mantle-Modelle können je Organisation unterschiedlich verfügbar sein.

Maßnahme:

```text
Mantle in Version 1 nur warnend prüfen.
Explizite Mantle-Allowlist unterstützen.
```

### LH-RISK-003 — False Positives bei Terraform

Dynamische Terraform-Ausdrücke können statisch schwer auswertbar sein.

Maßnahme:

```text
Unklare Fälle als WARN markieren.
Optional terraform plan JSON einlesen.
```

### LH-RISK-004 — Scheinbare Sicherheit

Das Tool kann eine falsche Sicherheit erzeugen.

Maßnahme:

```text
Dokumentieren, dass das Tool nur technische Konfigurationen prüft und keine rechtliche Garantie liefert.
```

---

## 17. Offene Punkte

### LH-OFF-001 — Exakte erlaubte AWS-Regionen

Es muss final entschieden werden, welche EU-Regionen erlaubt sind.

Vorschlag:

```text
eu-central-1 zunächst als einzige erlaubte Region.
```

### LH-OFF-002 — Erlaubte Claude-Modelle

Es muss final entschieden werden, welche Modelle erlaubt sind.

Vorschlag:

```text
Claude Sonnet 4.6 als Standard.
Claude Haiku 4.5 als günstige Option.
Opus nur manuell/fallweise.
```

### LH-OFF-003 — Mantle-Strategie

Es muss entschieden werden, ob Mantle erlaubt, verboten oder nur experimentell erlaubt wird.

Vorschlag:

```text
Mantle zunächst WARN, nicht FAIL.
Später explizite Allowlist.
```

### LH-OFF-004 — AWS-Account-Struktur

Es muss entschieden werden, welche AWS-Accounts für Entwicklung, Test und Produktion erlaubt sind.

---

## 18. Roadmap

### LH-RM-001 — Version 0.1

Lokaler Env-Check, Modell-ID-Check, Exit-Codes.

### LH-RM-002 — Version 0.2

Claude-Code-, Devcontainer-, Docker- und Terraform-Scanner.

### LH-RM-003 — Version 0.3

AWS-STS-, Bedrock- und IAM-Prüfungen.

### LH-RM-004 — Version 0.4

JSON-/SARIF-Ausgabe und CI-Integration.

### LH-RM-005 — Version 1.0

Stabile Team-Version mit Dokumentation, Tests, Docker-Image und Release-Artefakten.

---

## 19. Zusammenfassung

Das geplante Tool `bedrock-eu-check` soll Fehlkonfigurationen beim Betrieb von Claude Code über AWS Bedrock EU frühzeitig erkennen. Der Fokus liegt auf lokaler Entwicklerkonfiguration, AWS-Regionen, Bedrock-/Claude-Modellprofilen, direkter Anthropic-API-Nutzung, Terraform/IAM-Risiken und CI/CD-Fähigkeit.

Das Tool ersetzt keine juristische Prüfung, ist aber ein wichtiger technischer Kontrollmechanismus, um versehentliche Nutzung von US-/Global-Profilen, direkter Anthropic API oder unkontrollierten Bedrock-Rechten zu verhindern.

Die harten Kernpunkte sind: **direkte Anthropic API blockieren, Region explizit setzen, `us.*`/`global.*` Profile erkennen, Terraform/IAM scannen und in CI mit Exit-Code `1` abbrechen.**

---

## 20. Historie

| Version | Datum | Änderung | Verweis |
|---|---|---|---|
| 0.1.0 | 2026-06-24 | Initiale Fassung; Übernahme aus der Projektidee als vertragliches Lastenheft. `LH-FA-CLAUDE-004` (→ §8.2) und `LH-FA-AUTH-001` (→ §8.4) aus dem Roadmap-Anhang in ihre Fachsektionen einsortiert. | Harness-Bootstrap |

> **Referenz-Disziplin (SDP).** Dieses Lastenheft referenziert *nicht*
> abwärts auf ADRs oder Slices; die Begründung „warum so" lebt in den
> aufwärts zeigenden ADRs (`docs/plan/adr/`). Provenance ausschließlich in
> dieser Historie-Tabelle.
