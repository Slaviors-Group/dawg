package packager

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dawg-placeholder/dawg/engine/internal/dawgtypes"
	"github.com/dawg-placeholder/dawg/engine/internal/manifest"
)

const (
	ociImageManifestMediaType = "application/vnd.oci.image.manifest.v1+json"
	ociImageIndexMediaType    = "application/vnd.oci.image.index.v1+json"
)

type ociDescriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type ociImageManifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        ociDescriptor   `json:"config"`
	Layers        []ociDescriptor `json:"layers"`
}

type ociIndex struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Manifests     []ociDescriptor `json:"manifests"`
}

// Package builds an OCI Image Layout from a sanitized capture directory.
func Package(request dawgtypes.PackageRequest) (dawgtypes.PackagedArtifact, error) {
	if err := validateRequest(request); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	report, err := readSanitizeReport(request.SessionDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	if !report.ExportAllowed {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: %w by policy %s", dawgtypes.ErrExportBlocked, report.PolicyFile)
	}
	metadata, err := readCaptureMetadata(request.SessionDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	layers, err := buildLayers(request.SessionDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}

	if err := os.MkdirAll(filepath.Dir(request.OutputDirectory), 0o700); err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: create output parent directory: %w", err)
	}
	stagingDirectory, err := os.MkdirTemp(filepath.Dir(request.OutputDirectory), ".dawg-package-")
	if err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDirectory)
	if err := os.MkdirAll(filepath.Join(stagingDirectory, "blobs", "sha256"), 0o700); err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: create blob directory: %w", err)
	}
	for _, layer := range layers {
		if err := writeBlob(stagingDirectory, layer.spec.Digest, layer.contents); err != nil {
			return dawgtypes.PackagedArtifact{}, err
		}
	}

	manifestValue := buildManifest(request, metadata, report, layers)
	if request.SchemaPath == "" {
		request.SchemaPath, err = manifest.DefaultSchemaPath()
		if err != nil {
			return dawgtypes.PackagedArtifact{}, err
		}
	}
	logicalID, err := logicalIdentity(manifestValue)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	manifestValue.ID = logicalID
	if err := manifest.Validate(request.SchemaPath, manifestValue); err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: validate DAWG manifest: %w", err)
	}
	configContents, err := manifest.Marshal(manifestValue)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	configSpec := layerSpec(dawgtypes.MediaTypeManifest, configContents)
	if err := writeBlob(stagingDirectory, configSpec.Digest, configContents); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	if err := writeJSON(filepath.Join(stagingDirectory, "dawg-manifest.json"), manifestValue); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}

	ociManifestContents, err := marshalOCIManifest(configSpec, layers)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	ociManifestSpec := layerSpec(dawgtypes.MediaType(ociImageManifestMediaType), ociManifestContents)
	if err := writeBlob(stagingDirectory, ociManifestSpec.Digest, ociManifestContents); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	if err := writeJSON(filepath.Join(stagingDirectory, "oci-layout"), map[string]string{"imageLayoutVersion": "1.0.0"}); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	if err := writeJSON(filepath.Join(stagingDirectory, "index.json"), ociIndex{
		SchemaVersion: 2,
		MediaType:     ociImageIndexMediaType,
		Manifests:     []ociDescriptor{descriptor(ociManifestSpec)},
	}); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	if _, err := os.Stat(request.OutputDirectory); err == nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: output directory already exists: %s", request.OutputDirectory)
	} else if !errors.Is(err, os.ErrNotExist) {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: inspect output directory: %w", err)
	}
	if err := os.Rename(stagingDirectory, request.OutputDirectory); err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: publish OCI layout: %w", err)
	}
	return dawgtypes.PackagedArtifact{
		Directory:         request.OutputDirectory,
		ManifestPath:      filepath.Join(request.OutputDirectory, "dawg-manifest.json"),
		OCIManifestDigest: ociManifestSpec.Digest,
	}, nil
}

func validateRequest(request dawgtypes.PackageRequest) error {
	if request.SessionDirectory == "" || request.OutputDirectory == "" || request.Title == "" {
		return fmt.Errorf("packager: session directory, output directory, and title are required")
	}
	if request.Source.Reporter == "" || request.Source.Environment == "" || request.Source.RepoCommit == "" {
		return fmt.Errorf("packager: source reporter, environment, and repo commit are required")
	}
	if request.ExpectedOutcome.Type == "" || request.ExpectedOutcome.Description == "" || request.ExpectedOutcome.AssertionFile == "" {
		return fmt.Errorf("packager: complete expected outcome is required")
	}
	return nil
}

func readSanitizeReport(sessionDirectory string) (dawgtypes.SanitizeReport, error) {
	path := filepath.Join(sessionDirectory, "sanitize-report.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		return dawgtypes.SanitizeReport{}, fmt.Errorf("packager: read sanitization report %s: %w", path, err)
	}
	var report dawgtypes.SanitizeReport
	if err := json.Unmarshal(contents, &report); err != nil {
		return dawgtypes.SanitizeReport{}, fmt.Errorf("packager: decode sanitization report %s: %w", path, err)
	}
	return report, nil
}

func readCaptureMetadata(sessionDirectory string) (dawgtypes.CaptureMetadata, error) {
	path := filepath.Join(sessionDirectory, "meta.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		return dawgtypes.CaptureMetadata{}, fmt.Errorf("packager: read capture metadata %s: %w", path, err)
	}
	var metadata dawgtypes.CaptureMetadata
	if err := json.Unmarshal(contents, &metadata); err != nil {
		return dawgtypes.CaptureMetadata{}, fmt.Errorf("packager: decode capture metadata %s: %w", path, err)
	}
	return metadata, nil
}

func buildManifest(request dawgtypes.PackageRequest, metadata dawgtypes.CaptureMetadata, report dawgtypes.SanitizeReport, layers []layerBlob) manifest.Manifest {
	layerSpecs := make([]dawgtypes.LayerSpec, 0, len(layers))
	for _, layer := range layers {
		layerSpecs = append(layerSpecs, layer.spec)
	}
	return manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		CreatedAt:     metadata.StoppedAt.UTC(),
		Title:         request.Title,
		Source:        request.Source,
		Layers:        layerSpecs,
		Sanitize: dawgtypes.ManifestSanitize{
			PolicyVersion:  report.PolicyVersion,
			FieldsRedacted: report.FieldsRedacted,
		},
		Determinism:     metadata.Determinism,
		ExpectedOutcome: request.ExpectedOutcome,
	}
}

func logicalIdentity(value manifest.Manifest) (string, error) {
	value.ID = ""
	contents, err := manifest.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("packager: serialize logical identity: %w", err)
	}
	digest := sha256.Sum256(contents)
	return fmt.Sprintf("sha256:%x", digest), nil
}

func marshalOCIManifest(config dawgtypes.LayerSpec, layers []layerBlob) ([]byte, error) {
	descriptors := make([]ociDescriptor, 0, len(layers))
	for _, layer := range layers {
		descriptors = append(descriptors, descriptor(layer.spec))
	}
	return json.Marshal(ociImageManifest{
		SchemaVersion: 2,
		MediaType:     ociImageManifestMediaType,
		Config:        descriptor(config),
		Layers:        descriptors,
	})
}

func descriptor(spec dawgtypes.LayerSpec) ociDescriptor {
	return ociDescriptor{MediaType: string(spec.MediaType), Digest: spec.Digest, Size: spec.Size}
}

func writeBlob(layoutDirectory, digest string, contents []byte) error {
	digestPath, err := layerDigestPath(digest)
	if err != nil {
		return fmt.Errorf("packager: blob digest: %w", err)
	}
	path := filepath.Join(layoutDirectory, "blobs", "sha256", digestPath)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("packager: write blob %s: %w", path, err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("packager: serialize %s: %w", path, err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("packager: write %s: %w", path, err)
	}
	return nil
}
