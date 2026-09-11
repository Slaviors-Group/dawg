# DAWG Desktop Shell

The DAWG desktop application is a Tauri v2 shell for the DAWG engine. It is
built with **Rust**, **React 19**, **Vite**, and **Tailwind CSS**, and packages a
self-contained Windows/Linux runtime for capture artifact replay.

> Current generated desktop package version: `0.2.0`
>
> Current engine/schema release version: `0.2.0-naughty`

---

## Desktop Responsibilities

The desktop shell intentionally keeps pipeline behavior in the Go engine. It
provides the UI, starts engine commands through typed IPC, manages process
lifecycle, and resolves bundled resources.

- **Capture** — selects a target URL and asks `dawg capture` to begin an
  extension-driven capture session.
- **Replay** — invokes `dawg run <artifact>`, displays replay diagnostics, and
  exposes **Stop Replay** while Chromium is running.
- **Safe shutdown** — tracks replay and capture-daemon PIDs. Closing the DAWG
  window force-terminates DAWG-owned background processes, including the replay
  engine, Node.js, Chromium, mitmdump, and an active capture daemon.
- **Doctor** — invokes `dawg doctor` and displays component health for the
  staged/bundled runtime.

---

## Architecture

| Area | Location | Responsibility |
| --- | --- | --- |
| React UI | `src/` | Capture, replay, logs, artifact selection, settings, diagnostics |
| Typed IPC | `src/lib/engine.ts` | TypeScript bridge to Tauri commands |
| Engine context | `src/context/EngineContext.tsx` | Application state, sessions, artifacts, UI logs |
| Tauri backend | `src-tauri/src/lib.rs` | Engine resolution, resource environment, subprocess tracking/cancellation |
| Staged runtime | `src-tauri/resources/` | Engine, Node.js, mitmdump, Playwright modules, Chromium, schema, policy, extension |
| Bundle scripts | `build-bundle.ps1`, `build-bundle.sh` | Runtime staging, health check, platform package build |

### Replay Lifecycle

1. The UI invokes `run_replay` with the selected artifact path.
2. Tauri starts `dawg run` and records its PID.
3. The engine launches `replay-browser.cjs` through the bundled Node.js and
   Chromium runtime.
4. Replay logs include rrweb event shape, viewport metadata, in-page errors,
   and the visible-text length of the reconstructed document.
5. **Stop Replay** kills the tracked process tree. Closing the application does
   the same as a final safety measure.

A successful rrweb replay normally has one Meta event, one FullSnapshot, one
replay iframe, and a positive visible-text length. Warnings about malformed SVG
geometry from older artifacts indicate they were sanitized before the current
SVG-preservation fix; recapture those sessions.

---

## Development

### Prerequisites

- Node.js 20+
- Rust toolchain supported by Tauri v2
- Go 1.22+ to build the engine
- The DAWG browser extension loaded from `../extension/` for capture testing

### Install dependencies

```bash
npm install
```

### Run the native application with live reload

```bash
npm run tauri dev
```

To run the frontend alone in a browser:

```bash
npm run dev
```

The development shell resolves the engine from the staged resource directory
when present, then falls back to repository and PATH locations. Run the bundle
staging command after engine/runtime changes so desktop development uses current
assets.

---

## Build a Monolithic Distribution

The bundle scripts stage all replay dependencies, run `dawg doctor`, then build
the platform distribution.

### Windows

From this directory:

```powershell
.\build-bundle.ps1
```

For an already-populated cache/resource tree:

```powershell
.\build-bundle.ps1 -SkipDownload
```

The script invokes the local Tauri CLI directly, preserving the NSIS bundle
argument on Windows. The installer is written below:

```text
src-tauri/target/release/bundle/nsis/
```

### Linux

```bash
chmod +x ./build-bundle.sh
./build-bundle.sh
```

Linux package output is written below:

```text
src-tauri/target/release/bundle/
```

### Staged Resources

A healthy staged resource tree contains:

```text
src-tauri/resources/
├── binaries/dawg.exe            # Platform-specific engine executable
├── binaries/mitmdump/           # Standalone mitmdump
├── binaries/node/               # Portable Node.js
├── node_modules/                # Playwright and rrweb dependencies
├── browsers/chromium-*/         # Playwright Chromium
├── scripts/replay-browser.cjs   # Browser replay launcher
├── schema/                      # Manifest schema and OPA policy
└── extension/                   # Installable DAWG browser extension
```

---

## Versioning

Do not manually update package, Cargo, Tauri, or extension manifest versions.
[`../version.json`](../version.json) is the source of truth.

```bash
npm run sync:versions
npm run check:versions
```

The synchronizer accepts convenient labels such as `0.2-naughty` and emits
strict SemVer (`0.2.0-naughty`) wherever Tauri, Cargo, npm, and the artifact
schema require it. Chrome extension labels are emitted through `version_name`
while the extension's installable version remains numeric.

---

## Useful Validation Commands

```bash
# Verify generated metadata matches version.json
npm run check:versions

# Type-check the React frontend
npx tsc --noEmit

# Build the Rust desktop backend
cargo build --manifest-path src-tauri/Cargo.toml

# Verify staged runtime resources
.\src-tauri\resources\binaries\dawg.exe doctor
```

For project-wide workflow and CLI instructions, see the
[repository README](../README.md).
