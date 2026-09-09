package packager

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
)

func TestPackageCreatesValidOCIImageLayout(t *testing.T) {
	sessionDirectory := createSanitizedSession(t, true)
	outputDirectory := filepath.Join(t.TempDir(), "nested", "artifact")
	artifact, err := Package(packageRequest(sessionDirectory, outputDirectory))
	if err != nil {
		t.Fatalf("package sanitized session: %v", err)
	}
	if artifact.Directory != outputDirectory || artifact.OCIManifestDigest == "" {
		t.Fatalf("unexpected packaged artifact: %#v", artifact)
	}

	value, err := manifest.Read(artifact.ManifestPath)
	if err != nil {
		t.Fatalf("read DAWG manifest: %v", err)
	}
	if err := manifest.Validate(schemaPath(t), value); err != nil {
		t.Fatalf("validate DAWG manifest: %v", err)
	}
	if len(value.Layers) != 4 {
		t.Fatalf("expected four OCI layers, got %#v", value.Layers)
	}
	for _, layer := range value.Layers {
		assertBlobDigest(t, outputDirectory, layer.Digest, layer.Size)
	}

	indexContents, err := os.ReadFile(filepath.Join(outputDirectory, "index.json"))
	if err != nil {
		t.Fatalf("read OCI index: %v", err)
	}
	var index ociIndex
	if err := json.Unmarshal(indexContents, &index); err != nil {
		t.Fatalf("decode OCI index: %v", err)
	}
	if index.MediaType != ociImageIndexMediaType || len(index.Manifests) != 1 {
		t.Fatalf("unexpected OCI index: %#v", index)
	}
	if index.Manifests[0].Digest != artifact.OCIManifestDigest {
		t.Fatalf("OCI digest mismatch: index=%s result=%s", index.Manifests[0].Digest, artifact.OCIManifestDigest)
	}
	assertBlobDigest(t, outputDirectory, artifact.OCIManifestDigest, index.Manifests[0].Size)
}

func TestPackageRefusesDeniedSanitizationReport(t *testing.T) {
	sessionDirectory := createSanitizedSession(t, false)
	outputDirectory := filepath.Join(t.TempDir(), "artifact")
	_, err := Package(packageRequest(sessionDirectory, outputDirectory))
	if !errors.Is(err, dawgtypes.ErrExportBlocked) {
		t.Fatalf("expected export blocked error, got %v", err)
	}
	if _, err := os.Stat(outputDirectory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("denied export created output directory: %v", err)
	}
}

func TestPackageRefusesMissingSanitizationReport(t *testing.T) {
	sessionDirectory := t.TempDir()
	outputDirectory := filepath.Join(t.TempDir(), "artifact")
	_, err := Package(packageRequest(sessionDirectory, outputDirectory))
	if err == nil {
		t.Fatal("expected missing sanitization report error")
	}
	if _, err := os.Stat(outputDirectory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing report created output directory: %v", err)
	}
}

func TestBuildLayersIsDeterministicAndCompressedLayersRoundTrip(t *testing.T) {
	sessionDirectory := createSanitizedSession(t, true)
	first, err := buildLayers(sessionDirectory)
	if err != nil {
		t.Fatalf("build first layer set: %v", err)
	}
	second, err := buildLayers(sessionDirectory)
	if err != nil {
		t.Fatalf("build second layer set: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("layer count changed: %d != %d", len(first), len(second))
	}
	for index, layer := range first {
		if layer.spec != second[index].spec {
			t.Fatalf("layer %d was not deterministic: %#v != %#v", index, layer.spec, second[index].spec)
		}
		if layer.spec.MediaType == dawgtypes.MediaTypeEnvironment {
			continue
		}
		decoded, err := decompressZstd(layer.contents)
		if err != nil {
			t.Fatalf("decompress %s: %v", layer.spec.MediaType, err)
		}
		if len(decoded) == 0 {
			t.Fatalf("decompressed %s was empty", layer.spec.MediaType)
		}
	}
}

func createSanitizedSession(t *testing.T, exportAllowed bool) string {
	t.Helper()
	directory := t.TempDir()
	metadata := dawgtypes.CaptureMetadata{
		SessionID: "session-1",
		StartedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
		StoppedAt: time.Date(2026, 9, 1, 10, 5, 0, 0, time.UTC),
		Determinism: dawgtypes.DeterminismConfig{
			ClockFrozenAt: time.Date(2026, 9, 1, 9, 58, 0, 0, time.UTC),
			RandomSeed:    42,
		},
	}
	writeJSONFixture(t, filepath.Join(directory, "meta.json"), metadata)
	writeJSONFixture(t, filepath.Join(directory, "sanitize-report.json"), dawgtypes.SanitizeReport{
		PolicyVersion: "1.2.0",
		PolicyFile:    "policies/default.rego",
		OPAResult:     map[bool]string{true: "allow", false: "deny"}[exportAllowed],
		Redactions:    []dawgtypes.Redaction{},
		BlockedFields: []string{},
		ExportAllowed: exportAllowed,
	})
	writeTextFixture(t, filepath.Join(directory, "env", "lockfile.json"), "{\"backend\":\"sha256:pinned\"}\n")
	writeTextFixture(t, filepath.Join(directory, "db", "diff.jsonl"), "{\"table\":\"orders\"}\n")
	writeTextFixture(t, filepath.Join(directory, "traces", "rrweb.jsonl"), "{\"type\":2}\n")
	writeTextFixture(t, filepath.Join(directory, "http", "frontend.jsonl"), "{\"id\":\"request-1\"}\n")
	writeTextFixture(t, filepath.Join(directory, "logs", "structured.jsonl"), "{\"level\":\"info\"}\n")
	writeTextFixture(t, filepath.Join(directory, "cassettes", "thirdparty.jsonl"), "{\"id\":\"cassette-1\"}\n")
	return directory
}

func packageRequest(sessionDirectory, outputDirectory string) dawgtypes.PackageRequest {
	return dawgtypes.PackageRequest{
		SessionDirectory: sessionDirectory,
		OutputDirectory:  outputDirectory,
		SchemaPath:       schemaPathForPackage(),
		Title:            "checkout-negative-total",
		Source: dawgtypes.ManifestSource{
			Reporter:    "qa@example.com",
			Environment: "staging",
			RepoCommit:  "a1b2c3d",
		},
		ExpectedOutcome: dawgtypes.ExpectedOutcome{
			Type:          "assertion",
			Description:   "total must be non-negative",
			AssertionFile: "assertions/checkout.json",
		},
	}
}

func schemaPath(t *testing.T) string {
	t.Helper()
	return schemaPathForPackage()
}

func schemaPathForPackage() string {
	return filepath.Join("..", "..", "..", "schema", "manifest", "v0.1.3-alpha.json")
}

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	contents, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("serialize fixture %s: %v", path, err)
	}
	writeTextFixture(t, path, string(contents))
}

func writeTextFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create fixture parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

func assertBlobDigest(t *testing.T, directory, digest string, expectedSize int64) {
	t.Helper()
	digestPath, err := layerDigestPath(digest)
	if err != nil {
		t.Fatalf("parse digest %s: %v", digest, err)
	}
	contents, err := os.ReadFile(filepath.Join(directory, "blobs", "sha256", digestPath))
	if err != nil {
		t.Fatalf("read blob %s: %v", digest, err)
	}
	actual := sha256.Sum256(contents)
	if digest != fmt.Sprintf("sha256:%x", actual) || int64(len(contents)) != expectedSize {
		t.Fatalf("blob descriptor does not match content: %s", digest)
	}
}
