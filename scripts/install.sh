#!/usr/bin/env sh
# Ancla CLI installer — https://ancla.dev
#
# Usage:
#   curl -LsSf https://ancla.dev/install.sh | sh
#   curl -LsSf https://ancla.dev/install.sh | sh -s -- --version v0.5.0
#
# Environment variables:
#   ANCLA_INSTALL_DIR  — override install directory (default: /usr/local/bin or ~/.local/bin)
#   ANCLA_VERSION      — pin a specific release tag (e.g. v0.5.0)

set -eu

REPO="SideQuest-Group/ancla-client"
BINARY_NAME="ancla"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

say() {
  printf '%s\n' "$*"
}

err() {
  say "error: $*" >&2
  exit 1
}

need() {
  command -v "$1" > /dev/null 2>&1 || err "need '$1' (command not found)"
}

# ---------------------------------------------------------------------------
# Detect platform
# ---------------------------------------------------------------------------

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux"  ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) err "Windows is not supported by this installer — download from GitHub Releases" ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)   echo "amd64" ;;
    aarch64|arm64)   echo "arm64" ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------

VERSION="${ANCLA_VERSION:-latest}"

while [ $# -gt 0 ]; do
  case "$1" in
    --version)
      shift
      VERSION="${1:?--version requires a value}"
      ;;
    *)
      err "unknown argument: $1"
      ;;
  esac
  shift
done

# ---------------------------------------------------------------------------
# Resolve install directory
# ---------------------------------------------------------------------------

resolve_install_dir() {
  if [ -n "${ANCLA_INSTALL_DIR:-}" ]; then
    echo "$ANCLA_INSTALL_DIR"
    return
  fi
  if [ "$(id -u)" -eq 0 ]; then
    echo "/usr/local/bin"
  elif [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    echo "$HOME/.local/bin"
  else
    echo "/usr/local/bin"
  fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
  need curl
  need tar

  OS="$(detect_os)"
  ARCH="$(detect_arch)"
  INSTALL_DIR="$(resolve_install_dir)"

  say "Ancla CLI installer"
  say ""

  # Resolve version tag
  if [ "$VERSION" = "latest" ]; then
    say "Resolving latest release..."
    VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
    [ -n "$VERSION" ] || err "could not determine latest release"
  fi

  SEMVER="${VERSION#v}"
  ARCHIVE="ancla_${SEMVER}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

  say "  Platform:  ${OS}/${ARCH}"
  say "  Version:   ${VERSION}"
  say "  Archive:   ${ARCHIVE}"
  say "  Install:   ${INSTALL_DIR}/${BINARY_NAME}"
  say ""

  # Download and extract
  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  say "Downloading ${URL}..."
  curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}" \
    || err "download failed — check that release ${VERSION} exists for ${OS}/${ARCH}"

  say "Extracting..."
  tar xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

  # Install binary
  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  else
    say "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  fi
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  say ""
  say "Ancla CLI ${VERSION} installed to ${INSTALL_DIR}/${BINARY_NAME}"

  # Check PATH
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      say ""
      say "WARNING: ${INSTALL_DIR} is not in your PATH."
      say "Add it with:"
      say "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac

  say ""
  say "Run 'ancla login' to get started."
}

main
