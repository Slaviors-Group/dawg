package capture_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/capture"
)

func TestBrowserExtensionServerContractedTraceFiles(t *testing.T) {
	sessionDir := t.TempDir()
	server := &capture.ExtensionServer{
		ListenAddr: "127.0.0.1:0",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Start(ctx, sessionDir); err != nil {
		t.Fatalf("Start extension server: %v", err)
	}

	baseURL := "http://" + server.Addr()

	// Stream mock rrweb event
	rrwebData := []byte(`{"timestamp":1600000000,"type":2}`)
	resp, err := http.Post(baseURL+"/api/v1/stream/rrweb", "application/json", bytes.NewReader(rrwebData))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("post rrweb event failed: %v", err)
	}
	_ = resp.Body.Close()

	// Stream mock action event
	actionData := []byte(`{"type":"click","selector":"#checkout","timestamp":1600000001}`)
	resp, err = http.Post(baseURL+"/api/v1/stream/actions", "application/json", bytes.NewReader(actionData))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("post action event failed: %v", err)
	}
	_ = resp.Body.Close()

	// Stream mock frontend HTTP request
	httpData := []byte(`{"id":"req-1","timestamp":"2026-09-09T00:00:00Z","request":{"method":"GET","url":"http://api.local/data"},"response":{"status":200},"direction":"frontend-to-backend","durationMs":15}`)
	resp, err = http.Post(baseURL+"/api/v1/stream/http", "application/json", bytes.NewReader(httpData))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("post http event failed: %v", err)
	}
	_ = resp.Body.Close()

	if err := server.Stop(); err != nil {
		t.Fatalf("Stop extension server: %v", err)
	}

	assertJSONLFile(t, filepath.Join(sessionDir, "traces", "rrweb.jsonl"))
	assertJSONLFile(t, filepath.Join(sessionDir, "actions", "browser.jsonl"))
	assertJSONLFile(t, filepath.Join(sessionDir, "http", "frontend.jsonl"))
}

func assertJSONLFile(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("expected JSONL output in %s", path)
	}
	for _, line := range lines {
		var value map[string]any
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			t.Fatalf("decode JSONL %s: %v", path, err)
		}
	}
}
