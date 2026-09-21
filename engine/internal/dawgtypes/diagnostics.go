package dawgtypes

// DiagnosticProfile controls the fidelity requested for browser evidence capture.
type DiagnosticProfile string

const (
	DiagnosticProfileSafe     DiagnosticProfile = "safe"
	DiagnosticProfileEnhanced DiagnosticProfile = "enhanced"
)

// EvidenceState describes why a diagnostic value is or is not present.
type EvidenceState string

const (
	EvidenceCaptured      EvidenceState = "captured"
	EvidenceRedacted      EvidenceState = "redacted"
	EvidencePreviewOnly   EvidenceState = "preview-only"
	EvidenceTruncated     EvidenceState = "truncated"
	EvidenceBlocked       EvidenceState = "blocked"
	EvidenceUnavailable   EvidenceState = "unavailable"
	EvidenceNotRequested  EvidenceState = "not-requested"
	EvidenceCaptureFailed EvidenceState = "capture-failed"
)

// DiagnosticLimits are the policy limits applied while collecting evidence.
type DiagnosticLimits struct {
	MaxConsoleRecordBytes int64 `json:"maxConsoleRecordBytes"`
	MaxConsoleRecords     int64 `json:"maxConsoleRecords"`
	MaxNetworkRecords     int64 `json:"maxNetworkRecords"`
	MaxBodyBytes          int64 `json:"maxBodyBytes"`
	MaxBodiesBytes        int64 `json:"maxBodiesBytes"`
	MaxLayerBytes         int64 `json:"maxLayerBytes"`
}

// DefaultDiagnosticLimits are deliberately conservative release defaults.
func DefaultDiagnosticLimits() DiagnosticLimits {
	return DiagnosticLimits{
		MaxConsoleRecordBytes: 32 << 10,
		MaxConsoleRecords:     10_000,
		MaxNetworkRecords:     20_000,
		MaxBodyBytes:          256 << 10,
		MaxBodiesBytes:        20 << 20,
		MaxLayerBytes:         50 << 20,
	}
}

// DiagnosticsSummary declares the diagnostic evidence retained by an artifact.
type DiagnosticsSummary struct {
	FormatVersion          string            `json:"formatVersion"`
	Profile                DiagnosticProfile `json:"profile"`
	Sources                []string          `json:"sources"`
	ConsoleEvents          int               `json:"consoleEvents"`
	NetworkEvents          int               `json:"networkEvents"`
	ErrorEvents            int               `json:"errorEvents"`
	RetainedRequestBodies  int               `json:"retainedRequestBodies"`
	RetainedResponseBodies int               `json:"retainedResponseBodies"`
	RedactedBodies         int               `json:"redactedBodies"`
	BlockedBodies          int               `json:"blockedBodies"`
	TruncatedBodies        int               `json:"truncatedBodies"`
	PolicyVersion          string            `json:"policyVersion"`
	Limits                 DiagnosticLimits  `json:"limits"`
	Degradations           []string          `json:"degradations,omitempty"`
}

// EmptyDiagnosticsSummary makes absence of a diagnostic event stream explicit.
func EmptyDiagnosticsSummary(profile DiagnosticProfile, policyVersion string) DiagnosticsSummary {
	if profile != DiagnosticProfileEnhanced {
		profile = DiagnosticProfileSafe
	}
	return DiagnosticsSummary{
		FormatVersion: "1",
		Profile:       profile,
		Sources:       []string{},
		PolicyVersion: policyVersion,
		Limits:        DefaultDiagnosticLimits(),
		Degradations:  []string{},
	}
}
