#!/usr/bin/env sh

set -eu

repo="${BOLTZ_API_REPO:-boltz-bio/boltz-compute-api-cli}"
version="${BOLTZ_API_VERSION:-latest}"
install_dir="${BOLTZ_API_INSTALL_DIR:-}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "boltz-api installer requires $1" >&2
    exit 1
  fi
}

need curl
need find
need grep
need install
need sed

case "$(uname -s)" in
  Darwin) os="macos" ;;
  Linux) os="linux" ;;
  *)
    echo "Unsupported operating system: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  i386 | i686) arch="386" ;;
  armv*) arch="arm" ;;
  *)
    echo "Unsupported CPU architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

if [ "$version" = "latest" ]; then
  release_url="https://api.github.com/repos/${repo}/releases/latest"
else
  case "$version" in
    v*) tag="$version" ;;
    *) tag="v$version" ;;
  esac
  release_url="https://api.github.com/repos/${repo}/releases/tags/${tag}"
fi

release_json="$(curl -fsSL -H "Accept: application/vnd.github+json" "$release_url")"
tag="$(printf '%s\n' "$release_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
version_number="$(printf '%s' "$tag" | sed 's/^v//')"

if [ -z "$tag" ]; then
  echo "Could not determine the boltz-api release tag" >&2
  exit 1
fi

case "$os" in
  macos) ext="zip" ;;
  linux) ext="tar.gz" ;;
esac

asset_url="$(printf '%s\n' "$release_json" \
  | sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
  | grep "boltz-api_.*_${os}_${arch}.*\\.${ext}$" \
  | head -n 1)"

if [ -z "$asset_url" ]; then
  echo "No boltz-api release asset found for ${os}/${arch} in ${tag}" >&2
  exit 1
fi

if [ -z "$install_dir" ]; then
  if command -v boltz-api >/dev/null 2>&1; then
    existing_binary="$(command -v boltz-api)"
    install_dir="$(dirname "$existing_binary")"
  else
    existing_binary=""
    install_dir="${HOME}/.local/bin"
  fi
else
  existing_binary="${install_dir}/boltz-api"
fi

if [ -x "$existing_binary" ]; then
  current_version="$("$existing_binary" --version 2>/dev/null | sed -n 's/.* \([0-9][0-9]*\.[0-9][^ ]*\).*/\1/p' | head -n 1 || true)"
  if [ "$current_version" = "$version_number" ]; then
    echo "boltz-api ${tag} is already installed at ${existing_binary}"
    exit 0
  fi
fi

tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t boltz-api)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

archive="${tmpdir}/boltz-api.${ext}"
curl -fL "$asset_url" -o "$archive"

case "$ext" in
  zip)
    need unzip
    unzip -q "$archive" -d "$tmpdir"
    ;;
  tar.gz)
    need tar
    tar -xzf "$archive" -C "$tmpdir"
    ;;
esac

binary="$(find "$tmpdir" -type f -name boltz-api | head -n 1)"
if [ -z "$binary" ]; then
  echo "Downloaded archive did not contain boltz-api" >&2
  exit 1
fi

mkdir -p "$install_dir"
install -m 0755 "$binary" "${install_dir}/boltz-api"

echo "Installed boltz-api ${tag} to ${install_dir}/boltz-api"

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    echo "Add ${install_dir} to PATH to run boltz-api without the full path." >&2
    ;;
esac
