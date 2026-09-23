# CLI Reference

This reference describes the DAWG `0.3.3-middlechild` command-line interface.

## Usage

```bash
dawg [--output text|json] <command> [flags]
dawg --version
```

The global `--output` format defaults to `text`. `doctor` supports text and JSON output. `run` and `verify` define their own `--output` flags. `inspect` always prints formatted JSON, and the `artifacts` subcommands always print JSON.

## Commands

### `dawg init [directory]`

Create `dawg.config.yaml` in the current directory, or in an existing directory supplied as the argument.

```bash
dawg init [directory] [--force]
```

| Flag | Description |
|---|---|
| `--force` | Overwrite an existing `dawg.config.yaml` |

The command also creates the state capture directory and writes a default policy to the state policy directory. See [Configuration](#configuration) for the generated files and an important runtime caveat.

### `dawg capture`

Start an extension-driven capture session.

```bash
dawg capture --url <http-or-https-url> [flags]
```

| Flag | Required | Description |
|---|---:|---|
| `--url` | Yes | Exact HTTP(S) target URL for the browser extension |
| `--session-dir <directory>` | No | Capture session root; defaults to `~/.dawg/captures` |
| `--compose-file <file>` | No | Docker Compose file to snapshot |
| `--db-diff-file <file>` | No | Normalized ORM database-diff JSONL input |
| `--log-file <file>` | No | Structured JSON log file to tail |
| `--policy-file <file>` | No | Rego policy used by the sanitizer |
| `--unsafe-skip-sanitize` | No | Skip sanitization; accepted only for localhost targets |
| `--title <title>` | No | Artifact title |
| `--diagnostics-profile <safe\|enhanced>` | No | Diagnostic evidence profile; defaults to `safe` |

The launcher creates a session and starts a detached capture daemon. The daemon listens on `127.0.0.1:8082`; the DAWG Manifest V3 browser extension connects to it and records the selected tab. The extension must be installed and available to complete capture.

The session can contain:

- rrweb events in `traces/rrweb.jsonl`
- click and fill actions in `actions/browser.jsonl`
- frontend request/response metadata in `http/frontend.jsonl`
- an optional Compose snapshot in `env/compose.yaml` and `env/lockfile.json`
- optional database changes in `db/diff.jsonl`
- optional structured logs in `logs/structured.jsonl`

### Diagnostic profiles

`safe` is the default profile. It records bounded console/error and network
metadata from the selected tab without CDP body retrieval. Network body fields
state `not-requested` or `unavailable` when a body is not retained.

`enhanced` requests Chrome DevTools Protocol (CDP) diagnostics. It captures CDP
console APIs, exceptions, and network records, and may retain response bodies
only for eligible JSON, GraphQL, form-encoded, or text content within configured
size and concurrency limits. CDP attach failure falls back to Safe and records a
capture degradation; a later CDP detachment is also recorded. Enhanced capture
is an explicit opt-in in Desktop; CLI users select it with the flag above.

Diagnostic fields use evidence states: `captured`, `redacted`, `preview-only`,
`truncated`, `blocked`, `unavailable`, `not-requested`, and `capture-failed`.
Those states express what is retained; the CLI does not reconstruct an absent
original value.

### `dawg capture stop`

Ask the active browser extension to drain its buffered events, then sanitize, package, and register the capture.

```bash
dawg capture stop [--control-file <file>]
```

| Flag | Description |
|---|---|
| `--control-file <file>` | Capture control file; defaults to `~/.dawg/capture-control.json` |

A successful stop creates an OCI Image Layout artifact under the local artifact directory.

### `dawg inspect <artifact-directory>`

Validate an artifact and print its DAWG manifest as formatted JSON.

```bash
dawg inspect <artifact-directory>
```

The manifest includes the schema version, artifact identity, source metadata, layer descriptors, sanitization metadata, determinism metadata, and expected outcome.

### `dawg push <artifact-directory>`

Push a local OCI artifact to an OCI-compatible registry.

```bash
dawg push <artifact-directory> --registry <registry-reference>
```

| Flag | Required | Description |
|---|---:|---|
| `--registry <registry-reference>` | Yes | Destination OCI registry reference |

### `dawg pull <registry-reference>`

Pull an OCI artifact into a local directory.

```bash
dawg pull <registry-reference> --output <directory>
```

| Flag | Required | Description |
|---|---:|---|
| `--output <directory>` | Yes | Destination directory |

### `dawg artifacts list`

List validated OCI artifacts in the local artifact directory.

```bash
dawg artifacts list
```

The JSON result is backed by the local catalog, which can contain captured and imported artifacts and discovered legacy artifacts.

### `dawg artifacts export <artifact-directory>`

Export a local OCI artifact as a portable ZIP-based `.dawg` archive.

```bash
dawg artifacts export <artifact-directory> --output <file.dawg> [--force]
```

| Flag | Required | Description |
|---|---:|---|
| `--output <file.dawg>` | Yes | Destination archive |
| `--force` | No | Overwrite an existing destination |

### `dawg artifacts import <file.dawg>`

Import a portable `.dawg` archive into the local artifact store.

```bash
dawg artifacts import <file.dawg>
```

Import validates archive paths, symlinks, size and compression limits, OCI structure, descriptors, and content digests before accepting the artifact.

### `dawg diagnostics inspect <artifact-directory>`

Print the packaged, sanitized diagnostic evidence as JSON.

```bash
dawg diagnostics inspect <artifact-directory>
```

Legacy artifacts return empty diagnostic evidence. The command verifies declared
diagnostic blobs before reading their bounded entries.

### `dawg diagnostics export-har <artifact-directory>`

Write a sanitized HAR representation of retained network evidence.

```bash
dawg diagnostics export-har <artifact-directory> --output evidence.har
```

| Flag | Required | Description |
|---|---:|---|
| `--output <file>` | Yes | Destination HAR file |

The HAR includes retained request/response metadata and body state annotations;
it does not reconstruct empty, blocked, unavailable, or truncated bodies.

### `dawg diagnostics copy-curl <artifact-directory>`

Print a cURL command for one retained network request.

```bash
dawg diagnostics copy-curl <artifact-directory> --request-id <request-id>
```

| Flag | Required | Description |
|---|---:|---|
| `--request-id <id>` | Yes | Captured request ID |

The generated command omits sensitive authentication and cookie headers and
includes a retained request body only when it is available. Review it before
running: the request may mutate a live service.

### `dawg diagnostics remove <artifact-directory>`

Create a new, independently content-addressed OCI artifact without selected
sanitized diagnostic evidence. The source artifact is validated but never
modified; the reviewed copy has new digests and does not carry provenance from
the source.

```bash
dawg diagnostics remove <artifact-directory> \
  --output-dir <new-artifact-directory> \
  --remove-category console \
  --remove-body-ref diagnostics/bodies/response_request-id.json
```

| Flag | Required | Description |
|---|---:|---|
| `--output-dir <directory>` | Yes | New, non-existent artifact directory |
| `--remove-category <category>` | No | Repeatable: `console`, `network`, `errors`, or `bodies` |
| `--remove-body-ref <path>` | No | Repeatable retained body path below `diagnostics/bodies/` |

At least one category or body reference is required. Removing `network` also
removes retained bodies because they have no remaining request context. Removing
individual bodies rewrites matching network body states to `unavailable`; DAWG
does not infer replacement content.

### `dawg run <artifact-directory>`

Replay an artifact and write its replay outcome.

```bash
dawg run <artifact-directory> [--interactive] [--output text|json]
```

| Flag | Description |
|---|---|
| `--interactive` | Keep headed Chromium open with play/pause, skip, speed, and timeline controls. Desktop Replay uses this mode. |

Replay unpacks the OCI layers, optionally starts a captured Compose environment and restores a PostgreSQL fixture, optionally starts cassette replay, and uses Playwright Chromium to render the recorded rrweb timeline. Standard mode plays to completion, writes `outcome/screenshot.png`, and exits. Interactive mode remains open until Chromium is closed or the Desktop **Stop Replay** action terminates it; it does not produce the final screenshot. Press `Ctrl+Shift+I` to inspect the reconstructed rrweb document and `window.__DAWG_REPLAY__`. Refreshing the local replay page restarts it at the beginning.

`run` visualizes the recording; it does not re-execute recorded click/fill actions or navigate through the original application flow. See [Architecture](./architecture.md#replay) for platform and determinism boundaries.

### `dawg verify <artifact-directory>`

Replay an artifact and compare the resulting outcome with the recorded expectations.

```bash
dawg verify <artifact-directory> [--against <label>] [--output text|json]
```

| Flag | Default | Description |
|---|---|---|
| `--against <label>` | `local` | Label stored as `verifiedAgainst` in the result |
| `--output text|json` | `text` | Result format |

`--against` does not check out or select a branch, commit, directory, or other code target. Verification always checks a successful replay exit code. When available, it also compares ordered HTTP response status codes and performs an exact per-channel screenshot comparison with a tolerance of at most 100 differing pixels.

A verification result of `fail` means the recorded bug still reproduces and exits with status `1`.

### `dawg doctor`

Inspect the installed runtime and bundled assets.

```bash
dawg doctor [--output text|json]
```

The report checks:

- the engine
- `mitmdump`
- Node.js
- Playwright Chromium
- `replay-browser.cjs`
- the browser extension manifest
- the default Rego policy
- the current `0.3.3-middlechild` manifest schema

The report status is `ready` or `degraded`. A degraded component is recorded in the report; degradation alone does not currently make the command return an execution error.

## Exit Codes

| Code | Meaning |
|---:|---|
| `0` | Command completed successfully |
| `1` | `verify` concluded that the reproduction still fails |
| `2` | Cobra parsing, validation, or command execution error |

## State Directories

DAWG resolves its state root in this order:

1. `DAWG_STATE_DIR`, when set
2. `~/.dawg`
3. `.dawg` if the home directory cannot be resolved

The default layout under that root is:

```text
.dawg/
├── artifact-catalog.json
├── artifacts/
├── capture-control.json
├── capture-results/
├── captures/
└── policies/
```

Captured artifacts are stored under `artifacts/<timestamped-title>`.

## Configuration

`dawg init` generates a configuration with the current schema version and absolute state paths:

```yaml
schemaVersion: "0.3.3-middlechild"
capture:
  outputDir: "<absolute-state-directory>/captures"
  browser: "chromium"
sanitize:
  policyFile: "<absolute-state-directory>/policies/default.rego"
replay:
  composeFile: "compose.yaml"
```

It also generates this policy:

```text
package dawg.sanitizer

default allow = true
```

::: warning Current configuration boundary
The current CLI commands do not load `dawg.config.yaml`. They use command flags and runtime defaults directly. Treat the generated file as project metadata rather than an active command configuration source.
:::
