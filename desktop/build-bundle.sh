#!/usr/bin/env bash
# build-bundle.sh — Assembles the self-contained DAWG AppImage for Linux (Strategy A)
set -euo pipefail

SKIP_DOWNLOAD=false
SKIP_TAURI=false
MITM_VERSION="12.2.3"
NODE_VERSION="22.14.0"

for arg in "$@"; do
  case $arg in
    --skip-download)
      SKIP_DOWNLOAD=true
      ;;
    --skip-tauri)
      SKIP_TAURI=true
      ;;
    --mitm-version=*)
      MITM_VERSION="${arg#*=}"
      ;;
    --node-version=*)
      NODE_VERSION="${arg#*=}"
      ;;
    *)
      echo "Unknown argument: ${arg}" >&2
      exit 2
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
BROWSERS_TARGET_DIR="${RESOURCES_DIR}/browsers"

mkdir -p "${CACHE_DIR}" "${BINARIES_DIR}" "${MITM_TARGET_DIR}" "${NODE_TARGET_DIR}" "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}" "${BROWSERS_TARGET_DIR}"

# 1. Compile Go Engine
echo -e "\n[1/7] Compiling DAWG Go Engine for Linux..."
cd "${ENGINE_DIR}"
ENGINE_BIN="${BINARIES_DIR}/dawg"
# The engine is a Tauri resource, not the primary AppImage executable. Keeping
# it static prevents linuxdeploy from trying to rewrite its ELF RPATH with the
# older bundled patchelf, which crashes on Go binaries built with CGO.
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${ENGINE_BIN}" ./cmd/dawg
chmod +x "${ENGINE_BIN}"
echo "  -> Engine built at ${ENGINE_BIN}"

# 2. Stage Standalone mitmdump
echo -e "\n[2/7] Staging Standalone mitmdump (v${MITM_VERSION})..."
MITM_BIN="${MITM_TARGET_DIR}/mitmdump"
MITM_TAR="${CACHE_DIR}/mitmproxy-${MITM_VERSION}-linux.tar.gz"

if [ "${SKIP_DOWNLOAD}" = false ] || [ ! -x "${MITM_BIN}" ]; then
  if [ ! -f "${MITM_TAR}" ]; then
    if [ "${SKIP_DOWNLOAD}" = true ]; then
      echo "  Missing cached mitmdump archive: ${MITM_TAR}" >&2
      exit 1
    fi
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
echo -e "\n[3/7] Staging Portable Node.js (v${NODE_VERSION})..."
NODE_BIN="${NODE_TARGET_DIR}/node"
NODE_TAR="${CACHE_DIR}/node-v${NODE_VERSION}-linux-x64.tar.xz"

if [ "${SKIP_DOWNLOAD}" = false ] || [ ! -x "${NODE_BIN}" ]; then
  if [ ! -f "${NODE_TAR}" ]; then
    if [ "${SKIP_DOWNLOAD}" = true ]; then
      echo "  Missing cached Node.js archive: ${NODE_TAR}" >&2
      exit 1
    fi
    NODE_URL="https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz"
    echo "  Downloading ${NODE_URL}..."
    curl -fsSL "${NODE_URL}" -o "${NODE_TAR}"
  fi
  TEMP_NODE_EXTRACT="$(mktemp -d "${CACHE_DIR}/node_extract.XXXXXX")"
  tar -xf "${NODE_TAR}" -C "${TEMP_NODE_EXTRACT}" --strip-components=1
  cp "${TEMP_NODE_EXTRACT}/bin/node" "${NODE_BIN}"
  chmod +x "${NODE_BIN}"
  rm -rf "${TEMP_NODE_EXTRACT}"
fi
echo "  -> Node.js staged at ${NODE_BIN}"

# 4. Provision Playwright & rrweb
echo -e "\n[4/7] Provisioning Playwright & rrweb dependencies..."
RESOURCES_NODE_MODULES="${RESOURCES_DIR}/node_modules"
if [ -d "${ENGINE_DIR}/node_modules" ]; then
  rm -rf "${RESOURCES_NODE_MODULES}"
  cp -a "${ENGINE_DIR}/node_modules" "${RESOURCES_NODE_MODULES}"
else
  cp "${ENGINE_DIR}/package.json" "${RESOURCES_DIR}/package.json"
  cp "${ENGINE_DIR}/package-lock.json" "${RESOURCES_DIR}/package-lock.json"
  cd "${RESOURCES_DIR}"
  npm ci --omit=dev
fi
echo "  -> Node modules staged at ${RESOURCES_NODE_MODULES}"

# 5. Provision the Playwright browser into the resource tree. Playwright uses
# this directory at runtime via PLAYWRIGHT_BROWSERS_PATH set by the Tauri shell.
echo -e "\n[5/7] Provisioning bundled Chromium..."
PLAYWRIGHT_CLI="${RESOURCES_NODE_MODULES}/playwright/cli.js"
if [ "${SKIP_DOWNLOAD}" = false ]; then
  PLAYWRIGHT_BROWSERS_PATH="${BROWSERS_TARGET_DIR}" "${NODE_BIN}" "${PLAYWRIGHT_CLI}" install chromium
elif ! find "${BROWSERS_TARGET_DIR}" -mindepth 1 -maxdepth 1 -type d -name 'chromium-*' -print -quit | grep -q .; then
  echo "  Bundled Chromium is missing; rerun without --skip-download." >&2
  exit 1
fi
echo "  -> Chromium staged at ${BROWSERS_TARGET_DIR}"

# 6. Copy Engine Scripts & Schema
echo -e "\n[6/7] Copying engine scripts and schema..."
rm -rf "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}"
mkdir -p "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}"
cp -a "${ENGINE_DIR}/scripts/." "${SCRIPTS_TARGET_DIR}/"
cp -a "${SCHEMA_DIR}/." "${SCHEMA_TARGET_DIR}/"
echo "  -> Scripts & Schema staged into ${RESOURCES_DIR}"

# 7. Verify Staged Bundle via DAWG Doctor
echo -e "\n[7/7] Verifying Staged Monolithic Bundle with 'dawg doctor'..."
export DAWG_RESOURCES_DIR="${RESOURCES_DIR}"
export PLAYWRIGHT_BROWSERS_PATH="${BROWSERS_TARGET_DIR}"
"${ENGINE_BIN}" doctor

if [ "${SKIP_TAURI}" = false ]; then
  echo -e "\nBuilding Tauri Linux AppImage..."
  cd "${DESKTOP_DIR}"
  # Tauri's cached linuxdeploy ships an old strip that cannot read the RELR
  # sections used by current Fedora system libraries. These libraries are
  # already stripped; disabling linuxdeploy's redundant pass is safe.
  NO_STRIP=1 npm run tauri build -- --bundles appimage
  echo -e "\n🎉 Monolithic DAWG Desktop AppImage created successfully!"
else
  echo -e "\nBundle staging complete (Tauri build skipped)."
fi
