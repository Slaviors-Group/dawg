#!/usr/bin/env bash

set -euo pipefail

BUILD_HOST="${1:?usage: collect-artifacts.sh BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR}"
REMOTE_ARTIFACT_DIR="${2:?usage: collect-artifacts.sh BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR}"
OUTPUT_DIR="${3:?usage: collect-artifacts.sh BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR}"
TRANSFER_JOBS="${DAWG_CI_TRANSFER_JOBS:-2}"
TRANSFER_RETRIES="${DAWG_CI_TRANSFER_RETRIES:-8}"

case "${OUTPUT_DIR}" in
  ""|/|.)
    echo "Refusing unsafe artifact output directory: ${OUTPUT_DIR}" >&2
    exit 2
    ;;
esac
case "${REMOTE_ARTIFACT_DIR}" in
  *"'"*|*$'\n'*)
    echo "Refusing unsafe remote artifact directory: ${REMOTE_ARTIFACT_DIR}" >&2
    exit 2
    ;;
esac
if ! [[ "${TRANSFER_JOBS}" =~ ^[1-9][0-9]*$ ]] || ! [[ "${TRANSFER_RETRIES}" =~ ^[1-9][0-9]*$ ]]; then
  echo "Transfer job and retry counts must be positive integers." >&2
  exit 2
fi

# Preserve verified chunks across the Jenkins retry block. Failed individual
# transfers use temporary files, so an interrupted run never makes an
# incomplete chunk look reusable.
mkdir -p "${OUTPUT_DIR}/parts"

metadata_archive="${OUTPUT_DIR}/.metadata.tar"
metadata_temporary="${metadata_archive}.tmp"
rm -f "${metadata_temporary}"

metadata_downloaded=false
for ((attempt = 1; attempt <= TRANSFER_RETRIES; attempt++)); do
  echo "Downloading artifact metadata (attempt ${attempt}/${TRANSFER_RETRIES})..."
  if ssh \
      -o BatchMode=yes \
      -o ConnectTimeout=30 \
      -o ConnectionAttempts=3 \
      -o ServerAliveInterval=15 \
      -o ServerAliveCountMax=4 \
      "${BUILD_HOST}" \
      "tar -cf - -C '${REMOTE_ARTIFACT_DIR}' SHA256SUMS PARTS.sha256 doctor.txt build-metadata.txt" \
      > "${metadata_temporary}" && \
    tar -tf "${metadata_temporary}" >/dev/null 2>&1; then
    mv -f "${metadata_temporary}" "${metadata_archive}"
    metadata_downloaded=true
    break
  fi

  rm -f "${metadata_temporary}"
  if [ "${attempt}" -lt "${TRANSFER_RETRIES}" ]; then
    sleep $((attempt * 5))
  fi
done
if [ "${metadata_downloaded}" != true ]; then
  echo "Artifact metadata transfer failed after ${TRANSFER_RETRIES} attempts." >&2
  exit 1
fi

tar -xf "${metadata_archive}" -C "${OUTPUT_DIR}"

artifact_name="$(awk 'NR == 1 { print $2 }' "${OUTPUT_DIR}/SHA256SUMS")"
case "${artifact_name}" in
  DAWG_*_amd64.AppImage)
    ;;
  *)
    echo "Unsafe artifact name in SHA256SUMS: ${artifact_name}" >&2
    exit 1
    ;;
esac

part_list="${OUTPUT_DIR}/part-list"
awk '{ print $2 }' "${OUTPUT_DIR}/PARTS.sha256" > "${part_list}"
test -s "${part_list}"

part_count=0
while IFS= read -r part; do
  case "${part}" in
    "${artifact_name}.part."[0-9][0-9][0-9][0-9])
      ;;
    *)
      echo "Unsafe artifact part name: ${part}" >&2
      exit 1
      ;;
  esac
  part_count=$((part_count + 1))
done < "${part_list}"

unique_part_count="$(sort -u "${part_list}" | wc -l)"
if [ "${part_count}" -ne "${unique_part_count}" ]; then
  echo "PARTS.sha256 contains duplicate part names." >&2
  exit 1
fi

export BUILD_HOST REMOTE_ARTIFACT_DIR OUTPUT_DIR TRANSFER_RETRIES
xargs -r -P "${TRANSFER_JOBS}" -I '{}' bash -c '
  set -euo pipefail

  part="$1"
  expected_hash="$(grep -F -m1 "  ${part}" "${OUTPUT_DIR}/PARTS.sha256" | cut -d " " -f 1)"
  if ! [[ "${expected_hash}" =~ ^[0-9a-f]{64}$ ]]; then
    echo "Missing or invalid checksum for ${part}." >&2
    exit 1
  fi

  destination="${OUTPUT_DIR}/parts/${part}"
  if [ -f "${destination}" ] && \
    printf "%s  %s\n" "${expected_hash}" "${destination}" | sha256sum -c - >/dev/null 2>&1; then
    echo "Already verified ${part}; skipping."
    exit 0
  fi
  rm -f "${destination}"

  for ((attempt = 1; attempt <= TRANSFER_RETRIES; attempt++)); do
    temporary="${destination}.tmp.${BASHPID}"
    rm -f "${temporary}"
    echo "Downloading ${part} (attempt ${attempt}/${TRANSFER_RETRIES})..."

    if scp -q \
        -o BatchMode=yes \
        -o ConnectTimeout=30 \
        -o ConnectionAttempts=3 \
        -o ServerAliveInterval=15 \
        -o ServerAliveCountMax=4 \
        "${BUILD_HOST}:${REMOTE_ARTIFACT_DIR}/${part}" \
        "${temporary}" && \
      printf "%s  %s\n" "${expected_hash}" "${temporary}" | sha256sum -c - >/dev/null 2>&1; then
      mv -f "${temporary}" "${destination}"
      echo "Verified ${part}."
      exit 0
    fi

    rm -f "${temporary}"
    if [ "${attempt}" -lt "${TRANSFER_RETRIES}" ]; then
      sleep $((attempt * 5))
    fi
  done

  echo "Failed to download ${part} after ${TRANSFER_RETRIES} attempts." >&2
  exit 1
' bash '{}' < "${part_list}"

assembled_temporary="${OUTPUT_DIR}/${artifact_name}.tmp"
rm -f "${assembled_temporary}"
while IFS= read -r part; do
  cat "${OUTPUT_DIR}/parts/${part}" >> "${assembled_temporary}"
done < "${part_list}"
mv -f "${assembled_temporary}" "${OUTPUT_DIR}/${artifact_name}"

if ! (
  cd "${OUTPUT_DIR}"
  sha256sum -c SHA256SUMS
); then
  rm -f "${OUTPUT_DIR}/${artifact_name}"
  echo "Reassembled artifact failed its checksum; verified parts were retained for retry." >&2
  exit 1
fi

rm -rf "${OUTPUT_DIR}/parts"
rm -f "${part_list}" "${metadata_archive}"

echo "Collected and verified ${OUTPUT_DIR}/${artifact_name}."
