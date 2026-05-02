#!/usr/bin/env bash
set -euo pipefail

: "${SPR_PACKAGE:?SPR_PACKAGE is required}"
: "${SPR_VERSION:?SPR_VERSION is required}"
: "${SPR_REBUILT_ARTIFACT:?SPR_REBUILT_ARTIFACT is required}"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

echo "Checking OSS Rebuild availability for npm ${SPR_PACKAGE}@${SPR_VERSION}"

OSS_REBUILD_BIN=""

if command -v oss-rebuild >/dev/null 2>&1; then
  OSS_REBUILD_BIN="oss-rebuild"
elif command -v oss_rebuild >/dev/null 2>&1; then
  OSS_REBUILD_BIN="oss_rebuild"
else
  echo "oss-rebuild CLI is not installed" >&2
  exit 10
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker CLI is not available inside this container" >&2
  exit 10
fi

echo "Using OSS Rebuild binary: ${OSS_REBUILD_BIN}"

if ! "$OSS_REBUILD_BIN" list npm "$SPR_PACKAGE" --verify=false | grep -Fxq "$SPR_VERSION"; then
  echo "OSS Rebuild has no rebuild record for npm ${SPR_PACKAGE}@${SPR_VERSION}" >&2
  exit 10
fi

echo "Fetching OSS Rebuild Dockerfile"
"$OSS_REBUILD_BIN" get npm "$SPR_PACKAGE" "$SPR_VERSION" --verify=false --output=dockerfile > "$workdir/Dockerfile"

echo "Generated Dockerfile:"
sed -n "1,220p" "$workdir/Dockerfile"

echo "Building OSS Rebuild image"
image_id="$(docker build -q "$workdir")"

if [ -z "$image_id" ]; then
  echo "Docker build did not return an image id" >&2
  exit 10
fi

echo "Built image: $image_id"

echo "Running OSS Rebuild image and capturing rebuilt artifact"
mkdir -p "$(dirname "$SPR_REBUILT_ARTIFACT")"

docker run --rm "$image_id" > "$SPR_REBUILT_ARTIFACT"

if [ ! -s "$SPR_REBUILT_ARTIFACT" ]; then
  echo "Rebuilt artifact was not created or is empty" >&2
  exit 10
fi

echo "Rebuilt artifact written to $SPR_REBUILT_ARTIFACT"
file "$SPR_REBUILT_ARTIFACT" || true
ls -lh "$SPR_REBUILT_ARTIFACT"