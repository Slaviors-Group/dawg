# 🖥️ DAWG Desktop

DAWG Desktop is the Tauri v2 interface for the Go engine. It uses Rust, React 19,
Vite 7, Tailwind CSS 4, Framer Motion, and Phosphor icons.

> Application package: `0.2.3-naughty` · Tauri/Cargo package: `0.2.3`
>
> Bundled Node.js: `22.14.0` · Bundled mitmproxy: `12.2.3`

The current [`0.2.3-naughty` prerelease](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-2)
provides a Windows x64 NSIS installer. Linux targets are configured but must
currently be built from source.

## 🧭 Current Desktop Scope

The engine remains responsible for capture, sanitization, OCI packaging, replay,
verification, and registry operations. The desktop application currently
provides engine status, capture controls, a persistent artifact browser, archive
import/export, replay controls, logs, and runtime diagnostics.

- **Capture:** starts `dawg capture` with a URL and optional title, tracks the
  detached capture daemon, and waits for packaging when capture stops.
- **Artifact catalog:** loads validated artifacts from the engine at startup and
  refreshes after capture or import.
- **Import/export:** uses native `.dawg` file dialogs; the dashboard also accepts
  a dropped `.dawg` archive.
- **Replay:** runs `dawg run <artifact>` and exposes **Stop Replay** while the
  subprocess is active.
- **Process cleanup:** tracks replay and capture-daemon PIDs. Windows cancellation
  uses `taskkill /T /F`; application exit attempts to terminate tracked work.
- **Doctor:** displays `dawg doctor --output json` component status.
- **Preferences:** stores theme and motion settings in browser local storage.

The Sanitizer Policy screen currently displays built-in policy examples; it does
not edit the engine policy. The Diff & Verify screen is currently a UI preview
and does not execute verification. Use `dawg verify` from the CLI for the current
engine implementation.

## 🧱 Architecture

| Area | Location | Responsibility |
| --- | --- | --- |
| Application shell | `src/App.tsx` | Navigation, providers, settings, doctor modal |
| Feature views | `src/components/` | Dashboard, capture, replay, diagnostics, logs |
| Engine state | `src/context/EngineContext.tsx` | Capture state, catalog, imports/exports, shared logs |
| Typed IPC | `src/lib/engine.ts` | TypeScript wrappers for Tauri commands |
| Native backend | `src-tauri/src/lib.rs` | Engine discovery, resource environment, subprocesses, cancellation |
| Staged runtime | `src-tauri/resources/` | Engine and replay/capture dependencies assembled by bundle scripts |
| Packaging | `build-bundle.ps1`, `build-bundle.sh` | Runtime downloads/staging, doctor check, Tauri build |

Engine output is buffered until each command completes; the current UI does not
stream subprocess stdout/stderr live. Capture progress shown before stop/package
completion is a UI estimate rather than engine-reported progress.

## 🔁 Runtime Flow

### Capture

1. React invokes `start_capture` with the target URL and optional title.
2. Tauri executes `dawg capture --url ... --title ... --output json`.
3. The engine launches its detached capture daemon and returns its PID and
   control-file path.
4. Tauri registers the daemon PID for exit cleanup.
5. Stop invokes `dawg capture stop`; that command waits for extension drain,
   sanitization, OCI packaging, and catalog registration.
6. The desktop refreshes the catalog after a packaged artifact is returned.

### Replay

1. React invokes `run_replay` with the selected artifact directory.
2. Tauri starts `dawg run <artifact> --output json` and registers the PID.
3. The engine unpacks the OCI layers and launches the replay script with the
   staged Node.js/Playwright/Chromium runtime.
4. The completed command returns replay diagnostics and outcome metadata.
5. **Stop Replay** terminates the tracked process tree. Application exit performs
   the same cleanup for tracked replay and capture processes.

On Windows, replay runs in native compatibility mode: Docker Compose sandboxing
and database restoration are skipped. Supported non-Windows environment replay
requires rootless Docker.

## 🧑‍💻 Development

### Requirements

- Node.js `^20.19.0` or `>=22.12.0` (Vite 7 requirement)
- npm
- Go `1.25.1` to build the engine
- Rust `1.85+` with a Tauri v2-compatible toolchain
- Chrome or Chromium `116+` with `../extension/` loaded for capture testing

Additional native dependencies:

- **Windows:** MSVC Rust target, Visual Studio C++ Build Tools, WebView2, and
  PowerShell.
- **Linux:** a C/C++ toolchain plus the WebKitGTK, GTK, AppIndicator, librsvg,
  and GStreamer development/runtime packages required by Tauri and the configured
  AppImage media bundle. The bundle script also uses Bash, `curl`, `tar`, and xz.

### Install and run

```powershell
npm ci
npm run tauri dev
```

The native app resolves the engine from staged resources when available, then
checks executable-relative installation paths, repository build locations, and
`PATH`.

To build only the frontend:

```powershell
npm run build
```

`npm run dev` starts only Vite. Engine IPC and native file dialogs are unavailable
in a normal browser tab.

## 📦 Build a Self-Contained Distribution

The bundle scripts build the Go engine, stage Node.js, mitmdump, Playwright,
Chromium, replay scripts, schema/policy files, and the extension, run
`dawg doctor`, then optionally invoke Tauri packaging.

### Windows

```powershell
.\build-bundle.ps1
```

Use cached/staged dependencies:

```powershell
.\build-bundle.ps1 -SkipDownload
```

Stage resources without building the installer:

```powershell
.\build-bundle.ps1 -SkipTauri
```

The NSIS installer is written below:

```text
src-tauri/target/release/bundle/nsis/
```

### Linux

```bash
chmod +x ./build-bundle.sh
./build-bundle.sh
```

The helper script builds an AppImage under:

```text
src-tauri/target/release/bundle/appimage/
```

The Tauri configuration declares a Debian target, but this helper does not build
it.

Use `./build-bundle.sh --skip-tauri` to stage and check resources without
building packages.

### Staged resources

```text
src-tauri/resources/
├── binaries/
│   ├── dawg[.exe]
│   ├── mitmdump/
│   └── node/
├── node_modules/                # Playwright and rrweb packages
├── browsers/chromium-*/         # Playwright Chromium
├── scripts/replay-browser.cjs
├── schema/                      # Manifest schema and OPA policy
└── extension/                   # Installable MV3 extension
```

`src-tauri/resources/` is generated/staged content and may be absent in a clean
source checkout.

## 🔢 Versioning

[`../version.json`](../version.json) is the source for application, desktop,
extension, and bundled runtime versions.

```powershell
npm run sync:versions
npm run check:versions
```

The desktop npm package uses the labeled application version
`0.2.3-naughty`; Cargo and Tauri use numeric version `0.2.3`.

## ✅ Validation

```powershell
npm ci
npm run check:versions
npx biome check .
npm run build
cargo fmt --manifest-path src-tauri/Cargo.toml -- --check
cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets -- -D warnings
cargo test --manifest-path src-tauri/Cargo.toml
```

After staging a bundle, check its resources:

```powershell
.\src-tauri\resources\binaries\dawg.exe doctor
```

A native smoke test should cover capture start/stop, catalog persistence,
`.dawg` import/export, replay cancellation, and application exit during active
capture or replay.

For engine commands, release downloads, and the complete project workflow, see
the [repository README](../README.md).
