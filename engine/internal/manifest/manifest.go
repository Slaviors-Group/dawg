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

// SchemaVersion is the P0 DAWG manifest schema version.
const SchemaVersion = "0.1.4-alpha"

// Manifest is the OCI config document for a DAWG artifact.
type Manifest struct {
	SchemaVersion   string                      `json:"schemaVersion"`
	ID              string                      `json:"id"`
	CreatedAt       time.Time                   `json:"createdAt"`
	Title           string                      `json:"title"`
	Source          dawgtypes.ManifestSource    `json:"source"`
	Layers          []dawgtypes.LayerSpec       `json:"layers"`
	Sanitize        dawgtypes.ManifestSanitize  `json:"sanitize"`
	Determinism     dawgtypes.DeterminismConfig `json:"determinism"`
	Provenance      *Provenance                 `json:"provenance,omitempty"`
	ExpectedOutcome dawgtypes.ExpectedOutcome   `json:"expectedOutcome"`
}

// Provenance reserves the signed artifact fields implemented in P1.
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

// DefaultSchemaPath locates the development schema from the working directory or executable path.
func DefaultSchemaPath() (string, error) {
	if configuredPath := os.Getenv("DAWG_SCHEMA_PATH"); configuredPath != "" {
		return configuredPath, nil
	}
	if resDir := os.Getenv("DAWG_RESOURCES_DIR"); resDir != "" {
		candidate := filepath.Join(resDir, "schema", "manifest", "v0.1.4-alpha.json")
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
			candidate := filepath.Join(directory, "schema", "manifest", "v0.1.4-alpha.json")
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
			if parent := filepath.Dir(directory); parent == directory {
				break
			}
		}
	}
	return "", fmt.Errorf("manifest: locate schema: %w", os.ErrNotExist)
}
