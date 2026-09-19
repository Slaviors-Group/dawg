#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

if [ "$(uname -s)" != "Linux" ] || [ "$(uname -m)" != "x86_64" ]; then
  echo "The AppImage pipeline currently requires Linux on x86_64." >&2
  exit 1
fi

cd "${REPO_ROOT}/desktop"
exec ./build-bundle.sh
