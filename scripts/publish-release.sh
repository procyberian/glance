#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TAG="v11.0.6"
DIST_DIR="${DIST_DIR:-}"
NOTES_FILE="${NOTES_FILE:-}"
GITHUB_REPO="${GITHUB_REPO:-procyberian/glance}"
CODEBERG_REPO="${CODEBERG_REPO:-procyberian/glance}"
GITHUB_API="https://api.github.com"
CODEBERG_API="https://codeberg.org/api/v1"
PUBLISH_GITHUB=1
PUBLISH_CODEBERG=1
DRY_RUN=0

usage() {
  cat <<'EOF'
Usage: scripts/publish-release.sh [options] [tag]

Publishes or updates the release for the given tag on GitHub and Codeberg,
then uploads all packaged artifacts from dist/.

Options:
  --github-only     Publish only to GitHub
  --codeberg-only   Publish only to Codeberg
  --dist-dir PATH   Read release assets from PATH instead of ./dist
  --notes-file FILE Read release notes from FILE instead of ./release-notes/<tag>.md
  --dry-run         Print the planned actions without calling either API
  -h, --help        Show this help text

Required environment variables:
  GH_TOKEN          GitHub token with release write access when GitHub publishing is enabled
  CODEBERG_TOKEN    Codeberg token with repository release write access when Codeberg publishing is enabled

Optional environment variables:
  DIST_DIR          Directory containing release assets (default: ./dist)
  NOTES_FILE        Release notes markdown file (default: ./release-notes/<tag>.md)
  GITHUB_REPO       owner/repo for GitHub (default: procyberian/glance)
  CODEBERG_REPO     owner/repo for Codeberg (default: procyberian/glance)

Recommended token scopes:
  GitHub classic PAT: repo
  GitHub fine-grained PAT: repository Contents read/write
  Codeberg token: repository write access for releases/assets on the target repo

Expected assets:
  glance-linux-amd64.tar.gz
  glance-linux-arm64.tar.gz
  glance-darwin-amd64.tar.gz
  glance-darwin-arm64.tar.gz
  glance-windows-amd64.zip
  sha256sums.txt
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --github-only)
      PUBLISH_GITHUB=1
      PUBLISH_CODEBERG=0
      shift
      ;;
    --codeberg-only)
      PUBLISH_GITHUB=0
      PUBLISH_CODEBERG=1
      shift
      ;;
    --dist-dir)
      if [[ $# -lt 2 ]]; then
        printf 'missing value for --dist-dir\n\n' >&2
        usage >&2
        exit 1
      fi
      DIST_DIR="$2"
      shift 2
      ;;
    --notes-file)
      if [[ $# -lt 2 ]]; then
        printf 'missing value for --notes-file\n\n' >&2
        usage >&2
        exit 1
      fi
      NOTES_FILE="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      printf 'unknown option: %s\n\n' "$1" >&2
      usage >&2
      exit 1
      ;;
    *)
      TAG="$1"
      shift
      if [[ $# -gt 0 ]]; then
        printf 'unexpected positional arguments: %s\n\n' "$*" >&2
        usage >&2
        exit 1
      fi
      break
      ;;
  esac
done

DIST_DIR="${DIST_DIR:-$REPO_ROOT/dist}"
NOTES_FILE="${NOTES_FILE:-$REPO_ROOT/release-notes/$TAG.md}"

if [[ "$PUBLISH_GITHUB" -eq 0 && "$PUBLISH_CODEBERG" -eq 0 ]]; then
  printf 'nothing to do: no publishing target is enabled\n' >&2
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

require_file() {
  if [[ ! -f "$1" ]]; then
    printf 'missing required file: %s\n' "$1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq
require_file "$NOTES_FILE"

ASSETS=(
  "$DIST_DIR/glance-linux-amd64.tar.gz"
  "$DIST_DIR/glance-linux-arm64.tar.gz"
  "$DIST_DIR/glance-darwin-amd64.tar.gz"
  "$DIST_DIR/glance-darwin-arm64.tar.gz"
  "$DIST_DIR/glance-windows-amd64.zip"
  "$DIST_DIR/sha256sums.txt"
)

for asset in "${ASSETS[@]}"; do
  require_file "$asset"
done

describe_plan() {
  printf 'tag: %s\n' "$TAG"
  printf 'notes: %s\n' "$NOTES_FILE"
  printf 'assets:\n'
  printf '  %s\n' "${ASSETS[@]}"

  if [[ "$PUBLISH_GITHUB" -eq 1 ]]; then
    printf 'target: github (%s)\n' "$GITHUB_REPO"
  fi

  if [[ "$PUBLISH_CODEBERG" -eq 1 ]]; then
    printf 'target: codeberg (%s)\n' "$CODEBERG_REPO"
  fi
}

body_json() {
  jq -Rs . < "$1"
}

github_request() {
  local method="$1"
  local endpoint="$2"
  local data="${3:-}"
  local curl_args=(
    -fsSL
    -X "$method"
    -H "Accept: application/vnd.github+json"
    -H "Authorization: Bearer $GH_TOKEN"
  )

  if [[ -n "$data" ]]; then
    curl_args+=(-H "Content-Type: application/json" -d "$data")
  fi

  curl "${curl_args[@]}" "$GITHUB_API$endpoint"
}

codeberg_request() {
  local method="$1"
  local endpoint="$2"
  local data="${3:-}"
  local curl_args=(
    -fsSL
    -X "$method"
    -H "Accept: application/json"
    -H "Authorization: token $CODEBERG_TOKEN"
  )

  if [[ -n "$data" ]]; then
    curl_args+=(-H "Content-Type: application/json" -d "$data")
  fi

  curl "${curl_args[@]}" "$CODEBERG_API$endpoint"
}

upload_github_asset() {
  local upload_url_base="$1"
  local asset_path="$2"
  local asset_name
  asset_name="$(basename "$asset_path")"

  curl -fsSL \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @"$asset_path" \
    "$upload_url_base?name=$asset_name"
}

upload_codeberg_asset() {
  local release_id="$1"
  local asset_path="$2"
  local asset_name
  asset_name="$(basename "$asset_path")"

  curl -fsSL \
    -X POST \
    -H "Accept: application/json" \
    -H "Authorization: token $CODEBERG_TOKEN" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @"$asset_path" \
    "$CODEBERG_API/repos/$CODEBERG_REPO/releases/$release_id/assets?name=$asset_name"
}

delete_existing_github_asset() {
  local release_json="$1"
  local asset_name="$2"
  local asset_id
  asset_id="$(jq -r --arg name "$asset_name" '.assets[]? | select(.name == $name) | .id' <<< "$release_json")"
  if [[ -n "$asset_id" ]]; then
    github_request DELETE "/repos/$GITHUB_REPO/releases/assets/$asset_id" >/dev/null
  fi
}

delete_existing_codeberg_asset() {
  local release_json="$1"
  local asset_name="$2"
  local release_id
  local attachment_uuid

  release_id="$(jq -r '.id' <<< "$release_json")"
  attachment_uuid="$(jq -r --arg name "$asset_name" '.assets[]? | select(.name == $name) | .uuid // empty' <<< "$release_json")"
  if [[ -n "$attachment_uuid" ]]; then
    codeberg_request DELETE "/repos/$CODEBERG_REPO/releases/$release_id/assets/$attachment_uuid" >/dev/null
  fi
}

ensure_github_release() {
  local notes_json payload release_json
  notes_json="$(body_json "$NOTES_FILE")"

  if release_json="$(github_request GET "/repos/$GITHUB_REPO/releases/tags/$TAG" 2>/dev/null)"; then
    printf '%s' "$release_json"
    return 0
  fi

  payload="$(jq -n --arg tag "$TAG" --arg name "$TAG" --argjson body "$notes_json" '{tag_name:$tag, name:$name, body:$body, draft:false, prerelease:false, generate_release_notes:false}')"
  github_request POST "/repos/$GITHUB_REPO/releases" "$payload"
}

ensure_codeberg_release() {
  local notes_json payload release_json
  notes_json="$(body_json "$NOTES_FILE")"

  if release_json="$(codeberg_request GET "/repos/$CODEBERG_REPO/releases/tags/$TAG" 2>/dev/null)"; then
    printf '%s' "$release_json"
    return 0
  fi

  payload="$(jq -n --arg tag "$TAG" --arg name "$TAG" --argjson body "$notes_json" '{tag_name:$tag, name:$name, body:$body, draft:false, prerelease:false}')"
  codeberg_request POST "/repos/$CODEBERG_REPO/releases" "$payload"
}

publish_github() {
  local release_json upload_url_base asset asset_name

  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf 'dry-run github release for %s\n' "$TAG"
    for asset in "${ASSETS[@]}"; do
      printf 'github would upload: %s\n' "$(basename "$asset")"
    done
    return 0
  fi

  release_json="$(ensure_github_release)"
  upload_url_base="$(jq -r '.upload_url | sub("\\{\\?name,label\\}$"; "")' <<< "$release_json")"

  for asset in "${ASSETS[@]}"; do
    asset_name="$(basename "$asset")"
    delete_existing_github_asset "$release_json" "$asset_name"
    upload_github_asset "$upload_url_base" "$asset" >/dev/null
    printf 'github uploaded: %s\n' "$asset_name"
  done
}

publish_codeberg() {
  local release_json release_id asset asset_name

  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf 'dry-run codeberg release for %s\n' "$TAG"
    for asset in "${ASSETS[@]}"; do
      printf 'codeberg would upload: %s\n' "$(basename "$asset")"
    done
    return 0
  fi

  release_json="$(ensure_codeberg_release)"
  release_id="$(jq -r '.id' <<< "$release_json")"

  for asset in "${ASSETS[@]}"; do
    asset_name="$(basename "$asset")"
    delete_existing_codeberg_asset "$release_json" "$asset_name"
    upload_codeberg_asset "$release_id" "$asset" >/dev/null
    printf 'codeberg uploaded: %s\n' "$asset_name"
  done
}

if [[ "$DRY_RUN" -eq 1 ]]; then
  describe_plan
fi

if [[ "$PUBLISH_GITHUB" -eq 1 && "$DRY_RUN" -eq 0 && -z "${GH_TOKEN:-}" ]]; then
  printf 'GH_TOKEN is required\n' >&2
  exit 1
fi

if [[ "$PUBLISH_CODEBERG" -eq 1 && "$DRY_RUN" -eq 0 && -z "${CODEBERG_TOKEN:-}" ]]; then
  printf 'CODEBERG_TOKEN is required\n' >&2
  exit 1
fi

if [[ "$PUBLISH_GITHUB" -eq 1 ]]; then
  publish_github
fi

if [[ "$PUBLISH_CODEBERG" -eq 1 ]]; then
  publish_codeberg
fi

printf 'release publication completed for %s\n' "$TAG"