# DAWG — Digs Any Web-app Glitch

DAWG captures a browser reproduction, sanitizes sensitive data, packages the
result as a portable OCI-layout artifact, and replays the recorded DOM session
in an isolated Chromium instance. It provides both a Tauri desktop application
and a Go CLI.

> **Current release label:** `0.2.3-naughty`
>
> Generated application/schema version: `0.2.3-naughty`

---

## What DAWG Does

- **Extension-first capture** — the Manifest V3 DAWG Browser Extension records
  rrweb DOM events, browser actions, and request metadata from one
  desktop-selected tab. DAWG does not launch Playwright during capture.
- **Privacy-aware artifacts** — the sanitizer removes secrets and PII before an
  artifact is packaged. It preserves rrweb structural data, including document
  type and SVG geometry fields, so sanitized replays remain renderable.
- **Portable replay** — replay launches DAWG-bundled Chromium through
  Playwright, reconstructs the rrweb session, and writes a final screenshot.
- **Safe process lifecycle** — a replay can be stopped from the desktop UI. On
  Windows, stopping a replay or closing DAWG terminates the engine and its
  Node.js, Chromium, and mitmproxy child-process tree.
- **Runtime diagnostics** — `dawg doctor` validates the engine, Node.js,
  mitmdump, Chromium, replay script, schema, policy, and extension resources.
- **OCI-layout packaging** — captured sessions become digest-addressed artifacts
  that can be inspected, replayed, or verified later.

---

## Repository Layout

```text
dawg/
├── version.json          # Release/runtime version source of truth
├── engine/               # Go CLI, capture, sanitizer, packager, replay, verify
├── extension/            # Installable Chromium/Chrome Manifest V3 extension
├── desktop/              # Tauri v2 desktop shell and bundle assembly scripts
├── schema/               # Artifact manifest JSON Schema and OPA policy
├── tools/                # Version synchronization utility
├── test/                 # End-to-end smoke-test assets
└── bin/                  # Local built engine and development resources
```

---

## Capture → Replay Workflow

1. Install/reload the DAWG browser extension from `extension/`.
2. Start the desktop app and enter an `http://` or `https://` target URL.
3. Click **Start Capture**. DAWG focuses an existing matching browser tab or
   opens a new one, then the extension begins recording that tab.
4. Reproduce the issue in the browser.
5. Click **Stop Capture**. DAWG waits for the extension to flush, sanitizes the
   collected data, and packages an artifact under `~/.dawg/artifacts/`.
6. Open **Replay**, select the artifact, and click **Run Replay**.
7. Use **Stop Replay** to interrupt an in-flight replay. A successful replay
   closes Chromium automatically after it writes its screenshot.

The Replay execution log reports rrweb event counts, captured viewport metadata,
rendered text length, and in-page errors. A successful replay normally has one
Meta event, one FullSnapshot, one replay iframe, and a positive visible-text
length.

---

## Version Management

[`version.json`](version.json) is DAWG's source of truth for application,
desktop, extension, bundled Node.js, and bundled mitmproxy versions.

```json
{
  "appVersion": "0.2.3-naughty",
  "desktopVersion": "0.2.3",
  "extensionVersion": "0.2.3_naughty"
}
```

After editing it, run either command from `engine/` or `desktop/`:

```powershell
npm run sync:versions
npm run check:versions  # reports drift only; does not write files
```

Convenient release labels are accepted and normalized for the tools that require
strict SemVer:

| `version.json` value | Generated package/schema value |
| --- | --- |
| `0.2.3` | `0.2.3` |
| `0.2.3-naughty` | `0.2.3-naughty` |
| `0.2.3_naughty` for the extension | Chrome `version: "0.2.3"` plus `version_name: "0.2.3_naughty"` |

The synchronizer updates package metadata, Cargo/Tauri metadata, engine/schema
references and schema filename, the extension manifest, bundle defaults, and
package-lock root metadata. Third-party dependencies such as Playwright and
rrweb are intentionally **not** changed by this command; upgrade them through a
normal dependency update and compatibility-test workflow.

---

## Desktop App

The desktop distribution is a monolithic bundle. It stages the engine,
mitmdump, portable Node.js, Playwright dependencies, Chromium, schema/policy,
and the installable browser extension.

### Build a Windows installer

From the repository root:

```powershell
.\desktop\build-bundle.ps1
```

Use `-SkipDownload` only when the required cached/staged runtime assets already
exist:

```powershell
.\desktop\build-bundle.ps1 -SkipDownload
```

The script runs `dawg doctor` before packaging and creates the NSIS installer
under `desktop/src-tauri/target/release/bundle/nsis/`.

### Build Linux packages

```bash
chmod +x ./desktop/build-bundle.sh
./desktop/build-bundle.sh
```

The Linux bundle output is created below
`desktop/src-tauri/target/release/bundle/`.

For desktop-specific development, packaging, IPC, and runtime details, see
[`desktop/README.md`](desktop/README.md).

---

## Engine CLI

### Development prerequisites

- Go 1.22+
- Node.js 20+ for replay development
- mitmproxy / `mitmdump` when running unbundled capture/replay dependencies
- Chromium installed by Playwright, or a staged desktop bundle

### Diagnostics

```powershell
dawg doctor
```

Example of a healthy bundled installation:

```text
DAWG Engine Doctor Diagnostics (v0.2.3-naughty)
Resource Root: ...\resources (Bundled: true)

✅ dawg-engine
✅ mitmdump
✅ node
✅ playwright:chromium
✅ script:replay-browser.cjs
✅ browser-extension
✅ policy:default.rego
✅ schema:manifest

All required runtime components are healthy and ready!
```

### Commands

```powershell
# Initialize a DAWG config in a target project
dawg init

# Ask the browser extension to capture a target tab
dawg capture --url https://my-app.local

# Stop, sanitize, and package the active capture
dawg capture stop

# Inspect without replaying
dawg inspect .dawg/artifacts/<session-id>

# Replay an artifact
dawg run .dawg/artifacts/<session-id>

# Verify an artifact against a target
dawg verify .dawg/artifacts/<session-id> --against local
```

---

## Development

### Engine

```powershell
cd engine
go test ./...
go build -o dawg.exe ./cmd/dawg
```

### Desktop

```powershell
cd desktop
npm install
npm run tauri dev
```

### Extension

Load [`extension/`](extension/) as an unpacked extension from
`chrome://extensions` with **Developer mode** enabled. Reload the extension
there after changing extension source or installing a newly built desktop
bundle.

---

## License

Apache-2.0. See [LICENSE](LICENSE).
