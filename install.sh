#!/usr/bin/env bash
# install.sh — install monom: the mnmd engine plus its bash/zsh integration.
#
#   curl -fsSL https://raw.githubusercontent.com/adamgen/monom/main/install.sh | bash
#
# What it does, in order:
#   1. Downloads the prebuilt monom-<os>-<arch>.tar.gz for your machine from
#      the latest GitHub release (linux and darwin, amd64 and arm64) and checks
#      it against the release's checksums.txt (sha256). A mismatch aborts.
#   2. If that release has no tarball for your platform (or there is no
#      release yet) and Go is installed, builds mnmd from source with build.sh
#      instead. Without Go, it stops and says what to do. It never runs make.
#   3. Installs <dir>/bin/mnmd and <dir>/src/monom{,.bash,.zsh} into
#      ~/.local/share/monom, and links ~/.local/bin/mnmd to the binary.
#   4. Runs `mnmd install`, which adds one line to ~/.zshrc (zsh) or to
#      ~/.bash_profile if it exists, else ~/.bashrc (bash):
#        source "<dir>/src/monom"
#      It never adds that line twice. Nothing is run with sudo.
#
# Re-running upgrades in place. If an rc file already sources an earlier
# monom install that isn't a git checkout, that directory is upgraded instead
# of creating a second install.
#
# Environment overrides:
#   MONOM_VERSION        release tag to install, e.g. v0.1.0 (default: latest)
#   MONOM_INSTALL_DIR    install directory (default: ~/.local/share/monom)
#   MONOM_BIN_DIR        where to link mnmd (default: ~/.local/bin; empty: skip)
#   MONOM_NO_MODIFY_RC   1: don't touch rc files; print the line to add instead
#   MONOM_FROM_SOURCE    1: skip the download and build from source (needs Go)
#   MONOM_DOWNLOAD_BASE  URL of a directory holding the release tarballs and
#                        checksums.txt (mirrors, tests)
#   MONOM_SOURCE_URL     source .tar.gz for the build fallback (mirrors, tests)
set -euo pipefail

REPO="adamgen/monom"

say() { printf 'monom: %s\n' "$*"; }
warn() { printf 'monom: warning: %s\n' "$*" >&2; }
die() { printf 'monom: error: %s\n' "$*" >&2; exit 1; }
has() { command -v "$1" >/dev/null 2>&1; }

usage() {
  local self="${BASH_SOURCE[0]:-}"
  if [ -n "$self" ] && [ -f "$self" ]; then
    sed -n '2,/^set -euo/p' "$self" | sed '$d; s/^# \{0,1\}//'
  else
    printf 'The options are documented at the top of https://github.com/%s/blob/main/install.sh\n' "$REPO"
  fi
}

detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"
  case "$os" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    MINGW* | MSYS* | CYGWIN*) die "Windows is not supported (monom needs bash or zsh on Linux or macOS; WSL works)" ;;
    *) os="$(printf '%s' "$os" | tr '[:upper:]' '[:lower:]')" ;;
  esac
  case "$arch" in
    x86_64 | amd64) arch=amd64 ;;
    aarch64 | arm64) arch=arm64 ;;
  esac
  # An x86_64 shell under Rosetta on Apple silicon: prefer the native binary.
  if [ "$os" = darwin ] && [ "$arch" = amd64 ] &&
    [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" = 1 ]; then
    arch=arm64
  fi
  PLATFORM="$os-$arch"
}

# download <url> <dest> — fails (non-zero, reason on stderr) on any HTTP error.
download() {
  if has curl; then
    curl -fsSL --retry 2 -o "$2" "$1"
  elif has wget; then
    wget -q -O "$2" "$1"
  else
    die "need curl or wget to download monom"
  fi
}

sha256_of() {
  if has sha256sum; then
    sha256sum "$1" | cut -d' ' -f1
  elif has shasum; then
    shasum -a 256 "$1" | cut -d' ' -f1
  elif has openssl; then
    openssl dgst -sha256 "$1" | sed 's/^.*= *//'
  else
    die "need sha256sum, shasum or openssl to verify the download"
  fi
}

# find_existing_install [rc file...] — prints each directory the rc files
# (default: all bash and zsh ones) source monom from, one per line: the <dir>
# in `source "<dir>/src/monom"`.
find_existing_install() {
  local rc
  [ $# -gt 0 ] || set -- "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile"
  for rc in "$@"; do
    [ -f "$rc" ] || continue
    sed -nE 's#^[[:space:]]*(source|\.)[[:space:]]+"?([^"]*)/src/monom"?[[:space:]]*$#\2#p' "$rc" |
      sed -e "s#^~/#$HOME/#" -e "s#^\\\$HOME/#$HOME/#" -e "s#^\\\${HOME}/#$HOME/#"
  done | awk '!seen[$0]++'
}

# same_dir <a> <b> — whether a and b are the same directory (symlinks resolved).
same_dir() {
  [ -d "$1" ] && [ -d "$2" ] && [ "$(cd "$1" && pwd -P)" = "$(cd "$2" && pwd -P)" ]
}

resolve_install_dir() {
  local existing
  if [ -n "${MONOM_INSTALL_DIR:-}" ]; then
    INSTALL_DIR="$MONOM_INSTALL_DIR"
  else
    INSTALL_DIR="$HOME/.local/share/monom"
    while IFS= read -r existing; do
      [ -n "$existing" ] || continue
      if [ -d "$existing/.git" ]; then
        die "your shell already sources a monom git checkout at $existing.
  Update it with: cd '$existing' && git pull && ./build.sh
  To switch to a release install, remove that source line from your rc file and re-run this installer,
  or set MONOM_INSTALL_DIR to install a second copy alongside it."
      elif [ -d "$existing" ] && ! same_dir "$existing" "$INSTALL_DIR"; then
        say "found an earlier install at $existing; upgrading it in place"
        INSTALL_DIR="$existing"
      fi
    done <<<"$(find_existing_install)"
  fi
  case "$INSTALL_DIR" in
    /*) ;;
    *) INSTALL_DIR="$PWD/$INSTALL_DIR" ;;
  esac
  if [ -d "$INSTALL_DIR/.git" ]; then
    die "$INSTALL_DIR is a git checkout; update it with 'git pull && ./build.sh' there, or set MONOM_INSTALL_DIR to another directory"
  fi
}

# fetch_release <stage dir> — downloads, verifies and unpacks the release
# tarball into <stage dir>. Returns 1 (reason in RELEASE_ERR) when there is no
# tarball for this platform; aborts on a checksum problem.
fetch_release() {
  local stage="$1" base asset expected actual
  if [ -n "${MONOM_DOWNLOAD_BASE:-}" ]; then
    base="${MONOM_DOWNLOAD_BASE%/}"
  elif [ "$VERSION" = latest ]; then
    base="https://github.com/$REPO/releases/latest/download"
  else
    base="https://github.com/$REPO/releases/download/$VERSION"
  fi
  asset="monom-$PLATFORM.tar.gz"
  say "downloading $base/$asset"
  if ! download "$base/$asset" "$TMP/$asset" 2>"$TMP/err"; then
    RELEASE_ERR="no $asset at $base ($(tail -n 1 "$TMP/err" 2>/dev/null || true))"
    return 1
  fi
  download "$base/checksums.txt" "$TMP/checksums.txt" 2>"$TMP/err" ||
    die "downloaded $asset but not checksums.txt from $base ($(tail -n 1 "$TMP/err")); refusing to install an unverified binary"
  expected="$(awk -v f="$asset" '$2 == f || $2 == "*" f { print $1; exit }' "$TMP/checksums.txt")"
  [ -n "$expected" ] || die "checksums.txt at $base has no entry for $asset; refusing to install an unverified binary"
  actual="$(sha256_of "$TMP/$asset")"
  [ "$actual" = "$expected" ] ||
    die "checksum mismatch for $asset (expected $expected, got $actual); the download is corrupt or was tampered with"
  say "verified sha256 $actual"
  tar -xzf "$TMP/$asset" -C "$TMP"
  [ -x "$TMP/monom-$PLATFORM/bin/mnmd" ] || die "$asset does not contain monom-$PLATFORM/bin/mnmd"
  "$TMP/monom-$PLATFORM/bin/mnmd" version >/dev/null 2>&1 ||
    die "the mnmd in $asset does not run on this machine ($(uname -sm))"
  mv "$TMP/monom-$PLATFORM" "$stage"
}

# build_from_source <stage dir> — downloads the source tarball and builds mnmd
# with build.sh (or plain `go build` for trees that predate it).
build_from_source() {
  local stage="$1" url stamp src
  if [ -n "${MONOM_SOURCE_URL:-}" ]; then
    url="$MONOM_SOURCE_URL"
    stamp="source"
  elif [ "$VERSION" = latest ]; then
    url="https://github.com/$REPO/archive/refs/heads/main.tar.gz"
    stamp="main-source"
  else
    url="https://github.com/$REPO/archive/refs/tags/$VERSION.tar.gz"
    stamp="$VERSION"
  fi
  say "building from source with $(go version 2>/dev/null || printf 'go')"
  say "downloading $url"
  download "$url" "$TMP/source.tar.gz" 2>"$TMP/err" ||
    die "could not download the source from $url ($(tail -n 1 "$TMP/err"))"
  mkdir -p "$TMP/source"
  tar -xzf "$TMP/source.tar.gz" -C "$TMP/source"
  src="$(find "$TMP/source" -maxdepth 2 -name go.mod -exec dirname {} \; | head -n 1)"
  if [ -z "$src" ] || [ ! -f "$src/cmd/mnmd/main.go" ]; then
    die "$url does not look like the monom source"
  fi
  mkdir -p "$stage/bin" "$stage/src"
  if [ -f "$src/build.sh" ]; then
    VERSION="$stamp" OUT="$stage/bin/mnmd" bash "$src/build.sh" ||
      die "the build failed (monom needs Go $(sed -n 's/^go //p' "$src/go.mod") or newer: https://go.dev/dl/)"
  else
    (cd "$src" && GOWORK=off CGO_ENABLED=0 go build -o "$stage/bin/mnmd" ./cmd/mnmd) ||
      die "the build failed (monom needs Go $(sed -n 's/^go //p' "$src/go.mod") or newer: https://go.dev/dl/)"
  fi
  cp "$src/src/monom" "$src/src/monom.bash" "$src/src/monom.zsh" "$stage/src/"
}

# place <from> <to> — replaces <to> atomically, so a shell sourcing it or a
# running mnmd never sees a half-written file.
place() {
  cp "$1" "$2.new.$$"
  mv -f "$2.new.$$" "$2"
}

install_tree() {
  local stage="$1" f
  mkdir -p "$INSTALL_DIR/bin" "$INSTALL_DIR/src"
  place "$stage/bin/mnmd" "$INSTALL_DIR/bin/mnmd"
  chmod 755 "$INSTALL_DIR/bin/mnmd"
  for f in monom monom.bash monom.zsh; do
    place "$stage/src/$f" "$INSTALL_DIR/src/$f"
  done
  local version
  version="$("$INSTALL_DIR/bin/mnmd" version 2>/dev/null || true)"
  say "installed mnmd${version:+ $version} to $INSTALL_DIR"
}

link_binary() {
  local bin_dir="$1" link
  [ -n "$bin_dir" ] || return 0
  link="$bin_dir/mnmd"
  if [ -e "$link" ] && [ ! -L "$link" ]; then
    warn "$link exists and is not a symlink; leaving it alone"
    return 0
  fi
  mkdir -p "$bin_dir"
  ln -sfn "$INSTALL_DIR/bin/mnmd" "$link"
  case ":$PATH:" in
    *":$bin_dir:"*) say "linked $link" ;;
    *) say "linked $link (not on your PATH; the monom shell integration doesn't need it)" ;;
  esac
}

wire_shell() {
  local line="source \"$INSTALL_DIR/src/monom\""
  if [ "${MONOM_NO_MODIFY_RC:-}" = 1 ]; then
    say "MONOM_NO_MODIFY_RC=1: add this line to your ~/.zshrc or ~/.bashrc yourself:"
    printf '  %s\n' "$line"
    return 0
  fi
  # An rc line that already sources this install, in any form (~, $HOME, a
  # symlinked path), counts: `mnmd install` only recognises its own spelling.
  local existing rcs=()
  case "${SHELL:-}" in
    */zsh) rcs=("$HOME/.zshrc") ;;
    */bash) rcs=("$HOME/.bash_profile" "$HOME/.bashrc") ;;
  esac
  while IFS= read -r existing; do
    if [ -n "$existing" ] && same_dir "$existing" "$INSTALL_DIR"; then
      say "already installed (your shell sources $existing/src/monom)"
      zsh_completion_note
      return 0
    fi
  done <<<"$(find_existing_install ${rcs[@]+"${rcs[@]}"})"
  if ! MONOM_ACTIVE=1 "$INSTALL_DIR/bin/mnmd" install </dev/null | sed 's/^/monom: /'; then
    warn "could not add the source line for SHELL='${SHELL:-}' (monom supports bash and zsh)"
    warn "if you use bash or zsh, add this line to its rc file:"
    printf '  %s\n' "$line" >&2
    return 0
  fi
  zsh_completion_note
}

# zsh_completion_note — monom registers its zsh completion with compdef, which
# only exists after compinit. Say so when ~/.zshrc visibly never runs it.
zsh_completion_note() {
  case "${SHELL:-}" in */zsh) ;; *) return 0 ;; esac
  if ! grep -Eqs "compinit|oh-my-zsh|prezto|zinit|zimfw|antidote|antigen|zplug" "$HOME/.zshrc"; then
    say "note: ~/.zshrc doesn't run compinit, so Tab completion for monom stays off until it does."
    say "      Add this line above the monom source line:  autoload -Uz compinit && compinit"
  fi
}

main() {
  case "${1:-}" in
    -h | --help) usage; return 0 ;;
    "") ;;
    *) die "unknown argument '$1'; configure the installer with MONOM_* environment variables (see --help)" ;;
  esac

  [ -n "${HOME:-}" ] || die "HOME is not set"
  has tar || die "need tar to unpack monom"

  VERSION="${MONOM_VERSION:-latest}"
  case "$VERSION" in
    latest | v*) ;;
    [0-9]*) VERSION="v$VERSION" ;;
  esac

  detect_platform
  resolve_install_dir

  TMP="$(mktemp -d 2>/dev/null || mktemp -d -t monom)"
  # shellcheck disable=SC2064 # expand now: TMP is fixed for this run
  trap "rm -rf '$TMP'" EXIT

  local stage="$TMP/stage"
  RELEASE_ERR=""
  if [ "${MONOM_FROM_SOURCE:-}" = 1 ]; then
    has go || die "MONOM_FROM_SOURCE=1 needs Go on PATH (https://go.dev/dl/)"
    build_from_source "$stage"
  elif ! fetch_release "$stage"; then
    if has go; then
      say "$RELEASE_ERR"
      say "falling back to a source build"
      build_from_source "$stage"
    else
      die "$RELEASE_ERR, and Go is not installed to build mnmd from source.
  Either install Go (https://go.dev/dl/) and re-run this installer,
  or pick a release with a $PLATFORM build: MONOM_VERSION=vX.Y.Z (https://github.com/$REPO/releases)."
    fi
  fi

  install_tree "$stage"
  link_binary "${MONOM_BIN_DIR-$HOME/.local/bin}"
  wire_shell
  say "done. Open a new shell, cd into any git repo, and type: monom <Tab>"
}

main "$@"
