#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DESKTOP_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUNDLE_DIR="${DESKTOP_DIR}/src-tauri/target/release/bundle/appimage"
GSTREAMER_HOOK_SOURCE="${SCRIPT_DIR}/appimage/linuxdeploy-plugin-gstreamer.sh"
APPIMAGE_PLUGIN="${TAURI_LINUXDEPLOY_PLUGIN_APPIMAGE:-${HOME}/.cache/tauri/linuxdeploy-plugin-appimage.AppImage}"

mapfile -t appdirs < <(find "${BUNDLE_DIR}" -mindepth 1 -maxdepth 1 -type d -name '*.AppDir' -print)
mapfile -t appimages < <(find "${BUNDLE_DIR}" -maxdepth 1 -type f -name 'DAWG_*_amd64.AppImage' -print)

if [ "${#appdirs[@]}" -ne 1 ] || [ "${#appimages[@]}" -ne 1 ]; then
  echo "Expected exactly one AppDir and one DAWG amd64 AppImage." >&2
  echo "Found ${#appdirs[@]} AppDirs and ${#appimages[@]} AppImages." >&2
  exit 1
fi

appdir="${appdirs[0]}"
appimage="${appimages[0]}"
libdir="${appdir}/usr/lib"
gstreamer_hook="${appdir}/apprun-hooks/linuxdeploy-plugin-gstreamer.sh"

if [ ! -d "${libdir}" ] || [ ! -d "$(dirname "${gstreamer_hook}")" ]; then
  echo "Generated AppDir is missing its runtime library layout: ${appdir}" >&2
  exit 1
fi
if [ ! -x "${APPIMAGE_PLUGIN}" ]; then
  echo "Tauri's AppImage output plugin is unavailable: ${APPIMAGE_PLUGIN}" >&2
  exit 1
fi

echo "Removing host-coupled Wayland libraries from the AppDir..."
find "${libdir}" -maxdepth 1 \( -type f -o -type l \) \
  \( -name 'libwayland-client.so*' \
    -o -name 'libwayland-cursor.so*' \
    -o -name 'libwayland-egl.so*' \
    -o -name 'libwayland-server.so*' \) \
  -print -delete

install -m 0755 "${GSTREAMER_HOOK_SOURCE}" "${gstreamer_hook}"

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
  echo "Refusing to package host-coupled graphics libraries:" >&2
  printf '  %s\n' "${forbidden_graphics_libraries[@]}" >&2
  exit 1
fi

temporary_output="${appimage}.repacked"
runtime_file="$(mktemp "${BUNDLE_DIR}/.appimage-runtime.XXXXXX")"
trap 'rm -f "${temporary_output}" "${runtime_file}"' EXIT
rm -f "${temporary_output}"

# appimagetool normally downloads this stub from GitHub. Reuse the runtime from
# Tauri's original image so release builds stay deterministic and work on
# builders without outbound internet access.
appimage_offset="$("${appimage}" --appimage-offset)"
if [[ ! "${appimage_offset}" =~ ^[0-9]+$ ]] || [ "${appimage_offset}" -le 0 ]; then
  echo "Could not determine the embedded AppImage runtime size." >&2
  exit 1
fi
head -c "${appimage_offset}" "${appimage}" > "${runtime_file}"

echo "Repacking sanitized AppDir..."
APPIMAGE_EXTRACT_AND_RUN=1 \
  LDAI_OUTPUT="${temporary_output}" \
  LDAI_RUNTIME_FILE="${runtime_file}" \
  "${APPIMAGE_PLUGIN}" --appdir="${appdir}"

if [ ! -s "${temporary_output}" ]; then
  echo "AppImage repack did not create ${temporary_output}." >&2
  exit 1
fi

chmod +x "${temporary_output}"
mv -f "${temporary_output}" "${appimage}"
rm -f "${runtime_file}"
trap - EXIT

echo "Sanitized and repacked ${appimage}."
