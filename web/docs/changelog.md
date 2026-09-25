---
description: "DAWG release history, direct Windows and Linux downloads, compatibility notes, additions, changes, and fixes."
---

# Changelog

Notable changes to DAWG are grouped by release family and listed in descending release order. Released versions link to their corresponding GitHub releases. Download links appear only for assets currently published on that release.

## Middlechild family — `0.3.x`

### [0.3.5-middlechild](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-5) - 2026-09-24

**Desktop:** `0.3.5` · **Extension display version:** `0.3.5_middlechild`

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_x64-setup.exe) · [Linux AppImage](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_amd64.AppImage)

### Added

- Desktop replay controls now stay synchronized with the authoritative rrweb player in Chromium through a versioned bidirectional protocol.
- Console, error, and network evidence appears at its literal replay time, including request, first-byte, and completion stages, with an **All captured** fallback.
- Artifacts retain a sanitized browser device profile covering available browser/OS hints, viewport, screen, locale, timezone, hardware capacity, connection hints, page URL, and user agent; Replay exposes it in a dedicated Device tab.
- Prominent tag-specific Windows and Linux downloads are available from the homepage.
- GitHub Sponsors links are available in the website navigation/footer and Desktop open-source card.

### Fixed

- MV3 worker restarts preserve the private capture token and wait for state restoration before forwarding rrweb events.
- Capture startup now fails when rrweb is unavailable instead of publishing an empty recording.
- Packaging rejects replay traces without events, valid timestamps, or a FullSnapshot.
- The Desktop Timeline tab reads the packaged diagnostic timeline instead of looking for unavailable manifest summary fields.

### Changed

- The extension capture panel is status-only; captures stop from Desktop or CLI so pending events drain before packaging.
- Browser-restricted MAC, hostname, and reliable IP values are represented as unavailable; capture does not call an external discovery service.
- New artifacts use the explicit `v0.3.5-middlechild` schema while previous supported schemas remain readable.

---

### [0.3.3-middlechild](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-3) - 2026-09-23

**Desktop:** `0.3.3` · **Extension display version:** `0.3.1_middlechild`

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-3/DAWG_0.3.3_x64-setup.exe)

### Added

- Exported `.dawg` artifacts now register as **DAWG Artifact** documents in desktop bundles, using the DAWG paw application icon.
- Double-clicking an associated `.dawg` archive opens DAWG and imports it through the existing validated import flow.

### Changed

- New artifacts use the explicit `v0.3.3-middlechild` manifest schema; existing supported artifact schemas remain readable.

---

### [0.3.2-middlechild](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-2) - 2026-09-23

**Desktop:** `0.3.2` · **Extension display version:** `0.3.1_middlechild`

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-2/DAWG_0.3.2_x64-setup.exe) · [Linux AppImage](https://github.com/Slaviors-Group/dawg/releases/download/middlechild-2/DAWG_0.3.2_amd64.AppImage)

### Fixed

- Large rrweb events no longer fail diagnostic timeline extraction because of Go's default scanner token limit.
- Diagnostic record streams are no longer limited by the retained-body file-size limit.
- Oversized replay traces omit only optional timeline correlation; diagnostic inspection, HAR export, and cURL export remain available.
- HTTP verification now records response-comparison failures instead of silently omitting the check.

### Changed

- Capture, sanitization, diagnostic packaging, and HTTP verification now share a 16 MiB JSONL record limit.
- Diagnostic-layer decompression is bounded before data is retained in memory.
- New artifacts use the explicit `v0.3.2-middlechild` manifest schema; previous supported artifact schemas remain readable.

---

### [0.3.1-middlechild](https://github.com/Slaviors-Group/dawg/releases/tag/middlechild) - 2026-09-21

**Desktop:** `0.3.1` · **Extension display version:** `0.3.1_middlechild`

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/middlechild/DAWG_0.3.1_x64-setup.exe) · [Linux AppImage](https://github.com/Slaviors-Group/dawg/releases/download/middlechild/DAWG_0.3.1_amd64.AppImage)

### Added

- **Safe** capture as the default bounded diagnostic profile for console, error, and network metadata
- Consented **Enhanced Diagnostics** capture using CDP for console APIs, exceptions, and network evidence
- Eligible Enhanced CDP response-body retention for JSON, GraphQL, form-encoded, and text content, bounded by content type, record count, per-body size, aggregate-body size, layer size, and concurrent retrieval limits
- Explicit diagnostic evidence states: `captured`, `redacted`, `preview-only`, `truncated`, `blocked`, `unavailable`, `not-requested`, and `capture-failed`
- OCI diagnostic-record and diagnostic-body layers, manifest diagnostic summaries, and a strict `v0.3.1-middlechild` artifact schema
- `dawg diagnostics inspect`, `dawg diagnostics export-har`, and `dawg diagnostics copy-curl` commands
- `dawg diagnostics remove` to create an independently digest-addressed reviewed artifact with selected diagnostic categories or individual retained body references removed
- Desktop Replay evidence review with state filtering, sanitized HAR export, reviewed cURL copying, and explicit evidence-removal confirmation
- Approximate diagnostic-to-rrweb correlation and Desktop seeking for an active interactive replay when the record timestamp falls within the retained rrweb range

### Changed

- Sanitization now covers retained diagnostic console, network, error, and body evidence before packaging
- Enhanced CDP attach failures fall back to Safe and capture degradations are retained as diagnostic errors
- Enhanced capture attaches Chrome's debugger before rrweb takes its initial full snapshot, so captures use the viewport after Chrome applies its debugger infobar
- `dawg doctor` derives `schema:manifest` from the current schema constant and now reports `0.3.1-middlechild`
- Version synchronization distinguishes the application/artifact schema (`0.3.1-middlechild`), Desktop (`0.3.1`), and extension display version (`0.3.1_middlechild`)

### Privacy and fidelity boundaries

- Safe capture does not retrieve response bodies through CDP
- Enhanced capture retains only eligible bounded body content; unavailable, blocked, truncated, or failed content remains represented by its evidence state
- Removing a retained body rewrites its matching network reference to `unavailable`; DAWG never recreates removed content
- A reviewed artifact has recomputed OCI descriptors and logical identity. Source provenance is not carried over to modified evidence
- Replay correlation is approximate and only appears for in-range timestamps; records without a valid correlation do not expose a seek action
- Replay continues to render the rrweb recording rather than re-executing captured actions or restoring the original application runtime

### Compatibility

- Existing `0.2.3-naughty`, `0.2.5-naughty`, and `0.2.7-naughty` artifacts remain supported for inspect, import/export, and replay. They return empty diagnostics where no evidence layers exist.

---

## Naughty family — `0.2.x`

### [0.2.7-naughty](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-4) - 2026-09-20

**Tag:** `naughty-4`

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/naughty-4/DAWG_0.2.7_x64-setup.exe) · [Linux AppImage](https://github.com/Slaviors-Group/dawg/releases/download/naughty-4/DAWG_0.2.7_amd64.AppImage)

### Added

- Replay debugging through Chromium DevTools with named replay resources and a `window.__DAWG_REPLAY__` handle for the reconstructed rrweb document
- Artifact compatibility for supported manifests from `0.2.3-naughty` onward

### Changed

- Interactive replay now uses a stable local replay page that can be inspected and reloaded normally

### Fixed

- Refreshing interactive Chromium now reconstructs and restarts replay instead of leaving a blank page
- Artifact Inspector dialogs now remain above the sidebar and fit minimized or narrow windows

---

### [0.2.5-naughty](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-3) - 2026-09-19

**Tag:** `naughty-3`

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/naughty-3/DAWG_0.2.5_x64-setup.exe) · [Linux AppImage](https://github.com/Slaviors-Group/dawg/releases/download/naughty-3/DAWG_0.2.5_amd64.AppImage)

### Added

- A DAWG-styled extension popup using the real paw asset and locally hosted Outfit font
- Portable `.dawg` Replay import and export
- Chrome Web Store distribution for the browser extension

### Changed

- Dashboard **Replay** now navigates to Replay, preselects the chosen artifact, and does not start playback automatically
- Desktop replay now provides interactive play/pause, skip, speed, and timeline controls
- Desktop replay preserves the recorded viewport with corrected letterboxing and scaling

---

### [0.2.3-naughty](https://github.com/Slaviors-Group/dawg/releases/tag/naughty-2) - 2026-09-12

**Tag:** `naughty-2` · **Release:** Naughty 2, Naughty Boy

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/naughty-2/DAWG_0.2.3_x64-setup.exe)

### Added

- Portable `.dawg` artifact import and export through the engine, native desktop dialogs, and desktop drag-and-drop
- A persistent local artifact catalog with automatic discovery and refresh after captures and imports
- Readable, collision-safe capture names and optional artifact titles
- Artifact origin labels plus replay search and filtering by title, digest, target URL, path, and origin
- Staged archive extraction and validation for paths, links, size and compression limits, OCI descriptors, digests, and DAWG manifests

### Fixed

- Desktop export now preserves the path selected in the save dialog instead of creating an unintended file named `json`

---

### [0.2.0-naughty](https://github.com/Slaviors-Group/dawg/releases/tag/naughty) - 2026-09-11

**Tag:** `naughty` · **Release:** Naughty Boy is Here, Finally

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/naughty/DAWG_0.2.0_x64-setup.exe)

### Added

- Extension-first capture of rrweb events, browser actions, and frontend request metadata through a session-authenticated engine streaming server
- Real desktop replay execution, replay diagnostics, cancellation, and cleanup of DAWG-owned engine, Node.js, Chromium, mitmproxy, and capture processes
- Centralized version management through `version.json` with synchronization and consistency checks

### Fixed

- Capture and replay lifecycle races, shutdown deadlocks, cleanup ordering, and stop timeouts
- Replay corruption caused by sanitizing rrweb doctype names and SVG `viewBox` or `points` geometry
- Unwanted Windows console windows for background DAWG subprocesses

---

## Alpha family — `0.1.x`

### [0.1.3-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/v0.1.3) - 2026-09-09

**Tag:** `v0.1.3` · **Release:** v0.1.3-1/alpha

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/v0.1.3/DAWG_0.1.3_x64-setup.exe)

### Added

- Base Jenkins configuration for automated build and bundle validation

### Fixed

- Engine subprocesses now run without visible console windows on Windows
- The desktop Replay Engine view now starts replay through the engine bridge
- Bare Windows hosts now show the engine's native-compatibility replay status instead of an incorrect Linux or WSL2 blocker

---

### [0.1.2-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-3) - 2026-09-06

**Tag:** `alpha-3` · **Release:** v0.1.2-1/alpha

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/alpha-3/DAWG_0.1.2_x64-setup.exe)

### Changed

- Revamped the desktop UI and standardized control paths

### Fixed

- Access-denied errors while creating the DAWG artifact store

---

### [0.1.1-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha-2) - 2026-09-04

**Tag:** `alpha-2` · **Release:** v0.1.1-1/alpha

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/alpha-2/DAWG_0.1.1_x64-setup.exe)

### Changed

- Improved dynamic path integration and engine commands on Windows
- Deprecated MSI distribution in favor of the Windows NSIS bundle

---

### [0.1.0-alpha](https://github.com/Slaviors-Group/dawg/releases/tag/alpha) - 2026-09-04

**Tag:** `alpha` · **Release:** 0.1.0-1/alpha

**Downloads:** [Windows installer](https://github.com/Slaviors-Group/dawg/releases/download/alpha/DAWG_0.1.0_x64-setup.exe) · [Windows MSI](https://github.com/Slaviors-Group/dawg/releases/download/alpha/DAWG_0.1.0_x64_en-US.msi)

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
