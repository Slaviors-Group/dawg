# Installation

DAWG is available as a self-contained desktop application or as source for engine, desktop, and extension development. The desktop application is the recommended way to use the complete workflow.

## Current Release

The current prerelease is [`0.3.5-middlechild`](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-5). It includes desktop version
`0.3.5` and extension display version `0.3.5_middlechild`.

Download the installer for your platform directly from the `middlechild-5` GitHub release:

- [Download DAWG for Windows (x64 installer)](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_x64-setup.exe)
- [Download DAWG for Linux (x86_64 AppImage)](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_amd64.AppImage)

The Windows link downloads the NSIS installer. The Linux link downloads the
AppImage. After downloading the AppImage, make it executable and start it:

```bash
chmod +x DAWG_0.3.5_amd64.AppImage
./DAWG_0.3.5_amd64.AppImage
```

See [all releases](https://github.com/Slaviors-Group/dawg/releases) for
checksums, release notes, older prereleases, and other assets.

## Compatibility and Requirements

| Platform or feature | Supported install | Minimum requirement |
|---|---|---|
| Windows desktop app | x64 NSIS installer | Windows 10 or Windows 11, 64-bit |
| Linux desktop app | x86_64 AppImage | A 64-bit Linux distribution capable of running AppImage bundles |
| Browser capture | DAWG Browser Extension | Chrome or Chromium 116 or newer |

The current x86_64 AppImage has been tested on Fedora 44 and Arch Linux/CachyOS.
Other 64-bit Linux distributions may work but are not yet verified. If the
AppImage does not launch, build DAWG from source with
`desktop/build-bundle.sh`.

## Desktop Application

### Windows

1. Select **Download DAWG for Windows** above and run the downloaded x64 NSIS installer.
2. Launch DAWG.
3. Open the runtime diagnostics and confirm the engine reports **Ready**.
4. Install the browser extension before starting a capture.

### Linux

1. Select **Download DAWG for Linux** above.
2. Make the downloaded AppImage executable: `chmod +x DAWG_0.3.5_amd64.AppImage`.
3. Launch it: `./DAWG_0.3.5_amd64.AppImage`.
4. Open the runtime diagnostics and confirm the engine reports **Ready**.
5. Install the browser extension before starting a capture.

The package stages the Go engine, Node.js, mitmdump, Playwright and Chromium for replay, schemas, policies, and the browser extension. These components do not require separate system installation for the packaged application.

### Browser Extension

DAWG capture requires Chrome or Chromium **116 or newer**.

Install the DAWG Browser Extension from the [Chrome Web Store](https://chromewebstore.google.com/detail/peiigoeakholhhbbbbfkeojomekmmokj?utm_source=item-share-cb). This is the recommended installation method.

#### Unpacked installation (development or manual fallback)

Use an unpacked extension only for source development or when Chrome Web Store installation is unavailable:

1. Open `chrome://extensions`.
2. Enable **Developer mode**.
3. Select **Load unpacked**.
4. In DAWG, open **Engine Doctor**, find the `browser-extension` component, and select the parent directory of its displayed `manifest.json` path. For source development, select the repository's `extension/` directory.
5. Reload the extension after updating the desktop app or extension source.

The extension requests `<all_urls>` host access so the engine can select an
`http://` or `https://` target tab and observe that tab's browser events and
request metadata. It also declares Chrome's `debugger` permission for consented
Enhanced Diagnostics CDP collection. Recording starts only after the desktop
app or CLI creates a token-bearing capture session.

> `dawg doctor` checks that an installable extension manifest is present in the runtime resources. Browser installation and enablement must still be checked in `chrome://extensions`.

---

## Build and Run from Source

### Toolchain Requirements

| Tool | Required version or scope | Used for |
|---|---|---|
| **Go** | `1.25.1` | Engine build and tests (`engine/go.mod`) |
| **Node.js** | `^20.19.0` or `>=22.12.0` | Vite 7 desktop toolchain |
| **npm** | Compatible with the selected Node.js release | Locked JavaScript dependencies |
| **Rust** | `1.85+`, stable Tauri v2-compatible toolchain | Desktop native shell |
| **Chrome/Chromium** | `116+` | Extension-driven capture |
| **mitmdump** | Available on `PATH` for an unstaged CLI runtime | HTTP cassette replay and diagnostics |
| **Docker Engine + Compose** | Rootless, on supported non-Windows hosts | Captured environment replay and database restore |

Direct engine development also needs the Playwright and rrweb versions locked by `engine/package-lock.json` and a Playwright Chromium installation. The current self-contained bundle pins:

- Node.js `22.14.0`
- mitmproxy `12.2.3`
- Playwright `1.55.1`
- rrweb `2.0.0-alpha.18`

Python is not required to build the bundled desktop package because the bundle scripts download standalone mitmdump archives. If you run the CLI against a system mitmproxy installation, follow mitmproxy's platform installation requirements.

### Platform Build Dependencies

**Windows** development requires the MSVC Rust target, Visual Studio C++ Build Tools, WebView2, and PowerShell in addition to the toolchain above.

**Linux** development requires a C/C++ toolchain and the WebKitGTK, GTK, AppIndicator, librsvg, and GStreamer development/runtime packages used by Tauri and the configured AppImage media bundle. The bundle helper also uses Bash, `curl`, `tar`, and xz.

### Clone the Repository

```bash
git clone https://github.com/Slaviors-Group/dawg.git
cd dawg
```

### Engine Development Setup

Install the locked JavaScript dependencies, provision Chromium, test the Go engine, and build the executable.

**Windows PowerShell:**

```powershell
Set-Location engine
npm ci
npx playwright install chromium --no-shell
npm run check:versions
go test ./...
New-Item -ItemType Directory -Force ..\bin | Out-Null
go build -o ..\bin\dawg.exe ./cmd/dawg
..\bin\dawg.exe doctor
```

**Linux:**

```bash
cd engine
npm ci
npx playwright install chromium --no-shell
npm run check:versions
go test ./...
mkdir -p ../bin
go build -o ../bin/dawg ./cmd/dawg
../bin/dawg doctor
```

For this unstaged setup, make sure `mitmdump` is available on `PATH` before running `doctor` or replaying an artifact with an HTTP cassette.

### Desktop Development Setup

Build the engine into `bin/` first, then install the desktop dependencies and start the native Tauri application:

```bash
cd desktop
npm ci
npm run build
npm run tauri dev
```

`npm run dev` starts only the Vite frontend at `http://localhost:1420`. Tauri IPC, native import/export dialogs, and engine commands require `npm run tauri dev`.

For capture testing, load the repository's `extension/` directory as an unpacked extension and reload it after source changes.

---

## Build a Self-Contained Distribution

The bundle helpers compile the Go engine, stage Node.js and mitmdump, install the locked Playwright/rrweb dependencies and Chromium, copy the replay script, schema, policies, and extension, run `dawg doctor`, and then invoke the platform-specific Tauri package build.

Install desktop dependencies before running either helper:

```bash
cd desktop
npm ci
cd ..
```

### Windows NSIS Installer

Run from the repository root in PowerShell:

```powershell
.\desktop\build-bundle.ps1
```

Useful options:

```powershell
# Reuse complete, current cached/staged dependencies
.\desktop\build-bundle.ps1 -SkipDownload

# Stage and validate resources without invoking Tauri
.\desktop\build-bundle.ps1 -SkipTauri

# Override the versions normally sourced from version.json
.\desktop\build-bundle.ps1 -MitmVersion "12.2.3" -NodeVersion "22.14.0"
```

`-SkipDownload` fails when required caches or staged dependencies are missing or stale. The installer is written under:

```text
desktop/src-tauri/target/release/bundle/nsis/
```

### Linux AppImage

Run from the repository root:

```bash
chmod +x ./desktop/build-bundle.sh
./desktop/build-bundle.sh
```

Useful options:

```bash
./desktop/build-bundle.sh --skip-download
./desktop/build-bundle.sh --skip-tauri
./desktop/build-bundle.sh --mitm-version=12.2.3 --node-version=22.14.0
```

The helper explicitly builds an AppImage under:

```text
desktop/src-tauri/target/release/bundle/appimage/
```

Tauri also declares a Debian target, but `build-bundle.sh` does not build it.

### Staged Resources

A successful staging run creates a generated resource tree similar to:

```text
desktop/src-tauri/resources/
├── binaries/
│   ├── dawg[.exe]
│   ├── mitmdump/
│   └── node/
├── node_modules/
├── browsers/chromium-*/
├── scripts/replay-browser.cjs
├── schema/
└── extension/
```

`desktop/src-tauri/resources/` is generated content and may be absent from a clean source checkout.

---

## Runtime Storage and Platform Boundaries

By default, DAWG keeps runtime state in the user's home directory:

```text
~/.dawg/
├── artifacts/
├── captures/
├── policies/
├── artifact-catalog.json
├── capture-control.json
└── capture-result.json
```

Set `DAWG_STATE_DIR` to move this state root. Runtime resource overrides include `DAWG_RESOURCES_DIR`, `DAWG_NODE_PATH`, `DAWG_MITMDUMP_PATH`, `DAWG_CHROMIUM_EXECUTABLE_PATH`, and `DAWG_POLICY_PATH`.

On Windows, replay runs Chromium natively and skips Docker Compose isolation
and database fixture restoration. On supported non-Windows hosts, environment
replay requires rootless Docker. Replay renders the recorded rrweb timeline; it
does not execute captured browser actions or recreate the original application
runtime.

## Diagnostic Evidence

Capture defaults to **Safe**, which stores bounded console/error and network
metadata without CDP response-body retrieval. **Enhanced Diagnostics** requires
explicit desktop consent (or `dawg capture --diagnostics-profile enhanced` in the
CLI), uses CDP when available, and can retain only eligible JSON, GraphQL,
form-encoded, or text response bodies within limits. CDP attachment failure
falls back to Safe and is recorded as a degradation.

Packaged diagnostic values include an evidence state such as `captured`,
`redacted`, `preview-only`, `truncated`, `blocked`, `unavailable`,
`not-requested`, or `capture-failed`. These states communicate fidelity; they do
not recover missing values. The Desktop Replay workspace and the diagnostics CLI
commands can inspect retained evidence, export sanitized HAR, and produce a
reviewed cURL request.

## Troubleshooting

### Engine Not Found in Desktop Development

Build the engine into the repository's `bin/` directory before starting `npm run tauri dev`. The desktop checks bundled resources, executable-relative install locations, repository build locations, the Windows registry `PATH`, and the process `PATH`.

### Capture Waits for the Extension

Confirm that:

- Chrome or Chromium is version 116 or newer;
- the DAWG extension is enabled; if using an unpacked development copy, it was reloaded after the last update;
- the target begins with `http://` or `https://`;
- loopback port `8082` is not blocked.

### Runtime Reports Degraded

Run:

```bash
dawg doctor
```

For a source runtime, verify that Node.js, mitmdump, Playwright Chromium, `replay-browser.cjs`, the extension manifest, schema, and policy are discoverable. For a staged bundle, do not use the skip-download option until all cached and generated resources are complete and current.
