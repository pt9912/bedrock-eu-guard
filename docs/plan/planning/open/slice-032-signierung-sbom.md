# Slice slice-032: Signierte Releases und SBOM

**Status:** open (geplant, nicht priorisiert). **Verantwortlich:** w2 (Plan-Review: w1). **Priorisierung:** steht aus; der Slice startet erst nach der Entscheidung der Auftraggeber (Abschnitt 3).

**Welle:** [`welle-08-erweiterungen`](../welle-08-erweiterungen.md).

**Bezug:** [`docs/user/releasing.md`](../../../user/releasing.md) (Prüfsummen schützen nicht vor einem kompromittierten Ursprung; „Nicht Teil von 1.0: signierte Releases, SBOM“), [`slice-025`](../done/slice-025-release-create.md) (`release-create`), [`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md) (Abhängigkeiten unter `vendor/`).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Jedes Release trägt eine nachprüfbare Herkunft: Nutzer können feststellen, dass Binaries und Images aus diesem Repository und aus dem dafür vorgesehenen Workflow stammen, und nicht nur, dass sie zu einer Prüfsumme passen, die vom selben Ursprung kommt. Zusätzlich liegt eine Stückliste (SBOM) der Abhängigkeiten bei.

## 2. Ausgangslage

- Das Release enthält Binaries und `SHA256SUMS`; die Images liegen auf `ghcr.io`, das Feature als OCI-Artefakt. Nichts davon ist signiert oder mit einer SBOM versehen.
- `releasing.md` und `devcontainer.md` benennen die Grenze ausdrücklich: Prüfsummen schützen vor Korruption, nicht vor einem kompromittierten Ursprung.
- Die Veröffentlichung läuft über `release-create` (Schreibrecht nur im Job `create`, Umgebung `release`) und `publish` (`packages: write`). Eine Signatur braucht ein zusätzliches Recht (`id-token: write`), also eine **neue Fähigkeit** der Workflows.
- Die Abhängigkeiten liegen unter `vendor/` ([`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md)); `THIRD_PARTY_NOTICES.md` listet sie bereits.

## 3. Entscheidungen der Auftraggeber (vor dem Start)

Der Slice startet erst, wenn **beide Nutzer** Priorisierung und Festlegungen bestätigt haben. Vorschlag in Klammern; die Empfehlung ist die des Autors und offen.

1. **Verfahren** (Vertrauensmodell).
   - **a) GitHub Artifact Attestations:** eine Aktion, Prüfung mit `gh attestation verify`; Rechte `id-token: write` und `attestations: write`; Nachweise liegen bei GitHub. Prüfende brauchen die `gh`-CLI. **Zu klären im Slice:** ob die Attestierung bei einem öffentlichen Repository ebenfalls in ein öffentliches, nicht löschbares Log schreibt (nicht belegt).
   - **b) Sigstore mit `cosign` (keyless, GitHub-OIDC):** Signatur der Prüfsummen-Datei und der Binaries (`sign-blob`) sowie der Image-Digests (`sign`); Prüfung mit `cosign verify-blob` gegen die Identität des Workflows. Die Einträge landen im **öffentlichen Transparenzlog** und sind **nicht löschbar**. (Empfehlung: b, weil `releasing.md` Sigstore als Folgepunkt nennt und die Prüfung nicht an ein GitHub-Konto gebunden ist.)
   - **c) GPG mit eigenem Schlüssel:** Schlüsselverwaltung bei den Nutzern, ein Geheimnis im Repository; nicht empfohlen.
2. **Umfang.** (Binaries und `SHA256SUMS` sowie die beiden Images; das Feature in einem späteren Schritt.)
3. **SBOM.** Format (CycloneDX JSON oder SPDX JSON) und Erzeuger aus den Modul- und `vendor/`-Daten; Veröffentlichung als Release-Asset. Der Erzeuger ist ein Werkzeug in der Lieferkette: Es soll offline im gepinnten Image laufen können ([`ADR-0004`](../../adr/0004-abhaengigkeiten-und-offline-gates.md)) oder als gepinnte Aktion mit Begründung. (CycloneDX JSON für das Binary; Erzeuger offline im gepinnten Image; SBOM der Images optional später.)
4. **Rechte.** `id-token: write` nur im Job, der signiert; keine weitere Veröffentlichungsfähigkeit. Die Erweiterung der Workflows um diese Fähigkeit braucht das Ja beider Nutzer, wie bei `release-create`.
5. **Nachweis der Entscheidung.** Eine neue ADR (Proposed, Accepted nach Sign-off beider Nutzer) hält Verfahren, Identität, Umfang und die Unwiderruflichkeit der Transparenzlog-Einträge fest.

## 4. Definition of Done

**Teil 1: SBOM, ADR und Vorbereitung** (ohne Signatur, ohne Veröffentlichung)

- [ ] Beide Nutzer haben Priorisierung und die Festlegungen aus Abschnitt 3 bestätigt; die ADR ist `Accepted`.
- [ ] Die SBOM wird erzeugt (lokal oder in CI erprobbar), ist als Release-Asset vorgesehen und im Trockenlauf von `release-create` sichtbar; der Erzeuger ist gepinnt und aktuell (Alter gegen das aktuelle Release geprüft).
- [ ] Die Workflows sind für die Signatur vorbereitet (Rechte `id-token: write` nur im signierenden Job, Schritte vorhanden), signieren im **Trockenlauf aber nicht** (jede Signatur ist ein dauerhafter öffentlicher Eintrag).
- [ ] Entwurf der Prüfanleitung (Befehle mit der erwarteten Identität) in `releasing.md`, als „noch nicht an einem veröffentlichten Artefakt erprobt“ gekennzeichnet.
- [ ] `make verify-clean` und CI grün; Review-Report durch w1; Closure-Notiz.

**Teil 2: erste Signatur und Prüfanleitung** (nur mit einem freigegebenen Release; eigener Schritt mit dem Ja beider Nutzer zu Version, Commit und Zielen)

- [ ] Der scharfe Lauf eines freigegebenen Releases signiert die vereinbarten Artefakte.
- [ ] Die Prüfanleitung ist an dem veröffentlichten Artefakt erprobt (frischer Download, Prüfung gelingt; eine manipulierte Datei fällt durch) und die Kennzeichnung „nicht erprobt“ entfällt; README und `devcontainer.md` nennen die Grenze „Prüfsumme vom selben Ursprung“ entsprechend.
- [ ] Closure-Notiz für Teil 2 (Ergänzung der Closure-Notiz) mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Zuständig |
|---|---|---|
| neue ADR zum Signierverfahren und Index | neu | w2 (Review w1) |
| `.github/workflows/release-create.yml`, `publish.yml` | update | w2 |
| SBOM-Erzeugung (Make-Target oder Workflow-Schritt) | neu | w2 |
| `docs/user/releasing.md`, README, Prüfanleitung | update/neu | w2 |

## 6. Trigger

Entscheidung beider Nutzer (Abschnitt 3). Voraussetzung ist erfüllt: `release-create` (slice-025) liegt vor.

## 7. Closure-Trigger

Teil 1: DoD des Teils 1 + PR gemerged + Closure-Notiz; der Slice kann damit abgeschlossen werden, Teil 2 folgt mit dem nächsten freigegebenen Release (eigener Schritt, Ja beider Nutzer zu Version, Commit und Zielen) und ergänzt die Closure-Notiz.

## 8. Risiken und offene Punkte

- **Nicht rückholbar:** Transparenzlog-Einträge (Verfahren b) sind öffentlich und dauerhaft; Tests mit echter Signatur sind zu vermeiden.
- **Neue Fähigkeit:** Das OIDC-Recht erweitert, was die Workflows tun dürfen; es gehört in dieselbe Freigabepraxis wie `release-create`.
- **Werkzeug von außen:** Aktionen und Signierwerkzeuge sind Fremdcode in der Veröffentlichungskette; Pins per SHA und regelmäßige Aktualitätsprüfung sind Pflicht (Lehre aus slice-034).
- **Frühere Releases** (1.0.0 bis 1.2.0) bleiben unsigniert; eine nachträgliche Signatur der Prüfsummen-Dateien ist möglich, aber eine eigene Entscheidung.
- Eine Signatur belegt Herkunft, nicht Fehlerfreiheit; die Grenzen des Werkzeugs (kein Rechtsnachweis) bleiben.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung der bestehenden Release-Workflows um Signatur und SBOM, keine neue Sub-Area.
