# DAWG bundled resources

This directory is populated by `desktop/build-bundle.ps1` or
`desktop/build-bundle.sh` before creating a release package. It contains the
DAWG engine binaries, Node runtime, Playwright dependencies and Chromium,
replay script, schema, and browser extension.

The tracked file keeps Tauri's `resources/**/*` bundle glob valid in a clean
source checkout, allowing `cargo check`, the Rust language server, and other
development checks to run without staging the full production runtime.
