// Package manifest owns DAWG artifact manifest serialization and schema validation.
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// SchemaVersion is the current DAWG manifest schema version.
const SchemaVersion = "0.4.3-omega"

var schemaFiles = map[string]string{
	"0.2.3-naughty":     "v0.2.3-naughty.json",
	"0.2.5-naughty":     "v0.2.5-naughty.json",
	"0.2.7-naughty":     "v0.2.7-naughty.json",
	"0.3.1-middlechild": "v0.3.1-middlechild.json",
	"0.3.2-middlechild": "v0.3.2-middlechild.json",
	"0.3.3-middlechild": "v0.3.3-middlechild.json",
	"0.3.5-middlechild": "v0.3.5-middlechild.json",
	"0.4.1-omega":       "v0.4.1-omega.json",
	"0.4.2-omega":       "v0.4.2-omega.json",
	"0.4.3-omega":       "v0.4.3-omega.json",
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
	Review          *Review                       `json:"review,omitempty"`
}

// Review is portable review metadata for an immutable flagged artifact.
type Review struct {
	FormatVersion    string       `json:"formatVersion"`
	Kind             string       `json:"kind"`
	RootArtifactID   string       `json:"rootArtifactId"`
	ParentArtifactID string       `json:"parentArtifactId"`
	Revision         int          `json:"revision"`
	ReviewedAt       time.Time    `json:"reviewedAt"`
	Flags            []ReviewFlag `json:"flags"`
}

// ReviewFlag marks a point or range within the replay timeline.
type ReviewFlag struct {
	ID            string `json:"id"`
	StartOffsetMs int64  `json:"startOffsetMs"`
	EndOffsetMs   *int64 `json:"endOffsetMs,omitempty"`
	Title         string `json:"title"`
	Note          string `json:"note,omitempty"`
	Category      string `json:"category"`
	Severity      string `json:"severity"`
}

var reviewIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)
var artifactIDPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

// Validate checks review metadata, including ordering and an optional replay duration.
// A negative duration skips timeline-bound checks.
func (review Review) Validate(durationMs int64) error {
	if review.FormatVersion != "1" || review.Kind != "flagged" || !artifactIDPattern.MatchString(review.RootArtifactID) || !artifactIDPattern.MatchString(review.ParentArtifactID) || review.Revision < 1 {
		return fmt.Errorf("invalid review metadata")
	}
	if review.ReviewedAt.IsZero() || review.ReviewedAt.Location() != time.UTC || len(review.Flags) == 0 || len(review.Flags) > 500 {
		return fmt.Errorf("invalid review timestamp or flag count")
	}
	seen := make(map[string]struct{}, len(review.Flags))
	for index, flag := range review.Flags {
		if !reviewIDPattern.MatchString(flag.ID) || flag.StartOffsetMs < 0 || (durationMs >= 0 && flag.StartOffsetMs > durationMs) {
			return fmt.Errorf("invalid review flag %d offset or ID", index)
		}
		if flag.EndOffsetMs != nil && (*flag.EndOffsetMs <= flag.StartOffsetMs || (durationMs >= 0 && *flag.EndOffsetMs > durationMs)) {
			return fmt.Errorf("invalid review flag %d range", index)
		}
		if strings.TrimSpace(flag.Title) != flag.Title || flag.Title == "" || utf8.RuneCountInString(flag.Title) > 120 || utf8.RuneCountInString(flag.Note) > 2000 {
			return fmt.Errorf("invalid review flag %d text", index)
		}
		if flag.Category != "bug" && flag.Category != "error" && flag.Category != "network" && flag.Category != "console" && flag.Category != "action" && flag.Category != "note" {
			return fmt.Errorf("invalid review flag %d category", index)
		}
		if flag.Severity != "info" && flag.Severity != "warning" && flag.Severity != "error" {
			return fmt.Errorf("invalid review flag %d severity", index)
		}
		if _, duplicate := seen[flag.ID]; duplicate {
			return fmt.Errorf("duplicate review flag ID %q", flag.ID)
		}
		seen[flag.ID] = struct{}{}
		if index > 0 && (flag.StartOffsetMs < review.Flags[index-1].StartOffsetMs || (flag.StartOffsetMs == review.Flags[index-1].StartOffsetMs && flag.ID < review.Flags[index-1].ID)) {
			return fmt.Errorf("review flags are not deterministically ordered")
		}
	}
	return nil
}

// SortFlags orders flags deterministically before they are serialized.
func SortFlags(flags []ReviewFlag) {
	sort.Slice(flags, func(i, j int) bool {
		if flags[i].StartOffsetMs == flags[j].StartOffsetMs {
			return flags[i].ID < flags[j].ID
		}
		return flags[i].StartOffsetMs < flags[j].StartOffsetMs
	})
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
	value, err := Unmarshal(contents)
	if err != nil {
		return Manifest{}, err
	}
	if value.Review != nil {
		if err := value.Review.Validate(-1); err != nil {
			return Manifest{}, fmt.Errorf("%w: review validation: %v", dawgtypes.ErrInvalidManifest, err)
		}
	}
	return value, nil
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
