# Planning — bedrock-eu-guard

**Derivativ:** dokumentiert die Konvention; Quelle der Wahrheit sind die
Dateien in den Verzeichnissen selbst.

Slice-Lifecycle: `open/` → `next/` → `in-progress/` → `done/`.

Reine `git mv`-Commits beim Wechsel zwischen Verzeichnissen — siehe Hard
Rule „git mv + Inhaltsänderung = zwei Commits" in
[`../../../AGENTS.md`](../../../AGENTS.md) §3.3.

## Lifecycle-Bedeutungen

| Verzeichnis | Bedeutung |
|---|---|
| `open/` | Geplant, noch nicht priorisiert. Keine Garantie auf Umsetzung. |
| `next/` | Als Nächstes priorisiert. Verantwortlicher zugeordnet. |
| `in-progress/` | Branch / PR existiert. |
| `done/` | DoD erfüllt, gemerged, Closure-Notiz vorhanden. |

## Slices vs. Wellen — zwei Status-Mechanismen

- **Slices** tragen ihren Status über das **Verzeichnis** (open → … → done).
- Eine **Welle** wird **in der Roadmap** geführt
  ([`in-progress/roadmap.md`](in-progress/roadmap.md)); ihr Status lebt im
  `Status:`-Feld des Welle-Plans, der **flach** in `planning/`
  (`<welle-id>.md`) liegt. Welle-Closure: Lerneintrag in
  `done/<welle-id>-results.md`.

## Vorlagen

- Slice: [`slice.template.md`](slice.template.md)
- Welle: [`welle.template.md`](welle.template.md)

## Roadmap

Siehe [`in-progress/roadmap.md`](in-progress/roadmap.md).
