package dawgtypes

// LayerSpec describes a content-addressed OCI layer in a DAWG artifact.
type LayerSpec struct {
	MediaType MediaType `json:"mediaType"`
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
}

// ManifestSource identifies the origin of a DAWG artifact.
type ManifestSource struct {
	Reporter    string `json:"reporter"`
	Environment string `json:"environment"`
	RepoCommit  string `json:"repoCommit"`
}

// ManifestSanitize summarizes the export policy applied to the artifact.
type ManifestSanitize struct {
	PolicyVersion  string `json:"policyVersion"`
	FieldsRedacted int    `json:"fieldsRedacted"`
	ReviewedBy     string `json:"reviewedBy,omitempty"`
}

// ExpectedOutcome specifies the assertion used to determine whether the bug still reproduces.
type ExpectedOutcome struct {
	Type          string `json:"type"`
	Description   string `json:"description"`
	AssertionFile string `json:"assertionFile"`
}

// ManifestBuilder provides packager inputs before an OCI manifest digest exists.
type ManifestBuilder struct {
	Title           string
	Source          ManifestSource
	Layers          []LayerSpec
	Sanitize        ManifestSanitize
	Determinism     DeterminismConfig
	ExpectedOutcome ExpectedOutcome
}

// PackageRequest contains the sanitized session and user-supplied context required to build an artifact.
type PackageRequest struct {
	SessionDirectory string
	OutputDirectory  string
	SchemaPath       string
	Title            string
	Source           ManifestSource
	ExpectedOutcome  ExpectedOutcome
}

// PackagedArtifact identifies a completed local OCI Image Layout.
type PackagedArtifact struct {
	Directory         string `json:"directory"`
	ManifestPath      string `json:"manifestPath"`
	OCIManifestDigest string `json:"ociManifestDigest"`
}
