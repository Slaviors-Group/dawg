# Changelog

Notable changes to DAWG are listed below in descending release order. Each heading links to the corresponding GitHub release.

## [0.2.3-naughty](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-2) - 2026-09-12

**Tag:** `naughty-2` · **Release:** Naughty 2, Naughty Boy

### Added

- Portable `.dawg` artifact import and export through the engine, native desktop dialogs, and desktop drag-and-drop
- A persistent local artifact catalog with automatic discovery and refresh after captures and imports
- Readable, collision-safe capture names and optional artifact titles
- Artifact origin labels plus replay search and filtering by title, digest, target URL, path, and origin
- Staged archive extraction and validation for paths, links, size and compression limits, OCI descriptors, digests, and DAWG manifests

### Fixed

- Desktop export now preserves the path selected in the save dialog instead of creating an unintended file named `json`

---

## [0.2.0-naughty](https://github.com/Slaviors-Group/dawg/releases/tag/naughty) - 2026-09-11

**Tag:** `naughty` · **Release:** Naughty Boy is Here, Finally

### Added

- Extension-first capture of rrweb events, browser actions, and frontend request metadata through a session-authenticated engine streaming server
- Real desktop replay execution, replay diagnostics, cancellation, and cleanup of DAWG-owned engine, Node.js, Chromium, mitmproxy, and capture processes
- Centralized version management through `version.json` with synchronization and consistency checks

### Fixed

- Capture and replay lifecycle races, shutdown deadlocks, cleanup ordering, and stop timeouts
- Replay corruption caused by sanitizing rrweb doctype names and SVG `viewBox` or `points` geometry
- Unwanted Windows console windows for background DAWG subprocesses

---

## [0.1.3-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/v0.1.3) - 2026-09-09

**Tag:** `v0.1.3` · **Release:** v0.1.3-1/alpha

### Added

- Base Jenkins configuration for automated build and bundle validation

### Fixed

- Engine subprocesses now run without visible console windows on Windows
- The desktop Replay Engine view now starts replay through the engine bridge
- Bare Windows hosts now show the engine's native-compatibility replay status instead of an incorrect Linux or WSL2 blocker

---

## [0.1.2-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-3) - 2026-09-06

**Tag:** `alpha-3` · **Release:** v0.1.2-1/alpha

### Changed

- Revamped the desktop UI and standardized control paths

### Fixed

- Access-denied errors while creating the DAWG artifact store

---

## [0.1.1-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-2) - 2026-09-04

**Tag:** `alpha-2` · **Release:** v0.1.1-1/alpha

### Changed

- Improved dynamic path integration and engine commands on Windows
- Deprecated MSI distribution in favor of the Windows NSIS bundle

---

## [0.1.0-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha) - 2026-09-04

**Tag:** `alpha` · **Release:** 0.1.0-1/alpha

### Added

- A self-contained Windows desktop app with bundled runtimes and built-in `dawg doctor` diagnostics
- DOM, browser-action, HTTP, and structured-log capture
- OPA-driven privacy sanitization with gated PII redaction and reports
- OCI artifact packaging, native browser replay, verification, and registry push/pull commands

### Known limitations

- Prebuilt Linux packages were not included
- Container-isolated replay required WSL2 with rootless Docker on Windows; bare Windows used native mock sandbox mode

---

[Full release history on GitHub](https://github.com/Slaviors-Group/dawg/releases)
