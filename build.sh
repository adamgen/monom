#!/usr/bin/env bash
# build.sh — the one place that knows how to compile mnmd. No make needed.
#
#   ./build.sh                 build bin/mnmd for this machine
#   ./build.sh dist [DIR]      cross-compile the release tarballs and
#                              checksums.txt into DIR (default: dist/)
#
# Environment:
#   VERSION      stamped into `mnmd version` (default: git describe, else "dev")
#   OUT          output path for the host build (default: bin/mnmd)
#   TARGETS      dist only: space-separated os/arch list
#                (default: linux/amd64 linux/arm64 darwin/amd64 darwin/arm64)
#   GOOS GOARCH  honoured by the host build, as with `go build`
#   CGO_ENABLED  default 0: mnmd is pure Go, and a static binary runs anywhere
#
# Only ./cmd/mnmd from the root module is built. The root go.mod has no
# dependencies, so a build never downloads anything; the test-only modules
# (tests/harness) are never touched.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

export GOWORK=off
export CGO_ENABLED="${CGO_ENABLED:-0}"

die() { printf 'build.sh: %s\n' "$*" >&2; exit 1; }

version() {
  if [ -n "${VERSION:-}" ]; then
    printf '%s\n' "$VERSION"
  elif git rev-parse --git-dir >/dev/null 2>&1; then
    git describe --tags --always --dirty 2>/dev/null || printf 'dev\n'
  else
    printf 'dev\n'
  fi
}

# compile <output path> — builds ./cmd/mnmd with the stamped version.
compile() {
  command -v go >/dev/null 2>&1 || die "go not found on PATH (need Go $(go_min) or newer: https://go.dev/dl/)"
  mkdir -p "$(dirname "$1")"
  go build -trimpath -ldflags "-s -w -X main.version=$(version)" -o "$1" ./cmd/mnmd
}

go_min() { sed -n 's/^go //p' go.mod; }

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$@"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$@"
  else
    die "need sha256sum or shasum to write checksums"
  fi
}

# dist <dir> — one monom-<os>-<arch>.tar.gz per target, plus checksums.txt.
# Each tarball holds monom-<os>-<arch>/{bin/mnmd,src/monom,src/monom.bash,
# src/monom.zsh,README.md}: the <root>/bin + <root>/src layout that
# `mnmd install` and src/monom resolve paths against.
dist() {
  local out="$1" targets target os arch name stage
  targets="${TARGETS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64}"
  mkdir -p "$out"
  out="$(cd "$out" && pwd)"
  stage="$(mktemp -d)"
  # shellcheck disable=SC2064 # expand now: $stage is local
  trap "rm -rf '$stage'" EXIT

  rm -f "$out"/monom-*.tar.gz "$out/checksums.txt"
  for target in $targets; do
    os="${target%/*}"
    arch="${target#*/}"
    name="monom-$os-$arch"
    printf 'building %s (%s)\n' "$name" "$(version)"
    mkdir -p "$stage/$name/src"
    GOOS="$os" GOARCH="$arch" compile "$stage/$name/bin/mnmd"
    cp src/monom src/monom.bash src/monom.zsh "$stage/$name/src/"
    cp README.md "$stage/$name/"
    tar -C "$stage" -czf "$out/$name.tar.gz" "$name"
  done
  (cd "$out" && sha256 monom-*.tar.gz > checksums.txt)
  printf 'wrote %s\n' "$out"
  (cd "$out" && ls -1 monom-*.tar.gz checksums.txt)
}

case "${1:-}" in
  "")
    compile "${OUT:-bin/mnmd}"
    ;;
  dist)
    dist "${2:-dist}"
    ;;
  -h|--help)
    sed -n '2,/^set -euo/p' "$0" | sed '$d; s/^# \{0,1\}//'
    ;;
  *)
    die "unknown argument '$1' (try --help)"
    ;;
esac
