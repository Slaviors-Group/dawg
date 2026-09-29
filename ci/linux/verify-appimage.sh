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
  # Jenkins reaches this host through a userspace Tailscale SOCKS proxy. Split
  # the image so interrupted transfers can resume at a verified chunk boundary.
  split --bytes=16M --numeric-suffixes=0 --suffix-length=4 \
    "${artifact_name}" "${artifact_name}.part."
  sha256sum "${artifact_name}.part."* > PARTS.sha256
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
libdir="${appdir}/usr/lib"
gstreamer_plugins="${libdir}/gstreamer-1.0"
gstreamer_scanner="${libdir}/gstreamer1.0/gstreamer-1.0/gst-plugin-scanner"
gstreamer_hook="${appdir}/apprun-hooks/linuxdeploy-plugin-gstreamer.sh"

test -x "${engine}"
test -f "${resources}/extension/manifest.json"
test -f "${resources}/extension/content/popup-panel.js"
test -f "${resources}/scripts/replay-browser.cjs"

mapfile -t chromium_executables < <(
  find "${resources}/browsers" -type f -path '*/chrome-linux/chrome' -print
)
if [ "${#chromium_executables[@]}" -ne 1 ]; then
  echo "Expected exactly one bundled Chromium executable, found ${#chromium_executables[@]}." >&2
  exit 1
fi
bundled_node="${resources}/binaries/node/node"
test -x "${bundled_node}"
test -x "${chromium_executables[0]}"

mapfile -t forbidden_graphics_libraries < <(
  find "${libdir}" -maxdepth 1 \( -type f -o -type l \) \
    \( -name 'libwayland*.so*' \
      -o -name 'libEGL.so*' \
      -o -name 'libGL.so*' \
      -o -name 'libGLX.so*' \
      -o -name 'libGLdispatch.so*' \
      -o -name 'libOpenGL.so*' \
      -o -name 'libgbm.so*' \
      -o -name 'libdrm.so*' \) \
    -print
)
if [ "${#forbidden_graphics_libraries[@]}" -ne 0 ]; then
  echo "AppImage contains host-coupled graphics libraries:" >&2
  printf '  %s\n' "${forbidden_graphics_libraries[@]}" >&2
  exit 1
fi

test -x "${gstreamer_scanner}"
test -d "${gstreamer_plugins}"
test -x "${gstreamer_hook}"
bash -n "${gstreamer_hook}"
grep -Fq 'unset GST_PLUGIN_SYSTEM_PATH_1_0' "${gstreamer_hook}"

plugin_count=0
while IFS= read -r -d '' plugin; do
  plugin_count=$((plugin_count + 1))
  if ! file -Lb "${plugin}" | grep -Eq '^ELF 64-bit .* x86-64'; then
    echo "GStreamer plugin is not an x86-64 ELF object: ${plugin}" >&2
    file -L "${plugin}" >&2
    exit 1
  fi
done < <(find "${gstreamer_plugins}" -maxdepth 1 -type f -name '*.so' -print0)
if [ "${plugin_count}" -eq 0 ]; then
  echo "AppImage does not contain any GStreamer plugins." >&2
  exit 1
fi
if ! file -Lb "${gstreamer_scanner}" | grep -Eq '^ELF 64-bit .* x86-64'; then
  echo "GStreamer plugin scanner is not an x86-64 ELF executable." >&2
  file -L "${gstreamer_scanner}" >&2
  exit 1
fi

DAWG_RESOURCES_DIR="${resources}" \
PLAYWRIGHT_BROWSERS_PATH="${resources}/browsers" \
  "${engine}" doctor | tee "${ARTIFACT_DIR}/doctor.txt"

# AppRun puts AppDir libraries ahead of the host for WebKitGTK. Replay and
# Editor launch a separate Chromium process, which must strip those paths or a
# bundle built on Ubuntu can combine Ubuntu NSS/GLib with the user's Mesa/NSS
# stack and abort. Exercise the packaged browser and enforce that isolation.
timeout 30s env \
  APPDIR="${appdir}" \
  LD_LIBRARY_PATH="${appdir}/usr/lib:${appdir}/usr/lib/x86_64-linux-gnu:${appdir}/usr/lib64" \
  DAWG_CHROMIUM_EXECUTABLE_PATH="${chromium_executables[0]}" \
  "${bundled_node}" "${SCRIPT_DIR}/smoke-bundled-browser.cjs" "${resources}"

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
