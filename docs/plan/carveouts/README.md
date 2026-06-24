# Carveouts — bedrock-eu-guard

**Derivativ:** Quelle der Wahrheit sind die Carveout-Dateien; bei jedem
neuen/aufgelösten Carveout mitziehen.

Aktive Carveouts mit Auflösungs-Trigger. Aufgelöste Carveouts wandern nach
`done/` (reiner `git mv`).

## Aktive Carveouts

(noch keine)

> Der einzige reale Gate beim Bootstrap — `make doc-check` — ist grün,
> nicht strukturell rot; daher kein Carveout nötig. Sobald ein Gate
> dauerhaft rot wird (z. B. ein bewusst noch nicht erfüllter
> Coverage-Schwellwert), bekommt es hier einen `CO-<NNN>` mit Trigger und
> Folge-Slice, und die Sensors-Tabelle in `harness/README.md` verweist
> über die Bindung-Spalte darauf.

## Aufgelöste Carveouts

(noch keine)

## Konventionen

- Jeder aktive Carveout braucht: Trigger, Folge-Slice, letzten Prüf-Termin.
- Bei Welle-Closure: Carveout-Audit zwingend — welche gültig, welche aufgelöst?
- Vorlage: [`carveout.template.md`](carveout.template.md).
- Siehe [Kurs Modul 7](https://github.com/pt9912/ai-harness-course/blob/v1.4.0/kurs/de/02-planung/modul-07-carveouts.md).
