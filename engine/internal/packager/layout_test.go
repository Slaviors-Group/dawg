package packager

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	if len(value.Layers) != 6 {
		t.Fatalf("expected six OCI layers including diagnostics, got %#v", value.Layers)
	}
	for _, layer := range value.Layers {
		assertBlobDigest(t, outputDirectory, layer.Digest, layer.Size)
	}
	unpackedDirectory := t.TempDir()
	if err := Unpack(outputDirectory, unpackedDirectory); err != nil {
		t.Fatalf("unpack artifact: %v", err)
	}
	if contents, err := os.ReadFile(filepath.Join(unpackedDirectory, "actions", "browser.jsonl")); err != nil || !strings.Contains(string(contents), "click") {
		t.Fatalf("packaged artifact omitted browser actions: contents=%q err=%v", contents, err)
	}
	if contents, err := os.ReadFile(filepath.Join(unpackedDirectory, "diagnostics", "network.jsonl")); err != nil || !strings.Contains(string(contents), "request-1") {
		t.Fatalf("packaged artifact omitted diagnostic network evidence: contents=%q err=%v", contents, err)
	}
	evidence, err := ReadDiagnostics(outputDirectory)
	if err != nil || len(evidence.Console) != 1 || len(evidence.Network) != 1 || len(evidence.Bodies) != 1 || evidence.Device["userAgent"] != "fixture-agent" || evidence.Timeline == nil || evidence.Timeline.DurationMs != 2500 {
		t.Fatalf("read packaged diagnostics: evidence=%#v err=%v", evidence, err)
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

	reviewedDirectory := filepath.Join(t.TempDir(), "reviewed-artifact")
	reviewed, err := RepackageDiagnostics(DiagnosticRemovalRequest{InputDirectory: outputDirectory, OutputDirectory: reviewedDirectory, Categories: []string{"console", "network", "errors", "device", "bodies"}})
	if err != nil {
		t.Fatalf("repackage removed diagnostics: %v", err)
	}
	if reviewed.Directory != reviewedDirectory || reviewed.Directory == artifact.Directory {
		t.Fatalf("unexpected reviewed artifact: %#v", reviewed)
	}
	remaining, err := ReadDiagnostics(reviewedDirectory)
	if err != nil || len(remaining.Console) != 0 || len(remaining.Network) != 0 || len(remaining.Errors) != 0 || len(remaining.Device) != 0 || len(remaining.Bodies) != 0 {
		t.Fatalf("removed diagnostics remained: evidence=%#v err=%v", remaining, err)
	}
	original, err := ReadDiagnostics(outputDirectory)
	if err != nil || len(original.Network) != 1 {
		t.Fatalf("original artifact was changed: evidence=%#v err=%v", original, err)
	}
}

func TestRepackageDiagnosticsPreservesUnselectedEvidence(t *testing.T) {
	sessionDirectory := createSanitizedSession(t, true)
	input := filepath.Join(t.TempDir(), "artifact")
	if _, err := Package(packageRequest(sessionDirectory, input)); err != nil {
		t.Fatalf("package input: %v", err)
	}

	t.Run("remove console only", func(t *testing.T) {
		output := filepath.Join(t.TempDir(), "without-console")
		if _, err := RepackageDiagnostics(DiagnosticRemovalRequest{InputDirectory: input, OutputDirectory: output, Categories: []string{"console"}}); err != nil {
			t.Fatalf("repackage console: %v", err)
		}
		evidence, err := ReadDiagnostics(output)
		if err != nil {
			t.Fatalf("read reviewed evidence: %v", err)
		}
		if len(evidence.Console) != 0 || len(evidence.Network) != 1 || len(evidence.Errors) != 1 || len(evidence.Bodies) != 1 {
			t.Fatalf("unexpected reviewed evidence: %#v", evidence)
		}
	})

	t.Run("remove a single body reference", func(t *testing.T) {
		output := filepath.Join(t.TempDir(), "without-body")
		if _, err := RepackageDiagnostics(DiagnosticRemovalRequest{InputDirectory: input, OutputDirectory: output, BodyRefs: []string{"diagnostics/bodies/response-request-1.json"}}); err != nil {
			t.Fatalf("repackage body: %v", err)
		}
		evidence, err := ReadDiagnostics(output)
		if err != nil {
			t.Fatalf("read reviewed evidence: %v", err)
		}
		if len(evidence.Bodies) != 0 || len(evidence.Network) != 1 {
			t.Fatalf("unexpected body-removal evidence: %#v", evidence)
		}
		response, _ := evidence.Network[0]["response"].(map[string]any)
		body, _ := response["body"].(map[string]any)
		if body["state"] != "unavailable" || body["ref"] != nil {
			t.Fatalf("body reference was not rewritten: %#v", body)
		}
	})

	t.Run("reject existing output and invalid source OCI", func(t *testing.T) {
		if _, err := RepackageDiagnostics(DiagnosticRemovalRequest{InputDirectory: input, OutputDirectory: input, Categories: []string{"console"}}); err == nil {
			t.Fatal("expected existing input output rejection")
		}
		if _, err := RepackageDiagnostics(DiagnosticRemovalRequest{InputDirectory: input, OutputDirectory: t.TempDir(), Categories: []string{"console"}}); err == nil {
			t.Fatal("expected existing output rejection")
		}
		manifestValue, err := manifest.Read(filepath.Join(input, "dawg-manifest.json"))
		if err != nil {
			t.Fatalf("read input manifest: %v", err)
		}
		digestPath, err := layerDigestPath(manifestValue.Layers[0].Digest)
		if err != nil {
			t.Fatalf("parse fixture digest: %v", err)
		}
		if err := os.WriteFile(filepath.Join(input, "blobs", "sha256", digestPath), []byte("tampered"), 0o600); err != nil {
			t.Fatalf("tamper source blob: %v", err)
		}
		if _, err := RepackageDiagnostics(DiagnosticRemovalRequest{InputDirectory: input, OutputDirectory: filepath.Join(t.TempDir(), "rejected"), Categories: []string{"console"}}); err == nil {
			t.Fatal("expected invalid source OCI rejection")
		}
	})
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
	writeTextFixture(t, filepath.Join(directory, "traces", "rrweb.jsonl"), "{\"type\":2,\"timestamp\":1788256800000}\n{\"type\":3,\"timestamp\":1788256802500}\n")
	writeTextFixture(t, filepath.Join(directory, "actions", "browser.jsonl"), "{\"type\":\"click\",\"selector\":\"#checkout\"}\n")
	writeTextFixture(t, filepath.Join(directory, "http", "frontend.jsonl"), "{\"id\":\"request-1\"}\n")
	writeTextFixture(t, filepath.Join(directory, "logs", "structured.jsonl"), "{\"level\":\"info\"}\n")
	writeTextFixture(t, filepath.Join(directory, "cassettes", "thirdparty.jsonl"), "{\"id\":\"cassette-1\"}\n")
	writeTextFixture(t, filepath.Join(directory, "diagnostics", "profile.json"), "{\"profile\":\"safe\"}\n")
	writeTextFixture(t, filepath.Join(directory, "diagnostics", "console.jsonl"), "{\"id\":\"console-1\",\"source\":\"main-world\",\"text\":\"checkout failed\"}\n")
	writeTextFixture(t, filepath.Join(directory, "diagnostics", "network.jsonl"), "{\"id\":\"network-1\",\"source\":\"safe-extension\",\"requestId\":\"request-1\",\"request\":{\"body\":{\"state\":\"not-requested\"}},\"response\":{\"body\":{\"state\":\"captured\",\"ref\":\"diagnostics/bodies/response-request-1.json\",\"sha256\":\"sha256:fixture\",\"size\":23,\"contentType\":\"application/json\"}}}\n")
	writeTextFixture(t, filepath.Join(directory, "diagnostics", "errors.jsonl"), "{\"id\":\"error-1\",\"source\":\"main-world\",\"text\":\"checkout failed\"}\n")
	writeTextFixture(t, filepath.Join(directory, "diagnostics", "device.jsonl"), "{\"formatVersion\":\"1\",\"userAgent\":\"fixture-agent\",\"viewport\":{\"width\":1280,\"height\":720}}\n")
	writeTextFixture(t, filepath.Join(directory, "diagnostics", "bodies", "response-request-1.json"), "{\"message\":\"sanitized\"}\n")
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
	return filepath.Join("..", "..", "..", "schema", "manifest", "v0.3.5-middlechild.json")
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
