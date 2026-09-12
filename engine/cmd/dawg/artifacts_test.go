package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
)

func TestArtifactsExportImportRoundTrip(t *testing.T) {
	t.Setenv("DAWG_STATE_DIR", t.TempDir())
	source := buildTestOCIArtifact(t, filepath.Join(t.TempDir(), "artifact"))
	archive := filepath.Join(t.TempDir(), "artifact.dawg")
	if err := exportArtifact(source, archive); err != nil {
		t.Fatalf("export artifact: %v", err)
	}
	result, err := importArtifact(archive)
	if err != nil {
		t.Fatalf("import artifact: %v", err)
	}
	if result.Origin != "imported" || result.Status != "valid" {
		t.Fatalf("unexpected import result: %#v", result)
	}
	if _, err := validateOCIArtifact(result.Path); err != nil {
		t.Fatalf("validate imported artifact: %v", err)
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		t.Fatalf("read catalog: %v", err)
	}
	if entry := catalog.Artifacts[result.ID]; entry.Origin != "imported" || entry.ImportSource == "" {
		t.Fatalf("unexpected catalog entry: %#v", entry)
	}
}

func TestExtractArchiveRejectsTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "unsafe.dawg")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../outside")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("no")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := extractArchive(reader.File, t.TempDir()); err == nil {
		t.Fatal("expected traversal rejection")
	}
}

func buildTestOCIArtifact(t *testing.T, directory string) string {
	t.Helper()
	layer := []byte("layer")
	layerDigest := sha256.Sum256(layer)
	value := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		ID:            "sha256:" + fmt.Sprintf("%064x", 1),
		CreatedAt:     mustTime(t, "2026-09-01T10:00:00Z"),
		Title:         "test artifact",
		Source:        dawgtypes.ManifestSource{Reporter: "qa", Environment: "test", RepoCommit: "abc123"},
		Layers: []dawgtypes.LayerSpec{{
			MediaType: dawgtypes.MediaTypeEnvironment,
			Digest:    fmt.Sprintf("sha256:%x", layerDigest),
			Size:      int64(len(layer)),
		}},
		Sanitize:        dawgtypes.ManifestSanitize{PolicyVersion: "1", FieldsRedacted: 0},
		Determinism:     dawgtypes.DeterminismConfig{ClockFrozenAt: mustTime(t, "2026-09-01T09:00:00Z"), RandomSeed: 1},
		ExpectedOutcome: dawgtypes.ExpectedOutcome{Type: "assertion", Description: "test", AssertionFile: "assertions/test.json"},
	}
	config, err := manifest.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	configDigest := sha256.Sum256(config)
	image, err := json.Marshal(ociImageManifest{SchemaVersion: 2, MediaType: ociImageManifestMediaType, Config: ociDescriptor{MediaType: string(dawgtypes.MediaTypeManifest), Digest: fmt.Sprintf("sha256:%x", configDigest), Size: int64(len(config))}, Layers: []ociDescriptor{{MediaType: string(dawgtypes.MediaTypeEnvironment), Digest: fmt.Sprintf("sha256:%x", layerDigest), Size: int64(len(layer))}}})
	if err != nil {
		t.Fatal(err)
	}
	imageDigest := sha256.Sum256(image)
	if err := os.MkdirAll(filepath.Join(directory, "blobs", "sha256"), 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name string, contents []byte) {
		if err := os.WriteFile(name, contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(directory, "dawg-manifest.json"), config)
	write(filepath.Join(directory, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`))
	index, err := json.Marshal(ociIndex{SchemaVersion: 2, MediaType: ociImageIndexMediaType, Manifests: []ociDescriptor{{MediaType: ociImageManifestMediaType, Digest: fmt.Sprintf("sha256:%x", imageDigest), Size: int64(len(image))}}})
	if err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(directory, "index.json"), index)
	write(filepath.Join(directory, "blobs", "sha256", fmt.Sprintf("%x", layerDigest)), layer)
	write(filepath.Join(directory, "blobs", "sha256", fmt.Sprintf("%x", configDigest)), config)
	write(filepath.Join(directory, "blobs", "sha256", fmt.Sprintf("%x", imageDigest)), image)
	return directory
}

func mustTime(t *testing.T, value string) (resultTime time.Time) {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
