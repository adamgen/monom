#!/usr/bin/env bash
# install.sh — installs monom: the mnmd binary plus its bash/zsh integration.
#
#   curl -fsSL https://monom.dev/install.sh | bash
#
# Downloads the release tarball for your OS and CPU from GitHub, checks its
# sha256 against the release's checksums.txt, unpacks it into
# ~/.local/share/monom, links ~/.local/bin/mnmd, and runs `mnmd install`, which
# adds one `source` line to your ~/.zshrc or ~/.bashrc (never twice).
# Re-running upgrades. MONOM_VERSION=v1.2.3 pins a release. No sudo.
# Same file on GitHub: https://raw.githubusercontent.com/adamgen/monom/main/install.sh
set -eu

die() { echo "monom: error: $*" >&2; exit 1; }

main() {
  local repo="https://github.com/adamgen/monom" dir="$HOME/.local/share/monom"
  local os arch asset base sum ver
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    arm64 | aarch64) arch=arm64 ;;
    *) arch=$(uname -m) ;;
  esac
  asset="monom-$os-$arch.tar.gz"
  base="$repo/releases/latest/download"
  [ -z "${MONOM_VERSION:-}" ] || base="$repo/releases/download/$MONOM_VERSION"
  base="${MONOM_DOWNLOAD_BASE:-$base}" # tests point this at a local fake release

  tmp=$(mktemp -d) # global: the EXIT trap runs after main returns
  trap 'rm -rf "$tmp"' EXIT
  cd "$tmp"

  echo "monom: downloading $base/$asset"
  curl -fsSL -o "$asset" "$base/$asset" ||
    die "no monom release for $os/$arch. Build it from source instead (needs Go): $repo#install"
  curl -fsSL -o checksums.txt "$base/checksums.txt" || die "could not download checksums.txt"
  grep " $asset\$" checksums.txt > "$asset.sha256" || die "checksums.txt has no entry for $asset"
  if command -v sha256sum > /dev/null; then sum=sha256sum; else sum="shasum -a 256"; fi
  $sum -c "$asset.sha256" > /dev/null || die "sha256 mismatch for $asset; not installing it"

  tar -xzf "$asset"
  ver=$("./monom-$os-$arch/bin/mnmd" version) || die "the downloaded mnmd does not run here"
  mkdir -p "$dir" "$HOME/.local/bin"
  rm -rf "${dir:?}/bin" "${dir:?}/src"
  mv "monom-$os-$arch/bin" "monom-$os-$arch/src" "$dir/"
  ln -sf "$dir/bin/mnmd" "$HOME/.local/bin/mnmd"
  echo "monom: installed mnmd $ver in $dir"

  "$dir/bin/mnmd" install ||
    echo "monom: add this line to your shell's rc file: source \"$dir/src/monom\"" >&2
}

main "$@"
