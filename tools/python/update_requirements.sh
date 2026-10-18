#!/usr/bin/env bash
# Re-locks the uv workspace and regenerates requirements_lock.txt (what Bazel's pip.parse reads).
# Run after changing any pyproject.toml: bazel run //tools/python:requirements.update
set -euo pipefail

uv="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
cd "${BUILD_WORKSPACE_DIRECTORY:?run me with bazel run}"

"$uv" lock
{
  echo "# Generated from uv.lock by //tools/python:requirements.update; do not edit."
  echo "# uv.lock sha256: $(shasum -a 256 uv.lock | cut -d' ' -f1)"
  "$uv" export --frozen --all-packages --all-groups --no-emit-workspace --no-header --format requirements.txt
} > requirements_lock.txt
echo "updated uv.lock and requirements_lock.txt"
