---
description: "Understand DAWG OCI layouts, evidence layers, portable .dawg archives, integrity validation, and reviewed artifacts."
---

# Artifact Anatomy

A DAWG artifact is a portable, validated record of a browser reproduction. It
contains captured evidence after sanitization and packaging; it is not a copy of
the original application, browser profile, or runtime environment.

## From capture to artifact

1. DAWG records the selected browser tab and any configured local inputs into a
   private session directory.
2. The sanitizer rewrites supported diagnostic and capture streams, then the OPA
   policy decides whether packaging may continue.
3. The packager validates the replay trace and records the artifact as an OCI
   Image Layout with content-addressed blobs.
4. DAWG registers the resulting directory in the local catalog. You can replay
   it locally, export it as a `.dawg` archive, or transfer it through an OCI
   registry.

## OCI image layout

A packaged artifact directory has this shape:

```text
<artifact-directory>/
├── dawg-manifest.json
├── oci-layout
├── index.json
└── blobs/
    └── sha256/
        └── <digest>
```

`oci-layout` and `index.json` identify the directory as an OCI Image Layout.
Each blob is addressed by its SHA-256 digest. `dawg-manifest.json` is DAWG's
validated description of the artifact: its schema version, logical ID, title,
capture metadata, sanitization result, determinism settings, expected outcome,
diagnostic summary, and layer descriptors.

The artifact ID and descriptors identify specific content. Changing retained
evidence produces a new artifact identity rather than modifying the original
one in place.

## Layers and evidence

The manifest lists only the layers present in a capture. The current layer types
are:

| Layer | Contents | Notes |
|---|---|---|
| Environment | Environment lock data | Optional tar layer. |
| Database fixture | Normalized database fixture | Optional zstd-compressed tar layer. |
| Trace | rrweb events, browser actions, frontend HTTP data, and structured logs | zstd-compressed tar layer. A usable trace needs rrweb events, timestamps, and a FullSnapshot. |
| Cassette | Third-party replay cassette data | Optional zstd-compressed tar layer. |
| Diagnostics | Console, network, error, and browser device-profile records | zstd-compressed tar layer. Network records retain explicit evidence states. |
| Diagnostic bodies | Eligible retained Enhanced Diagnostics bodies | Optional zstd-compressed tar layer containing already-sanitized body files. |

The trace provides a visual rrweb replay. It does not re-execute recorded clicks,
restore cookies, or start the original application code. Diagnostic evidence is
available for review and export, but is not injected into the replayed page.

## Portable `.dawg` archives

A `.dawg` file is a ZIP-based transport archive containing the OCI layout. It is
portable between DAWG installations, but importing it does not trust its
contents automatically. DAWG stages and validates archive paths, links, entry
counts, decompressed size, compression ratio, OCI descriptors, digests, and the
DAWG manifest before adding it to the local artifact store.

## Integrity and compatibility

DAWG verifies blob digests against the descriptors in the manifest before it
reads evidence. The schema version controls which artifact fields are required;
newer artifacts do not make older supported schemas unreadable. Artifacts from
`0.2.3-naughty` onward remain supported for inspection, import/export, and
replay, though older artifacts can have no diagnostic evidence layers.

## Privacy and fidelity boundaries

Sanitization runs before packaging, but it is not a confidentiality guarantee.
Review an artifact before sharing it. In particular:

- rrweb masks DOM input fields, while the separate action stream and request
  metadata can contain entered values before sanitization;
- **Safe** diagnostics do not retrieve response bodies through CDP;
- **Enhanced Diagnostics** retains only eligible bounded text or structured
  bodies, and records `captured`, `redacted`, `preview-only`, `truncated`,
  `blocked`, `unavailable`, `not-requested`, or `capture-failed` when needed;
- browser device metadata includes only browser-exposed hints. MAC addresses,
  hostnames, and reliable IP addresses are unavailable, and DAWG does not call
  an external IP-discovery service.

See [Sanitizer Policy](/docs/sanitizer-policy) for the policy and redaction
boundaries.

## Reviewed artifacts

Replay lets you remove complete diagnostic categories or selected retained body
references. DAWG creates a separate reviewed artifact for that operation. It
repackages the remaining evidence with recomputed descriptors and a new logical
identity; the source artifact remains unchanged. Removed bodies are represented
as `unavailable` in matching network records and are never recreated.

## Inspect an artifact

Use the Desktop **Inspect** action or these CLI commands:

```bash
dawg inspect <artifact-directory>
dawg diagnostics inspect <artifact-directory>
```

Use `dawg artifacts export` to create a portable `.dawg` archive, or
`dawg artifacts import` to validate and add one to the local catalog.
