# DAWG (Digs Any Web-app Glitch)

**DAWG** packages sanitized, deterministic web-app bug reproductions into portable OCI artifacts for automated verification, replay, and instant debugging.

---

## Features

- 🎥 **Full-fidelity Capture**: Records DOM mutations via rrweb, executable browser action traces (Playwright), frontend HTTP streams, and backend API interactions (`mitmdump`).
- 🛡️ **Automated Sanitization**: Hard-gated PII redaction and synthetic value replacement powered by Open Policy Agent (OPA) Rego policies.
- 📦 **Standardized OCI Packaging**: Packages reproduction sessions into standard OCI Image Layouts with SHA-256 digest-pinned layers.
- 🩺 **Built-in Diagnostics**: Integrated `dawg doctor` probe to verify runtime readiness and bundled assets.
- 🖥️ **Desktop GUI & CLI**: Interactive Tauri v2 desktop application and headless CLI for developers and CI/CD pipelines.

---

## Repository Layout

```text
dawg/
├── engine/              # Go Engine CLI implementation (cmd/dawg, capture, packager, replay, verify)
├── desktop/             # Tauri v2 Desktop Shell (Rust backend + React 19 / Vite frontend)
├── schema/              # OCI Manifest JSON Schema and OPA Rego sanitization policies
├── test/                # End-to-end smoke tests and testing environments
└── bin/                 # Compiled binary outputs (dawg.exe / dawg)
```

---

## Quick Start: Desktop App (Self-Contained Installer)

The desktop application is distributed as a **monolithic standalone bundle** (zero manual prerequisites required — all runtimes including `mitmdump`, `node`, Playwright scripts, and schemas are embedded).

1. Download the latest installer for your operating system:
   - **Windows**: `DAWG_x64-setup.exe` / `.msi`
   - **Linux**: `DAWG_amd64.AppImage` / `.deb`
2. Launch **DAWG**. The header will display `Engine Ready [Bundled]` with all runtimes automatically initialized.
3. Enter your web application target URL, click **Start Capture**, reproduce the issue in the browser, and click **Stop Capture** to generate your reproduction artifact.

---

## Quick Start: Engine CLI

### Prerequisites for Standalone CLI Development

- **Go 1.22+**
- **Node.js 20+** (with Playwright)
- **mitmproxy** (`mitmdump` on PATH, or bundled via `dawg doctor`)

### Verification & Diagnostics

Run the built-in diagnostic tool to inspect your environment:

```powershell
dawg doctor
```

```text
DAWG Engine Doctor Diagnostics (v0.1.3-alpha)
Resource Root: D:\Next Project\pandi-petualang\dawg\bin (Bundled: false)

✅ dawg-engine          (v0.1.3-alpha) -> D:\...\bin\dawg.exe
✅ mitmdump             (vMitmproxy: 12.2.3) -> ...\mitmdump.exe
✅ node                 (vv22.14.0) -> ...\node.exe
✅ script:capture-browser.cjs [bundled] -> ...\scripts\capture-browser.cjs
✅ script:capture-proxy.py [bundled] -> ...\scripts\capture-proxy.py
✅ script:replay-browser.cjs [bundled] -> ...\scripts\replay-browser.cjs
✅ policy:default.rego  -> ...\schema\policies\default.rego
✅ schema:manifest      (v0.1.3-alpha)

All required runtime components are healthy and ready! 🚀
```

### Basic CLI Commands

```powershell
# Initialize config in a target project
dawg init

# Capture a web-app session
dawg capture --url https://my-app.local

# Stop an active capture session
dawg capture stop

# Inspect an OCI reproduction artifact without replaying
dawg inspect .dawg/artifacts/<session-id>

# Replay a reproduction artifact
dawg run .dawg/artifacts/<session-id>

# Verify reproduction against local codebase
dawg verify .dawg/artifacts/<session-id> --against local
```

---

## Building & Packaging

### 1. Build the Monolithic Desktop Bundle

To compile the Go engine, stage all standalone runtimes (`mitmdump`, `node`, `playwright`), and build the final desktop installer:

- **Windows (PowerShell)**:

  ```powershell
  .\desktop\build-bundle.ps1 -SkipDownload
  ```

  _Output:_ `desktop/src-tauri/target/release/bundle/nsis/DAWG_x64-setup.exe`

- **Linux (Bash)**:
  ```bash
  chmod +x ./desktop/build-bundle.sh
  ./desktop/build-bundle.sh
  ```
  _Output:_ `desktop/src-tauri/target/release/bundle/appimage/DAWG_amd64.AppImage`

### 2. Development Mode

#### Running the Go Engine Tests

```powershell
cd engine
go test ./...
```

#### Running the Desktop App in Development

```powershell
cd desktop
npm install
npm run dev      # Runs Vite web dev server (http://localhost:1420)
npm run tauri dev # Runs Tauri window with live-reloading
```

#### Running the End-to-End Smoke Test

```powershell
.\test\smoke_test.ps1
```

---

## License

Apache-2.0. See [LICENSE](LICENSE) for details.
