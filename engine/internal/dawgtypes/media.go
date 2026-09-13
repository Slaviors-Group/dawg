// Package dawgtypes defines the contracts shared across DAWG pipeline stages.
package dawgtypes

// MediaType identifies a DAWG OCI artifact layer format.
type MediaType string

const (
	// MediaTypeManifest is the OCI config media type for a DAWG manifest.
	MediaTypeManifest MediaType = "application/vnd.dawg.manifest.v0+json"
	// MediaTypeEnvironment is the layer containing the environment lockfile.
	MediaTypeEnvironment MediaType = "application/vnd.dawg.env.lockfile+tar"
	// MediaTypeDatabaseFixture is the layer containing the sanitized DB fixture.
	MediaTypeDatabaseFixture MediaType = "application/vnd.dawg.db.fixture+tar.zstd"
	// MediaTypeTrace is the layer containing browser, HTTP, and log traces.
	MediaTypeTrace MediaType = "application/vnd.dawg.trace.rrweb+jsonl.zstd"
	// MediaTypeCassette is the layer containing third-party API cassettes.
	MediaTypeCassette MediaType = "application/vnd.dawg.cassette+tar.zstd"
)

// CanonicalMediaTypes maps schema keys to their authoritative media types.
func CanonicalMediaTypes() map[string]MediaType {
	return map[string]MediaType{
		"manifest":    MediaTypeManifest,
		"envLockfile": MediaTypeEnvironment,
		"dbFixture":   MediaTypeDatabaseFixture,
		"trace":       MediaTypeTrace,
		"cassette":    MediaTypeCassette,
	}
}
