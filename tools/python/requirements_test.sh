#!/usr/bin/env bash
# Fails when uv.lock changed without regenerating requirements_lock.txt.
set -euo pipefail

want="$(grep '^# uv.lock sha256: ' "$2" | cut -d' ' -f4)"
got="$(shasum -a 256 "$1" | cut -d' ' -f1)"
if [[ "$want" != "$got" ]]; then
  echo "requirements_lock.txt is stale (uv.lock $got, recorded $want)." >&2
  echo "Run: bazel run //tools/python:requirements.update" >&2
  exit 1
fi
