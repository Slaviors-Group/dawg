# 🧩 DAWG Browser Extension

The DAWG Browser Extension is the capture component used by the desktop
application and CLI. It records one engine-selected tab and delivers events to
the local capture daemon.

> Manifest: `3` · Minimum Chrome/Chromium: `116`
>
> Install version: `0.3.1` · Display version: `0.3.1_middlechild`

Playwright and bundled Chromium are replay dependencies. Capture runs in the
user's installed Chrome or Chromium browser.

## 📥 Install

For normal use, install the [DAWG Browser Extension from the Chrome Web Store](https://chromewebstore.google.com/detail/peiigoeakholhhbbbbfkeojomekmmokj?utm_source=item-share-cb).

### Install for Development

Use an unpacked extension only for source development:

1. Open `chrome://extensions`.
2. Enable **Developer mode**.
3. Select **Load unpacked**.
4. Select this `extension/` directory.
5. Reload the extension after source changes or a desktop application update.

The bundle scripts stage a distribution copy at
`desktop/src-tauri/resources/extension/`. Use this source directory while
working on the extension.

## 🔐 Permissions

The manifest declares:

- `activeTab`, `tabs`, and `scripting` to find, focus/open, and initialize the
  selected target tab;
- `webRequest` to collect Safe-capture request and response metadata;
- `debugger` for consented Enhanced Diagnostics CDP collection;
- `storage` to retain capture state across Manifest V3 worker restarts;
- `<all_urls>` host access for content-script and request observation.

The recorder is initialized only for the tab selected by the active DAWG
session. Content-script injection defaults to the top-level document because the
manifest does not enable `all_frames`.

## 🔌 Capture Protocol

The engine listens on loopback:

```text
WebSocket: ws://127.0.0.1:8082/ws
HTTP fallback: http://127.0.0.1:8082/api/v1/stream/event
```

A normal session proceeds as follows:

1. `dawg capture --url <target>` starts the local daemon.
2. The extension connects and sends `DAWG_EXTENSION_READY`.
3. The engine sends `DAWG_COMMAND_START` with the target URL and a random session
   token.
4. The extension focuses an exact normalized URL match or opens a new tab,
   starts the recorder, and sends `DAWG_SESSION_START` with the active Safe or
   Enhanced Diagnostics profile. If CDP attachment for Enhanced fails, it records
   a degradation and continues in Safe mode.
5. It delivers `DAWG_RRWEB_EVENT`, `DAWG_ACTION_EVENT`, and `DAWG_HTTP_EVENT`
   envelopes. A `DAWG_KEEPALIVE` is sent every 20 seconds while connected.
6. On stop, the extension stops rrweb, waits for pending deliveries, sends
   `DAWG_SESSION_STOP`, and waits up to 3.5 seconds for
   `DAWG_SESSION_STOP_ACK`.
7. The engine drains the stream, sanitizes the capture, packages the OCI
   artifact, and updates the local catalog.

When WebSocket delivery is unavailable, envelopes are POSTed to the HTTP
fallback with a two-second request timeout. The engine accepts data only for the
active session token; WebSocket delivery is also bound to the selected extension
connection.

## 🎥 Recorded Data

### rrweb stream

- Initial DOM snapshot and incremental rrweb events
- Input fields masked through `maskAllInputs: true`

### Action stream

- Click timestamp and a basic element selector
- Input timestamp, selector, field name, input type, and entered value

### Diagnostic evidence profiles

**Safe** is the default. It records bounded console/error and network metadata
from the selected tab. Network body fields are marked `not-requested` or
`unavailable`; Safe does not use CDP to retrieve response bodies.

**Enhanced Diagnostics** requires explicit desktop consent and attaches Chrome
DevTools Protocol (CDP) to the selected tab. It adds CDP console APIs,
exceptions, and network records. For completed network responses, it requests a
body only when the response is JSON, GraphQL, form-encoded, or text, and only
within the per-body and concurrent-fetch limits. Binary, oversized, unavailable,
and failed body retrievals are recorded as states rather than retained values.

The extension detaches CDP on stop. If CDP cannot attach or is detached during
capture, it emits a capture-degradation record; an initial attach failure falls
back to Safe.

### Evidence states

Diagnostic values state whether they are `captured`, `redacted`, `preview-only`,
`truncated`, `blocked`, `unavailable`, `not-requested`, or `capture-failed`.
A state describes retained fidelity; it is not a promise that the original value
can be recovered.

## ⚠️ Data Handling

rrweb input masking does not apply to the separate action stream. Input action
values, available headers, and request bodies reach the local engine before its
sanitizer processes the session. The engine applies secret/PII classification
and the configured OPA policy before packaging, but captured artifacts should
still be inspected before distribution.

The sanitizer also processes diagnostic console, network, error, and retained
body records before packaging. It preserves rrweb document-type names and SVG
`viewBox`/`points` geometry because those fields are required to reconstruct
valid DOM/SVG nodes.

## 🩺 Troubleshooting

- If capture remains waiting, confirm the extension is enabled and reloaded, the
  target uses `http://` or `https://`, and loopback port `8082` is not blocked.
- The popup reports current state and can stop an active recording. Capture must
  be started by the desktop app or CLI so the extension receives a daemon-issued
  session token.
- Closing the selected tab sends a session error and stops extension recording.
- A service-worker restart restores session-scoped capture state; the content
  script wakes the worker while a normal page remains open.

## 🔢 Versioning

Do not edit `manifest.json` version fields separately. Update
[`../version.json`](../version.json), then run:

```powershell
cd ..\engine
npm run sync:versions
npm run check:versions
```

Chrome requires a numeric install version. The synchronizer converts
`0.3.1_middlechild` to `version: "0.3.1"` and
`version_name: "0.3.1_middlechild"`.

## ✅ Validation

From the repository root:

```powershell
node --check extension/background/service_worker.js
node --check extension/content/recorder.js
node --check extension/popup/popup.js
node --test extension/tests/service_worker.test.cjs
```

The automated test uses mocked Chrome APIs. Complete validation still requires a
manual Chrome/Chromium capture, stop, package, and replay cycle.
