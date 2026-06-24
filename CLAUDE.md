# CLAUDE.md — Einstieg für Claude Code

Dieses Repo wird über ein generisches Agent-Briefing gesteuert. Die
**kanonische Quelle für jede AI-Session** (Hard Rules, Source Precedence,
Quality Gates, Agent-Workflow) ist [`AGENTS.md`](AGENTS.md) — diese Datei
ist nur der Claude-Code-Einstiegspunkt und dupliziert deren Inhalt
bewusst **nicht**, um Drift zu vermeiden.

## Bevor du etwas änderst

1. [`AGENTS.md`](AGENTS.md) lesen — insbesondere die Harten Regeln (§3)
   und den Minimal Agent Workflow (§6).
2. Source Precedence beachten: bei Konflikt gilt die kanonische Quelle,
   nicht dieses Briefing (Reihenfolge in [`AGENTS.md`](AGENTS.md) §2).
3. Struktur-, ID- und Konventionsregeln stehen in
   [`harness/conventions.md`](harness/conventions.md); der Harness-Einstieg
   in [`harness/README.md`](harness/README.md).

## Gates vor dem Handoff

Quality-Gates laufen über `make` (Details und reale Targets in
[`AGENTS.md`](AGENTS.md) §4):

```bash
make gates      # alle aktuell lauffaehigen Gates (heute nur doc-check)
make doc-check  # Doku-Referenzen pruefen (Links/Anker/IDs, netzlos)
```

Keine Erfolgsmeldung ohne ausgeführten Gate-Lauf.
