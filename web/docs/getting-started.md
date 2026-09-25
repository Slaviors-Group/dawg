---
description: "Capture, sanitize, inspect, export, import, and replay your first portable DAWG bug-reproduction artifact."
---

# Getting Started

DAWG records one browser tab, sanitizes the captured data, and packages the session as a portable OCI artifact. The recommended workflow uses the desktop app together with the DAWG Browser Extension.

> **Current prerelease:** [`0.3.5-middlechild`](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-5) — desktop `0.3.5`; extension display version `0.3.5_middlechild`.

## Quick Start: Desktop App

### 1. Install DAWG

Download the `middlechild-5` release directly:

- [Windows x64 installer](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_x64-setup.exe)
- [Linux x86_64 AppImage](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_amd64.AppImage)

See [Installation](/docs/installation) for Linux AppImage launch steps and the
[GitHub release](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-5)
for checksums and release notes.

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
4. Leave **Safe** selected for bounded standard diagnostics, or select
   **Enhanced Diagnostics** and confirm its warning when additional CDP evidence
   is appropriate.
5. Select **Start Capture**. DAWG focuses an existing tab with the exact
   normalized URL or opens a new tab. Reproduce the issue in that tab.
6. Select **Stop Capture**. Wait while the engine drains extension events, sanitizes the session, packages the OCI layout, and registers it in the local catalog.

The extension records:

- rrweb DOM snapshots and incremental events, with DOM input fields masked;
- click and input actions;
- bounded console, error, and network metadata from the Safe profile;
- CDP console, exception, and network evidence when consented Enhanced
  Diagnostics is available;
- eligible text or structured response bodies only in Enhanced, subject to
  content-type, size, concurrency, and aggregate limits.

Safe does not retrieve response bodies through CDP. Enhanced records a body
state—such as `captured`, `redacted`, `truncated`, `blocked`, `unavailable`, or
`capture-failed`—when it cannot retain an eligible body. If CDP cannot attach,
DAWG records a degradation and continues in Safe mode. Input action values and
request metadata reach the local engine before sanitization, so inspect an
artifact before sharing it.

### 5. Inspect and Manage Artifacts

The dashboard refreshes after a successful capture. Under **Recent Artifacts**, you can:

- select **Inspect** to read the manifest;
- select **Export** to create a portable `.dawg` archive;
- select **Import** to choose an existing `.dawg` archive;
- drop a `.dawg` archive onto the dashboard to import it.

DAWG stores validated OCI artifact directories under `~/.dawg/artifacts/` and keeps their captured, imported, or legacy origin in `~/.dawg/artifact-catalog.json`. Untitled captures receive a hostname-based title, and artifact directories use readable timestamped names with numeric suffixes for collisions.

A `.dawg` file is a ZIP-based transport archive. Import validates its paths, links, size limits, compression ratio, OCI descriptors, digests, and DAWG manifest before publishing it to the artifact store. See [Artifact Anatomy](/docs/artifact-anatomy) for the OCI layout, evidence layers, integrity checks, and reviewed-artifact behavior.

### 6. Replay

Open **Replay Engine**, search or filter the catalog, and select an artifact.
The Replay Diagnostics workspace shows packaged console, network, and error
evidence, including evidence states. After review confirmation, it can export a
sanitized HAR or copy cURL for a selected request; copied cURL can mutate a live
service. Choose **Run Replay** to open an interactive Chromium replay with
play/pause, skip, speed, and timeline controls synchronized with the Desktop
player. Console, error, and network evidence appears as replay reaches its
recorded time; **All captured** keeps the full review available. **Stop Replay** cancels the
in-flight replay and terminates the tracked engine/browser process tree. The
standard CLI replay remains the noninteractive option that saves a final
screenshot.

On Windows, browser replay runs in native compatibility mode and skips Docker Compose isolation and database restoration. Supported non-Windows environment replay uses rootless Docker.

---

## Quick Start: Engine CLI

The CLI uses the same extension-driven capture and artifact catalog as the desktop app. Ensure the extension is loaded and `dawg doctor` reports the required runtime resources before starting.

```bash
# Start an extension-driven capture
dawg capture --url https://example.test --title "Checkout timeout" --diagnostics-profile safe

# Reproduce the issue in the selected browser tab, then stop and package
dawg capture stop

# Discover validated artifacts in the local catalog
dawg artifacts list

# Inspect and replay an artifact directory
dawg inspect <artifact-directory>
dawg diagnostics inspect <artifact-directory>
dawg diagnostics export-har <artifact-directory> --output evidence.har
dawg diagnostics copy-curl <artifact-directory> --request-id <request-id>
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

Replay uses the bundled Playwright Chromium runtime to reconstruct the rrweb
session. Diagnostic evidence supports investigation but is not injected into or
used to re-execute the original application. If the artifact includes
environment, database, or cassette layers, the engine restores or serves them
where the host platform supports those operations.

## Next Steps

- [Installation](/docs/installation) — Release installation, source setup, and bundle requirements
- [Architecture](/docs/architecture) — How capture, sanitization, packaging, and replay fit together
- [Artifact Anatomy](/docs/artifact-anatomy) — OCI layout, evidence layers, and artifact integrity
- [Sanitizer Policy](/docs/sanitizer-policy) — Review the data-handling policy before sharing artifacts
- [Contributing](/docs/contributing) — Development, validation, and versioning commands
