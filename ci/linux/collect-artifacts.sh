#!/usr/bin/env bash

set -euo pipefail

BUILD_HOST="${1:?usage: collect-artifacts.sh BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR}"
REMOTE_ARTIFACT_DIR="${2:?usage: collect-artifacts.sh BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR}"
OUTPUT_DIR="${3:?usage: collect-artifacts.sh BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR}"
TRANSFER_JOBS="${DAWG_CI_TRANSFER_JOBS:-6}"

case "${OUTPUT_DIR}" in
  ""|/|.)
    echo "Refusing unsafe artifact output directory: ${OUTPUT_DIR}" >&2
    exit 2
    ;;
esac

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}/parts"

ssh -o BatchMode=yes "${BUILD_HOST}" \
  "tar -cf - -C '${REMOTE_ARTIFACT_DIR}' SHA256SUMS doctor.txt build-metadata.txt" |
  tar -xf - -C "${OUTPUT_DIR}"

artifact_name="$(awk 'NR == 1 { print $2 }' "${OUTPUT_DIR}/SHA256SUMS")"
case "${artifact_name}" in
  DAWG_*_amd64.AppImage)
    ;;
  *)
    echo "Unsafe artifact name in SHA256SUMS: ${artifact_name}" >&2
    exit 1
    ;;
esac

ssh -o BatchMode=yes "${BUILD_HOST}" \
  "find '${REMOTE_ARTIFACT_DIR}' -maxdepth 1 -type f \
    -name '${artifact_name}.part.*' -printf '%f\\n' | sort" \
  > "${OUTPUT_DIR}/part-list"
test -s "${OUTPUT_DIR}/part-list"

export BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR
xargs -r -P "${TRANSFER_JOBS}" -I '{}' bash -c '
  set -euo pipefail
  part="$1"
  scp -q \
    -o BatchMode=yes \
    -o ServerAliveInterval=30 \
    -o ServerAliveCountMax=6 \
    "${BUILD_HOST}:${REMOTE_ARTIFACT_DIR}/${part}" \
    "${OUTPUT_DIR}/parts/${part}"
' bash '{}' < "${OUTPUT_DIR}/part-list"

while IFS= read -r part; do
  cat "${OUTPUT_DIR}/parts/${part}"
done < "${OUTPUT_DIR}/part-list" > "${OUTPUT_DIR}/${artifact_name}"

rm -rf "${OUTPUT_DIR}/parts"
rm -f "${OUTPUT_DIR}/part-list"
(
  cd "${OUTPUT_DIR}"
  sha256sum -c SHA256SUMS
)

echo "Collected and verified ${OUTPUT_DIR}/${artifact_name}."
