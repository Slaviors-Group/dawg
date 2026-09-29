#!/usr/bin/env bash

# Keep WebKit on the media framework shipped with the AppImage when that
# framework is complete. If packaging ever omits it, fall back to the host
# defaults instead of pointing GStreamer at a missing directory and poisoning
# its plugin registry with an empty scan.
if [ -z "${APPDIR:-}" ]; then
  APPDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  export APPDIR
fi

plugins_dir="${APPDIR}/usr/lib/gstreamer-1.0"
scanner="${APPDIR}/usr/lib/gstreamer1.0/gstreamer-1.0/gst-plugin-scanner"
ptp_helper="${APPDIR}/usr/lib/gstreamer1.0/gstreamer-1.0/gst-ptp-helper"

if [ -d "${plugins_dir}" ] && [ -x "${scanner}" ]; then
  export GST_REGISTRY_REUSE_PLUGIN_SCANNER=no
  export GST_PLUGIN_SYSTEM_PATH_1_0="${plugins_dir}"
  export GST_PLUGIN_PATH_1_0="${plugins_dir}"
  export GST_PLUGIN_SCANNER_1_0="${scanner}"

  if [ -x "${ptp_helper}" ]; then
    export GST_PTP_HELPER_1_0="${ptp_helper}"
  else
    unset GST_PTP_HELPER_1_0
  fi
else
  unset GST_REGISTRY_REUSE_PLUGIN_SCANNER
  unset GST_PLUGIN_SYSTEM_PATH_1_0
  unset GST_PLUGIN_PATH_1_0
  unset GST_PLUGIN_SCANNER_1_0
  unset GST_PTP_HELPER_1_0
fi
