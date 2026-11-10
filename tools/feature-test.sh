#!/bin/sh
# Test des Devcontainer-Features gegen ein lokal gebautes "Release" (braucht Docker,
# aber keinen Netzzugriff: die Installation nutzt file://). Aufruf: tools/feature-test.sh
set -eu

root=$(cd "$(dirname "$0")/.." && pwd)
feature="$root/devcontainer-feature/src/bedrock-eu-check"
image=${GO_IMAGE:?GO_IMAGE fehlt (Make-Variable)}
version=9.9.9
fail() { echo "FEHLER: $*" >&2; exit 1; }

# 1. Struktur: Metadaten, Syntax.
python3 -m json.tool "$feature/devcontainer-feature.json" >/dev/null || fail "devcontainer-feature.json ist kein gueltiges JSON"
grep -q '"id": "bedrock-eu-check"' "$feature/devcontainer-feature.json" || fail "id passt nicht zum Ordnernamen"
head -1 "$feature/install.sh" | grep -q '^#!/bin/sh' || fail "install.sh ohne Shebang"
sh -n "$feature/install.sh" || fail "install.sh hat einen Syntaxfehler"
[ -x "$feature/install.sh" ] || fail "install.sh ist nicht ausfuehrbar"

# 2. Lokales "Release" bauen.
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
arch=$(uname -m); case "$arch" in x86_64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; esac
name="bedrock-eu-check_${version}_linux_${arch}"
mkdir -p "$work/release/v$version"
make -C "$root" build VERSION=$version COMMIT=featuretest >/dev/null
cp "$root/bin/bedrock-eu-check" "$work/release/v$version/$name"
(cd "$work/release/v$version" && sha256sum "$name" > SHA256SUMS)
chmod -R a+rX "$work"

run() {
  docker run --rm --network none \
    -v "$work/release:/release:ro" -v "$feature:/feature:ro" \
    -e VERSION=$version -e BASEURL=file:///release \
    "$image" sh -c "$1"
}

# 3. Positiv: Installation und Aufruf.
out=$(run 'cd /feature && ./install.sh >/dev/null && bedrock-eu-check version') || fail "Installation schlug fehl"
[ "$out" = "bedrock-eu-check $version (featuretest)" ] || fail "unerwartete Ausgabe: $out"

# 4. Negativ: manipulierte Pruefsumme -> Exit != 0, Fehlermeldung, kein Binary.
sed -i 's/^./0/' "$work/release/v$version/SHA256SUMS"
rc=0
out=$(run 'cd /feature && ./install.sh 2>&1; code=$?; [ -e /usr/local/bin/bedrock-eu-check ] && echo "BINARY-INSTALLIERT"; exit $code') || rc=$?
[ "$rc" -ne 0 ] || fail "manipulierte Pruefsumme wurde akzeptiert"
echo "$out" | grep -q 'Pruefsumme stimmt nicht' || fail "keine Fehlermeldung zur Pruefsumme: $out"
echo "$out" | grep -q 'BINARY-INSTALLIERT' && fail "trotz falscher Pruefsumme wurde installiert"

# 5. Negativ: ungueltige Version.
rc=0; docker run --rm --network none -v "$feature:/feature:ro" -e 'VERSION=1.0/../x' "$image" sh -c 'cd /feature && ./install.sh' >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] || fail "ungueltige Version wurde akzeptiert"

echo "OK: Devcontainer-Feature"
