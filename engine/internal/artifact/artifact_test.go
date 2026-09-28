package artifact

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/klauspost/compress/zstd"
)

func TestDeriveReviewPreservesEvidenceBlobs(t *testing.T) {
	sourceDirectory := buildReviewableArtifact(t, filepath.Join(t.TempDir(), "source"))
	source, err := Validate(sourceDirectory)
	if err != nil {
		t.Fatalf("validate source: %v", err)
	}
	originalBlob, err := os.ReadFile(filepath.Join(sourceDirectory, "blobs", "sha256", source.Layers[0].Digest[7:]))
	if err != nil {
		t.Fatal(err)
	}
	flags := []manifest.ReviewFlag{{ID: "11111111111111111111111111111111", StartOffsetMs: 100, Title: "Visible failure", Category: "bug", Severity: "error"}}
	review := NewReview(source, flags, time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC))
	outputDirectory := filepath.Join(t.TempDir(), "review")
	artifactTitle := "Checkout failure review"
	_, derived, err := DeriveReview(DeriveReviewRequest{InputDirectory: sourceDirectory, OutputDirectory: outputDirectory, ArtifactTitle: artifactTitle, Review: review})
	if err != nil {
		t.Fatalf("derive review: %v", err)
	}
	if derived.ID == source.ID || derived.Title != artifactTitle || derived.Review == nil || derived.Review.ParentArtifactID != source.ID {
		t.Fatalf("unexpected derived manifest: %#v", derived)
	}
	if _, err := Validate(outputDirectory); err != nil {
		t.Fatalf("validate derived artifact: %v", err)
	}
	derivedBlob, err := os.ReadFile(filepath.Join(outputDirectory, "blobs", "sha256", source.Layers[0].Digest[7:]))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(originalBlob, derivedBlob) {
		t.Fatal("review derivation changed an evidence blob")
	}
	unchangedSource, err := Validate(sourceDirectory)
	if err != nil || unchangedSource.Title != source.Title || unchangedSource.Review != nil {
		t.Fatalf("source artifact was modified: %#v, %v", unchangedSource, err)
	}
}

func TestTraceDurationAcceptsJSONLRecordOverFourMiB(t *testing.T) {
	var trace bytes.Buffer
	trace.WriteString(`{"timestamp":1000,"data":"`)
	trace.WriteString(strings.Repeat("x", (4<<20)+1))
	trace.WriteString(`"}` + "\n")
	trace.WriteString(`{"timestamp":1500}` + "\n")
	if trace.Len() >= dawgtypes.MaxJSONLRecordBytes {
		t.Fatalf("test record exceeds shared JSONL limit: %d", trace.Len())
	}

	duration, err := traceDuration(&trace)
	if err != nil {
		t.Fatalf("trace duration: %v", err)
	}
	if duration != 500 {
		t.Fatalf("unexpected duration: %d", duration)
	}
}

func TestTraceDurationAcceptsRecordAtSharedLimit(t *testing.T) {
	prefix := `{"timestamp":1000,"data":"`
	suffix := `"}`
	first := prefix + strings.Repeat("x", dawgtypes.MaxJSONLRecordBytes-len(prefix)-len(suffix)) + suffix
	if len(first) != dawgtypes.MaxJSONLRecordBytes {
		t.Fatalf("test record has unexpected size: %d", len(first))
	}

	duration, err := traceDuration(strings.NewReader(first + "\n{\"timestamp\":1500}\n"))
	if err != nil {
		t.Fatalf("trace duration: %v", err)
	}
	if duration != 500 {
		t.Fatalf("unexpected duration: %d", duration)
	}
}

func buildReviewableArtifact(t *testing.T, directory string) string {
	t.Helper()
	traceLayer := reviewTraceLayer(t)
	traceDigest := sha256.Sum256(traceLayer)
	diagnostics := dawgtypes.EmptyDiagnosticsSummary(dawgtypes.DiagnosticProfileSafe, "1")
	value := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		CreatedAt:     time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
		Title:         "reviewable artifact",
		Source:        dawgtypes.ManifestSource{Reporter: "qa", Environment: "test", RepoCommit: "abc123"},
		Layers: []dawgtypes.LayerSpec{{
			MediaType: dawgtypes.MediaTypeTrace,
			Digest:    fmt.Sprintf("sha256:%x", traceDigest),
			Size:      int64(len(traceLayer)),
		}},
		Sanitize:        dawgtypes.ManifestSanitize{PolicyVersion: "1", FieldsRedacted: 0},
		Determinism:     dawgtypes.DeterminismConfig{ClockFrozenAt: time.Date(2026, time.September, 1, 9, 0, 0, 0, time.UTC), RandomSeed: 1},
		ExpectedOutcome: dawgtypes.ExpectedOutcome{Type: "assertion", Description: "test", AssertionFile: "assertions/test.json"},
		Diagnostics:     &diagnostics,
	}
	identity, err := Identity(value)
	if err != nil {
		t.Fatal(err)
	}
	value.ID = identity
	config, err := manifest.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	configDescriptor := descriptor(dawgtypes.MediaTypeManifest, config)
	layers := []Descriptor{{MediaType: string(dawgtypes.MediaTypeTrace), Digest: value.Layers[0].Digest, Size: value.Layers[0].Size}}
	imageBytes, err := json.Marshal(imageManifest{SchemaVersion: 2, MediaType: ociImageManifestMediaType, Config: configDescriptor, Layers: layers})
	if err != nil {
		t.Fatal(err)
	}
	imageDescriptor := descriptor(dawgtypes.MediaType(ociImageManifestMediaType), imageBytes)
	indexBytes, err := json.Marshal(index{SchemaVersion: 2, MediaType: ociImageIndexMediaType, Manifests: []Descriptor{imageDescriptor}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "blobs", "sha256"), 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(path string, contents []byte) {
		if err := os.WriteFile(path, contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(directory, "dawg-manifest.json"), config)
	write(filepath.Join(directory, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`))
	write(filepath.Join(directory, "index.json"), indexBytes)
	write(filepath.Join(directory, "blobs", "sha256", traceDigestString(traceDigest)), traceLayer)
	write(filepath.Join(directory, "blobs", "sha256", configDescriptor.Digest[7:]), config)
	write(filepath.Join(directory, "blobs", "sha256", imageDescriptor.Digest[7:]), imageBytes)
	return directory
}

func reviewTraceLayer(t *testing.T) []byte {
	t.Helper()
	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	contents := []byte("{\"timestamp\":1000}\n{\"timestamp\":1500}\n")
	if err := writer.WriteHeader(&tar.Header{Name: "traces/rrweb.jsonl", Mode: 0o600, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	encoder, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := encoder.Write(archive.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatal(err)
	}
	return compressed.Bytes()
}

func traceDigestString(digest [sha256.Size]byte) string {
	return fmt.Sprintf("%x", digest)
}
