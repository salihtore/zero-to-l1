#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# build-subnet-evm-plugin.sh
#
# Build and install a Subnet-EVM plugin binary that is RPCChainVM-protocol-
# compatible with AvalancheGo v1.15.0 (Helicon).
#
# Background
# ──────────
#   AvalancheGo v1.15.0  →  RPCChainVM protocol v46
#   Subnet-EVM v0.8.0    →  RPCChainVM protocol v44  (shipped by avalanche-cli v1.9.6)
#
#   The Subnet-EVM source was merged into the AvalancheGo monorepo in Dec 2025
#   under the directory returned by find_subnetevm_dir() below.
#
# Usage (run inside WSL):
#   bash cli/scripts/build-subnet-evm-plugin.sh [CHAIN_NAME] [AVALANCHEGO_TAG]
#
#   CHAIN_NAME      – name you used with `avalanche blockchain create` (default: zrgchain)
#   AVALANCHEGO_TAG – git tag to clone/checkout                        (default: v1.15.0)
#
# Environment variables (optional overrides):
#   AVALANCHEGO_SOURCE_DIR   – where to clone/find the avalanchego repo
#   PLUGIN_DIR               – where to install the binary
#                              (auto-detected from flags.json if not set)
#   SKIP_BUILD               – set to 1 to skip compilation (dry-run for CI)
# ──────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ── Colour helpers ─────────────────────────────────────────────────────────────
RED='\033[0;31m'; YELLOW='\033[1;33m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
info()    { echo -e "${CYAN}ℹ️  $*${NC}"; }
success() { echo -e "${GREEN}✅ $*${NC}"; }
warn()    { echo -e "${YELLOW}⚠️  $*${NC}"; }
die()     { echo -e "${RED}❌ ERROR: $*${NC}" >&2; exit 1; }

# ── Arguments / defaults ───────────────────────────────────────────────────────
CHAIN_NAME="${1:-zrgchain}"
AVALANCHEGO_TAG="${2:-v1.15.0}"
SOURCE_DIR="${AVALANCHEGO_SOURCE_DIR:-$HOME/avalanchego-src}"
SKIP_BUILD="${SKIP_BUILD:-0}"

# Minimum Go version required to build AvalancheGo v1.15.x
MIN_GO_MAJOR=1
MIN_GO_MINOR=22

info "Chain name      : $CHAIN_NAME"
info "AvalancheGo tag : $AVALANCHEGO_TAG"
info "Source dir      : $SOURCE_DIR"

# ── Prerequisite checks ────────────────────────────────────────────────────────
for cmd in avalanche git go; do
  command -v "$cmd" >/dev/null 2>&1 || die "'$cmd' is not on PATH — please install it first."
done

GO_VERSION="$(go env GOVERSION)"          # e.g. go1.22.5
GO_MAJOR="$(echo "$GO_VERSION" | sed -E 's/go([0-9]+)\..*/\1/')"
GO_MINOR="$(echo "$GO_VERSION" | sed -E 's/go[0-9]+\.([0-9]+).*/\1/')"

if (( GO_MAJOR < MIN_GO_MAJOR || (GO_MAJOR == MIN_GO_MAJOR && GO_MINOR < MIN_GO_MINOR) )); then
  die "Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ required; found ${GO_VERSION}. Install from https://go.dev/dl/"
fi
info "Go version      : $GO_VERSION ✓"

# ── Resolve plugin directory ───────────────────────────────────────────────────
resolve_plugin_dir() {
  local default_dir="$HOME/.avalanche-cli/local/${CHAIN_NAME}-local-node-fuji"

  # Try to read plugin-dir from the node's flags.json
  shopt -s nullglob
  local flags_files=("$default_dir"/*/flags.json)
  shopt -u nullglob

  if (( ${#flags_files[@]} > 0 )); then
    local configured
    configured="$(sed -nE 's/.*"plugin-dir"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/p' \
                  "${flags_files[0]}" | head -n 1)"
    [[ -n "$configured" ]] && { echo "$configured"; return; }
  fi

  # Fallback: avalanchego global plugins dir
  echo "$HOME/.avalanchego/plugins"
}

PLUGIN_DIR="${PLUGIN_DIR:-$(resolve_plugin_dir)}"
info "Plugin dir      : $PLUGIN_DIR"

# ── Resolve VMID ──────────────────────────────────────────────────────────────
info "Resolving VM ID for '$CHAIN_NAME'..."
# Strip ANSI escape codes, then extract the VM ID column from the describe table
VM_ID="$(avalanche blockchain describe "$CHAIN_NAME" 2>/dev/null \
  | perl -pe 's/\e\[[0-9;]*[[:alpha:]]//g' \
  | sed -nE 's/.*\|[[:space:]]*VM ID[[:space:]]*\|[[:space:]]*([A-Za-z0-9]+)[[:space:]]*\|.*/\1/p' \
  | head -n 1)"

[[ -n "$VM_ID" ]] || die "Could not determine VM ID for '$CHAIN_NAME'. Is the blockchain created?"
info "VM ID           : $VM_ID"

# ── Clone / verify source ──────────────────────────────────────────────────────
if [[ -d "$SOURCE_DIR/.git" ]]; then
  info "Source dir exists — verifying tag..."
  CURRENT_TAG="$(git -C "$SOURCE_DIR" describe --exact-match --tags HEAD 2>/dev/null || true)"
  if [[ "$CURRENT_TAG" != "$AVALANCHEGO_TAG" ]]; then
    warn "Repo at $SOURCE_DIR is on '$CURRENT_TAG', not '$AVALANCHEGO_TAG'. Fetching and checking out..."
    git -C "$SOURCE_DIR" fetch --tags --quiet
    git -C "$SOURCE_DIR" checkout "$AVALANCHEGO_TAG" --quiet
  fi
else
  info "Cloning AvalancheGo $AVALANCHEGO_TAG into $SOURCE_DIR (shallow clone)..."
  git clone --depth 1 --branch "$AVALANCHEGO_TAG" \
    https://github.com/ava-labs/avalanchego.git "$SOURCE_DIR"
fi
success "Source at $AVALANCHEGO_TAG"

# ── Find Subnet-EVM directory inside the monorepo ─────────────────────────────
find_subnetevm_dir() {
  local repo="$1"
  # Post-Dec-2025: source lives under graft/subnet-evm or plugin/evm
  for candidate in \
      "$repo/graft/subnet-evm" \
      "$repo/plugin/evm" \
      "$repo/subnet-evm"; do
    [[ -f "$candidate/go.mod" || -f "$candidate/main.go" ]] && { echo "$candidate"; return; }
  done
  # Broader fallback: find any directory named subnet-evm with a go.mod
  find "$repo" -maxdepth 4 -type d -name 'subnet-evm' | while read -r d; do
    [[ -f "$d/go.mod" ]] && { echo "$d"; return; }
  done
}

SUBNETEVM_DIR="$(find_subnetevm_dir "$SOURCE_DIR")"
[[ -n "$SUBNETEVM_DIR" ]] || die "Could not locate Subnet-EVM source inside $SOURCE_DIR."
info "Subnet-EVM dir  : $SUBNETEVM_DIR"

# ── Build ──────────────────────────────────────────────────────────────────────
BINARY_PATH="$SOURCE_DIR/build/subnet-evm-${AVALANCHEGO_TAG}"
mkdir -p "$(dirname "$BINARY_PATH")"

if [[ "$SKIP_BUILD" == "1" ]]; then
  warn "SKIP_BUILD=1 — skipping compilation (dry-run mode)."
else
  info "Building Subnet-EVM plugin (this may take a few minutes)..."

  # Prefer the upstream build script if it exists, otherwise go build directly
  BUILD_SCRIPT="$(find "$SUBNETEVM_DIR" -maxdepth 3 -name 'build.sh' | head -n 1 || true)"
  if [[ -n "$BUILD_SCRIPT" ]]; then
    info "Using build script: $BUILD_SCRIPT"
    bash "$BUILD_SCRIPT" "$BINARY_PATH"
  else
    info "No build.sh found — using 'go build'..."
    (
      cd "$SUBNETEVM_DIR"
      go build -o "$BINARY_PATH" .
    )
  fi

  [[ -f "$BINARY_PATH" ]] || die "Build completed but binary not found at $BINARY_PATH"
  success "Binary built: $BINARY_PATH"
fi

# ── Install ────────────────────────────────────────────────────────────────────
mkdir -p "$PLUGIN_DIR"
install -m 0755 "$BINARY_PATH" "$PLUGIN_DIR/$VM_ID"
success "Installed: $PLUGIN_DIR/$VM_ID"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "┌─────────────────────────────────────────────────────────────────────────┐"
echo "│  🎉  Subnet-EVM plugin installed successfully                           │"
echo "├─────────────────────────────────────────────────────────────────────────┤"
printf "│  Chain     : %-58s │\n" "$CHAIN_NAME"
printf "│  VM ID     : %-58s │\n" "$VM_ID"
printf "│  Plugin    : %-58s │\n" "$PLUGIN_DIR/$VM_ID"
printf "│  AG tag    : %-58s │\n" "$AVALANCHEGO_TAG"
echo "├─────────────────────────────────────────────────────────────────────────┤"
echo "│  Next steps:                                                            │"
echo "│  1. Restart your local AvalancheGo node.                               │"
echo "│  2. avalanche node local track <cluster> --chain $CHAIN_NAME           │"
echo "│  3. avalanche contract initValidatorManager $CHAIN_NAME --fuji         │"
echo "└─────────────────────────────────────────────────────────────────────────┘"
