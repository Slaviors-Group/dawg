# CLI Reference

The DAWG CLI provides headless access to all capture, replay, and verification features.

## Usage

```bash
dawg <command> [flags]
```

## Commands

### `dawg init`

Generate a `dawg.config.yaml` in the current project directory.

```bash
dawg init
```

Creates a default configuration file with sanitization policies and capture settings.

---

### `dawg doctor`

Run built-in diagnostics to verify runtime readiness and bundled assets.

```bash
dawg doctor
```

**Output:** Human-readable diagnostic report showing status of:
- Engine binary version
- mitmproxy availability
- Node.js and Playwright installation
- Bundled scripts (capture-browser, capture-proxy, replay-browser)
- OPA policies and JSON schemas

---

### `dawg capture`

Start a web-app session capture.

```bash
dawg capture --url <target-url>
```

**Flags:**

| Flag | Description |
|---|---|
| `--url` | Target web application URL to capture |

**What it does:**
1. Launches a Playwright-controlled browser with rrweb DOM recording
2. Starts mitmproxy to capture backend HTTP traffic
3. Records all DOM mutations, browser actions, and network requests
4. Waits for `dawg capture stop` to finalize

---

### `dawg capture stop`

Stop an active capture session and trigger automatic sanitization + packaging.

```bash
dawg capture stop
```

**What it does:**
1. Stops the browser, proxy, and all capture agents
2. Runs OPA-based sanitization on captured data
3. Packages into OCI Image Layout artifact
4. Stores artifact at `.dawg/artifacts/<session-id>`

---

### `dawg inspect`

Display an artifact's manifest without replaying it.

```bash
dawg inspect <artifact-path>
```

**Example:**

```bash
dawg inspect .dawg/artifacts/abc123
```

**Output:** JSON manifest showing:
- Schema version, artifact ID, creation timestamp
- Layer list with media types, digests, and sizes
- Sanitization report (fields redacted, policy version)
- Determinism config (frozen time, random seed)
- Provenance metadata

---

### `dawg push`

Push an artifact to an OCI-compatible registry.

```bash
dawg push <artifact-path> --registry <registry-url>
```

**Flags:**

| Flag | Description |
|---|---|
| `--registry` | OCI registry URL (e.g., `ghcr.io/org/repo`) |

---

### `dawg pull`

Pull an artifact from an OCI-compatible registry.

```bash
dawg pull <artifact-ref> --registry <registry-url>
```

**Flags:**

| Flag | Description |
|---|---|
| `--registry` | OCI registry URL to pull from |

---

### `dawg run`

Replay an artifact in a sandboxed environment.

```bash
dawg run <artifact-path>
```

**What it does:**
1. Verifies sandbox is available (Docker required)
2. Spins up environment from the artifact's lockfile
3. Restores database fixtures
4. Replays browser events and HTTP interactions
5. Returns exit code: `0` if replay succeeded, non-zero if bug reproduced

---

### `dawg verify`

Verify a reproduction artifact against your local codebase.

```bash
dawg verify <artifact-path> --against <source>
```

**Flags:**

| Flag | Description |
|---|---|
| `--against` | Comparison target: `local` (current branch) or a specific path |

**What it does:**
1. Replays the artifact against the specified codebase
2. Compares outcomes (screenshots, HTTP responses, exit codes)
3. Generates a diff report showing pass/fail per check
4. Returns exit code: `0` if verification passed

---

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | General error |
| `2` | Sandbox not available |
| `3` | Sanitization policy violation (export blocked) |
| `4` | Artifact not found or invalid |
| `5` | Registry authentication failed |

---

## Configuration

DAWG looks for configuration in this order:

1. `./dawg.config.yaml` (current directory)
2. `~/.config/dawg/config.yaml` (user home)
3. Bundled defaults (in standalone/bundled mode)

### Sample `dawg.config.yaml`

```yaml
capture:
  browser: chromium
  timeout: 300s
  
sanitize:
  policy: default.rego
  deny_unmatched: true
  
packaging:
  compression: zstd
  format: oci-image-layout
  
replay:
  sandbox: docker
  rootless: true
```
