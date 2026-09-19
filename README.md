![DAWG paw icon](desktop/src-tauri/icons/128x128.png)

# DAWG — Digs Any Web-app Glitch

DAWG records a browser reproduction, sanitizes captured data, packages the
session as an OCI Image Layout, and replays its rrweb DOM trace in Chromium. The
repository contains a Go CLI, a Tauri desktop application, and a Chromium
Manifest V3 extension.

> **Current prerelease:** [`0.2.5-naughty`](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-3)
>
> Application and schema: `0.2.5-naughty` · Desktop: `0.2.5` · Extension: `0.2.5_naughty`

## 📦 Install the Current Release

Download the available desktop installer assets from the
[`0.2.5-naughty` (`naughty-3`) GitHub release](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-3).

Linux targets are configured in Tauri; build the AppImage from source with
`desktop/build-bundle.sh` when an AppImage is not attached to the release.

See [all releases](https://github.com/Slaviors-Group/dawg/releases) for older
prereleases and their assets.

## 🔎 Current Capabilities

- **Extension-driven capture:** the desktop app or CLI selects one `http://` or
  `https://` tab. The extension records rrweb events, click/input actions, and
  frontend request/response metadata.
- **Sanitization before packaging:** heuristic secret/PII rules and an OPA policy
  process supported JSONL streams before an artifact is published.
- **OCI artifacts:** sessions are stored as digest-addressed OCI Image Layouts
  under `~/.dawg/artifacts/` by default.
- **Portable `.dawg` archives:** validated artifacts can be exported as ZIP-based
  `.dawg` files and imported into another DAWG installation.
- **Persistent catalog:** `~/.dawg/artifact-catalog.json` tracks captured,
  imported, and previously uncataloged local artifacts.
- **Browser replay:** Desktop launches interactive Playwright Chromium with
  play/pause, skip, speed, and timeline controls. Standard CLI replay records
  diagnostics, writes a final screenshot, and exits.
- **Desktop process control:** an active replay can be cancelled. On Windows,
  cancellation and application shutdown terminate the tracked engine process
  tree, including Node.js, Chromium, and mitmdump descendants.
- **Runtime checks:** `dawg doctor` reports the engine, Node.js, mitmdump,
  Chromium, replay script, schema, policy, and extension status.
- **OCI registry transport:** the CLI can push and pull OCI layouts with ORAS.

## ⚠️ Security and Runtime Boundaries

- rrweb masks input fields in DOM snapshots. Separate action events contain the
  entered value, and request metadata can contain headers and request bodies,
  until the engine sanitizes the capture at stop time. Review an artifact before
  sharing it; sanitization reduces exposure but is not a confidentiality proof.
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
3. Open the target page, enter its URL in DAWG, and optionally enter an artifact
   title.
4. Select **Start Capture**. The extension focuses an exact matching tab or opens
   the URL and starts recording the top-level document.
5. Reproduce the issue, then select **Stop Capture**. The engine waits for the
   extension to drain events, sanitizes the session, packages it, and registers
   it in the local catalog.
6. Select **Replay** on a Dashboard artifact to open Replay with that artifact
   preselected, then choose **Run Replay** when ready. Use the interactive
   Chromium controls to inspect the recording, or **Stop Replay** to terminate
   it.
7. Export or import an existing `.dawg` archive from the Dashboard or Replay
   page; the Dashboard also accepts drag-and-drop imports.

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
dawg capture --url https://example.test --title "Checkout timeout"

# Stop, sanitize, package, and register the active capture
dawg capture stop

# Manage the local artifact catalog
dawg artifacts list
dawg artifacts export <artifact-directory> --output <file.dawg> [--force]
dawg artifacts import <file.dawg>

# Inspect, replay, and verify an artifact
dawg inspect <artifact-directory>
dawg run <artifact-directory> [--interactive]
dawg verify <artifact-directory> --against local

# Transfer OCI layouts through a registry
dawg push <artifact-directory> --registry <registry-reference>
dawg pull <registry-reference> --output <directory>

# Check runtime resources
dawg doctor [--output json]
```

Capture also accepts optional `--compose-file`, `--db-diff-file`, `--log-file`,
`--policy-file`, and `--session-dir` inputs. `--unsafe-skip-sanitize` is limited
to localhost targets.

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
  "appVersion": "0.2.5-naughty",
  "desktopVersion": "0.2.5",
  "extensionVersion": "0.2.5_naughty",
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

The synchronizer normalizes labels for npm, Cargo, Tauri, the schema, and Chrome.
Chrome receives numeric `version: "0.2.5"` plus display label
`version_name: "0.2.5_naughty"`. Dependency versions remain managed by package
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
