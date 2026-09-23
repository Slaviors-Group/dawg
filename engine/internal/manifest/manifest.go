// Package manifest owns DAWG artifact manifest serialization and schema validation.
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// SchemaVersion is the current DAWG manifest schema version.
const SchemaVersion = "0.3.3-middlechild"

var schemaFiles = map[string]string{
	"0.2.3-naughty":     "v0.2.3-naughty.json",
	"0.2.5-naughty":     "v0.2.5-naughty.json",
	"0.2.7-naughty":     "v0.2.7-naughty.json",
	"0.3.1-middlechild": "v0.3.1-middlechild.json",
	"0.3.2-middlechild": "v0.3.2-middlechild.json",
	"0.3.3-middlechild": "v0.3.3-middlechild.json",
}

// Manifest is the OCI config document for a DAWG artifact.
type Manifest struct {
	SchemaVersion   string                        `json:"schemaVersion"`
	ID              string                        `json:"id"`
	CreatedAt       time.Time                     `json:"createdAt"`
	Title           string                        `json:"title"`
	Source          dawgtypes.ManifestSource      `json:"source"`
	Layers          []dawgtypes.LayerSpec         `json:"layers"`
	Sanitize        dawgtypes.ManifestSanitize    `json:"sanitize"`
	Determinism     dawgtypes.DeterminismConfig   `json:"determinism"`
	Provenance      *Provenance                   `json:"provenance,omitempty"`
	ExpectedOutcome dawgtypes.ExpectedOutcome     `json:"expectedOutcome"`
	Diagnostics     *dawgtypes.DiagnosticsSummary `json:"diagnostics,omitempty"`
}

// Provenance contains optional artifact-signing metadata.
type Provenance struct {
	SignedBy    string `json:"signedBy"`
	Attestation string `json:"attestation"`
}

// Marshal serializes a manifest as stable indented JSON.
func Marshal(value Manifest) ([]byte, error) {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("manifest: serialize: %w", err)
	}
	return append(contents, '\n'), nil
}

// Unmarshal decodes a manifest without validating its schema.
func Unmarshal(contents []byte) (Manifest, error) {
	var value Manifest
	if err := json.Unmarshal(contents, &value); err != nil {
		return Manifest{}, fmt.Errorf("%w: decode JSON: %v", dawgtypes.ErrInvalidManifest, err)
	}
	return value, nil
}

// Write serializes a manifest to a path.
func Write(path string, value Manifest) error {
	contents, err := Marshal(value)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("manifest: write %s: %w", path, err)
	}
	return nil
}

// Read loads and decodes a manifest from a path.
func Read(path string) (Manifest, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("manifest: read %s: %w", path, err)
	}
	return Unmarshal(contents)
}

// ReadValidated loads a manifest, validates its raw JSON against the schema
// declared by schemaVersion, and then decodes it.
func ReadValidated(path string) (Manifest, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("manifest: read %s: %w", path, err)
	}

	var header struct {
		SchemaVersion string `json:"schemaVersion"`
	}
	if err := json.Unmarshal(contents, &header); err != nil {
		return Manifest{}, fmt.Errorf("%w: decode JSON: %v", dawgtypes.ErrInvalidManifest, err)
	}

	schemaPath, err := SchemaPath(header.SchemaVersion)
	if err != nil {
		return Manifest{}, err
	}
	if err := ValidateJSON(schemaPath, contents); err != nil {
		return Manifest{}, err
	}
	return Unmarshal(contents)
}

// SchemaPath locates the schema for an explicitly supported manifest version.
func SchemaPath(version string) (string, error) {
	filename, supported := schemaFiles[version]
	if !supported {
		return "", fmt.Errorf("%w: unsupported schema version %q", dawgtypes.ErrInvalidManifest, version)
	}

	if configuredPath := os.Getenv("DAWG_SCHEMA_PATH"); configuredPath != "" {
		if version == SchemaVersion {
			return configuredPath, nil
		}
		candidate := filepath.Join(filepath.Dir(configuredPath), filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if resDir := os.Getenv("DAWG_RESOURCES_DIR"); resDir != "" {
		candidate := filepath.Join(resDir, "schema", "manifest", filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	starts := []string{}
	if workingDirectory, err := os.Getwd(); err == nil {
		starts = append(starts, workingDirectory)
	}
	if executable, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(executable))
	}
	for _, start := range starts {
		for directory := start; ; directory = filepath.Dir(directory) {
			candidate := filepath.Join(directory, "schema", "manifest", filename)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
			if parent := filepath.Dir(directory); parent == directory {
				break
			}
		}
	}
	return "", fmt.Errorf("manifest: locate schema for version %q: %w", version, os.ErrNotExist)
}

// DefaultSchemaPath locates the current development schema.
func DefaultSchemaPath() (string, error) {
	return SchemaPath(SchemaVersion)
}
