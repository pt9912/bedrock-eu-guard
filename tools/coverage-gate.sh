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

min="${1:?usage: coverage-gate.sh <mindestwert-in-prozent>}"
case "$min" in
  ''|*[!0-9.]*) echo "coverage-gate: ungueltiger Mindestwert '$min'" >&2; exit 2 ;;
esac

profile=.tmp/coverage.out
mkdir -p .tmp

pkgs=$(go list ./... | grep -v -e '/hexagon/coretest' -e '/cmd/' | paste -sd, -)
[ -n "$pkgs" ] || { echo "coverage-gate: keine Pakete gefunden" >&2; exit 2; }

go test -coverpkg="$pkgs" -coverprofile="$profile" ./... >/dev/null

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
