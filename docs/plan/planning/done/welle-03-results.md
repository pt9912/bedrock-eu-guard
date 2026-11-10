# Welle welle-03-aws-checks: Ergebnisse

**Abgeschlossen:** 2026-11-10. **Meilenstein:** M3 — AWS-Live-Checks.

## Closure-Trigger

- Alle Slices done: slice-008 bis slice-010.
- `make verify-clean`, `make deps-verify` und CI grün auf `main`, ohne Netz und ohne AWS-Zugang (Fake-`aws`).
- `bedrock-eu-check aws` hat je Regel ein Positiv- und ein Negativbeispiel in den Tests; ohne Credentials meldet der Modus FAIL (Exit 1), ohne `aws`-Binary ein Tool-Fehler (Exit 2); die lokalen Regeln laufen bei `all` weiter.
- Kein Credential-Wert und keine CLI-Ausgabe in Meldungen ([`LH-SEC-001`](../../../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials), [`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)), per Test geprüft.
- Die Mindestrechte des Werkzeugs sind dokumentiert (`docs/user/aws-permissions.md`, [`LH-SEC-004`](../../../../spec/lastenheft.md#lh-sec-004--least-privilege)).

## Lieferumfang

Befehle `aws` und `all` (`local` + `terraform` + `aws`), `--allowed-account`; Adapter `awscli` (STS, Bedrock, IAM) hinter drei kleinen Schnittstellen (`AWSReader`, `BedrockReader`, `IAMReader`); Regeln für Identität, Profil, Konto, Authentifizierungsart, Bedrock-Profile, Verfügbarkeit und IAM; gemeinsames Paket `iampolicy`.

## Steering-Loop

1. Eine Fähigkeit, eine kleine Schnittstelle: neue Quellen additiv hinzufügen, statt bestehende Ports zu erweitern.
2. Externe Prozesse hinter einem Port mit Fake-Skript testen; stderr dient nur der Klassifikation und erscheint nie in Meldungen.
3. Gemeinsame Auslegung (Policy-Verflachung) einmal bauen und von allen Adaptern nutzen lassen.
4. Closure- und Doku-PRs sofort mergen, sobald die CI grün ist, damit `main` den Stand widerspiegelt.

## Folgeaufgaben

- „Verwendet“ auch aus Claude-Settings und Konfigurationsdateien (Zusammenführung bei `all`), Nicht-EU-Probing für sichtbare Profile.
- SCP-/Permission-Boundary-Näherung der Simulation dokumentieren oder verbessern; weitere Condition-Schlüssel.
- `apiKeyHelper` (Spec-Präzisierung zu CLAUDE-003).
- Erprobung des Adapters und der Beispiel-Policy gegen ein echtes AWS-Konto.
