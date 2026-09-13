# Changelog

All notable changes to DAWG will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [v0.1.2-alpha] - 2026-09-06

### Changed

- Major Desktop UI revamp with improved layout and visual hierarchy
- Standardized control paths for consistent behavior across platforms

### Fixed

- Access denied error when creating DAWG artifacts folder store
- Control path resolution on Windows environments

### Contributors

- [@mamatqurtifa](https://github.com/mamatqurtifa)

---

## [v0.1.1-alpha] - 2026-09-04

### Changed

- Dynamic path integration enhancement for Windows environments
- Deprecated MSI bundler distribution — now focused on NSIS bundle for Windows
- Enhanced engine commands with improved error handling and output

### Fixed

- Path resolution issues on non-standard Windows installations

### Contributors

- [@REZ3X](https://github.com/REZ3X)
- [@mamatqurtifa](https://github.com/mamatqurtifa)

---

## [v0.1.0-alpha] - 2026-09-04

### Added

#### Desktop Application
- Self-contained monolithic desktop app with zero-config setup
- Double-click installer — all runtimes embedded (mitmdump, node, Playwright, schemas, OPA policies)
- Windows NSIS and MSI installers

#### Diagnostics
- Built-in `dawg doctor` diagnostics command (CLI and Desktop UI)
- Pre-flight health probe for runtime readiness, bundled scripts, and schema validation

#### Capture Engine
- Full-fidelity DOM & visual trace recording via rrweb
- Executable browser action trace via Playwright
- Frontend & backend HTTP interception via mitmproxy
- Structured log capture (stdout/file tailing)

#### Sanitization
- OPA-driven privacy sanitization with Rego policies
- Hard-gated PII redaction — export blocked if policy fails
- Synthetic value substitution for format-preserving replacement
- Detailed `sanitize-report.json` with deterministic redactions

#### Packaging & Replay
- Standardized OCI Image Layout packaging with SHA-256 digest-pinned layers
- zstd compression for efficient artifact sizes
- Pre-flight sandbox verification
- Native browser replay engine (`dawg run` / `dawg verify`)

#### CLI
- `dawg init` — Generate project configuration
- `dawg capture` — Start/stop browser session capture
- `dawg inspect` — Display artifact manifest
- `dawg push` / `dawg pull` — OCI registry push/pull
- `dawg run` — Replay artifact in sandbox
- `dawg verify` — Verify reproduction against local codebase

### Known Limitations

- Linux/GNU builds not distributed as prebuilt binaries — build from source required
- Replay sandboxing requires WSL2 with Rootless Docker on Windows
- Bare Windows hosts run in native mock sandbox mode

### Contributors

- [@REZ3X](https://github.com/REZ3X)

---

[Full release history on GitHub](https://github.com/Slaviors-Group/dawg/releases)
