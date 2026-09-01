package dawgtypes

import "time"

// CheckType identifies a verification comparison.
type CheckType string

const (
	// CheckTypeHTTPResponse compares a replayed HTTP response with its expected assertion.
	CheckTypeHTTPResponse CheckType = "http-response-diff"
	// CheckTypeScreenshot compares a replayed screenshot with a captured state.
	CheckTypeScreenshot CheckType = "screenshot-diff"
	// CheckTypeExitCode compares process exit status.
	CheckTypeExitCode CheckType = "exit-code"
)

// VerifyResult reports whether an artifact still reproduces against a target revision.
type VerifyResult struct {
	ArtifactID      string    `json:"artifactId"`
	VerifiedAt      time.Time `json:"verifiedAt"`
	VerifiedAgainst string    `json:"verifiedAgainst"`
	Result          string    `json:"result"`
	Checks          []Check   `json:"checks"`
	Summary         string    `json:"summary"`
}

// Check is one observable verification result.
type Check struct {
	Type       CheckType `json:"type"`
	Endpoint   string    `json:"endpoint,omitempty"`
	File       string    `json:"file,omitempty"`
	Expected   any       `json:"expected"`
	Actual     any       `json:"actual"`
	DiffPixels int       `json:"diffPixels,omitempty"`
	Threshold  int       `json:"threshold,omitempty"`
	Passed     bool      `json:"passed"`
}
