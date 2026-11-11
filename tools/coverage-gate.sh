#!/bin/sh
# Coverage-Gate (slice-031, ADR-0005): misst die Testabdeckung paketübergreifend
# (-coverpkg) über die Produktionspakete und bricht unter dem Mindestwert ab.
# Läuft im gepinnten Go-Image ohne Netz (siehe Makefile: coverage-gate).
#
# Aufruf: tools/coverage-gate.sh <mindestwert-in-prozent>
#
# Ausgeschlossen (Begründung in der ADR): internal/hexagon/coretest/* (Test-Doubles)
# und cmd/ (nur main).
set -eu

# Untergrenze aus ADR-0005: Der Mindestwert darf nicht darunter liegen. Eine Aenderung
# dieser Zahl ist eine neue ADR (AGENTS.md 3.6), kein gewoehnlicher PR.
ADR_FLOOR=90

min="${1:?usage: coverage-gate.sh <mindestwert-in-prozent>}"
# Strikt: Ganzzahl oder Dezimalzahl mit einem Punkt, nicht ueber 100.
if ! printf '%s\n' "$min" | grep -Eq '^[0-9]+(\.[0-9]+)?$' || \
   ! awk -v m="$min" 'BEGIN { exit !(m + 0 <= 100) }'; then
  echo "coverage-gate: ungueltiger Mindestwert '$min'" >&2
  exit 2
fi
if ! awk -v m="$min" -v f="$ADR_FLOOR" 'BEGIN { exit !(m + 0 >= f + 0) }'; then
  echo "coverage-gate: Mindestwert $min % liegt unter der ADR-Untergrenze $ADR_FLOOR % (ADR-0005)" >&2
  exit 2
fi

profile=.tmp/coverage.out
mkdir -p .tmp

pkgs=$(go list ./... | grep -v -e '/hexagon/coretest' -e '/cmd/' | paste -sd, -)
[ -n "$pkgs" ] || { echo "coverage-gate: keine Pakete gefunden" >&2; exit 2; }

# Testausgabe nur bei Fehlern zeigen (sonst stumm), damit ein Testfehler nicht ohne Meldung endet.
if ! go test -coverpkg="$pkgs" -coverprofile="$profile" ./... >.tmp/coverage-test.log 2>&1; then
  cat .tmp/coverage-test.log >&2
  echo "coverage-gate: Tests fehlgeschlagen, Abdeckung nicht bewertet" >&2
  exit 1
fi

# Je Paket: Anweisungen insgesamt und abgedeckt (Blöcke einmal zählen, höchster Zähler gewinnt).
echo "Abdeckung je Paket (paketübergreifend gemessen):"
awk -F'[: ,]' 'NR > 1 {
  key = $1 ":" $2 "," $3; stmts = $4; cnt = $5
  if (!(key in n) || cnt > c[key]) { c[key] = cnt }
  n[key] = stmts; f[key] = $1
}
END {
  for (k in n) {
    p = f[k]; sub(/\/[^\/]*$/, "", p)
    tot[p] += n[k]; if (c[k] > 0) cov[p] += n[k]
  }
  for (p in tot) printf "  %6.1f %%  %s\n", 100 * cov[p] / tot[p], p
}' "$profile" | sort -k3

total=$(go tool cover -func="$profile" | awk '/^total:/ { sub(/%/, "", $NF); print $NF }')
[ -n "$total" ] || { echo "coverage-gate: Gesamtwert nicht ermittelt" >&2; exit 2; }
echo "Gesamt: ${total} % (Mindestwert: ${min} %)"

if awk -v t="$total" -v m="$min" 'BEGIN { exit !(t + 0 >= m + 0) }'; then
  echo "coverage-gate: OK"
else
  echo "coverage-gate: FEHLER, ${total} % liegt unter dem Mindestwert ${min} %" >&2
  exit 1
fi
