package dawgtypes

import "errors"

var (
	// ErrExportBlocked indicates that an artifact failed sanitization policy.
	ErrExportBlocked = errors.New("sanitizer: export blocked by policy")
	// ErrSandboxRequired indicates that replay cannot enforce its sandbox boundary.
	ErrSandboxRequired = errors.New("replay: sandbox-required environment unavailable")
	// ErrInvalidManifest indicates that a manifest cannot satisfy the DAWG schema.
	ErrInvalidManifest = errors.New("manifest: invalid")
	// ErrInvalidOutput indicates an unsupported CLI output mode.
	ErrInvalidOutput = errors.New("cli: invalid output format")
	// ErrConfigExists indicates that init refused to overwrite an existing config.
	ErrConfigExists = errors.New("init: configuration already exists")
	// ErrDockerUnavailable indicates that environment capture cannot contact Docker.
	ErrDockerUnavailable = errors.New("capture: Docker unavailable")
	// ErrCaptureAlreadyRunning indicates that a capture session was started twice.
	ErrCaptureAlreadyRunning = errors.New("capture: session already running")
	// ErrCaptureNotRunning indicates that a capture session was stopped without a start.
	ErrCaptureNotRunning = errors.New("capture: session is not running")
)
