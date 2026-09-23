# Architecture

DAWG `0.3.3-middlechild` captures a browser session through an extension, sanitizes supported JSONL streams and retained diagnostic bodies, packages them as an OCI Image Layout, and renders the recording for replay and verification.

## Pipeline Overview

```mermaid
flowchart TD
    A[CLI capture launcher] --> B[Detached capture daemon]
    B --> C[Manifest V3 browser extension]
    C --> D[Session files]
    E[Optional Compose DB diff and logs] --> D
    D --> F[Heuristic sanitizer]
    F --> G[OPA allow gate]
    G --> H[OCI Image Layout]
    H --> I[Local catalog or OCI registry]
    H --> J[Replay renderer]
    J --> K[Verification checks]
```

| Stage | Main responsibility |
|---|---|
| Capture | Record rrweb events, browser actions, Safe or Enhanced diagnostic evidence, and optional environment inputs |
| Sanitize | Replace recognized secrets and PII in supported JSONL files and retained diagnostic bodies, then evaluate a Rego allow decision |
| Package | Validate metadata and create content-addressed OCI layers, including bounded diagnostics when present |
| Store | Register locally, export/import a `.dawg` archive, or push/pull through an OCI registry |
| Replay | Restore available environment data and render the rrweb timeline in Playwright Chromium |
| Verify | Compare replay exit status and any available HTTP and screenshot outcomes |

## Capture

The current `dawg capture` path is extension-driven. Although the source tree contains other browser and proxy helpers, the CLI does not instantiate them for capture.

1. The short-lived CLI launcher creates a session and starts a detached engine daemon.
2. The daemon starts the extension server on `127.0.0.1:8082`.
3. The Manifest V3 extension connects with a token-bearing WebSocket handshake.
4. The extension selects a tab whose HTTP(S) URL exactly matches `--url`, or opens one.
5. `dawg capture stop` asks the extension to drain buffered data before finalization.
6. The daemon sanitizes the session, packages it, and adds it to the local catalog.

The extension requires Chrome 116 or newer and requests access to `<all_urls>`. It records only requests associated with the selected tab. Its `debugger` permission is used only when an Enhanced capture is requested.

### Session Data

| Path | Contents |
|---|---|
| `traces/rrweb.jsonl` | rrweb DOM and interaction events |
| `actions/browser.jsonl` | Separate click and fill action records |
| `http/frontend.jsonl` | Frontend request headers/body and response status/headers |
| `env/compose.yaml` | Optional Compose snapshot |
| `env/lockfile.json` | Optional resolved environment lock data |
| `db/diff.jsonl` | Optional normalized ORM database-diff stream |
| `logs/structured.jsonl` | Optional structured JSON logs |
| `diagnostics/console.jsonl` | Bounded console evidence |
| `diagnostics/network.jsonl` | Bounded network evidence with per-body evidence states |
| `diagnostics/errors.jsonl` | Browser errors and capture degradations |
| `diagnostics/bodies/` | Sanitized bodies retained by Enhanced capture only |

rrweb runs with `maskAllInputs: true`. The separate action stream can still contain entered values, so sanitization includes both streams.

### Diagnostic evidence fidelity

**Safe** is the default profile. It collects bounded main-world/extension console,
error, and network metadata without using CDP to retrieve response bodies.

**Enhanced** is a consented CDP profile. It adds CDP console, exception, and
network records and requests response bodies only when their MIME type is JSON,
GraphQL, form-encoded, or text. Per-body, aggregate-body, record, and layer
limits bound what can be retained. CDP cannot attach or may detach during a
session; these conditions are recorded as degradations, and initial attach
failure continues in Safe mode.

Every diagnostic value can state `captured`, `redacted`, `preview-only`,
`truncated`, `blocked`, `unavailable`, `not-requested`, or `capture-failed`.
These are fidelity markers, not recovery mechanisms: DAWG does not reconstruct
body content that was not retained.

## Sanitization and Policy Gate

The sanitizer rewrites a fixed set of JSONL files in place through temporary files and sanitizes retained diagnostic bodies before packaging. It uses field-path and value heuristics for secrets, PII, and payment-card data. Email addresses, phone numbers, and names receive deterministic SHA-256-derived synthetic values; other detections become `[REDACTED]`.

After rewriting, the engine evaluates `data.dawg.sanitizer.allow`. The policy receives only redaction counters and blocked-field names—not the complete capture. Packaging requires an allowed `sanitize-report.json`.

This process reduces accidental exposure but does not establish that an artifact is free of confidential data. See [Sanitizer Policy](./sanitizer-policy.md) for the exact detection and policy boundaries.

## OCI Packaging

A packaged artifact is an OCI Image Layout:

```text
<artifact-directory>/
├── dawg-manifest.json
├── oci-layout
├── index.json
└── blobs/
    └── sha256/
        └── <digest>
```

The DAWG manifest is validated against the schema matching its declared version. DAWG `0.3.3-middlechild` requires diagnostic summary metadata in addition to the schema version, SHA-256 artifact ID, creation time, non-empty title, source, at least one layer, sanitization metadata, determinism metadata, and expected outcome. The summary identifies the selected profile, sources, counts, retained/redacted/blocked/truncated body totals, limits, and any degradations.

### Current Layer Types

| Concern | Media type | Packaging |
|---|---|---|
| Environment | `application/vnd.dawg.env.lockfile+tar` | Uncompressed tar |
| Database | `application/vnd.dawg.db.fixture+tar.zstd` | zstd-compressed tar |
| Trace | `application/vnd.dawg.trace.rrweb+jsonl.zstd` | zstd-compressed tar |
| Cassette | `application/vnd.dawg.cassette+tar.zstd` | zstd-compressed tar |
| Diagnostics | `application/vnd.dawg.diagnostics+tar.zstd` | zstd-compressed tar of console, network, and error records |
| Diagnostic bodies | `application/vnd.dawg.diagnostic-bodies+tar.zstd` | zstd-compressed tar of sanitized retained bodies |

The trace layer groups `traces/`, `actions/`, `http/`, and `logs/`. Descriptors and blobs use SHA-256 content addressing.

Artifacts default to the state directory at `~/.dawg/artifacts/<timestamped-title>` unless `DAWG_STATE_DIR` changes the state root. The local catalog distinguishes captured and imported artifacts and can discover legacy artifacts. Portable `.dawg` exports are ZIP archives containing the OCI layout.

Import applies archive traversal, symlink, entry-count, size, compression-ratio, OCI-structure, descriptor, and digest validation before registration.

## Replay

Replay is a rendering pipeline rather than action re-execution:

1. Validate and unpack the OCI layers into a temporary directory.
2. Read the DAWG manifest.
3. Export `FAKETIME=@<clockFrozenAt>` when the recorded value is nonzero.
4. On supported non-Windows hosts, start `env/compose.yaml` with Docker Compose.
5. Restore `db/fixture.sql`, when present, into `<project>-db-1` with `pg_restore --data-only --no-owner --no-privileges -d postgres`.
6. Start `mitmdump` server replay with `--server-replay-kill-extra` when `cassettes/cassette.yaml` exists.
7. Use `engine/scripts/replay-browser.cjs` and Playwright Chromium to reconstruct the rrweb DOM timeline in a local replay page.
8. In standard CLI and verification mode, wait for the trace duration plus two seconds, then write `outcome/screenshot.png` and diagnostics. In Desktop interactive mode, keep Chromium open with in-page playback controls until the user closes it or selects **Stop Replay**.

Interactive replay renders the recorded viewport in a fixed stage and scales or letterboxes that stage when the Chromium window is resized or maximized. `FAKETIME` only affects a target environment that consumes it. A random seed is represented in the manifest types but is not applied during replay.

Recorded action events are not executed, and replay does not navigate or run through the original application's workflow. It visualizes the captured rrweb recording; diagnostic evidence is available for inspection and export, not replayed into the original application.

Desktop derives an **approximate** replay offset from a diagnostic's capture
millisecond timestamp (network records prefer `timing.startedAt`) and the first
rrweb timestamp. It only offers a seek when the value falls within the recorded
rrweb range. A newline-delimited local control message is forwarded to the
interactive player, which clamps the offset before seeking. Missing or
out-of-range timestamps deliberately show no correlation.

## Evidence review and removal

`dawg diagnostics remove` validates the complete source OCI layout, unpacks it
into a private staging directory, removes requested complete categories or
individual retained body files, and packages a fresh OCI layout. This leaves the
source unchanged, recomputes descriptors and logical artifact identity, and
drops any source provenance rather than implying it still attests to modified
evidence. Removing a body updates any matching network reference to the explicit
`unavailable` state; DAWG never reconstructs removed content.

### Platform Boundaries

On non-Windows systems, Compose replay requires digest-pinned service images and rootless Docker. Windows currently skips sandbox startup and database restore. These differences mean environment replay is not equivalent across platforms.

## Verification

`dawg verify` runs replay first and always checks for replay exit code `0`. It then attempts available comparisons:

| Check | Current comparison |
|---|---|
| HTTP | Ordered response status codes only |
| Screenshot | Exact per-channel pixels; passes with at most 100 differing pixels |
| Replay | Process exit code equals `0` |

HTTP and screenshot comparison errors are currently omitted from the check list. Screenshot comparison is attempted only when the manifest has a non-empty `assertionFile`; captures currently write `"none"`, so a usable baseline may be absent and the screenshot check omitted.

The `--against` value is report metadata (`verifiedAgainst`) only. It does not select, check out, or execute another branch, commit, or path. A verification result of `fail` means the bug still reproduces.

## Repository Structure

```text
dawg/
├── version.json
├── engine/
│   ├── cmd/dawg/                 # Cobra command definitions
│   ├── internal/
│   │   ├── capture/
│   │   ├── dawgenv/
│   │   ├── dawgtypes/
│   │   ├── manifest/
│   │   ├── packager/
│   │   ├── procutil/
│   │   ├── registry/
│   │   ├── replay/
│   │   ├── sanitize/
│   │   └── verify/
│   └── scripts/replay-browser.cjs
├── extension/                    # Chrome Manifest V3 capture extension
├── desktop/
│   ├── src/                      # React UI
│   └── src-tauri/                # Tauri host
├── schema/
│   ├── manifest/v0.3.3-middlechild.json
│   ├── mediatypes.json
│   └── policies/default.rego
├── tools/sync-versions.cjs
└── web/                          # VitePress documentation
```

## Toolchain

| Area | Technology | Version |
|---|---|---:|
| Engine | Go | 1.25.1 |
| CLI | Cobra | 1.9.1 |
| Policy | OPA | 0.70.0 |
| Registry | ORAS | 2.3.1 |
| Compression | klauspost/compress | 1.17.11 |
| Schema validation | jsonschema/v6 | 6.0.1 |
| Browser automation | Playwright | 1.55.1 |
| Recording | rrweb | 2.0.0-alpha.18 |
| Cassette replay | mitmproxy | 12.2.3 |
| JavaScript runtime | Node.js | 22.14.0 |
| Desktop | Tauri | 2 |
| Desktop UI | React | 19.1 |
| Desktop language | TypeScript | 5.8.3 |
| Desktop build | Vite | 7.0.4 |
| Styling | Tailwind CSS | 4.3.3 |
