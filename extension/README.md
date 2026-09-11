# DAWG Browser Extension

This directory contains the DAWG Browser Extension (Manifest V3). It captures
rrweb DOM events, user actions, and request/response metadata from exactly one
desktop-selected tab. Capture no longer launches or connects to a Playwright
browser; Playwright/Chromium remain part of DAWG only for artifact replay.

Load this directory as an unpacked extension from `chrome://extensions` (enable
Developer mode first). After updating the source or installing a new desktop
bundle, click **Reload** on the extension before starting a capture.

The desktop engine listens on `ws://127.0.0.1:8082/ws`. The extension announces
itself, waits for the desktop's `DAWG_COMMAND_START`, focuses or opens the target
URL, and acknowledges only after the content recorder is running. Either the
desktop or the popup may stop capture; in both cases the extension drains events
and completes the stop/ack handshake before artifact packaging begins.
