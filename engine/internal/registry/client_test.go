package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexedManifestDigestReturnsOnlyDescriptor(t *testing.T) {
	directory := t.TempDir()
	writeIndex(t, directory, `{"manifests":[{"digest":"sha256:abc"}]}`)
	digest, err := indexedManifestDigest(directory)
	if err != nil {
		t.Fatalf("read indexed manifest digest: %v", err)
	}
	if digest != "sha256:abc" {
		t.Fatalf("unexpected digest: %q", digest)
	}
}

func TestIndexedManifestDigestRejectsMultipleDescriptors(t *testing.T) {
	directory := t.TempDir()
	writeIndex(t, directory, `{"manifests":[{"digest":"sha256:a"},{"digest":"sha256:b"}]}`)
	_, err := indexedManifestDigest(directory)
	if err == nil || !strings.Contains(err.Error(), "one manifest") {
		t.Fatalf("expected index validation error, got %v", err)
	}
}

func writeIndex(t *testing.T, directory, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, "index.json"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write OCI index: %v", err)
	}
}
