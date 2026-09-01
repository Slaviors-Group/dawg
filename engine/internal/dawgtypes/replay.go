package dawgtypes

import "time"

// SandboxConfig describes the isolated environment required for replay.
type SandboxConfig struct {
	ComposeFile     string            `json:"composeFile"`
	ImageDigests    map[string]string `json:"imageDigests"`
	Environment     map[string]string `json:"environment"`
	Determinism     DeterminismConfig `json:"determinism"`
	RequireRootless bool              `json:"requireRootless"`
}

// ReplayOutput records files and sandbox details produced by one replay run.
type ReplayOutput struct {
	ArtifactID string         `json:"artifactId"`
	ReplayedAt time.Time      `json:"replayedAt"`
	Sandbox    ReplaySandbox  `json:"sandbox"`
	Outcomes   ReplayOutcomes `json:"outcomes"`
	Status     string         `json:"status"`
}

// ReplaySandbox identifies the isolated compose project used by replay.
type ReplaySandbox struct {
	ContainerID    string `json:"containerId"`
	ComposeProject string `json:"composeProject"`
}

// ReplayOutcomes points to replay output files.
type ReplayOutcomes struct {
	Screenshots   []string `json:"screenshots"`
	HTTPResponses string   `json:"httpResponses"`
	ExitCode      int      `json:"exitCode"`
	AppLogs       string   `json:"appLogs"`
}
