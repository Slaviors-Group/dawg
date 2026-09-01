# DAWG

DAWG (Digs Any Web-app Glitch) packages a sanitized web-app bug reproduction into a portable OCI artifact.

## P0 status

This repository currently contains the engine and schema foundation. Capture, sanitization, packaging, replay, registry, verification, and the desktop shell are implemented in later P0 tasks.

## Repository layout

- `engine/`: Go CLI and pipeline implementation.
- `schema/`: Versioned manifest schema and canonical artifact media types.

The Tauri desktop shell is intentionally deferred until the engine JSON CLI contract is implemented. Browser extensions and CI action wrappers are out of scope until P1 and P2.

## Prerequisites

- Go 1.22 or later.
- For future replay support: Linux or WSL2 with rootless Docker. Bare Windows hosts are not supported for replay because DAWG must enforce its sandbox boundary.

## Development

```powershell
Set-Location engine
go test ./...
go vet ./...
go build ./cmd/dawg
```

Run the built CLI from `engine`:

```powershell
.\dawg.exe --help
.\dawg.exe --version
```

Create a baseline configuration in the current directory or a specified target repository:

```powershell
.\dawg.exe init
.\dawg.exe init ..\example-app
```

`dawg init` will not overwrite an existing `dawg.config.yaml` unless invoked with `--force`.

An active capture daemon is stopped through its restricted local control state:

```powershell
.\dawg.exe capture stop
```

Architectural decisions are recorded in `DECISIONS.md`. Product and implementation planning documents live in the sibling `agent/` directory.
