# DAWG

DAWG (Digs Any Web-app Glitch) packages a sanitized web-app bug reproduction into a portable OCI artifact.

## P0 status

The engine foundation, capture components, sanitizer, local OCI packager, artifact inspection, registry client, and replay sandbox preflight are implemented. The end-to-end capture-to-artifact workflow, registry authentication/round-trip testing, replay execution, verification, and desktop shell remain in progress.

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

`dawg capture` currently requires Node.js with Playwright/Chromium and mitmproxy. Replay is supported only on Linux or WSL2 with rootless Docker; bare Windows refuses replay to preserve the sandbox requirement.
