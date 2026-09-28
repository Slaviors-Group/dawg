package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/artifact"
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
	entry, found := catalog.Artifacts[result.InstanceID]
	if !found || entry.Origin != "imported" || entry.ImportSource == "" || entry.ArtifactID != result.ID {
		t.Fatalf("unexpected catalog entry: %#v", entry)
	}
}

func TestReadArtifactCatalogMigratesV1Entries(t *testing.T) {
	state := t.TempDir()
	t.Setenv("DAWG_STATE_DIR", state)
	storedPath := filepath.Join(state, "artifacts", "legacy")
	if err := os.MkdirAll(filepath.Dir(storedPath), 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := map[string]any{"artifacts": map[string]any{
		"sha256:" + fmt.Sprintf("%064x", 9): map[string]any{
			"origin": "imported", "importedAt": "2026-09-01T10:00:00Z", "storedPath": storedPath, "importSource": "archive.dawg",
		},
	}}
	contents, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, artifactCatalogName), contents, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		t.Fatalf("migrate catalog: %v", err)
	}
	if catalog.Version != artifactCatalogVersion || len(catalog.Artifacts) != 1 {
		t.Fatalf("unexpected migrated catalog: %#v", catalog)
	}
	for _, entry := range catalog.Artifacts {
		if entry.InstanceID == "" || entry.ArtifactID == "" || entry.AddedAt.Format(time.RFC3339) != "2026-09-01T10:00:00Z" {
			t.Fatalf("unexpected migrated entry: %#v", entry)
		}
	}
	persisted, err := os.ReadFile(filepath.Join(state, artifactCatalogName))
	if err != nil || !bytes.Contains(persisted, []byte(`"version": 2`)) {
		t.Fatalf("catalog v2 migration was not persisted: %s, %v", persisted, err)
	}
}

func TestDeleteArtifactOnlyDeletesManagedLocalInstance(t *testing.T) {
	t.Setenv("DAWG_STATE_DIR", t.TempDir())
	root, err := artifactRoot()
	if err != nil {
		t.Fatal(err)
	}
	local := buildTestOCIArtifact(t, filepath.Join(root, "managed"))
	if err := registerCapturedArtifact(local, "http://localhost:3000"); err != nil {
		t.Fatalf("register artifact: %v", err)
	}
	result, err := deleteArtifact(local)
	if err != nil {
		t.Fatalf("delete managed artifact: %v", err)
	}
	if result.Status != "deleted" || result.InstanceID == "" {
		t.Fatalf("unexpected delete result: %#v", result)
	}
	if _, err := os.Stat(local); !os.IsNotExist(err) {
		t.Fatalf("artifact was not deleted: %v", err)
	}
	outside := buildTestOCIArtifact(t, filepath.Join(t.TempDir(), "outside"))
	if _, err := deleteArtifact(outside); err == nil {
		t.Fatal("expected deletion outside the artifact root to be rejected")
	}
}

func TestDefaultReviewArtifactTitle(t *testing.T) {
	if got := defaultReviewArtifactTitle("Checkout failure", 3); got != "Checkout failure review 3" {
		t.Fatalf("unexpected default review title: %q", got)
	}
}

func TestEditorArtifactTitle(t *testing.T) {
	tests := []struct {
		name    string
		raw     json.RawMessage
		want    string
		wantErr bool
	}{
		{name: "custom", raw: json.RawMessage(`"Regression findings"`), want: "Regression findings"},
		{name: "normalizes whitespace", raw: json.RawMessage(`"  Regression findings  "`), want: "Regression findings"},
		{name: "blank uses default", raw: json.RawMessage(`"   "`), want: "Checkout failure review 2"},
		{name: "missing uses default", want: "Checkout failure review 2"},
		{name: "rejects non-string", raw: json.RawMessage(`42`), wantErr: true},
		{name: "rejects null", raw: json.RawMessage(`null`), wantErr: true},
		{name: "rejects over 120 characters", raw: json.RawMessage(`"` + strings.Repeat("x", 121) + `"`), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := editorArtifactTitle(test.raw, "Checkout failure", 2)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error, got title %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("editor artifact title: %v", err)
			}
			if got != test.want {
				t.Fatalf("unexpected title: got %q, want %q", got, test.want)
			}
		})
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
	diagnostics := dawgtypes.EmptyDiagnosticsSummary(dawgtypes.DiagnosticProfileSafe, "1")
	value := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,

		CreatedAt: mustTime(t, "2026-09-01T10:00:00Z"),
		Title:     "test artifact",
		Source:    dawgtypes.ManifestSource{Reporter: "qa", Environment: "test", RepoCommit: "abc123"},
		Layers: []dawgtypes.LayerSpec{{
			MediaType: dawgtypes.MediaTypeEnvironment,
			Digest:    fmt.Sprintf("sha256:%x", layerDigest),
			Size:      int64(len(layer)),
		}},
		Sanitize:        dawgtypes.ManifestSanitize{PolicyVersion: "1", FieldsRedacted: 0},
		Determinism:     dawgtypes.DeterminismConfig{ClockFrozenAt: mustTime(t, "2026-09-01T09:00:00Z"), RandomSeed: 1},
		ExpectedOutcome: dawgtypes.ExpectedOutcome{Type: "assertion", Description: "test", AssertionFile: "assertions/test.json"},
		Diagnostics:     &diagnostics,
	}
	logicalID, err := artifact.Identity(value)
	if err != nil {
		t.Fatal(err)
	}
	value.ID = logicalID
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
