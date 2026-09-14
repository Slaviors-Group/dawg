# Contributing

Thank you for your interest in contributing to DAWG! This guide covers the current repository layout, development setup, validation commands, version management, and pull request workflow.

## Development Setup

### Requirements

| Tool | Version or scope | Purpose |
|---|---|---|
| **Go** | `1.25.1` | Engine compilation, vetting, and tests |
| **Node.js** | `^20.19.0` or `>=22.12.0` | Vite 7 desktop toolchain |
| **npm** | Compatible with the selected Node.js release | Locked engine and desktop dependencies |
| **Rust** | `1.85+`, stable Tauri v2-compatible toolchain | Desktop backend |
| **Chrome/Chromium** | `116+` | Extension capture testing |
| **Docker Engine + Compose** | Rootless where environment replay is supported | Non-Windows sandbox and database restore testing |

Windows desktop work also requires the MSVC Rust target, Visual Studio C++ Build Tools, WebView2, and PowerShell. Linux desktop work requires the native WebKitGTK, GTK, AppIndicator, librsvg, and GStreamer dependencies expected by Tauri and the AppImage build.

### Clone and Install

```bash
git clone https://github.com/Slaviors-Group/dawg.git
cd dawg
```

Install each package from its lockfile:

```bash
cd engine
npm ci

cd ../desktop
npm ci
```

Build the engine into the repository's `bin/` directory before running the native desktop app. For complete source runtime diagnostics, also install Playwright Chromium and make `mitmdump` available on `PATH`; see [Installation](/docs/installation).

### Run the Desktop App

```bash
cd desktop
npm run build
npm run tauri dev
```

`npm run dev` starts only the Vite frontend. Use `npm run tauri dev` when testing engine IPC, native `.dawg` file dialogs, drag-and-drop, capture, or replay.

Load `extension/` as an unpacked extension from `chrome://extensions` for capture development. Reload it after changing extension source.

### Optional Pre-commit Hooks

The checked-in `.pre-commit-config.yaml` defines local hooks for:

- `gofmt -w` on Go files;
- `go vet ./...` from `engine/`;
- `npx biome check --write` on JavaScript, JSX, TypeScript, TSX, and JSON files.

After installing `pre-commit`, enable and run the hooks with:

```bash
pre-commit install
pre-commit run --all-files
```

---

## Project Structure

```text
dawg/
├── version.json          # Application, desktop, extension, and runtime versions
├── engine/               # Go CLI and capture/sanitize/package/replay pipeline
├── extension/            # Chromium Manifest V3 capture extension
├── desktop/              # React 19 frontend and Tauri v2 backend
├── schema/               # Artifact schema, media types, and OPA policy
├── tools/                # Version synchronization utility
└── web/                  # VitePress documentation site
```

### Language by Concern

| Concern | Language | Location |
|---|---|---|
| Engine CLI and pipeline | Go | `engine/` |
| Browser capture | JavaScript | `extension/` |
| Desktop UI | TypeScript/React | `desktop/src/` |
| Desktop native backend | Rust | `desktop/src-tauri/` |
| Artifact validation and policy | JSON Schema/Rego | `schema/` |
| Documentation | Markdown/Vue/TypeScript | `web/` |

Keep changes within the component that owns the behavior. Desktop IPC commands should remain thin wrappers around engine operations where practical.

---

## Coding Standards

### Go

- Format with `gofmt` and keep `go vet ./...` clean.
- Use table-driven tests where multiple cases exercise the same behavior.
- Keep packages focused and errors actionable.
- Use the existing structured logging and error types instead of introducing parallel patterns.

### TypeScript and React

- Use the checked-in Biome configuration.
- Prefer functional components and hooks.
- Keep typed IPC payloads aligned with the engine's emitted JSON field names.
- Use the existing UI primitives and Tailwind styles.

### Rust

- Run `cargo fmt`, `cargo clippy`, and tests for native backend changes.
- Keep subprocess ownership and cleanup explicit.
- Return actionable `Result` errors across Tauri command boundaries.

### Browser Extension

- Keep the Manifest V3 service worker restart-safe.
- Preserve session-token checks and selected-tab scoping.
- Remember that rrweb input masking does not mask the separate action stream; sanitization occurs in the engine before packaging.

---

## Validation

Run the checks for every component you change. The commands below match the current package scripts and CI coverage.

### Engine and Version Metadata

From `engine/`:

```bash
npm ci
npm run check:versions
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/dawg
```

### Browser Extension

From the repository root:

```bash
node --check extension/background/service_worker.js
node --check extension/content/recorder.js
node --check extension/popup/popup.js
node --test extension/tests/service_worker.test.cjs
```

The automated extension test mocks Chrome APIs. Capture changes still require a manual Chrome/Chromium 116+ start, event delivery, stop, package, and replay cycle.

### Desktop

From `desktop/`:

```bash
npm ci
npx --no-install biome check .
npm run build
cargo fmt --manifest-path src-tauri/Cargo.toml -- --check
cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets -- -D warnings
cargo test --manifest-path src-tauri/Cargo.toml
```

For CI parity, also ensure the Rust package passes a locked check:

```bash
cargo check --locked --manifest-path src-tauri/Cargo.toml
```

### Documentation Site

From `web/`:

```bash
npm ci
npm run docs:build
```

### Bundle and Native Smoke Testing

Stage the complete runtime when a change affects discovery, packaging, or native process behavior:

```powershell
.\desktop\build-bundle.ps1 -SkipTauri
```

```bash
./desktop/build-bundle.sh --skip-tauri
```

Do not use the skip-download option unless the runtime cache and staged dependencies are already complete and current.

A manual native smoke test should cover the paths affected by the change, including capture start/stop, catalog persistence, `.dawg` import/export, replay, replay cancellation, and application exit during active work.

---

## Version Management

[`version.json`](https://github.com/Slaviors-Group/dawg/blob/main/version.json) is the source for application, desktop, extension, Node.js, and mitmproxy versions.

The current release values are:

| Field | Value |
|---|---|
| Application and schema | `0.2.3-naughty` |
| Desktop/Tauri/Cargo | `0.2.3` |
| Extension display version | `0.2.3_naughty` |
| Extension install version | `0.2.3` |
| Bundled Node.js | `22.14.0` |
| Bundled mitmproxy | `12.2.3` |

After changing `version.json`, run the synchronizer and consistency check from either `engine/` or `desktop/`:

```bash
npm run sync:versions
npm run check:versions
```

The synchronizer updates the engine and desktop package metadata, lockfiles, CLI constant, Cargo and Tauri versions, extension manifest, bundle runtime pins, settings display, schema filename, and schema references. Do not edit those generated version fields independently.

Chrome requires a numeric extension version, so `0.2.3_naughty` becomes `version: "0.2.3"` with `version_name: "0.2.3_naughty"`. Dependency versions remain managed by package manifests and lockfiles.

---

## Pull Request Workflow

### 1. Create a Focused Branch

The repository integrates work through `staging` before `main`. Start from the current remote staging branch:

```bash
git fetch origin
git switch --create feature/my-feature origin/staging
```

Use a descriptive prefix such as `feature/`, `fix/`, or `docs/`.

### 2. Make Focused Changes

- Follow the component boundaries and coding standards above.
- Add or update tests for behavior changes.
- Update documentation when commands, requirements, security boundaries, or user workflows change.
- Avoid committing generated build outputs or staged runtime resources.

### 3. Verify Locally

Run the narrowest relevant checks first, then the broader component checks. Include native and manual browser validation when automated tests cannot exercise the changed path.

### 4. Commit Clearly

Use a short imperative subject that describes the change. Add a concise body only when it explains useful context, constraints, or migration details that the subject cannot capture.

### 5. Open the Pull Request

Target `staging` and include:

- a concise description of the behavior changed;
- the related issue, when applicable;
- screenshots for visible desktop or documentation changes;
- the exact validation commands and manual checks performed;
- known platform-specific limitations.

## Reporting Issues

Open an issue on [GitHub Issues](https://github.com/Slaviors-Group/dawg/issues) and include:

- reproduction steps and target URL shape without secrets;
- expected and actual behavior;
- operating system and browser version;
- `dawg doctor` output;
- relevant desktop/engine logs;
- whether the artifact was captured, imported, or pulled.

Do not attach an artifact until you have inspected it for sensitive data. Sanitization reduces exposure but is not a confidentiality guarantee.

## License

By contributing, you agree that your contribution is licensed under the repository's [GNU General Public License v3.0](https://github.com/Slaviors-Group/dawg/blob/main/LICENSE).
