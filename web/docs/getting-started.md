# Getting Started

DAWG records one browser tab, sanitizes the captured data, and packages the session as a portable OCI artifact. The recommended workflow uses the desktop app together with the DAWG Browser Extension.

> **Current prerelease:** [`0.2.7-naughty`](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-4) (`naughty-4`) — desktop `0.2.7`; extension display version `0.2.5_naughty`.

## Quick Start: Desktop App

### 1. Install DAWG

[Download DAWG 0.2.7 for Windows x64](https://github.com/Slaviors-Group/dawg/releases/download/naughty-4/DAWG_0.2.7_x64-setup.exe), or visit the [0.2.7-naughty GitHub release](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-4) for other assets.

When an AppImage is not attached to the release, see [Installation](/docs/installation) to build it from source.

The packaged desktop app includes the engine, Node.js, mitmdump, Playwright, Chromium for replay, schemas, policies, and an installable copy of the extension. You do not need to install those runtimes separately.

### 2. Install the Browser Extension

Capture is extension-driven. Chrome or Chromium **116 or newer** is required.

Install the [DAWG Browser Extension from the Chrome Web Store](https://chromewebstore.google.com/detail/peiigoeakholhhbbbbfkeojomekmmokj?utm_source=item-share-cb). This is the recommended installation method.

For source development or when Chrome Web Store installation is unavailable, use an unpacked extension instead:

1. Open `chrome://extensions`.
2. Enable **Developer mode**.
3. Select **Load unpacked**.
4. In DAWG, open **Engine Doctor**, find the `browser-extension` component, and select the parent directory of its displayed `manifest.json` path. When running from source, select the repository's `extension/` directory.
5. Reload the unpacked extension after updating DAWG or changing extension files.

The extension cannot start a capture by itself. Start from the desktop app or CLI so the local engine can issue a session token and select the target tab.

### 3. Check the Runtime

Launch DAWG and check the engine status on the dashboard. The bundled installation should report the engine as **Ready** and **bundled**. The diagnostics view checks the engine, Node.js, mitmdump, Playwright Chromium, replay script, schema, policy, and bundled extension manifest.

If you are using the CLI directly, run:

```bash
dawg doctor
```

`dawg doctor` confirms that an extension manifest is available to install; it cannot confirm that the extension is enabled in your browser.

### 4. Capture a Bug

1. Open **Capture** in the desktop app.
2. Enter an `http://` or `https://` target URL.
3. Optionally enter an artifact name.
4. Select **Start Capture**.
5. DAWG focuses an existing tab with the exact normalized URL or opens a new tab. Reproduce the issue in that tab.
6. Select **Stop Capture**. Wait while the engine drains extension events, sanitizes the session, packages the OCI layout, and registers it in the local catalog.

The extension records:

- rrweb DOM snapshots and incremental events, with DOM input fields masked;
- click and input actions;
- request and response metadata available through Chrome's `webRequest` API;
- request bodies when Chrome exposes them.

Response bodies are not captured. Input action values and request metadata reach the local engine before sanitization, so inspect an artifact before sharing it.

### 5. Inspect and Manage Artifacts

The dashboard refreshes after a successful capture. Under **Recent Artifacts**, you can:

- select **Inspect** to read the manifest;
- select **Export** to create a portable `.dawg` archive;
- select **Import** to choose an existing `.dawg` archive;
- drop a `.dawg` archive onto the dashboard to import it.

DAWG stores validated OCI artifact directories under `~/.dawg/artifacts/` and keeps their captured, imported, or legacy origin in `~/.dawg/artifact-catalog.json`. Untitled captures receive a hostname-based title, and artifact directories use readable timestamped names with numeric suffixes for collisions.

A `.dawg` file is a ZIP-based transport archive. Import validates its paths, links, size limits, compression ratio, OCI descriptors, digests, and DAWG manifest before publishing it to the artifact store.

### 6. Replay

Open **Replay Engine**, search or filter the catalog, select an artifact, and choose **Run Replay**. DAWG opens an interactive Chromium replay with play/pause, skip, speed, and timeline controls. **Stop Replay** cancels the in-flight replay and terminates the tracked engine/browser process tree. The standard CLI replay remains the noninteractive option that saves a final screenshot.

On Windows, browser replay runs in native compatibility mode and skips Docker Compose isolation and database restoration. Supported non-Windows environment replay uses rootless Docker.

---

## Quick Start: Engine CLI

The CLI uses the same extension-driven capture and artifact catalog as the desktop app. Ensure the extension is loaded and `dawg doctor` reports the required runtime resources before starting.

```bash
# Start an extension-driven capture
dawg capture --url https://example.test --title "Checkout timeout"

# Reproduce the issue in the selected browser tab, then stop and package
dawg capture stop

# Discover validated artifacts in the local catalog
dawg artifacts list

# Inspect and replay an artifact directory
dawg inspect <artifact-directory>
dawg run <artifact-directory>

# Export it for another DAWG installation
dawg artifacts export <artifact-directory> --output checkout-timeout.dawg

# Import a portable archive into the local store
dawg artifacts import checkout-timeout.dawg
```

Use `--force` with `dawg artifacts export` only when you intend to replace an existing output file.

The CLI also supports OCI registry transport:

```bash
dawg push <artifact-directory> --registry <registry-reference>
dawg pull <registry-reference> --output <directory>
```

`dawg verify <artifact-directory> --against <value>` replays the artifact and labels the generated report with `<value>`. It does **not** check out or launch that branch, commit, or path.

---

## What Happens During Capture?

When capture starts, DAWG:

1. starts a local capture daemon and creates a token-bound session;
2. asks the extension to focus or open the target tab;
3. receives rrweb, action, and frontend HTTP event streams from that tab;
4. optionally includes supplied Compose, database-diff, and structured-log inputs;
5. sanitizes supported JSONL streams with secret/PII rules and the selected OPA policy;
6. packages the result as a digest-addressed OCI Image Layout;
7. registers the validated artifact in the persistent local catalog.

Replay uses the bundled Playwright Chromium runtime to reconstruct the rrweb session. If the artifact includes environment, database, or cassette layers, the engine restores or serves them where the host platform supports those operations.

## Next Steps

- [Installation](/docs/installation) — Release installation, source setup, and bundle requirements
- [Architecture](/docs/architecture) — How capture, sanitization, packaging, and replay fit together
- [Sanitizer Policy](/docs/sanitizer-policy) — Review the data-handling policy before sharing artifacts
- [Contributing](/docs/contributing) — Development, validation, and versioning commands
