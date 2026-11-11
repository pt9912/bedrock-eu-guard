# Slice slice-024: Devcontainer-Feature mit der devcontainer-CLI erproben und korrigieren

**Status:** open.

**Welle:** [`welle-06-release-haertung`](../welle-06-release-haertung.md).

**Bezug:** [`LH-TECH-005`](../../../../spec/lastenheft.md#lh-tech-005--devcontainer-integration), [`welle-05-results.md`](../done/welle-05-results.md) (bekannter Mangel im veröffentlichten Feature).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Das Devcontainer-Feature wird mit der echten `devcontainer`-CLI installiert und geprüft. Gefundene Mängel werden behoben, darunter die veraltete Option-Beschreibung. Das Ergebnis erscheint als neue **Feature-Version** (Vorschlag 1.0.1); das Werkzeug bleibt 1.0.0.

## 2. Definition of Done

- [ ] CI-Job (Netz, nicht Teil von `make gates`), der das veröffentlichte Feature mit der `devcontainer`-CLI (Version gepinnt) in einen Beispiel-Container baut und `bedrock-eu-check version` ausführt.
- [ ] Mängel aus der Erprobung behoben; die Option-Beschreibung „zeigt auf ein Release, das noch nicht existiert“ ist korrigiert.
- [ ] `feature-test` (lokal, `file://`) bleibt grün; `devcontainer.md` nennt den erprobten Weg und entfernt „nicht erprobt“.
- [ ] Neue Feature-Version in `devcontainer-feature.json`; Veröffentlichung über den Workflow `publish` (Schalter `feature`) **nur nach dem Ja beider Nutzer** zu Feature-Version, Commit und Ziel.
- [ ] `make verify-clean` und CI grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.github/workflows/feature-cli.yml` | neu | Erprobung mit der echten CLI, eigener Job mit Netz |
| `devcontainer-feature/src/bedrock-eu-check/devcontainer-feature.json` | update | Option-Beschreibung, Feature-Version |
| `docs/user/devcontainer.md`, `README.md` | update | erprobter Weg |

## 4. Trigger

slice-025 done (nutzt `publish` und den gehärteten Ablauf). Mängel werden gebündelt, damit das Feature nur einmal neu veröffentlicht wird.

## 5. Closure-Trigger

DoD vollständig + PR gemerged + Closure-Notiz geschrieben.

## 6. Risiken und offene Punkte

- Die Neuveröffentlichung des Features ist nicht rückholbar (Tags `1`, `1.0`, `latest` wandern); nur mit Ja beider Nutzer.
- Die `devcontainer`-CLI braucht Container-Laufzeit und Netz; die Version der CLI wird gepinnt (Runner-Abhängigkeit).
- Feature-Version und Tool-Version sind entkoppelt; der Default `version` des Features bleibt 1.0.0.

## 7. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 8. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung des bestehenden Release- und CI-Ablaufs ohne neue Sub-Area.
