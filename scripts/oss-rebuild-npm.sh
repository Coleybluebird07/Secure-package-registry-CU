mkdir -p scripts
cat > scripts/oss-rebuild-npm.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

: "${SPR_PACKAGE:?SPR_PACKAGE is required}"
: "${SPR_VERSION:?SPR_VERSION is required}"
: "${SPR_REBUILT_ARTIFACT:?SPR_REBUILT_ARTIFACT is required}"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

echo "Checking OSS Rebuild availability for npm ${SPR_PACKAGE}@${SPR_VERSION}"

if ! command -v oss-rebuild >/dev/null 2>&1; then
  echo "oss-rebuild CLI is not installed" >&2
  exit 10
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker CLI is not available inside this container" >&2
  exit 10
fi

if ! oss-rebuild list npm "$SPR_PACKAGE" --verify=false | grep -Fxq "$SPR_VERSION"; then
  echo "OSS Rebuild has no rebuild record for npm ${SPR_PACKAGE}@${SPR_VERSION}" >&2
  exit 10
fi

echo "Fetching OSS Rebuild Dockerfile"
oss-rebuild get npm "$SPR_PACKAGE" "$SPR_VERSION" --verify=false --output=dockerfile > "$workdir/Dockerfile"

echo "Building OSS Rebuild image"
image_id="$(docker buildx build -q "$workdir")"

echo "Running OSS Rebuild image"
container_id="$(docker create "$image_id")"
trap 'docker rm -f "$container_id" >/dev/null 2>&1 || true; rm -rf "$workdir"' EXIT

mkdir -p "$(dirname "$SPR_REBUILT_ARTIFACT")"

# OSS Rebuild Dockerfiles commonly leave the rebuilt package artifact somewhere
# in the container filesystem. Try common archive locations first, then fall back
# to searching for package archive formats.
candidate_paths="$(
  docker export "$container_id" | tar -tf - | grep -E '\.(tgz|tar\.gz|whl|crate)$' || true
)"

if [ -z "$candidate_paths" ]; then
  echo "OSS Rebuild completed but no archive artifact was found in the image" >&2
  exit 10
fi

artifact_path="$(printf '%s\n' "$candidate_paths" | grep -E '\.tgz$|\.tar\.gz$' | head -n 1)"
if [ -z "$artifact_path" ]; then
  artifact_path="$(printf '%s\n' "$candidate_paths" | head -n 1)"
fi

echo "Copying rebuilt artifact from image path: $artifact_path"
docker export "$container_id" | tar -xO "$artifact_path" > "$SPR_REBUILT_ARTIFACT"

if [ ! -s "$SPR_REBUILT_ARTIFACT" ]; then
  echo "Rebuilt artifact was not created or is empty" >&2
  exit 10
fi

echo "Rebuilt artifact written to $SPR_REBUILT_ARTIFACT"
EOF

chmod +x scripts/oss-rebuild-npm.sh