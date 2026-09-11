# DAWG Browser Extension

The DAWG Browser Extension is the **capture-side** component of DAWG. It is a
Manifest V3 extension that records one desktop-selected browser tab and streams
capture data to the local DAWG engine.

> Current configured label: `0.2_naughty`
>
> Chrome install version: `0.2.0` (`version_name` retains the release label)

Playwright and DAWG-bundled Chromium are used only for **artifact replay**.
Capture does not launch, control, or connect to a Playwright browser.

---

## Install for Development

1. Open `chrome://extensions` in Chrome, Chromium, or a compatible browser.
2. Enable **Developer mode**.
3. Choose **Load unpacked**.
4. Select this `extension/` directory.
5. After changing extension source or installing a newly built DAWG desktop
   bundle, click **Reload** for the DAWG extension before capturing again.

The desktop bundle also stages a copy at
`desktop/src-tauri/resources/extension/`; use the source directory during
extension development and the staged copy when distributing the desktop app.

---

## Capture Protocol

The engine starts a local extension server at:

```text
ws://127.0.0.1:8082/ws
```

The extension connects to that local server and performs a session handshake:

1. The desktop/CLI starts `dawg capture --url <target>`.
2. The engine waits for the extension and sends `DAWG_COMMAND_START` with a
   target URL and session token.
3. The extension focuses a matching tab or opens the target URL.
4. It injects the rrweb recorder into the **top-level document only** and
   acknowledges capture after the recorder is running.
5. It streams `DAWG_RRWEB_EVENT`, `DAWG_ACTION_EVENT`, and
   `DAWG_HTTP_EVENT` envelopes to the engine.
6. On stop, the extension stops recording, drains pending deliveries, sends a
   session-stop message, and waits for the engine acknowledgement before the
   artifact is sanitized and packaged.

The extension includes an HTTP delivery fallback when its WebSocket is not
available. Capture data is accepted only for the active session token and the
selected tab.

---

## What Is Recorded

- **rrweb DOM trace** — a full initial snapshot plus subsequent mutations and
  browser interaction state needed for deterministic visual replay.
- **User actions** — click and input metadata for the selected tab.
- **Frontend request metadata** — request/response records associated with the
  active capture tab.

Input values are masked by the recorder. The engine applies its additional
sanitizer and policy gate before packaging any artifact. Structural rrweb values
such as the DOCTYPE and SVG geometry are preserved so sanitization cannot break
replay rendering.

---

## Operational Notes

- The extension records only the tab selected by DAWG; it is not a general
  browser-history recorder.
- If the extension was installed before a desktop update, reload it from
  `chrome://extensions` to use the updated manifest and scripts.
- If capture waits for the extension, verify that it is enabled, the target tab
  is an `http://` or `https://` page, and no firewall/security tool blocks the
  local loopback connection to port `8082`.
- Closing the desktop application terminates DAWG-owned background work. If a
  capture remains active, start DAWG again and verify the extension state before
  beginning the next session.

---

## Versioning

Do not edit `manifest.json` version fields independently. Update
[`../version.json`](../version.json) and run:

```powershell
cd ..\engine
npm run sync:versions
```

For an extension label such as `0.2_naughty`, the synchronizer generates the
Chrome-required numeric `version` and stores the label in `version_name`.
