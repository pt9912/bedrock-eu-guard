# Slice slice-033: Binaries für macOS und Windows (Spike zuerst)

**Status:** open (geplant, nicht priorisiert). **Verantwortlich:** w2 (Plan-Review: w1). **Priorisierung:** steht aus; der Slice startet erst nach der Entscheidung der Auftraggeber (Abschnitt 3). **Empfehlung beider Agenten (Stand 2026-11-11):** zurückstellen; bei Bedarf nur der Spike aus Abschnitt 4, ohne Veröffentlichung.

**Welle:** [`welle-08-erweiterungen`](../welle-08-erweiterungen.md).

**Bezug:** [`LH-NF-001`](../../../../spec/lastenheft.md#lh-nf-001--plattform) (das Tool muss unter Linux lauffähig sein), [`docs/user/releasing.md`](../../../user/releasing.md) („Andere Plattformen (macOS, Windows) werden nicht gebaut und nicht getestet“), [`LH-NF-002`](../../../../spec/lastenheft.md#lh-nf-002--containerfähigkeit).

**Autor:** w2 (Projektteam bedrock-eu-guard). **Datum:** 2026-11-11.

---

## 1. Ziel

Das Werkzeug wird zusätzlich als Binary für macOS und Windows bereitgestellt, mit klarer Aussage, was dort getestet ist. Zuerst wird in einem **Spike** gemessen, ob und mit welchem Aufwand das Werkzeug dort überhaupt wie unter Linux arbeitet; ein Support-Versprechen gibt es erst nach der Entscheidung der Auftraggeber.

## 2. Ausgangslage

- Das Lastenheft verlangt nur Linux ([`LH-NF-001`](../../../../spec/lastenheft.md#lh-nf-001--plattform)). Weitere Plattformen wären eine Erweiterung, für die die Auftraggeber eine neue Anforderung mit Wortlaut festlegen müssten (wie bei [`LH-FA-IAM-004`](../../../../spec/lastenheft.md#lh-fa-iam-004--iam-policy-dokumente-json-prüfen)).
- `make dist` baut nur `linux/amd64` und `linux/arm64`; Gates, Tests und Smoke-Tests laufen in einem Linux-Container. Auf macOS und Windows ist das Werkzeug nie gelaufen.
- Das Binary wird mit `CGO_ENABLED=0` gebaut; ein Cross-Build mit `GOOS=darwin` und `GOOS=windows` ist grundsätzlich möglich, sein Lauf aber nicht auf dem Linux-Rechner prüfbar.
- Bekannte plattformabhängige Stellen: Pfade und Trenner (die Ausgabe nutzt `/`), Symlinks (die Suche folgt keinen), das atomare Ersetzen der Ausgabedatei (`--output`), das Home-Verzeichnis (`~/.claude/settings.json`), Zeilenenden, Dateirechte, der Aufruf von `aws` über `os/exec`, und die Tests mit Fake-`aws` als Shell-Skript.
- Das Docker-Image deckt die lokale Nutzung auf macOS und Windows über Docker Desktop ab ([`docker.md`](../../../user/docker.md)).

## 3. Entscheidungen der Auftraggeber (vor dem Start)

Vorschlag in Klammern; die Empfehlung ist die des Autors und offen.

1. **Ob überhaupt.** (Zurückstellen; nur bei konkretem Bedarf.) Das Docker-Image ist heute der unterstützte Weg außerhalb von Linux.
2. **Spike ohne Veröffentlichung.** (Ja, falls Bedarf besteht.) Er verändert weder Release noch Lastenheft und kostet nur CI-Zeit.
3. **Neue Anforderung** für ein Support-Versprechen (Nummer und Wortlaut legen die Auftraggeber fest; nur nach dem Spike). Vorschlag für den Inhalt: „Das Tool soll zusätzlich als Binary für macOS und Windows bereitgestellt werden; die Prüfungen verhalten sich dort wie unter Linux.“
4. **Zielplattformen** (nur nach dem Spike): `darwin/arm64`, `darwin/amd64`, `windows/amd64`; `windows/arm64` optional.
5. **Aussage zur Erprobung.** Zwischen „gebaut“, „rauchgetestet“ (Version, Prüfung der Testdaten, Exit-Codes) und „vollständig getestet“ (die ganze Testsuite) unterscheiden; Vorschlag: nur dann „vollständig“, wenn die Suite dort grün läuft.
6. **Signierung und Notarisierung.** Unsignierte Binaries lösen auf macOS (Gatekeeper) und Windows (SmartScreen) Warnungen aus; eine Notarisierung braucht ein Apple-Entwicklerkonto und Geheimnisse und ist nicht Teil dieses Slices. Die Signatur der Release-Artefakte regelt [`slice-032`](slice-032-signierung-sbom.md).

## 4. Definition of Done

**Teil 1: Spike** (veröffentlicht nichts; ohne Lastenheft-Änderung)

- [ ] Ein Workflow auf GitHub-Runnern `macos-latest` und `windows-latest` (nur Leserechte) baut das Werkzeug und führt `version` sowie die Prüfung der Testdaten `good`/`bad` aus (Exit-Codes `0` und `1`) und versucht die Testsuite.
- [ ] Ergebnisbericht in der Closure-Notiz: was läuft, was scheitert und warum (Pfade, Shell-Skripte, Symlinks, Rechte), mit geschätztem Aufwand für Anpassungen und Wartung (CI-Runner).
- [ ] Empfehlung an die Auftraggeber, ob Teil 2 sinnvoll ist.

**Teil 2: Bereitstellung** (nur nach Entscheidung und neuer Anforderung)

- [ ] Neue Anforderung im Lastenheft mit bestätigtem Wortlaut; [`LH-NF-001`](../../../../spec/lastenheft.md#lh-nf-001--plattform) bleibt für Linux gültig.
- [ ] `make dist` erzeugt die Binaries der vereinbarten Plattformen (Namensschema `bedrock-eu-check_<version>_<os>_<arch>`, unter Windows mit `.exe`); `SHA256SUMS` enthält sie; die Reproduzierbarkeit ist belegt.
- [ ] Plattformabhängige Stellen sind angepasst und getestet; die CI-Matrix hält die zugesagte Erprobung (Smoke-Test oder Suite) dauerhaft ein.
- [ ] Dokumentation: `releasing.md`, README, Handbuch (Bereitstellung, Grenzen, Gatekeeper/SmartScreen).
- [ ] `make verify-clean` und CI grün; Review-Report durch den anderen Agenten.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 5. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Teil |
|---|---|---|
| `.github/workflows/platforms-spike.yml` (nur Lesen, `workflow_dispatch`) | neu | 1 |
| `spec/lastenheft.md` (neue Anforderung) | update nach Wortlaut der Auftraggeber | 2 (w1) |
| `Makefile` (`dist`), Release-Workflows | update | 2 |
| Adapter und Tests bei Portabilitätsfunden | update | 2 |
| `docs/user/releasing.md`, README, Handbuch | update | 2 |

## 6. Trigger

Entscheidung beider Nutzer (Abschnitt 3, Nr. 1 und 2). Teil 2 zusätzlich: Ergebnis des Spikes und bestätigter Wortlaut der Anforderung.

## 7. Closure-Trigger

Teil 1: DoD des Spikes + Ergebnisbericht in der Closure-Notiz; der Slice kann damit abgeschlossen werden, auch wenn Teil 2 nicht folgt. Teil 2: DoD vollständig + PR gemerged + Closure-Notiz.

## 8. Risiken und offene Punkte

- **Versprechen ohne Nachprüfung:** Ein Binary, das nie dort lief, weckt eine Erwartung; deshalb zuerst der Spike und eine ausdrückliche Aussage zur Erprobung.
- **Wartungskosten:** Dauerhafte CI-Runner für macOS und Windows kosten Laufzeit und Pflege (Runner-Images wechseln).
- **Portabilität der Tests:** Die vorhandene Testsuite nutzt POSIX-Shell-Skripte; unter Windows sind Anpassungen wahrscheinlich.
- **Sicherheitsaussagen:** Pfad- und Rechte-Logik der Secret-Regeln ([`LH-NF-004`](../../../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)) darf auf anderen Plattformen nicht schwächer werden.
- **Sicherheitsmeldungen des Betriebssystems** bei unsignierten Binaries sind Nutzern zu erklären, nicht zu umgehen.

## 9. Closure-Notiz (nach `done/`)

<!-- Erst nach Abschluss füllen. -->

## 10. Sub-Area-Modus-Begründung

Alle berührten Sub-Areas **GF**: Erweiterung von Build und CI; Anpassungen im Code nur bei Portabilitätsfunden.
