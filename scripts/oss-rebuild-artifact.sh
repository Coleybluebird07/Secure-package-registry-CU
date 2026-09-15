#!/usr/bin/env bash
set -euo pipefail

: "${SPR_PACKAGE:?SPR_PACKAGE is required}"
: "${SPR_VERSION:?SPR_VERSION is required}"
: "${SPR_ECOSYSTEM:?SPR_ECOSYSTEM is required}"
: "${SPR_REBUILT_ARTIFACT:?SPR_REBUILT_ARTIFACT is required}"

workdir="$(mktemp -d)"
container_id=""
trap 'if [ -n "$container_id" ]; then docker rm -f "$container_id" >/dev/null 2>&1 || true; fi; rm -rf "$workdir"' EXIT

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

case "$SPR_ECOSYSTEM" in
  npm)
    OSS_ECOSYSTEM="npm"
    ;;
  pypi)
    OSS_ECOSYSTEM="pypi"
    ;;
  cargo)
    OSS_ECOSYSTEM="cratesio"
    ;;
  go)
    echo "OSS Rebuild does not currently support Go modules in this SPR worker" >&2
    exit 10
    ;;
  *)
    echo "Unsupported ecosystem for OSS Rebuild: $SPR_ECOSYSTEM" >&2
    exit 10
    ;;
esac

escape_regex() {
  printf '%s' "$1" | sed 's/[][\/.^$*+?{}()|]/\\&/g'
}

escaped_package="$(escape_regex "$SPR_PACKAGE")"
escaped_version="$(escape_regex "$SPR_VERSION")"

echo "Checking OSS Rebuild availability for ${OSS_ECOSYSTEM} ${SPR_PACKAGE}@${SPR_VERSION}"
echo "Using OSS Rebuild binary: ${OSS_REBUILD_BIN}"

oss_records="$("$OSS_REBUILD_BIN" list "$OSS_ECOSYSTEM" "$SPR_PACKAGE" || true)"

if ! printf '%s\n' "$oss_records" | grep -Eq "^${OSS_ECOSYSTEM}/${escaped_package}/${escaped_version}/"; then
  supported_version="$(
    printf '%s\n' "$oss_records" \
      | awk -F/ -v eco="$OSS_ECOSYSTEM" -v pkg="$SPR_PACKAGE" '$1 == eco && $2 == pkg { print $3 }' \
      | sort -V \
      | tail -n 1
  )"

  echo "OSS Rebuild has no rebuild record for ${OSS_ECOSYSTEM} ${SPR_PACKAGE}@${SPR_VERSION}" >&2

  if [ -n "$supported_version" ]; then
    echo "Newest supported OSS Rebuild version: ${supported_version}" >&2
  fi

  exit 10
fi

echo "Fetching OSS Rebuild Dockerfile"
"$OSS_REBUILD_BIN" get "$OSS_ECOSYSTEM" "$SPR_PACKAGE" "$SPR_VERSION" --output=dockerfile > "$workdir/Dockerfile"

echo "Generated Dockerfile:"
sed -n "1,220p" "$workdir/Dockerfile"

echo "Building OSS Rebuild image"
image_id="$(docker build -q "$workdir")"

if [ -z "$image_id" ]; then
  echo "Docker build did not return an image id" >&2
  exit 10
fi

echo "Built image: $image_id"

echo "Running OSS Rebuild image"
container_id="$(docker create "$image_id")"

docker start -a "$container_id"

echo "Copying rebuilt artifact from /out"
mkdir -p "$workdir/out"
docker cp "$container_id:/out/." "$workdir/out" 2>/dev/null || {
  echo "OSS Rebuild container did not produce an /out directory" >&2
  exit 10
}

echo "Files produced by OSS Rebuild:"
find "$workdir/out" -maxdepth 3 -type f -print

artifact_path="$(
  find "$workdir/out" -type f \( \
    -name "*.tgz" -o \
    -name "*.tar.gz" -o \
    -name "*.whl" -o \
    -name "*.crate" \
  \) | sort | head -n 1
)"

if [ -z "$artifact_path" ]; then
  echo "OSS Rebuild completed but no supported artifact file was found in /out" >&2
  exit 10
fi

mkdir -p "$(dirname "$SPR_REBUILT_ARTIFACT")"
cp "$artifact_path" "$SPR_REBUILT_ARTIFACT"

if [ ! -s "$SPR_REBUILT_ARTIFACT" ]; then
  echo "Rebuilt artifact was not created or is empty" >&2
  exit 10
fi

echo "Rebuilt artifact copied from $artifact_path"
echo "Rebuilt artifact written to $SPR_REBUILT_ARTIFACT"
file "$SPR_REBUILT_ARTIFACT" || true
ls -lh "$SPR_REBUILT_ARTIFACT"