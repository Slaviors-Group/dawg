# Installation

DAWG can be installed as a standalone desktop application (recommended) or used as a CLI tool for development and CI/CD integration.

## Desktop Application (Recommended)

The desktop app is distributed as a **monolithic standalone bundle** — zero manual prerequisites required. All runtimes including `mitmdump`, `node`, Playwright scripts, and schemas are embedded.

### Windows

Download the installer from the release page for your version:

| Version | Date | Installer | Download |
|---|---|---|---|
| **v0.1.2-alpha** (latest) | 2026-09-06 | `DAWG_0.1.2-alpha_x64-setup.exe` (NSIS) | [alpha-3](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-3) |
| v0.1.1-alpha | 2026-09-04 | `DAWG_0.1.1-alpha_x64-setup.exe` (NSIS) | [alpha-2](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-2) |
| v0.1.0-alpha | 2026-09-04 | `DAWG_0.1.0-alpha_x64-setup.exe` (NSIS) / `DAWG_0.1.0-alpha_x64_en-US.msi` | [alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha) |

Run the installer and follow the prompts. DAWG will be installed to your Program Files directory with Start Menu shortcuts.

> **Note:** MSI distribution was deprecated after v0.1.0-alpha — newer releases ship NSIS (`.exe`) only.

### Linux

Prebuilt Linux packages are not yet available. Build from source:

```bash
# Clone the repository
git clone https://github.com/Slaviors-Group/dawg.git
cd dawg/engine

# Build the Go engine
go build -o ../bin/dawg ./cmd/dawg

# Build the desktop app
cd ../desktop
npm install
npm run build
```

See [Building from Source](#building-from-source) for detailed instructions.

---

## Engine CLI (Standalone)

For developers who want to use DAWG from the command line or integrate it into CI/CD pipelines.

### Prerequisites

| Dependency | Version | Purpose |
|---|---|---|
| **Go** | 1.22+ | Engine compilation |
| **Node.js** | 20+ | Playwright browser automation |
| **Python** | 3.10+ | mitmproxy for HTTP capture |
| **mitmproxy** | 12+ | Backend HTTP traffic interception |

### Install from Source

```bash
# Clone the repository
git clone https://github.com/Slaviors-Group/dawg.git
cd dawg/engine

# Build the engine binary
go build -o ../bin/dawg ./cmd/dawg

# Install Node.js dependencies (for Playwright scripts)
npm install

# Add to PATH (optional)
export PATH="$PWD/../bin:$PATH"
```

### Verify Installation

Run the built-in diagnostic tool to verify your environment:

```bash
dawg doctor
```

Expected output:

```
DAWG Engine Doctor Diagnostics (v0.1.2-alpha)
Resource Root: /path/to/dawg/bin (Bundled: false)

✅ dawg-engine          (v0.1.2-alpha)
✅ mitmdump             (v12.2.3)
✅ node                 (v22.14.0)
✅ script:capture-browser.cjs
✅ script:capture-proxy.py
✅ script:replay-browser.cjs
✅ policy:default.rego
✅ schema:manifest      (v0.1.2-alpha)

All required runtime components are healthy and ready!
```

---

## Bundling from Source

Building DAWG is **one linear pipeline**, not separate builds: compile the engine → stage the runtimes (mitmdump, Node) → copy scripts, schemas, and node modules into the bundle → verify → build the installer. The `build-bundle` scripts automate the whole flow.

### Prerequisites

| Tool | Version | Needed for |
|---|---|---|
| Go | 1.22+ | Compiling the engine |
| Node.js + npm | 20+ | Desktop frontend, Playwright deps |
| Rust toolchain + Tauri CLI | stable | Building the installer |
| Docker (rootless) | 24+ | Replay sandbox (runtime, not build) |

```bash
# Clone the repository
git clone https://github.com/Slaviors-Group/dawg.git
cd dawg
```

### Option A — One-command bundle (recommended)

**Windows (PowerShell):**

```powershell
# Full pipeline: compile + stage runtimes + build the Tauri installer
.\desktop\build-bundle.ps1

# Reuse already-staged runtimes instead of re-downloading
.\desktop\build-bundle.ps1 -SkipDownload

# Stage the bundle only, build the installer manually later
.\desktop\build-bundle.ps1 -SkipTauri
cd desktop
npm run tauri build

# Override the pinned runtime versions (defaults shown)
.\desktop\build-bundle.ps1 -MitmVersion "12.2.3" -NodeVersion "22.14.0"
```

**Linux (Bash):**

```bash
chmod +x ./desktop/build-bundle.sh

# Full pipeline
./desktop/build-bundle.sh

# Same flags, sh-style
./desktop/build-bundle.sh --skip-download
./desktop/build-bundle.sh --skip-tauri
./desktop/build-bundle.sh --mitm-version=12.2.3 --node-version=22.14.0
```

| Flag (ps1 / sh) | Effect |
|---|---|
| `-SkipDownload` / `--skip-download` | Reuses staged binaries and `.cache/` archives instead of re-downloading |
| `-SkipTauri` / `--skip-tauri` | Stops after staging; run `npm run tauri build` inside `desktop/` yourself |
| `-MitmVersion` / `--mitm-version=` | Pin mitmproxy version (default `12.2.3`) |
| `-NodeVersion` / `--node-version=` | Pin portable Node.js version (default `22.14.0`) |

**Installer outputs:**

| OS | Output |
|---|---|
| Windows | `desktop/src-tauri/target/release/bundle/nsis/DAWG_x64-setup.exe` |
| Linux | `desktop/src-tauri/target/release/bundle/appimage/DAWG_amd64.AppImage` (+ `.deb`) |

### Option B — Manual walkthrough (same pipeline, step by step)

Use this when debugging the bundle or porting to a new platform. Each step feeds the next — don't skip ahead.

**Step 1 — Compile the engine into the bundle**

```bash
cd engine
go test ./...   # engine test suite must pass first
```

```powershell
# Windows: binary lands straight in the staged bundle
go build -o ../desktop/src-tauri/resources/binaries/dawg.exe ./cmd/dawg
```

```bash
# Linux: same, without the .exe extension
go build -o ../desktop/src-tauri/resources/binaries/dawg ./cmd/dawg
chmod +x ../desktop/src-tauri/resources/binaries/dawg
```

**Step 2 — Stage standalone mitmdump**

No system mitmproxy install needed — fetch the portable archive (cached under `desktop/.cache/` on repeat runs):

```powershell
# Windows: zip from downloads.mitmproxy.org (fallback: GitHub Releases),
# extract mitmdump.exe into the bundle (or copy your local mitmdump.exe)
Invoke-WebRequest -Uri "https://downloads.mitmproxy.org/12.2.3/mitmproxy-12.2.3-windows-x86_64.zip" -OutFile "desktop/.cache/mitmproxy-12.2.3.zip"
Expand-Archive -Path "desktop/.cache/mitmproxy-12.2.3.zip" -DestinationPath "desktop/.cache/mitm_extract"
Copy-Item "desktop/.cache/mitm_extract/mitmdump.exe" "desktop/src-tauri/resources/binaries/mitmdump/mitmdump.exe"
```

```bash
# Linux: tarball, extract only the mitmdump binary
curl -fsSL "https://downloads.mitmproxy.org/12.2.3/mitmproxy-12.2.3-linux-x86_64.tar.gz" -o "desktop/.cache/mitmproxy-12.2.3-linux.tar.gz"
tar -xzf "desktop/.cache/mitmproxy-12.2.3-linux.tar.gz" -C "desktop/src-tauri/resources/binaries/mitmdump" mitmdump
chmod +x "desktop/src-tauri/resources/binaries/mitmdump/mitmdump"
```

**Step 3 — Stage portable Node.js**

```powershell
# Windows
Invoke-WebRequest -Uri "https://nodejs.org/dist/v22.14.0/node-v22.14.0-win-x64.zip" -OutFile "desktop/.cache/node-v22.14.0-win-x64.zip"
Expand-Archive -Path "desktop/.cache/node-v22.14.0-win-x64.zip" -DestinationPath "desktop/.cache/node_extract"
Copy-Item "desktop/.cache/node_extract/*/node.exe" "desktop/src-tauri/resources/binaries/node/node.exe"
```

```bash
# Linux
curl -fsSL "https://nodejs.org/dist/v22.14.0/node-v22.14.0-linux-x64.tar.xz" -o "desktop/.cache/node-v22.14.0-linux-x64.tar.xz"
mkdir -p "desktop/.cache/node_extract"
tar -xf "desktop/.cache/node-v22.14.0-linux-x64.tar.xz" -C "desktop/.cache/node_extract" --strip-components=1
cp "desktop/.cache/node_extract/bin/node" "desktop/src-tauri/resources/binaries/node/node"
chmod +x "desktop/src-tauri/resources/binaries/node/node"
```

**Step 4 — Provision Playwright & rrweb modules**

If `engine/node_modules` already exists it gets copied over; otherwise install directly into the bundle:

```bash
cd engine
npm install   # playwright + rrweb (or: hoisted from a previous install)
cp -r node_modules ../desktop/src-tauri/resources/node_modules
# ...or fresh: cp package.json ../desktop/src-tauri/resources/ && (cd ../desktop/src-tauri/resources && npm install --omit=dev)
```

**Step 5 — Copy engine scripts and schema**

```bash
# From the repo root:
cp -r engine/scripts/* desktop/src-tauri/resources/scripts/
cp -r schema/* desktop/src-tauri/resources/schema/
```

The staged bundle now looks like this:

```text
desktop/src-tauri/resources/
├── binaries/
│   ├── dawg(.exe)              # ← step 1
│   ├── mitmdump/mitmdump(.exe) # ← step 2
│   └── node/node(.exe)         # ← step 3
├── node_modules/               # ← step 4 (playwright, rrweb)
├── scripts/                    # ← step 5 (capture-browser.cjs, capture-proxy.py, replay-browser.cjs)
└── schema/                     # ← step 5 (manifest JSON schema, OPA policies)
```

**Step 6 — Verify the staged bundle**

Point `dawg doctor` at the staged resources — every check must pass before building the installer:

```bash
DAWG_RESOURCES_DIR=desktop/src-tauri/resources ./desktop/src-tauri/resources/binaries/dawg doctor
```

```powershell
$env:DAWG_RESOURCES_DIR = "desktop/src-tauri/resources"
& "desktop/src-tauri/resources/binaries/dawg.exe" doctor
```

**Final — Build the installer**

```bash
cd desktop
npm run tauri build
```

Outputs are listed in [Option A](#option-a-one-command-bundle-recommended).

### Desktop development mode

For UI iteration (no bundling needed — uses your system toolchain):

```bash
cd desktop
npm install
npm run dev        # Vite web dev server (http://localhost:1420)
npm run tauri dev  # Tauri window with live-reloading
```

### Run the end-to-end smoke test

```powershell
.\test\smoke_test.ps1
```

---

## Docker Sandbox (for Replay)

The replay engine requires Docker for sandboxed execution:

- **Windows**: WSL2 with Rootless Docker
- **Linux**: Docker with rootless mode enabled

```bash
# Verify Docker is running
docker info

# Enable rootless mode (if not already)
dockerd-rootless-setuptool.sh install
```

On bare Windows hosts without WSL2, DAWG falls back to native mock sandbox mode.

---

## Troubleshooting

### "CLI Not Detected" in Desktop App

The Desktop app looks for the `dawg` binary in these locations (in order):

1. Bundled resources (inside the app)
2. `./bin/dawg` (relative to working directory)
3. `$PATH`
4. Windows Registry (Windows only)

Build the engine first: `cd engine && go build -o ../bin/dawg ./cmd/dawg`

### Playwright Browser Not Found

```bash
# Install Playwright browsers
npx playwright install chromium
```

### mitmproxy Not Found

```bash
# Install via pip
pip install mitmproxy

# Or on Windows, ensure mitmdump.exe is in PATH
```
