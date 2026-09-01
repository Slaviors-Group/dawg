package capture

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureDBDiffsWritesValidatedEntries(t *testing.T) {
	directory := t.TempDir()
	source := strings.NewReader(`{"timestamp":"2026-09-01T10:00:00Z","operation":"UPDATE","table":"orders","primaryKey":{"id":"order-1"},"before":{"total":30},"after":{"total":-15},"query":"UPDATE orders","source":"orm-tap"}` + "\n")
	result, err := CaptureDBDiffs(context.Background(), source, directory, nil)
	if err != nil {
		t.Fatalf("capture DB diffs: %v", err)
	}
	if result.Entries != 1 || result.Skipped {
		t.Fatalf("unexpected capture result: %#v", result)
	}
	contents, err := os.ReadFile(filepath.Join(directory, "db", "diff.jsonl"))
	if err != nil {
		t.Fatalf("read DB output: %v", err)
	}
	if !strings.Contains(string(contents), `"operation":"UPDATE"`) {
		t.Fatalf("unexpected DB output: %s", contents)
	}
}

func TestCaptureDBDiffsSkipsWhenAdapterIsNotConfigured(t *testing.T) {
	result, err := CaptureDBDiffs(context.Background(), nil, t.TempDir(), nil)
	if err != nil {
		t.Fatalf("skip DB diffs: %v", err)
	}
	if !result.Skipped || result.Reason == "" {
		t.Fatalf("expected skipped result, got %#v", result)
	}
}

func TestCaptureDBDiffsRejectsMalformedEntry(t *testing.T) {
	_, err := CaptureDBDiffs(context.Background(), strings.NewReader("not-json\n"), t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "decode DB diff") {
		t.Fatalf("expected malformed DB diff error, got %v", err)
	}
}
