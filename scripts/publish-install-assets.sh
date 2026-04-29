#!/usr/bin/env bash

set -euo pipefail

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "boltz-api install asset publisher requires $1" >&2
    exit 1
  fi
}

need aws
need find
need sort

dist_dir="${DIST_DIR:-dist}"
tag="${GITHUB_REF_NAME:-${BOLTZ_API_VERSION:-}}"
base_url="${BOLTZ_API_INSTALL_BASE_URL:-https://install.boltz.bio/boltz-api}"
s3_uri="${BOLTZ_API_INSTALL_S3_URI:-s3://boltz-platform-cli-install-assets/boltz-api}"
distribution_id="${BOLTZ_API_INSTALL_CLOUDFRONT_DISTRIBUTION_ID:-}"

if [ -z "$tag" ]; then
  echo "Set GITHUB_REF_NAME or BOLTZ_API_VERSION to the release tag, for example v0.10.0" >&2
  exit 1
fi

case "$tag" in
  v*) ;;
  *) tag="v${tag}" ;;
esac

version="${tag#v}"
base_url="${base_url%/}"
s3_uri="${s3_uri%/}"

if [ ! -d "$dist_dir" ]; then
  echo "Distribution directory not found: ${dist_dir}" >&2
  exit 1
fi

archives=()
while IFS= read -r archive; do
  archives+=("$archive")
done < <(
  find "$dist_dir" -maxdepth 1 -type f \
    \( -name "boltz-api_${version}_*.tar.gz" -o -name "boltz-api_${version}_*.zip" \) |
    sort
)

if [ "${#archives[@]}" -eq 0 ]; then
  echo "No boltz-api release archives found in ${dist_dir} for ${tag}" >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

checksum_file="${tmpdir}/checksums.txt"
(
  cd "$dist_dir"
  archive_names=()
  for archive in "${archives[@]}"; do
    archive_names+=("$(basename "$archive")")
  done

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${archive_names[@]}"
  else
    shasum -a 256 "${archive_names[@]}"
  fi
) >"$checksum_file"

release_json="${tmpdir}/release.json"
{
  printf '{\n'
  printf '  "tag_name": "%s",\n' "$tag"
  printf '  "assets": [\n'
  for i in "${!archives[@]}"; do
    name="$(basename "${archives[$i]}")"
    comma=","
    if [ "$i" -eq "$((${#archives[@]} - 1))" ]; then
      comma=""
    fi
    printf '    {"name": "%s", "browser_download_url": "%s/releases/%s/%s"}%s\n' "$name" "$base_url" "$tag" "$name" "$comma"
  done
  printf '  ]\n'
  printf '}\n'
} >"$release_json"

release_s3_uri="${s3_uri}/releases/${tag}"
mutable_cache_control="public, max-age=300"
immutable_cache_control="public, max-age=31536000, immutable"

for archive in "${archives[@]}"; do
  aws s3 cp "$archive" "${release_s3_uri}/$(basename "$archive")" \
    --content-type "application/octet-stream" \
    --cache-control "$immutable_cache_control"
done

aws s3 cp "$checksum_file" "${release_s3_uri}/checksums.txt" \
  --content-type "text/plain; charset=utf-8" \
  --cache-control "$immutable_cache_control"

aws s3 cp "$release_json" "${release_s3_uri}/release.json" \
  --content-type "application/json; charset=utf-8" \
  --cache-control "$immutable_cache_control"

aws s3 cp "scripts/install.sh" "${s3_uri}/install.sh" \
  --content-type "text/x-shellscript; charset=utf-8" \
  --cache-control "$mutable_cache_control"

aws s3 cp "scripts/install.ps1" "${s3_uri}/install.ps1" \
  --content-type "text/plain; charset=utf-8" \
  --cache-control "$mutable_cache_control"

# Publish latest metadata last so installers never observe a latest.json that
# points at release artifacts that have not been uploaded yet.
aws s3 cp "$release_json" "${s3_uri}/latest.json" \
  --content-type "application/json; charset=utf-8" \
  --cache-control "$mutable_cache_control"

if [ -n "$distribution_id" ]; then
  aws cloudfront create-invalidation \
    --distribution-id "$distribution_id" \
    --paths "/boltz-api/install.sh" "/boltz-api/install.ps1" "/boltz-api/latest.json" >/dev/null
fi

echo "Published boltz-api ${tag} install assets to ${s3_uri}"
