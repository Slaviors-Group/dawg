#!/usr/bin/env bash

set -euo pipefail

NODE_VERSION="${DAWG_CI_NODE_VERSION:-24.18.0}"
NVM_DIR="${NVM_DIR:-${HOME}/.nvm}"

if [ ! -s "${NVM_DIR}/nvm.sh" ]; then
  echo "nvm was not found at ${NVM_DIR}/nvm.sh" >&2
  exit 1
fi

# nvm is installed per-user on the dedicated build account.
# shellcheck source=/dev/null
source "${NVM_DIR}/nvm.sh"
nvm use --silent "${NODE_VERSION}"

export PATH="/usr/local/go/bin:${HOME}/.cargo/bin:${PATH}"

for command_name in node npm go rustc cargo; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "Required tool is unavailable: ${command_name}" >&2
    exit 1
  fi
done

exec "$@"
