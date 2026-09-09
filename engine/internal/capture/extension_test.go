package capture_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/capture"
)

func TestExtensionServerCapturesStreamedData(t *testing.T) {
	sessionDir := t.TempDir()
	server := &capture.ExtensionServer{
		ListenAddr: "127.0.0.1:0",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Start(ctx, sessionDir); err != nil {
		t.Fatalf("ExtensionServer.Start failed: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	addr := server.Addr()
	if addr == "" {
		t.Fatalf("expected non-empty server address")
	}

	baseURL := "http://" + addr

	// 1. Post RRWeb stream
	rrwebPayload := []byte(`{"type":1,"data":{"href":"http://localhost"}}`)
	resp, err := http.Post(baseURL+"/api/v1/stream/rrweb", "application/json", bytes.NewReader(rrwebPayload))
	if err != nil {
		t.Fatalf("failed to post rrweb: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 2. Post Event envelope
	envelopePayload, _ := json.Marshal(map[string]any{
		"type": "DAWG_ACTION_EVENT",
		"data": map[string]any{"type": "click", "selector": "#btn"},
	})
	resp, err = http.Post(baseURL+"/api/v1/stream/event", "application/json", bytes.NewReader(envelopePayload))
	if err != nil {
		t.Fatalf("failed to post event: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 3. Stop server
	if err := server.Stop(); err != nil {
		t.Fatalf("ExtensionServer.Stop failed: %v", err)
	}

	// Verify rrweb.jsonl content
	rrwebContent, err := os.ReadFile(filepath.Join(sessionDir, "traces", "rrweb.jsonl"))
	if err != nil {
		t.Fatalf("failed to read rrweb.jsonl: %v", err)
	}
	if !bytes.Contains(rrwebContent, []byte(`"href":"http://localhost"`)) {
		t.Errorf("unexpected rrweb file content: %s", string(rrwebContent))
	}

	// Verify browser.jsonl content
	actionsContent, err := os.ReadFile(filepath.Join(sessionDir, "actions", "browser.jsonl"))
	if err != nil {
		t.Fatalf("failed to read browser.jsonl: %v", err)
	}
	if !bytes.Contains(actionsContent, []byte(`"#btn"`)) {
		t.Errorf("unexpected actions file content: %s", string(actionsContent))
	}
}

