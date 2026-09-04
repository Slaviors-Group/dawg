# DAWG Desktop Shell

This directory contains the desktop GUI application for DAWG, built with **Tauri v2**, **Rust**, and **React 19 / Vite / Tailwind CSS**.

---

## Architecture Overview

- **Frontend (`src/`)**: React 19 single-page application with tabs for Capture Controls, Live Session Logs, Replay Viewer, Verification Diff Reports, and Sanitizer Policy Configuration.
- **IPC Bridge (`src/lib/engine.ts`)**: Strongly typed TypeScript interface invoking Tauri Rust commands.
- **Rust Backend (`src-tauri/src/lib.rs`)**: Manages sidecar engine binary resolution (`dawg.exe` / `dawg`), injects `DAWG_RESOURCES_DIR`, and handles subprocess execution and health probes.
- **Bundled Resources (`src-tauri/resources/`)**: Staging root for embedded standalone runtimes (`mitmdump`, `node`, Playwright scripts, and schemas).

---

## Development

### 1. Install Dependencies
```bash
npm install
```

### 2. Start Live Development Server
```bash
# Start frontend only (in browser)
npm run dev

# Launch native Tauri desktop window with live reload
npm run tauri dev
```

### 3. Build & Package (Monolithic Installer)
To package the desktop application with all embedded dependencies:

- **Windows**:
  ```powershell
  ..\desktop\build-bundle.ps1 -SkipDownload
  ```
- **Linux**:
  ```bash
  chmod +x ./build-bundle.sh && ./build-bundle.sh
  ```

---

## Recommended VS Code Setup

- [Tauri Extension](https://marketplace.visualstudio.com/items?itemName=tauri-apps.tauri-vscode)
- [rust-analyzer](https://marketplace.visualstudio.com/items?itemName=rust-lang.rust-analyzer)
- [Tailwind CSS IntelliSense](https://marketplace.visualstudio.com/items?itemName=bradlc.vscode-tailwindcss)
- [Biome](https://marketplace.visualstudio.com/items?itemName=biomejs.biome)
