# Architecture

DAWG is a pipeline-based system that captures, sanitizes, packages, replays, and verifies web-app bug reproductions.

## Pipeline Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Capture   │ ──▶│  Sanitizer  │ ──▶│  Packager   │ ──▶│   Replay    │ ──▶│   Verify    │
│    Agent    │    │    (OPA)    │    │   (OCI)     │    │   Engine    │    │   (Diff)    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

| Stage | Runs on | Purpose |
|---|---|---|
| **Capture** | Reporter's machine | Record full bug reproduction state |
| **Sanitize** | Reporter's machine | Redact PII & secrets before export |
| **Package** | Reporter's machine | Bundle into portable OCI artifact |
| **Registry** | Shared storage | Versioned push/pull (optional) |
| **Replay** | Fixer's machine | Deterministic re-execution in sandbox |
| **Verify** | Fixer's machine | Compare outcome against expectation |

---

## Repository Structure

```
dawg/
├── engine/          # Go Engine CLI — core pipeline logic
│   ├── cmd/dawg/    # CLI entry point (Cobra)
│   └── internal/    # Pipeline components
│       ├── capture/     # Browser, proxy, DB, env capture
│       ├── sanitize/    # OPA rules, faker, secret detection
│       ├── packager/    # OCI layout, zstd compression
│       ├── registry/    # ORAS-based push/pull
│       ├── replay/      # Docker sandbox, event replay
│       └── verify/      # Diff engine, report generation
├── desktop/         # Tauri v2 Desktop Shell
│   ├── src/         # React 19 + Vite frontend
│   └── src-tauri/   # Rust backend (IPC bridge)
├── schema/          # OCI Manifest JSON Schema + OPA policies
├── web/             # VitePress documentation site
└── test/            # End-to-end smoke tests
```

---

## Component Responsibilities

### Capture Agent

The capture agent records everything needed to reproduce a bug:

| Sub-component | Technology | What it captures |
|---|---|---|
| **Browser recorder** | rrweb + Playwright | DOM mutations, user actions, screenshots |
| **Network interceptor** | Playwright `page.route()` | Frontend HTTP requests/responses |
| **Backend proxy** | mitmproxy | Backend API traffic, third-party calls |
| **DB tap** | ORM middleware | Database row changes (not full dump) |
| **Env snapshot** | Docker Compose inspection | Environment definition, image digests |
| **Log capture** | stdout/file tailing | Structured JSON application logs |

### Sanitizer

Removes sensitive data before artifact export:

| Component | Purpose |
|---|---|
| **Rule engine** | Field-name matching (`*.password`, `*.email`) + regex patterns |
| **Secret detection** | Gitleaks-pattern detection for API keys, tokens, JWTs |
| **Faker** | Format-preserving synthetic data replacement |
| **OPA gate** | Hard export policy — blocks export if policy fails |

**Principle:** Deny-by-default. If a field doesn't match a known public schema, it gets redacted. Export is blocked (not warned) if OPA policy fails.

### Packager

Bundles capture output into standard OCI artifacts:

| Feature | Implementation |
|---|---|
| **Format** | OCI Image Layout spec |
| **Layering** | Per-concern layers (env, DB fixture, events, cassettes) |
| **Compression** | zstd (better ratio & speed than gzip) |
| **Content addressing** | SHA-256 digests for dedup |
| **Push/pull** | ORAS (OCI Registry As Storage) |

### Replay Engine

Re-executes the bug in a sandboxed environment:

| Feature | Implementation |
|---|---|
| **Environment** | Docker Compose with pinned image digests |
| **Sandbox** | Rootless Docker + seccomp (MVP) |
| **DB restore** | `pg_restore --data-only` filtered by fixture |
| **Time freeze** | Fake timers injected before app boot |
| **Random seed** | Deterministic `Math.random` / `random.seed()` |
| **Network replay** | mitmproxy cassette mode — never hits real APIs |

### Diff/Verify

Compares replayed outcome against expectations:

| Check type | Method |
|---|---|
| **Screenshot** | Pixel-level diff (pixelmatch) |
| **HTTP response** | Status code, headers, body comparison |
| **Exit code** | Pass/fail signal for CI integration |

---

## Technology Stack

### Core Engine

| Technology | Version | Purpose |
|---|---|---|
| Go | 1.25.1 | Primary language, CLI binary |
| Cobra | v1.9.1 | CLI framework |
| OPA | v0.70.0 | Policy evaluation (in-process) |
| ORAS | v2.3.1 | OCI registry client |
| klauspost/compress | v1.17.11 | zstd compression |
| jsonschema/v6 | v6.0.1 | Manifest validation |

### Capture & Replay

| Technology | Purpose |
|---|---|
| Playwright 1.55.1 | Browser automation |
| rrweb 2.0.0-alpha.18 | DOM mutation recording/replay |
| mitmproxy 12.2.3 | HTTP traffic capture |
| Docker | Sandbox isolation |
| PostgreSQL pg_restore | Database fixture restore |

### Desktop Shell

| Technology | Version | Purpose |
|---|---|---|
| Tauri | v2 | Desktop framework (Rust + native webview) |
| React | 19.1.0 | UI layer |
| TypeScript | 5.8.3 | Type safety |
| Vite | 7.0.4 | Build tool |
| Tailwind CSS | 4.3.3 | Utility-first CSS |
| Framer Motion | 12.0.0 | Animations |

---

## Design Decisions

### Why Go for the engine?

Go produces a single static binary with fast startup, making it ideal for CLI distribution across OS platforms. The engine needs to orchestrate subprocesses (Node.js, Python), not run a long-lived server.

### Why Tauri over Electron?

Tauri uses the native OS webview (~600KB) instead of bundling Chromium (~120MB). This aligns with the "desktop app is not a browser" principle — we launch the user's real browser via Playwright, not an embedded one.

### Why OCI Image Layout?

Reusing the OCI standard means existing container tooling (registries, signing, RBAC) works out of the box. No custom format to maintain.

### Why deny-by-default sanitization?

Security-sensitive by default. If the sanitizer can't identify a field as public, it redacts it. This is safer than trying to detect every possible PII pattern and missing edge cases.

---

## Non-Determinism Handling

DAWG explicitly targets deterministic replay for web-app bugs. Here's how it handles non-determinism:

| Source | Mitigation | Phase |
|---|---|---|
| System time | Fake timers injected before app boot | P0 |
| Random numbers | Seed injection (`Math.random`, `random.seed()`) | P0 |
| External API calls | Cassette replay (never hit real service) | P0 |
| UUID generation | Override sequence generators | P1 |
| Race conditions | Out of scope for v1 | — |

**Known limitation:** DAWG does not guarantee 100% reproduction for genuinely non-deterministic bugs (race conditions under real load, external service flakiness).
