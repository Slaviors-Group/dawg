# CLI Reference

This reference describes the DAWG `0.2.3-naughty` command-line interface.

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

The launcher creates a session and starts a detached capture daemon. The daemon listens on `127.0.0.1:8082`; the DAWG Manifest V3 browser extension connects to it and records the selected tab. The extension must be installed and available to complete capture.

The session can contain:

- rrweb events in `traces/rrweb.jsonl`
- click and fill actions in `actions/browser.jsonl`
- frontend request/response metadata in `http/frontend.jsonl`
- an optional Compose snapshot in `env/compose.yaml` and `env/lockfile.json`
- optional database changes in `db/diff.jsonl`
- optional structured logs in `logs/structured.jsonl`

Frontend network records include request headers and bodies plus response status and headers. Response bodies are currently empty.

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

### `dawg run <artifact-directory>`

Replay an artifact and write its replay outcome.

```bash
dawg run <artifact-directory> [--output text|json]
```

Replay unpacks the OCI layers, optionally starts a captured Compose environment and restores a PostgreSQL fixture, optionally starts cassette replay, and uses Playwright Chromium to render the recorded rrweb timeline. It produces `outcome/screenshot.png` and diagnostics.

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
- the `0.2.3-naughty` manifest schema

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
schemaVersion: "0.2.3-naughty"
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
