#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CI_CACHE_DIR="${DAWG_CI_CACHE_DIR:-${HOME}/.cache/dawg-ci}"
OUTPUT_DIR="${REPO_ROOT}/.ci-output"

export npm_config_cache="${CI_CACHE_DIR}/npm"
export CARGO_TERM_COLOR=always

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}" "${npm_config_cache}"

echo "==> Engine, extension, and version validation"
cd "${REPO_ROOT}/engine"
npm ci --no-audit --no-fund
npm run check:versions
node --check "${REPO_ROOT}/extension/background/service_worker.js"
node --check "${REPO_ROOT}/extension/content/recorder.js"
node --check "${REPO_ROOT}/extension/content/popup-panel.js"
node --test "${REPO_ROOT}/extension/tests/service_worker.test.cjs"

unformatted_files="$(gofmt -l .)"
if [ -n "${unformatted_files}" ]; then
  echo "The following Go files need gofmt:" >&2
  echo "${unformatted_files}" >&2
  exit 1
fi

go vet ./...
go test -race ./...
CGO_ENABLED=0 go build -trimpath -o "${OUTPUT_DIR}/dawg" ./cmd/dawg

echo "==> Desktop validation"
cd "${REPO_ROOT}/desktop"
npm ci --no-audit --no-fund
npx --no-install biome check .
npm run build

cd "${REPO_ROOT}/desktop/src-tauri"
cargo fmt --all -- --check
cargo clippy --locked --all-targets -- -D warnings
cargo test --locked

echo "==> Documentation validation"
cd "${REPO_ROOT}/web"
npm ci --no-audit --no-fund
npm run docs:build
npm run docs:check-seo

echo "All validation checks passed."
