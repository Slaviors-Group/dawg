![DAWG paw icon](desktop/src-tauri/icons/128x128.png)

# DAWG — Digs Any Web-app Glitch

DAWG records a browser reproduction, sanitizes captured data, packages the
session as an OCI Image Layout, and replays its rrweb DOM trace in Chromium. The
repository contains a Go CLI, a Tauri desktop application, and a Chromium
Manifest V3 extension.

**Website and documentation:** [dawg.slaviors.id](https://dawg.slaviors.id/)

> **Current release:** [`0.4.3-omega`](https://github.com/Slaviors-Group/dawg/releases/tag/omega-3)
>
> Application and artifact schema: `0.4.3-omega` · Desktop/npm/Cargo/Tauri: `0.4.3` · Extension: `0.3.5_middlechild` (install `0.3.5`, unchanged)

## 📦 Install the Current Release

DAWG [`0.4.3-omega`](https://github.com/Slaviors-Group/dawg/releases/tag/omega-3) is the current application and artifact-schema release.

- [Download DAWG for Windows (x64 installer)](https://github.com/Slaviors-Group/dawg/releases/download/omega-3/DAWG_0.4.3_x64-setup.exe)
- [Download DAWG for Linux (x86_64 AppImage)](https://github.com/Slaviors-Group/dawg/releases/download/omega-3/DAWG_0.4.3_amd64.AppImage)

These links download the matching asset from the `omega-3` GitHub release directly.
After downloading the Linux AppImage, make it executable with
`chmod +x DAWG_0.4.3_amd64.AppImage` and run it with
`./DAWG_0.4.3_amd64.AppImage`. Visit the
[GitHub releases page](https://github.com/Slaviors-Group/dawg/releases) for
checksums, release notes, and older assets.

## 🔎 Current Capabilities

> **Browser compatibility:** DAWG `0.4.3-omega` records only Chromium-based
> desktop browsers such as Chrome and Edge, version 116 or newer. Firefox and Safari recording are not
> supported in this release. Replay and review use bundled Playwright Chromium.

- **Extension-driven capture:** the desktop app or CLI selects one `http://` or
  `https://` tab. The extension records rrweb events, click/input actions, and
  frontend request/response metadata and a sanitized browser device profile.
- **Sanitization before packaging:** heuristic secret/PII rules and an OPA policy
  process supported JSONL streams before an artifact is published.
- **Diagnostic evidence:** **Safe** capture records bounded console/error, network, and browser device metadata without CDP body retrieval. **Enhanced Diagnostics** is an explicit, consented CDP opt-in that adds CDP console, exception, and network evidence; it can retain eligible text or structured response bodies within capture limits. Retained diagnostic evidence is sanitized before packaging, and CDP degradation or safe fallback is recorded.
- **OCI artifacts:** sessions are stored as digest-addressed OCI Image Layouts
  under `~/.dawg/artifacts/` by default.
- **Portable `.dawg` archives:** validated artifacts can be exported as ZIP-based
  `.dawg` files and imported into another DAWG installation.
- **Welcome and guided tour:** four welcome slides introduce capture, replay,
  artifact sharing, and project support. Skipping or finishing opens a 14-step
  tour across Dashboard, Capture, Artifacts, Replay, and Editor. Stop, close, or
  Escape ends the tour; the Dashboard **Guide** button reopens it. Welcome and
  tour completion are stored in local storage. Opening a `.dawg` file through
  file association bypasses welcome for that launch without marking it complete.
- **Dashboard summary:** the Dashboard highlights one most recently added local
  artifact instance, shows its import/flag/revision state, and summarizes local
  captures, flagged artifacts, diagnostics, and Doctor compatibility. Import a
  `.dawg` archive or open the latest artifact with **Replay** or **Edit**.
- **Artifact catalog:** the **Artifacts** page searches and filters local
  instances, groups review revisions by lineage, and supports inspection,
  replay, editing, `.dawg` import/export, and confirmed local deletion. A local
  instance can be deleted only from DAWG's managed artifact store. Enter in a
  nonempty search inspects the first matching artifact.
- **Chromium review editor:** add point or range flags with a title, optional
  note, category, and severity. Save drafts locally and publish an immutable
  flagged revision when the review is ready; the source artifact and its rrweb
  and diagnostic evidence remain unchanged.
- **Browser replay and review:** Desktop launches interactive Playwright Chromium
  with play/pause, skip, speed, and timeline controls synchronized with the
  Desktop player. Flagged artifacts add clickable timeline markers, previous and
  next flag navigation, selected-flag playback, and **Review flags only** mode.
  Console and network evidence follows literal replay time and remains available
  for full-capture review. Its Replay workspace reviews packaged diagnostic
  evidence and can export sanitized HAR or copy a reviewed cURL request.
  Replay artifact search includes `instanceId`; Enter in a nonempty search
  selects the first match without running it. Choose **Run Replay** explicitly.
  Standard CLI replay writes a final screenshot and exits.
- **Desktop process control:** an active replay can be cancelled. On Windows,
  cancellation and application shutdown terminate the tracked engine process
  tree, including Node.js, Chromium, and mitmdump descendants.
- **Runtime checks:** `dawg doctor` reports the engine, Node.js, mitmdump,
  Chromium, replay script, schema, policy, and extension status.
- **OCI registry transport:** the CLI can push and pull OCI layouts with ORAS.

## ⚠️ Security and Runtime Boundaries

- Diagnostic evidence is a bounded packaged record, not a browser archive. It
  does not restore the original application JavaScript runtime, source maps,
  cookies, Chrome DevTools Network history, or unavailable/blocked/truncated
  body content. DAWG does not reconstruct missing evidence after capture.
- **Safe** does not retrieve response bodies through CDP. **Enhanced** requests
  them only for eligible JSON, GraphQL, form, or text response content; binary,
  oversized, unavailable, and failed retrievals are represented by an evidence
  state instead of a body value.
- rrweb masks input fields in DOM snapshots. Separate action events contain the
  entered value, and request metadata can contain headers and request bodies,
  until the engine sanitizes the capture at stop time. Review an artifact before
  sharing it; sanitization reduces exposure but is not a confidentiality proof.
- Device evidence includes browser/OS hints, viewport, screen, locale, timezone,
  approximate browser-exposed hardware capacity, connection hints, page URL, and
  user agent. Browser APIs do not expose MAC addresses, hostnames, or reliable IP
  addresses; DAWG records those fields as unavailable and does not call an external
  IP-discovery service.
- The extension requests `<all_urls>` access and captures only the tab selected
  for the active token-bearing session.
- Imported `.dawg` archives are extracted into staging and checked for traversal,
  symlinks, excessive size, entry count, and compression ratio before they are
  published.
- Windows replay runs Chromium natively and skips Docker Compose isolation and
  database restoration. Non-Windows environment replay requires rootless Docker.
- `dawg verify --against <value>` records the supplied value in the report; it
  does not check out or launch that branch or commit.
- `dawg init` writes a starter configuration and policy. Current commands use
  their flags and runtime defaults directly rather than loading that config.

## 🗂️ Repository Layout

```text
dawg/
├── version.json          # Application and bundled-runtime version source
├── engine/               # Go CLI, capture, sanitizer, OCI, replay, verify, registry
├── extension/            # Chromium/Chrome Manifest V3 capture extension
├── desktop/              # React/Tauri desktop application and bundle scripts
├── schema/               # Artifact JSON Schema, media types, and OPA policy
├── tools/                # Version synchronization utility
└── test/                 # End-to-end smoke-test assets
```

## 🎥 Capture and Replay

1. Install the desktop application or build the engine and desktop resources.
2. Install the [DAWG Browser Extension from the Chrome Web Store](https://chromewebstore.google.com/detail/peiigoeakholhhbbbbfkeojomekmmokj?utm_source=item-share-cb). Chrome or Chromium 116 or newer is required. Use an unpacked `extension/` directory through `chrome://extensions` only for development.
3. Open the target page, enter its URL in DAWG, optionally enter an artifact
   title, and choose **Safe** (default) or consent to **Enhanced Diagnostics**.
4. Select **Start Capture**. The extension focuses an exact matching tab or opens
   the URL and starts recording the top-level document. If Enhanced CDP setup is
   unavailable, the capture falls back to Safe and records the degradation.
5. Reproduce the issue, then select **Stop Capture**. The engine waits for the
   extension to drain events, sanitizes the session, packages it, and registers
   it in the local catalog.
6. Use the Dashboard to review the latest artifact and catalog summary, or open
   **Artifacts** to search, inspect, replay, edit, export, import, or safely
   delete a local artifact instance. The Artifacts page also accepts drag-and-
   drop `.dawg` imports.
7. Select **Replay** to open Replay with the artifact preselected, then choose
   **Run Replay** when ready. Flagged revisions expose timeline markers,
   previous/next navigation, selected-flag playback, and **Review flags only**
   mode in Chromium. Use **Stop Replay** to terminate it.
8. Select **Edit** to open Chromium Editor. Add point or range flags, save a
   local draft if needed, then save the artifact to publish a new immutable
   flagged revision without modifying its source.

Untitled captures receive a hostname-based title. Artifact directories use a
readable timestamped name such as `20260912-143025-checkout-timeout`; collisions
receive numeric suffixes.

## 🧰 CLI Reference

Run commands from `engine/` during source development or use the installed
`dawg` executable.

```powershell
# Create dawg.config.yaml and a default local policy
dawg init [directory] [--force]

# Start an extension-driven capture
dawg capture --url https://example.test --title "Checkout timeout" --diagnostics-profile safe

# Stop, sanitize, package, and register the active capture
dawg capture stop

# Manage the local artifact catalog
dawg artifacts list
dawg artifacts export <artifact-directory> --output <file.dawg> [--force]
dawg artifacts import <file.dawg>
dawg artifacts delete <artifact-directory>
dawg artifacts review <artifact-directory> --review-file review.json

# Open Chromium Editor and publish review flags
dawg editor <artifact-directory> --interactive

# Inspect, replay, and verify an artifact
dawg inspect <artifact-directory>
dawg diagnostics inspect <artifact-directory>
dawg diagnostics export-har <artifact-directory> --output evidence.har
dawg diagnostics copy-curl <artifact-directory> --request-id <request-id>
dawg diagnostics remove <artifact-directory> --output-dir <new-artifact-directory> --remove-category console --remove-body-ref diagnostics/bodies/response_request-id.json
dawg run <artifact-directory> [--interactive]
dawg verify <artifact-directory> --against local

# Transfer OCI layouts through a registry
dawg push <artifact-directory> --registry <registry-reference>
dawg pull <registry-reference> --output <directory>

# Check runtime resources
dawg doctor [--output json]
```

Capture also accepts optional `--compose-file`, `--db-diff-file`, `--log-file`,
`--policy-file`, and `--session-dir` inputs. `--diagnostics-profile` accepts
`safe` (default) or `enhanced`. `--unsafe-skip-sanitize` is limited to localhost
targets.

## 🧑‍💻 Development

### Requirements

- Go `1.25.1` (from `engine/go.mod`)
- Node.js `^20.19.0` or `>=22.12.0` for the Vite 7 desktop toolchain
- npm
- Rust `1.85+` with the platform dependencies required by Tauri v2
- Chrome or Chromium `116+` for extension capture
- Docker Engine and Docker Compose for environment capture/replay on supported
  non-Windows hosts

The self-contained bundle stages Node.js `22.14.0`, mitmproxy `12.2.3`,
Playwright `1.55.1`, rrweb `2.0.0-alpha.18`, and its Chromium revision. Users of
the packaged application do not install those components separately.

### Engine

```powershell
cd engine
npm ci
npm run check:versions
go test ./...
go build -o dawg.exe ./cmd/dawg
.\dawg.exe doctor
```

### Desktop

```powershell
cd desktop
npm ci
npm run build
npm run tauri dev
```

`npm run dev` starts only the Vite frontend; Tauri IPC, native dialogs, and engine
commands require `npm run tauri dev`.

### Extension

Load `extension/` as an unpacked extension. Run its automated test with:

```powershell
node --test extension/tests/service_worker.test.cjs
```

## 🏗️ Build a Distribution

### Windows NSIS installer

```powershell
.\desktop\build-bundle.ps1
```

Use `-SkipDownload` only when the runtime cache and staged dependencies are
already populated. The installer is written under
`desktop/src-tauri/target/release/bundle/nsis/`.

### Linux AppImage

```bash
chmod +x ./desktop/build-bundle.sh
./desktop/build-bundle.sh
```

The AppImage is written under `desktop/src-tauri/target/release/bundle/appimage/`.
The Tauri configuration also declares a Debian target, but the helper script does
not build it. See
[`desktop/README.md`](desktop/README.md) for native package prerequisites and
staged resource details.

## 🔢 Version Management

[`version.json`](version.json) is the source for application, desktop, extension,
Node.js, and mitmproxy versions.

```json
{
  "appVersion": "0.4.3-omega",
  "desktopVersion": "0.4.3",
  "extensionVersion": "0.3.5_middlechild",
  "runtime": {
    "mitmproxy": "12.2.3",
    "node": "22.14.0"
  }
}
```

After changing it, run either package script:

```powershell
npm run sync:versions
npm run check:versions
```

The synchronizer uses numeric `0.4.3` for Desktop/npm/Cargo/Tauri and labeled
`0.4.3-omega` for the application/engine and artifact schema. New artifacts use
`v0.4.3-omega.json`, retaining the `0.4.2-omega` schema structure. Historical
supported schemas remain readable; no artifact migration is required.
Chrome continues to receive numeric `version: "0.3.5"` plus display label
`version_name: "0.3.5_middlechild"`; Omega does not change the extension.
Dependency versions remain managed by package
manifests and lockfiles.

## ✅ Validation

The repository CI checks Go formatting, vetting, tests, and builds; Biome and the
frontend build; desktop bundle staging; and Cargo compilation. Before sharing
source changes within the project, run the relevant local checks:

```powershell
cd engine
npm run check:versions
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/dawg

cd ..\extension
node --test tests/service_worker.test.cjs

cd ..\desktop
npm ci
npx biome check .
npm run build
cargo test --manifest-path src-tauri/Cargo.toml
```

Bundle creation and a manual capture/import/export/replay smoke test are required
to validate staged runtime assets and native process handling.

## 🤝 Contributions

DAWG's code is public, but external contributions are not currently accepted.
Please do not submit pull requests. See [CONTRIBUTING.md](CONTRIBUTING.md) for
feedback channels and the current contribution policy.

## 📄 License

GNU GENERAL PUBLIC LICENSE Version 3, 29 June 2007. See [`LICENSE`](LICENSE).
