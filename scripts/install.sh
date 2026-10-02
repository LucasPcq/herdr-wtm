#!/bin/sh
# Installs bin/herdr-wtm for the version in herdr-plugin.toml: downloads the
# prebuilt release binary (checksum-verified), or builds from source with Go
# when no usable release exists. Run by herdr as the plugin's [[build]] step.
set -eu

# CDPATH= keeps an exported CDPATH from resolving the relative "scripts/.."
# (herdr runs `sh scripts/install.sh`) to some other directory.
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BASE_URL=${HERDR_WTM_RELEASE_BASE_URL:-https://github.com/LucasPcq/herdr-wtm/releases/download}

log() { printf 'herdr-wtm install: %s\n' "$*" >&2; }

# build_from_source <reason>: go build, or fail when Go is missing.
build_from_source() {
	if ! command -v go >/dev/null 2>&1; then
		log "$1, and Go is not installed to build from source"
		exit 1
	fi
	log "$1; building from source with Go"
	(cd "$ROOT" && go build -o bin/herdr-wtm ./cmd/herdr-wtm)
	log "built from source"
	exit 0
}

# fetch <url> <dest>
fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -q "$1" -O "$2"
	else
		return 1
	fi
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

version=$(sed -n 's/^version *= *"\(.*\)".*/\1/p' "$ROOT/herdr-plugin.toml" | head -n 1)
if [ -z "$version" ]; then
	log "no version found in herdr-plugin.toml"
	exit 1
fi

if [ "${HERDR_WTM_BUILD_FROM_SOURCE:-}" = 1 ]; then
	build_from_source "source build requested"
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) arch=$(uname -m) ;;
esac

archive="herdr-wtm_${version}_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if ! fetch "$BASE_URL/v$version/$archive" "$tmp/$archive" || ! fetch "$BASE_URL/v$version/checksums.txt" "$tmp/checksums.txt"; then
	build_from_source "no prebuilt binary for v$version ($os/$arch)"
fi

expected=$(awk -v f="$archive" '$2 == f {print $1}' "$tmp/checksums.txt")
if [ -z "$expected" ] || [ "$expected" != "$(sha256 "$tmp/$archive")" ]; then
	build_from_source "checksum verification failed for $archive"
fi

tar -xzf "$tmp/$archive" -C "$tmp" herdr-wtm
mkdir -p "$ROOT/bin"
mv -f "$tmp/herdr-wtm" "$ROOT/bin/herdr-wtm"
chmod +x "$ROOT/bin/herdr-wtm"
log "installed prebuilt v$version ($os/$arch)"
