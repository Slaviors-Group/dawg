# Getting Started

DAWG packages sanitized, deterministic web-app bug reproductions into portable OCI artifacts. This guide walks you through your first capture and replay.

## Quick Start: Desktop App

The fastest way to get started is with the desktop application.

### 1. Download & Install

Download the installer from the release page:

| Platform | Installer | Download |
|---|---|---|
| **Windows 10/11 (x64)** | `DAWG_0.1.2-alpha_x64-setup.exe` (NSIS) | [v0.1.2-alpha (latest)](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-3) |
| **Linux** | Build from source | See [Installation](/docs/installation) |

Older releases: [v0.1.1-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-2) · [v0.1.0-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha)

### 2. Launch & Verify

Launch DAWG. The header will display **Engine Ready [Bundled]** with all runtimes automatically initialized.

### 3. Capture a Bug

1. Enter your web application target URL (e.g., `https://example.com` or `http://localhost:3000`)
2. Click **Start Capture**
3. Reproduce the issue in the browser window
4. Click **Stop Capture** to generate your reproduction artifact

### 4. Inspect Your Artifact

Your packaged OCI artifact is stored at `.dawg/artifacts/<session-id>`. You can view it in the Desktop UI or inspect it via CLI:

```bash
dawg inspect .dawg/artifacts/<session-id>
```

---

## Quick Start: Engine CLI

For developers who prefer the command line or want to integrate DAWG into CI/CD pipelines.

### Prerequisites

- **Go 1.22+**
- **Node.js 20+** (with Playwright)
- **mitmproxy** (`mitmdump` on PATH)

### Step-by-step

```bash
# 1. Initialize DAWG in your project
dawg init

# 2. Verify your environment
dawg doctor

# 3. Start capturing a web-app session
dawg capture --url https://my-app.local

# 4. Reproduce the bug in the opened browser, then stop
dawg capture stop

# 5. Inspect the generated artifact
dawg inspect .dawg/artifacts/<session-id>

# 6. Replay the artifact to verify it reproduces the bug
dawg run .dawg/artifacts/<session-id>

# 7. Verify the fix against your local codebase
dawg verify .dawg/artifacts/<session-id> --against local
```

---

## What Just Happened?

When you run `dawg capture`, DAWG:

1. **Launches a real browser** via Playwright with rrweb DOM recording injected
2. **Starts a proxy** (mitmproxy) to capture all HTTP traffic between frontend and backend
3. **Records everything**: DOM mutations, browser actions, network requests/responses
4. **Sanitizes** the captured data using OPA policies to remove PII and secrets
5. **Packages** everything into a standard OCI Image Layout artifact

When you run `dawg run`, DAWG:

1. **Spins up** a sandboxed environment matching the original (via Docker Compose)
2. **Restores** any database fixtures captured during the session
3. **Replays** the browser events and HTTP interactions in deterministic order
4. **Freezes** non-determinism (time, random seeds, UUIDs)

---

## Next Steps

- [Installation](/docs/installation) — Detailed setup for all platforms
- [CLI Reference](/docs/cli-reference) — All available commands and flags
- [Architecture](/docs/architecture) — How the pipeline works under the hood
