#!/usr/bin/env bash
# build-bundle.sh — Assembles the self-contained monolithic DAWG desktop bundle for Linux (Strategy A)
set -euo pipefail

SKIP_DOWNLOAD=false
SKIP_TAURI=false
MITM_VERSION="12.2.3"
NODE_VERSION="22.14.0"

for arg in "$@"; do
  case $arg in
    --skip-download)
      SKIP_DOWNLOAD=true
      shift
      ;;
    --skip-tauri)
      SKIP_TAURI=true
      shift
      ;;
    --mitm-version=*)
      MITM_VERSION="${arg#*=}"
      shift
      ;;
    --node-version=*)
      NODE_VERSION="${arg#*=}"
      shift
      ;;
  esac
done

echo "========================================================"
echo "  DAWG Desktop Linux Monolithic Assembly Pipeline"
echo "========================================================"

DESKTOP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${DESKTOP_DIR}/.." && pwd)"
ENGINE_DIR="${REPO_ROOT}/engine"
SCHEMA_DIR="${REPO_ROOT}/schema"
CACHE_DIR="${DESKTOP_DIR}/.cache"
RESOURCES_DIR="${DESKTOP_DIR}/src-tauri/resources"

BINARIES_DIR="${RESOURCES_DIR}/binaries"
MITM_TARGET_DIR="${BINARIES_DIR}/mitmdump"
NODE_TARGET_DIR="${BINARIES_DIR}/node"
SCRIPTS_TARGET_DIR="${RESOURCES_DIR}/scripts"
SCHEMA_TARGET_DIR="${RESOURCES_DIR}/schema"

mkdir -p "${CACHE_DIR}" "${BINARIES_DIR}" "${MITM_TARGET_DIR}" "${NODE_TARGET_DIR}" "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}"

# 1. Compile Go Engine
echo -e "\n[1/6] Compiling DAWG Go Engine for Linux..."
cd "${ENGINE_DIR}"
ENGINE_BIN="${BINARIES_DIR}/dawg"
go build -o "${ENGINE_BIN}" ./cmd/dawg
chmod +x "${ENGINE_BIN}"
echo "  -> Engine built at ${ENGINE_BIN}"

# 2. Stage Standalone mitmdump
echo -e "\n[2/6] Staging Standalone mitmdump (v${MITM_VERSION})..."
MITM_BIN="${MITM_TARGET_DIR}/mitmdump"
MITM_TAR="${CACHE_DIR}/mitmproxy-${MITM_VERSION}-linux.tar.gz"

if [ ! -f "${MITM_BIN}" ] || [ "${SKIP_DOWNLOAD}" = false ]; then
  if [ ! -f "${MITM_TAR}" ]; then
    MITM_URL="https://downloads.mitmproxy.org/${MITM_VERSION}/mitmproxy-${MITM_VERSION}-linux-x86_64.tar.gz"
    echo "  Downloading ${MITM_URL}..."
    curl -fsSL "${MITM_URL}" -o "${MITM_TAR}" || \
    curl -fsSL "https://github.com/mitmproxy/mitmproxy/releases/download/v${MITM_VERSION}/mitmproxy-${MITM_VERSION}-linux-x86_64.tar.gz" -o "${MITM_TAR}"
  fi
  tar -xzf "${MITM_TAR}" -C "${MITM_TARGET_DIR}" mitmdump
  chmod +x "${MITM_BIN}"
fi
echo "  -> mitmdump staged at ${MITM_BIN}"

# 3. Stage Portable Node.js
echo -e "\n[3/6] Staging Portable Node.js (v${NODE_VERSION})..."
NODE_BIN="${NODE_TARGET_DIR}/node"
NODE_TAR="${CACHE_DIR}/node-v${NODE_VERSION}-linux-x64.tar.xz"

if [ ! -f "${NODE_BIN}" ] || [ "${SKIP_DOWNLOAD}" = false ]; then
  if [ ! -f "${NODE_TAR}" ]; then
    NODE_URL="https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz"
    echo "  Downloading ${NODE_URL}..."
    curl -fsSL "${NODE_URL}" -o "${NODE_TAR}"
  fi
  TEMP_NODE_EXTRACT="${CACHE_DIR}/node_extract"
  mkdir -p "${TEMP_NODE_EXTRACT}"
  tar -xf "${NODE_TAR}" -C "${TEMP_NODE_EXTRACT}" --strip-components=1
  cp "${TEMP_NODE_EXTRACT}/bin/node" "${NODE_BIN}"
  chmod +x "${NODE_BIN}"
  rm -rf "${TEMP_NODE_EXTRACT}"
fi
echo "  -> Node.js staged at ${NODE_BIN}"

# 4. Provision Playwright & rrweb
echo -e "\n[4/6] Provisioning Playwright & rrweb dependencies..."
RESOURCES_NODE_MODULES="${RESOURCES_DIR}/node_modules"
if [ -d "${ENGINE_DIR}/node_modules" ]; then
  cp -r "${ENGINE_DIR}/node_modules" "${RESOURCES_NODE_MODULES}"
else
  cp "${ENGINE_DIR}/package.json" "${RESOURCES_DIR}/package.json"
  cd "${RESOURCES_DIR}"
  npm install --omit=dev
fi
echo "  -> Node modules staged at ${RESOURCES_NODE_MODULES}"

# 5. Copy Engine Scripts & Schema
echo -e "\n[5/6] Copying engine scripts and schema..."
cp -r "${ENGINE_DIR}/scripts/"* "${SCRIPTS_TARGET_DIR}/"
cp -r "${SCHEMA_DIR}/"* "${SCHEMA_TARGET_DIR}/"
echo "  -> Scripts & Schema staged into ${RESOURCES_DIR}"

# 6. Verify Staged Bundle via DAWG Doctor
echo -e "\n[6/6] Verifying Staged Monolithic Bundle with 'dawg doctor'..."
export DAWG_RESOURCES_DIR="${RESOURCES_DIR}"
"${ENGINE_BIN}" doctor

if [ "${SKIP_TAURI}" = false ]; then
  echo -e "\nBuilding Tauri Linux Distribution Packages (.deb, .AppImage)..."
  cd "${DESKTOP_DIR}"
  npm run tauri build
  echo -e "\n🎉 Monolithic DAWG Desktop Linux packages created successfully!"
else
  echo -e "\nBundle staging complete (Tauri build skipped)."
fi

