#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="smtp-relay"
INSTALL_BIN_PATH="/usr/local/bin/${BINARY_NAME}"
CONFIG_PATH="/etc/${BINARY_NAME}.yaml"
SERVICE_PATH="/etc/systemd/system/${BINARY_NAME}.service"
RAW_BASE="https://raw.githubusercontent.com/0x2321/go-smtp-relay/main"
API_LATEST="https://api.github.com/repos/0x2321/go-smtp-relay/releases/latest"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: required command '$1' not found. Please install it and re-run." >&2
    exit 1
  }
}

require_linux() {
  if [ "$(uname -s)" != "Linux" ]; then
    echo "Error: this installer currently supports Linux only." >&2
    exit 1
  fi
}

map_arch() {
  local uarch
  uarch="$(uname -m)"
  case "$uarch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)
      echo "Error: unsupported architecture '$uarch'. Supported: amd64, arm64." >&2
      exit 1
      ;;
  esac
}

require_root_for() {
  local path="$1"
  if [ ! -w "$(dirname "$path")" ] || { [ -e "$path" ] && [ ! -w "$path" ]; }; then
    if [ "$EUID" -ne 0 ]; then
      echo "This step requires elevated privileges. Re-running with sudo..."
      exec sudo -E bash "$0" "$@"
    fi
  fi
}

fetch_latest_asset_url() {
  local arch="$1"
  # Avoid requiring jq; use grep/sed to find the asset URL
  # We look for browser_download_url entries ending with smtp-relay-linux-${arch}
  local url
  url=$(curl -fsSL "$API_LATEST" \
    | grep -oE '"browser_download_url"\s*:\s*"[^"]+"' \
    | sed -E 's/"browser_download_url"\s*:\s*"([^"]+)"/\1/' \
    | grep "/${BINARY_NAME}-linux-${arch}$" || true)
  if [ -z "$url" ]; then
    echo "Error: could not find latest release asset for linux/${arch}." >&2
    exit 1
  fi
  echo "$url"
}

backup_if_exists() {
  local file="$1"
  if [ -f "$file" ]; then
    local backup="${file}.bak.$(date +%Y%m%d%H%M%S)"
    echo "Backing up existing $(basename "$file") to $backup"
    cp -a "$file" "$backup"
  fi
}

main() {
  require_linux
  need_cmd curl
  need_cmd install
  need_cmd systemctl

  local arch
  arch=$(map_arch)
  echo "Detected architecture: $arch"

  echo
  echo "Fetching latest release asset URL..."
  local asset_url
  asset_url=$(fetch_latest_asset_url "$arch")
  echo "Latest asset: $asset_url"

  local tmpdir
  tmpdir=$(mktemp -d)
  # trap 'rm -rf "$tmpdir"' EXIT

  echo
  echo "Downloading binary..."
  curl -fsSL "$asset_url" -o "$tmpdir/${BINARY_NAME}"
  chmod +x "$tmpdir/${BINARY_NAME}"

  echo "Installing binary to ${INSTALL_BIN_PATH}"
  if [ "$EUID" -ne 0 ]; then
    sudo install -m 0755 -d "$(dirname "$INSTALL_BIN_PATH")"
    sudo install -m 0755 "$tmpdir/${BINARY_NAME}" "$INSTALL_BIN_PATH"
  else
    install -m 0755 -d "$(dirname "$INSTALL_BIN_PATH")"
    install -m 0755 "$tmpdir/${BINARY_NAME}" "$INSTALL_BIN_PATH"
  fi

  echo
  echo "Installing systemd service to ${SERVICE_PATH}"
  service_installed=0
  if [ -f "$SERVICE_PATH" ]; then
    echo "Service file already exists; not overwriting."
  else
    if [ "$EUID" -ne 0 ]; then
      curl -fsSL "${RAW_BASE}/${BINARY_NAME}.service" | sudo tee "$SERVICE_PATH" >/dev/null
      sudo chmod 0644 "$SERVICE_PATH"
    else
      curl -fsSL "${RAW_BASE}/${BINARY_NAME}.service" -o "$SERVICE_PATH"
      chmod 0644 "$SERVICE_PATH"
    fi
    service_installed=1
  fi

  if [ "$service_installed" -eq 1 ]; then
    echo "Reloading systemd daemon..."
    if [ "$EUID" -ne 0 ]; then
      sudo systemctl daemon-reload
    else
      systemctl daemon-reload
    fi
    echo "Systemd service installed at ${SERVICE_PATH}."
  else
    echo "Service file already exists; not overwriting."
  fi

  rm -rf "$tmpdir"
  echo
  echo "Installation complete."
  echo "Binary: ${INSTALL_BIN_PATH}"
  echo "Service: ${SERVICE_PATH}"
  echo
  echo "Next steps:"
  echo "  1) Create and edit your configuration at ${CONFIG_PATH}. For example:"
  echo "     sudo curl -fsSL \"${RAW_BASE}/config.yaml\" -o \"${CONFIG_PATH}\" && sudo chmod 0644 \"${CONFIG_PATH}\""
  echo "     Then edit it to your environment (secrets, ports, relays, etc.)."
  echo "  2) Enable and start the service:"
  echo "     sudo systemctl enable ${BINARY_NAME}"
  echo "     sudo systemctl start ${BINARY_NAME}"
  echo "  3) Check status and logs:"
  echo "     sudo systemctl status ${BINARY_NAME}"
  echo "     sudo journalctl -u ${BINARY_NAME} -f"
}

main "$@"
