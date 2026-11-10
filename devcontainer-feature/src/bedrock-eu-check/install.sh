#!/bin/sh
# Devcontainer-Feature: installiert bedrock-eu-check (LH-TECH-005).
#
# Lädt <baseUrl>/v<version>/bedrock-eu-check_<version>_linux_<arch> und prüft es
# gegen <baseUrl>/v<version>/SHA256SUMS. Ohne passende Prüfsumme wird nichts
# installiert. Die Prüfsumme kommt vom selben Ursprung wie das Binary: Sie schützt
# vor Korruption und Teil-Downloads, nicht vor einem kompromittierten Ursprung.
# Optionen kommen als Umgebungsvariablen (VERSION, BASEURL) aus
# devcontainer-feature.json.
set -eu

VERSION="${VERSION:-1.0.0}"
BASEURL="${BASEURL:-https://github.com/pt9912/bedrock-eu-guard/releases/download}"
DEST="/usr/local/bin/bedrock-eu-check"

fail() { echo "bedrock-eu-check (Feature): $*" >&2; exit 1; }

# Nur verschluesselte oder lokale Quellen (kein http://, kein ftp://).
case "$BASEURL" in
  https://*|file://*) ;;
  *) fail "baseUrl muss mit https:// oder file:// beginnen" ;;
esac

case "$VERSION" in
  ''|*[!0-9A-Za-z.+_-]*) fail "ungueltige Version '$VERSION'" ;;
esac

case "$(uname -m)" in
  x86_64|amd64)  ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) fail "nicht unterstuetzte Architektur '$(uname -m)'" ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update -y >/dev/null && apt-get install -y --no-install-recommends curl ca-certificates >/dev/null
  else
    fail "curl fehlt und kann nicht installiert werden"
  fi
fi

NAME="bedrock-eu-check_${VERSION}_linux_${ARCH}"
URL="${BASEURL%/}/v${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL/$NAME" -o "$TMP/$NAME" || fail "Download von $NAME fehlgeschlagen"
curl -fsSL "$URL/SHA256SUMS" -o "$TMP/SHA256SUMS" || fail "Download von SHA256SUMS fehlgeschlagen"

# Erwartete Pruefsumme der Datei aus SHA256SUMS ("<hash>  <datei>").
EXPECTED="$(awk -v f="$NAME" '$2 == f { print $1 }' "$TMP/SHA256SUMS")"
[ -n "$EXPECTED" ] || fail "keine Pruefsumme fuer $NAME in SHA256SUMS"
ACTUAL="$(sha256sum "$TMP/$NAME" | awk '{ print $1 }')"
[ "$EXPECTED" = "$ACTUAL" ] || fail "Pruefsumme stimmt nicht (erwartet $EXPECTED, erhalten $ACTUAL)"

install -m 0755 "$TMP/$NAME" "$DEST"
"$DEST" version
