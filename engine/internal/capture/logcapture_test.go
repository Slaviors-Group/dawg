package capture

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureStructuredLogsCopiesJSONStream(t *testing.T) {
	directory := t.TempDir()
	entries, err := CaptureStructuredLogs(context.Background(), strings.NewReader("{\"level\":\"info\",\"message\":\"ready\"}\n"), directory, nil)
	if err != nil {
		t.Fatalf("capture structured logs: %v", err)
	}
	if entries != 1 {
		t.Fatalf("expected one entry, got %d", entries)
	}
	contents, err := os.ReadFile(filepath.Join(directory, "logs", "structured.jsonl"))
	if err != nil {
		t.Fatalf("read structured log output: %v", err)
	}
	if !strings.Contains(string(contents), `"message":"ready"`) {
		t.Fatalf("unexpected structured log output: %s", contents)
	}
}

func TestCaptureStructuredLogsRejectsNonJSONLine(t *testing.T) {
	_, err := CaptureStructuredLogs(context.Background(), strings.NewReader("not-json\n"), t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "decode structured log") {
		t.Fatalf("expected invalid log error, got %v", err)
	}
}

func TestTailStructuredLogFileCopiesAppendedLines(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "application.jsonl")
	context, cancel := context.WithCancel(context.Background())
	result := make(chan struct {
		entries int
		err     error
	}, 1)
	go func() {
		entries, err := TailStructuredLogFile(context, sourcePath, directory, 10*time.Millisecond, nil)
		result <- struct {
			entries int
			err     error
		}{entries, err}
	}()
	if err := os.WriteFile(sourcePath, []byte("{\"message\":\"first\"}\n"), 0o600); err != nil {
		t.Fatalf("append source log: %v", err)
	}
	waitForNonEmptyFile(t, filepath.Join(directory, "logs", "structured.jsonl"), time.Second)
	cancel()
	select {
	case captured := <-result:
		if captured.err != nil || captured.entries != 1 {
			t.Fatalf("unexpected file tail result: %#v", captured)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out stopping structured log tail")
	}
}
