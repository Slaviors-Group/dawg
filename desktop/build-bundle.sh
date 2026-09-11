#!/usr/bin/env bash
# build-bundle.sh — Assembles the self-contained DAWG AppImage for Linux
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
EXTENSION_DIR="${REPO_ROOT}/extension"
CACHE_DIR="${DESKTOP_DIR}/.cache"
RESOURCES_DIR="${DESKTOP_DIR}/src-tauri/resources"

BINARIES_DIR="${RESOURCES_DIR}/binaries"
MITM_TARGET_DIR="${BINARIES_DIR}/mitmdump"
NODE_TARGET_DIR="${BINARIES_DIR}/node"
SCRIPTS_TARGET_DIR="${RESOURCES_DIR}/scripts"
SCHEMA_TARGET_DIR="${RESOURCES_DIR}/schema"
EXTENSION_TARGET_DIR="${RESOURCES_DIR}/extension"
BROWSERS_TARGET_DIR="${RESOURCES_DIR}/browsers"

mkdir -p "${CACHE_DIR}" "${BINARIES_DIR}" "${MITM_TARGET_DIR}" "${NODE_TARGET_DIR}" "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}" "${BROWSERS_TARGET_DIR}"

# Do not package stale Windows executables when this checkout has previously
# been used to assemble an NSIS bundle.
rm -f \
  "${BINARIES_DIR}/dawg.exe" \
  "${MITM_TARGET_DIR}/mitmdump.exe" \
  "${NODE_TARGET_DIR}/node.exe"

# 1. Compile Go Engine
echo -e "\n[1/7] Compiling DAWG Go Engine for Linux..."
cd "${ENGINE_DIR}"
ENGINE_BIN="${BINARIES_DIR}/dawg"
# A CGO-free PIE stays self-contained, while exposing the ELF interpreter and
# dynamic table that linuxdeploy expects. A plain static executable makes
# linuxdeploy treat ldd's expected non-zero result as fatal; Fedora's CGO PIE,
# on the other hand, uses RELR sections unsupported by its bundled patchelf.
CGO_ENABLED=0 go build -trimpath -buildmode=pie -ldflags="-s -w" -o "${ENGINE_BIN}" ./cmd/dawg
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

# 4. Provision Playwright for bundled replay. Capture uses the user's extension.
echo -e "\n[4/7] Provisioning browser dependencies..."
RESOURCES_NODE_MODULES="${RESOURCES_DIR}/node_modules"
if [ -d "${ENGINE_DIR}/node_modules" ]; then
  rm -rf "${RESOURCES_NODE_MODULES}"
  cp -a "${ENGINE_DIR}/node_modules" "${RESOURCES_NODE_MODULES}"
elif [ "${SKIP_DOWNLOAD}" = true ]; then
  if [ ! -f "${RESOURCES_NODE_MODULES}/playwright/cli.js" ] || \
     ! cmp -s "${ENGINE_DIR}/package-lock.json" "${RESOURCES_DIR}/package-lock.json"; then
    echo "  Cached Node dependencies are missing or stale; rerun without --skip-download." >&2
    exit 1
  fi
  npm --prefix "${RESOURCES_DIR}" ls --omit=dev >/dev/null
else
  cp "${ENGINE_DIR}/package.json" "${RESOURCES_DIR}/package.json"
  cp "${ENGINE_DIR}/package-lock.json" "${RESOURCES_DIR}/package-lock.json"
  cd "${RESOURCES_DIR}"
  npm ci --omit=dev
fi
echo "  -> Node modules staged at ${RESOURCES_NODE_MODULES}"

# 5. Provision Chromium into the resource tree used by replay at runtime.
echo -e "\n[5/7] Provisioning bundled Chromium..."
PLAYWRIGHT_CLI="${RESOURCES_NODE_MODULES}/playwright/cli.js"
if [ "${SKIP_DOWNLOAD}" = false ]; then
  # A shared/dual-boot checkout may contain a Chromium build for another OS.
  # Recreate the directory so the AppImage contains Linux assets only.
  rm -rf "${BROWSERS_TARGET_DIR}"
  mkdir -p "${BROWSERS_TARGET_DIR}"
  PLAYWRIGHT_BROWSERS_PATH="${BROWSERS_TARGET_DIR}" "${NODE_BIN}" "${PLAYWRIGHT_CLI}" install chromium --no-shell
elif ! find "${BROWSERS_TARGET_DIR}" -mindepth 1 -maxdepth 1 -type d -name 'chromium-*' -print -quit | grep -q .; then
  echo "  Bundled Chromium is missing; rerun without --skip-download." >&2
  exit 1
fi
# Replay passes the full Chromium executable explicitly and DAWG does not
# record video. Playwright's headless shell and ffmpeg downloads are therefore
# redundant; ffmpeg is also a static ELF that linuxdeploy cannot inspect.
find "${BROWSERS_TARGET_DIR}" -mindepth 1 -maxdepth 1 -type d -name 'chromium_headless_shell-*' -exec rm -rf -- {} +
find "${BROWSERS_TARGET_DIR}" -mindepth 1 -maxdepth 1 -type d -name 'ffmpeg-*' -exec rm -rf -- {} +
echo "  -> Chromium staged at ${BROWSERS_TARGET_DIR}"

# 6. Copy Engine Scripts, Schema & Browser Extension
echo -e "\n[6/7] Copying engine scripts, schema, and browser extension..."
rm -rf "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}" "${EXTENSION_TARGET_DIR}"
mkdir -p "${SCRIPTS_TARGET_DIR}" "${SCHEMA_TARGET_DIR}" "${EXTENSION_TARGET_DIR}"
cp "${ENGINE_DIR}/scripts/replay-browser.cjs" "${SCRIPTS_TARGET_DIR}/"
cp -a "${SCHEMA_DIR}/." "${SCHEMA_TARGET_DIR}/"
cp -a "${EXTENSION_DIR}/." "${EXTENSION_TARGET_DIR}/"
rm -rf "${EXTENSION_TARGET_DIR}/tests"
echo "  -> Scripts, schema, and installable extension staged into ${RESOURCES_DIR}"

# 7. Verify Staged Bundle via DAWG Doctor
echo -e "\n[7/7] Verifying staged bundle with 'dawg doctor'..."
export DAWG_RESOURCES_DIR="${RESOURCES_DIR}"
export PLAYWRIGHT_BROWSERS_PATH="${BROWSERS_TARGET_DIR}"
"${ENGINE_BIN}" doctor

if [ "${SKIP_TAURI}" = false ]; then
  echo -e "\nBuilding Tauri Linux AppImage..."
  cd "${DESKTOP_DIR}"
  # Tauri and linuxdeploy each reuse generated resource staging across attempts.
  # Recreate both layers so removed runtimes or P1 resources cannot leak into
  # the next AppImage.
  rm -rf \
    "${DESKTOP_DIR}/src-tauri/target/release/resources" \
    "${DESKTOP_DIR}/src-tauri/target/release/bundle/appimage/DAWG.AppDir" \
    "${DESKTOP_DIR}/src-tauri/target/release/bundle/appimage_deb"
  # Tauri's GStreamer helper knows Debian's multiarch directory but not
  # Fedora's lib64/libexec layout. Point it at the native locations when they
  # exist so WebKit does not start with missing elements such as appsink.
  if [ -d /usr/lib64/gstreamer-1.0 ]; then
    export GSTREAMER_PLUGINS_DIR=/usr/lib64/gstreamer-1.0
  fi
  if [ -d /usr/libexec/gstreamer-1.0 ]; then
    export GSTREAMER_HELPERS_DIR=/usr/libexec/gstreamer-1.0
  fi
  # Current Fedora libraries use RELR sections that the cached linuxdeploy
  # strip pass cannot parse. They are already stripped, so skip that redundant
  # pass while linuxdeploy assembles the AppDir.
  # The GTK plugin also copies both i386 /usr/lib and native /usr/lib64 GIO
  # modules on multilib Fedora. Prefix its narrowly scoped find shim so only
  # the native libgiognutls module enters an x86_64 AppDir.
  PATH="${DESKTOP_DIR}/scripts/linuxdeploy:${PATH}" \
    NO_STRIP=1 npm run tauri build -- --bundles appimage
  echo -e "\n🎉 DAWG Desktop AppImage created successfully!"
else
  echo -e "\nBundle staging complete (Tauri build skipped)."
fi
