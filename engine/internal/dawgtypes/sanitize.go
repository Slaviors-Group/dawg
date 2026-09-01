package dawgtypes

// SanitizeReport records redactions and the OPA export decision for a capture directory.
type SanitizeReport struct {
	PolicyVersion  string      `json:"policyVersion"`
	PolicyFile     string      `json:"policyFile"`
	OPAResult      string      `json:"opaResult"`
	FieldsScanned  int         `json:"fieldsScanned"`
	FieldsRedacted int         `json:"fieldsRedacted"`
	Redactions     []Redaction `json:"redactions"`
	BlockedFields  []string    `json:"blockedFields"`
	ExportAllowed  bool        `json:"exportAllowed"`
}

// Redaction describes one field changed or blocked by sanitization.
type Redaction struct {
	File           string `json:"file"`
	Line           int    `json:"line"`
	Field          string `json:"field"`
	Reason         string `json:"reason"`
	Action         string `json:"action"`
	SyntheticValue string `json:"syntheticValue,omitempty"`
}

// OPAResult is the result of evaluating a sanitizer policy.
type OPAResult struct {
	Allowed    bool     `json:"allowed"`
	PolicyFile string   `json:"policyFile"`
	Reasons    []string `json:"reasons,omitempty"`
}
