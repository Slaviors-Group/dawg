package capture

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestCaptureDBDiffsHonorsContextCancellationOnBlockedReader is a regression
// test for the scanner.Scan() blocking bug.
//
// Before the fix, scanner.Scan() blocked in a syscall and ctx.Err() was only
// checked between completed lines. A Stop() call on a live DB pipe was silently
// ignored until the next line arrived — which might never happen.
//
// The test uses a pipe whose write end is never written to, so Scan() would
// block forever under the old code. Cancelling the context must unblock
// CaptureDBDiffs within a short deadline.
func TestCaptureDBDiffsHonorsContextCancellationOnBlockedReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pr, _ := io.Pipe() // write end intentionally never written to or closed

	done := make(chan error, 1)
	go func() {
		_, err := CaptureDBDiffs(ctx, pr, t.TempDir(), nil)
		done <- err
	}()

	// Give the goroutine a moment to reach Scan() before cancelling.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected context error, got nil")
		}
		if !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("expected context canceled error, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CaptureDBDiffs did not unblock within 5s after context cancel — scanner.Scan() is still blocking")
	}
}
