#!/bin/sh
# Smoke-Test für ein gebautes Image (braucht Docker, aber keinen Netzzugriff zur Laufzeit).
# Aufruf: tools/image-smoke.sh <image> <slim|full>
set -eu

img=${1:?Image fehlt}
variant=${2:?Variante fehlt (slim|full)}
root=$(cd "$(dirname "$0")/.." && pwd)
good="$root/internal/adapter/driving/cli/testdata/good/project"
bad="$root/internal/adapter/driving/cli/testdata/bad/project"

fail() { echo "FEHLER: $*" >&2; exit 1; }
run() { docker run --rm --network none --read-only "$@"; }
envs="-e AWS_REGION=eu-central-1 -e CLAUDE_CODE_USE_BEDROCK=1"

# 1. Kein Root.
user=$(docker inspect --format '{{.Config.User}}' "$img")
[ "$user" = "65532:65532" ] || fail "Image läuft als '$user', erwartet 65532:65532"

# 2. version
out=$(run "$img" version)
case "$out" in "bedrock-eu-check "*) ;; *) fail "unerwartete version-Ausgabe: $out" ;; esac

# 3. local: gutes Beispiel -> Exit 0, schlechtes -> Exit 1
# shellcheck disable=SC2086
run $envs -v "$good:/work:ro" "$img" local --root /work >/dev/null || fail "local good: Exit != 0"
set +e
# shellcheck disable=SC2086
run $envs -v "$bad:/work:ro" "$img" local --root /work >/dev/null
code=$?
set -e
[ "$code" -eq 1 ] || fail "local bad: Exit $code, erwartet 1"

# 4. terraform + JSON parsebar
# shellcheck disable=SC2086
run -v "$bad:/work:ro" "$img" terraform --root /work --format json | grep -q '"version": 1' || fail "terraform json: Ausgabe ohne version"

# 5. aws: slim -> Tool-Fehler (kein aws-Binary), full -> Findings ohne Credentials (Exit 1)
set +e
# shellcheck disable=SC2086
run -e AWS_EC2_METADATA_DISABLED=true "$img" aws >/dev/null 2>&1
code=$?
set -e
case "$variant" in
  slim) [ "$code" -eq 2 ] || fail "slim aws: Exit $code, erwartet 2 (AWS_CLI_MISSING)" ;;
  full)
    [ "$code" -eq 1 ] || fail "full aws ohne Credentials: Exit $code, erwartet 1"
    ver=$(run --entrypoint aws "$img" --version 2>&1) || fail "aws --version"
    case "$ver" in "aws-cli/2.39.7 "*) ;; *) fail "unerwartete aws-Version: $ver" ;; esac
    ;;
  *) fail "unbekannte Variante $variant" ;;
esac

# 6. all --ci auf dem Beispiel-Repository: slim -> Tool-Fehler (kein aws-Binary), full -> FAIL
#    (keine Credentials). In beiden Fällen erscheinen die offline Befunde (local + terraform).
errfile=$(mktemp)
trap 'rm -f "$errfile"' EXIT
set +e
# shellcheck disable=SC2086
out=$(run $envs -e AWS_EC2_METADATA_DISABLED=true -v "$good:/work:ro" "$img" all --ci --root /work 2>"$errfile")
code=$?
set -e
err=$(cat "$errfile")
echo "$out" | grep -q 'LH-FA-LENV-001' || fail "all --ci: keine local-Befunde"
echo "$out" | grep -q 'LH-FA-TF-001' || fail "all --ci: keine terraform-Befunde"
echo "$err" | grep -q '^bedrock-eu-check: .* (Exit ' || fail "all --ci: keine Zusammenfassung auf stderr: $err"
case "$variant" in
  slim) [ "$code" -eq 2 ] || fail "slim all --ci: Exit $code, erwartet 2" ;;
  full)
    [ "$code" -eq 1 ] || fail "full all --ci ohne Credentials: Exit $code, erwartet 1"
    echo "$out" | grep -q 'LH-FA-AWS-001' || fail "full all --ci: keine aws-Befunde"
    ;;
esac

echo "OK: $img ($variant)"
