#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
BUNDLE_DIR="${REPO_ROOT}/desktop/src-tauri/target/release/bundle/appimage"
ARTIFACT_DIR="${REPO_ROOT}/.ci-artifacts"

mapfile -t appimages < <(find "${BUNDLE_DIR}" -maxdepth 1 -type f -name 'DAWG_*_amd64.AppImage' -print)
if [ "${#appimages[@]}" -ne 1 ]; then
  echo "Expected exactly one DAWG amd64 AppImage, found ${#appimages[@]}." >&2
  exit 1
fi

appimage="${appimages[0]}"
rm -rf "${ARTIFACT_DIR}"
mkdir -p "${ARTIFACT_DIR}"
cp "${appimage}" "${ARTIFACT_DIR}/"

artifact_name="$(basename "${appimage}")"
(
  cd "${ARTIFACT_DIR}"
  sha256sum "${artifact_name}" > SHA256SUMS
  # Jenkins reaches this host through a userspace Tailscale SOCKS proxy. A
  # handful of parallel SSH streams is substantially faster than one stream.
  split --bytes=16M --numeric-suffixes=0 --suffix-length=4 \
    "${artifact_name}" "${artifact_name}.part."
)

extract_root="$(mktemp -d)"
trap 'rm -rf "${extract_root}"' EXIT
(
  cd "${extract_root}"
  "${ARTIFACT_DIR}/${artifact_name}" --appimage-extract >/dev/null
)

appdir="${extract_root}/squashfs-root"
resources="${appdir}/usr/lib/DAWG/resources"
engine="${resources}/binaries/dawg"

test -x "${engine}"
test -f "${resources}/extension/manifest.json"
test -f "${resources}/extension/content/popup-panel.js"
test -f "${resources}/scripts/replay-browser.cjs"

DAWG_RESOURCES_DIR="${resources}" \
PLAYWRIGHT_BROWSERS_PATH="${resources}/browsers" \
  "${engine}" doctor | tee "${ARTIFACT_DIR}/doctor.txt"

{
  echo "git_commit=${DAWG_CI_GIT_COMMIT:-unknown}"
  echo "built_at_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "builder=$(uname -srmo)"
  echo "node=$(node --version)"
  echo "go=$(go version)"
  echo "rustc=$(rustc --version)"
  echo "cargo=$(cargo --version)"
} > "${ARTIFACT_DIR}/build-metadata.txt"

echo "Verified ${artifact_name}."
