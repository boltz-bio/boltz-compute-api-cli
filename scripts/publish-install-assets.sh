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
s3_path="${s3_uri#s3://}"
s3_bucket="${s3_path%%/*}"
s3_key_prefix="${s3_path#*/}"
if [ "$s3_key_prefix" = "$s3_path" ]; then
  s3_key_prefix=""
fi
latest_key="${s3_key_prefix:+${s3_key_prefix}/}latest.json"

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

extract_release_tag() {
  sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$1" | head -n 1
}

version_gte() {
  local left="${1#v}"
  local right="${2#v}"
  local la lb lc ra rb rc

  if [ -z "$right" ]; then
    return 0
  fi

  IFS=. read -r la lb lc _ <<EOF
$left
EOF
  IFS=. read -r ra rb rc _ <<EOF
$right
EOF

  la="${la%%[!0-9]*}"
  lb="${lb%%[!0-9]*}"
  lc="${lc%%[!0-9]*}"
  ra="${ra%%[!0-9]*}"
  rb="${rb%%[!0-9]*}"
  rc="${rc%%[!0-9]*}"

  la="${la:-0}"
  lb="${lb:-0}"
  lc="${lc:-0}"
  ra="${ra:-0}"
  rb="${rb:-0}"
  rc="${rc:-0}"

  if ((10#$la > 10#$ra)); then return 0; fi
  if ((10#$la < 10#$ra)); then return 1; fi
  if ((10#$lb > 10#$rb)); then return 0; fi
  if ((10#$lb < 10#$rb)); then return 1; fi
  if ((10#$lc >= 10#$rc)); then return 0; fi
  return 1
}

publish_latest_json() {
  local attempt current_json current_etag current_tag put_args

  for attempt in 1 2 3 4 5; do
    current_json="${tmpdir}/current-latest-${attempt}.json"
    current_etag=""
    current_tag=""

    if current_etag="$(aws s3api get-object \
      --bucket "$s3_bucket" \
      --key "$latest_key" \
      "$current_json" \
      --query ETag \
      --output text 2>/dev/null)"; then
      current_tag="$(extract_release_tag "$current_json")"

      if ! version_gte "$tag" "$current_tag"; then
        echo "Skipping latest.json update because current latest ${current_tag} is newer than ${tag}."
        return 1
      fi

      put_args=(--if-match "$current_etag")
    else
      put_args=(--if-none-match "*")
    fi

    if aws s3api put-object \
      --bucket "$s3_bucket" \
      --key "$latest_key" \
      --body "$release_json" \
      --content-type "application/json; charset=utf-8" \
      --cache-control "$mutable_cache_control" \
      "${put_args[@]}" >/dev/null 2>&1; then
      echo "Published latest.json for boltz-api ${tag}."
      return 0
    fi

    echo "latest.json changed while publishing ${tag}; retrying (${attempt}/5)." >&2
    sleep "$attempt"
  done

  echo "Could not update latest.json for ${tag} after concurrent write retries." >&2
  exit 1
}

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
# points at release artifacts that have not been uploaded yet. Use S3
# conditional writes so concurrent release jobs cannot roll latest backward.
latest_updated=0
if publish_latest_json; then
  latest_updated=1
fi

if [ -n "$distribution_id" ]; then
  invalidation_paths=("/boltz-api/install.sh" "/boltz-api/install.ps1")
  if [ "$latest_updated" -eq 1 ]; then
    invalidation_paths+=("/boltz-api/latest.json")
  fi
  aws cloudfront create-invalidation \
    --distribution-id "$distribution_id" \
    --paths "${invalidation_paths[@]}" >/dev/null
fi

echo "Published boltz-api ${tag} install assets to ${s3_uri}"
