package dawgtypes

import "time"

// CaptureMetadata describes the files and execution context captured for one session.
type CaptureMetadata struct {
	SessionID          string             `json:"sessionId"`
	StartedAt          time.Time          `json:"startedAt"`
	StoppedAt          time.Time          `json:"stoppedAt"`
	TargetURL          string             `json:"targetUrl"`
	CapturedComponents []string           `json:"capturedComponents"`
	Environment        CaptureEnvironment `json:"environment"`
	Determinism        DeterminismConfig  `json:"determinism"`
	ActionTrace        *ActionTrace       `json:"actionTrace,omitempty"`
}

// CaptureEnvironment identifies the application revision and images used during capture.
type CaptureEnvironment struct {
	RepoCommit        string            `json:"repoCommit"`
	Branch            string            `json:"branch"`
	NodeVersion       string            `json:"nodeVersion"`
	GoVersion         string            `json:"goVersion"`
	DockerComposeFile string            `json:"dockerComposeFile"`
	ImageDigests      map[string]string `json:"imageDigests"`
}

// DeterminismConfig records controls applied during replay.
type DeterminismConfig struct {
	ClockFrozenAt time.Time `json:"clockFrozenAt"`
	RandomSeed    int64     `json:"randomSeed"`
}

// ActionTrace identifies the executable interaction trace paired with an rrweb trace.
type ActionTrace struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// HTTPPair records one request-response exchange.
type HTTPPair struct {
	ID         string       `json:"id"`
	Timestamp  time.Time    `json:"timestamp"`
	Request    HTTPRequest  `json:"request"`
	Response   HTTPResponse `json:"response"`
	Direction  string       `json:"direction"`
	DurationMS int64        `json:"durationMs"`
}

// HTTPRequest is the captured request portion of an HTTPPair.
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// HTTPResponse is the captured response portion of an HTTPPair.
type HTTPResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// DBDiff records a database row-level change or pre-replay fixture row.
type DBDiff struct {
	Timestamp  time.Time      `json:"timestamp"`
	Operation  string         `json:"operation"`
	Table      string         `json:"table"`
	PrimaryKey map[string]any `json:"primaryKey"`
	Before     map[string]any `json:"before"`
	After      map[string]any `json:"after"`
	Query      string         `json:"query"`
	Source     string         `json:"source"`
}

// CassetteEntry records a third-party exchange and its replay match key.
type CassetteEntry struct {
	HTTPPair
	MatchKey CassetteMatchKey `json:"matchKey"`
}

// CassetteMatchKey determines a cassette response without contacting the real service.
type CassetteMatchKey struct {
	Method          string `json:"method"`
	PathPattern     string `json:"pathPattern"`
	BodyFingerprint string `json:"bodyFingerprint"`
	Ordinal         int    `json:"ordinal"`
}
