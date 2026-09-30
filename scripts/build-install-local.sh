#!/usr/bin/env bash
# Build Wox from this fork's source with pinned toolchains and locked
# dependencies, then install it.
#
#   Windows: run in the MSYS2 MINGW64 shell (upstream's Makefile requires it).
#            Needs make, mingw-w64 gcc (x86_64, plus i686 for the 32-bit hook),
#            Go, Node 24, uv. Installs to %LOCALAPPDATA%\Programs\Wox\wox.exe.
#   macOS:   needs Xcode CLT, perl, create-dmg, Go, Node 24, uv. Leaves the
#            .dmg in release/ (INSTALL=1 copies the app to ~/Applications).
#
# Options (env): WINDOW_HOOK_CC32=<path to i686-w64-mingw32-gcc>, INSTALL=0|1,
#                VERIFY_OFFLINE=1 (rebuild the Go core with GOPROXY=off).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"
# shellcheck source=toolchain.env
source scripts/toolchain.env

fail() { echo "error: $*" >&2; exit 1; }

case "$(uname -s)" in
  MINGW64_NT*) PLATFORM=windows ;;
  Darwin)      PLATFORM=macos ;;
  *)           fail "run this in MSYS2 MINGW64 on Windows, or on macOS" ;;
esac

# ── Toolchains: pinned, and checked rather than trusted ────────────────────────
# Go fetches exactly this release if needed and verifies it via sum.golang.org.
export GOTOOLCHAIN="go$GO_VERSION"
export GOFLAGS="-mod=readonly"
go version | grep -q "go$GO_VERSION " || fail "Go $GO_VERSION not available ($(go version))"

node_major="$(node -p 'process.versions.node.split(".")[0]' 2>/dev/null || echo none)"
[[ "$node_major" == "$NODE_MAJOR" ]] || fail "Node $NODE_MAJOR required (found $node_major)"

uv_version="$(uv --version 2>/dev/null | awk '{print $2}')"
[[ "$uv_version" == "$UV_VERSION" ]] || fail "uv $UV_VERSION required (found ${uv_version:-none})"
export UV_PYTHON="$PYTHON_VERSION"

# pnpm switches itself to the version pinned in package.json and refuses to
# resolve anything that isn't in the lockfile (CI mode = frozen lockfile).
export CI=true
export npm_config_frozen_lockfile=true

# ── Inputs that aren't built from source ───────────────────────────────────────
dll="wox.core/resource/others/webview/WebView2Loader.dll"
actual="$( (sha256sum "$dll" 2>/dev/null || shasum -a 256 "$dll") | awk '{print $1}')"
[[ "$actual" == "$WEBVIEW2_LOADER_SHA256" ]] || fail "$dll hash changed: $actual (review before trusting it)"

echo "Verifying Go modules against go.sum..."
(cd wox.core && go mod verify)

# ── Build ──────────────────────────────────────────────────────────────────────
echo "Building Wox ($PLATFORM) from $(git rev-parse --short HEAD)..."
make_args=(build)
if [[ "$PLATFORM" == windows ]]; then
  make_args+=(WINDOW_HOOK_REQUIRE_32=1)
  [[ -n "${WINDOW_HOOK_CC32:-}" ]] && make_args+=("WINDOW_HOOK_CC32=$WINDOW_HOOK_CC32")
fi
make "${make_args[@]}"

if [[ "${VERIFY_OFFLINE:-0}" == 1 ]]; then
  echo "Rebuilding the Go core with the network off (GOPROXY=off)..."
  (cd wox.core && GOPROXY=off go build -tags sqlite_fts5 -o /dev/null .)
fi

# ── Record what was built ──────────────────────────────────────────────────────
# Files only (release/ also holds Wox.app, a folder).
(cd release && find . -maxdepth 1 -type f ! -name SHA256SUMS -print0 | sort -z |
  xargs -0 sh -c 'sha256sum "$@" 2>/dev/null || shasum -a 256 "$@"' _ | tee SHA256SUMS)

# ── Install ────────────────────────────────────────────────────────────────────
if [[ "$PLATFORM" == windows && "${INSTALL:-1}" == 1 ]]; then
  dest="$(cygpath -u "$LOCALAPPDATA")/Programs/Wox"
  mkdir -p "$dest"
  taskkill //IM wox.exe //F >/dev/null 2>&1 || true
  cp release/wox-windows-amd64.exe "$dest/wox.exe"
  echo "Installed to $(cygpath -w "$dest")\\wox.exe (unsigned: Smart App Control or AV may ask)."
elif [[ "$PLATFORM" == macos && "${INSTALL:-0}" == 1 ]]; then
  app="$(find release -maxdepth 2 -name 'Wox*.app' -print -quit)"
  [[ -n "$app" ]] || fail "no .app found in release/"
  ditto "$app" "$HOME/Applications/$(basename "$app")"
  echo "Installed to ~/Applications/$(basename "$app")"
fi
