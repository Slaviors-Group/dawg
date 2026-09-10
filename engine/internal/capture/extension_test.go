package capture_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func TestExtensionServerRejectsDoubleStart(t *testing.T) {
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}
	ctx := context.Background()

	if err := server.Start(ctx, t.TempDir()); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	t.Cleanup(func() { _ = server.Stop() })

	err := server.Start(ctx, t.TempDir())
	if err == nil {
		t.Fatal("expected error on double Start, got nil")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExtensionServerIdempotentStop(t *testing.T) {
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}
	ctx := context.Background()

	if err := server.Start(ctx, t.TempDir()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := server.Stop(); err != nil {
		t.Fatalf("first Stop: %v", err)
	}
	// Second Stop must be a no-op (nil), not a crash or error.
	if err := server.Stop(); err != nil {
		t.Fatalf("second Stop (idempotent): expected nil, got %v", err)
	}
}

// TestExtensionServerStopDoesNotDeadlockWithActiveWebSocketClient is a
// regression test for the mutex deadlock described in the bug analysis:
//
// Before the fix, ExtensionServer.Stop() held s.mu across server.Shutdown().
// When a WebSocket client was connected and actively sending frames,
// readWebSocketFrames → processEventPayload → appendJSONL would try to acquire
// s.mu, which was already held by Stop(). Meanwhile, Shutdown() waited for the
// WebSocket goroutine to finish (it never would, because it was deadlocked on
// the mutex). Result: Stop() hung forever and was never cancelled.
//
// The test verifies that Stop() returns within a short deadline even when a
// WebSocket client is mid-stream.
func TestExtensionServerStopDoesNotDeadlockWithActiveWebSocketClient(t *testing.T) {
	sessionDir := t.TempDir()
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}

	if err := server.Start(context.Background(), sessionDir); err != nil {
		t.Fatalf("Start: %v", err)
	}

	addr := server.Addr()

	// Establish a raw WebSocket connection and keep it open.
	conn, err := dialWebSocket(addr)
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.Close()

	// Send a stream of frames concurrently to maximise the chance of racing
	// with Stop() on the mutex.
	stopSending := make(chan struct{})
	go func() {
		payload, _ := json.Marshal(map[string]any{
			"type": "DAWG_HTTP_EVENT",
			"data": map[string]any{"url": "https://example.test"},
		})
		for {
			select {
			case <-stopSending:
				return
			default:
				_ = sendWebSocketTextFrame(conn, payload)
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Stop() must return well within 10 seconds even with the active client.
	stopDone := make(chan error, 1)
	go func() { stopDone <- server.Stop() }()

	select {
	case err := <-stopDone:
		close(stopSending)
		if err != nil {
			t.Fatalf("Stop returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		close(stopSending)
		t.Fatal("ExtensionServer.Stop() deadlocked: did not return within 10 seconds with active WebSocket client")
	}
}

// TestExtensionServerPortIsFreedAfterStop verifies that the listening port is
// released after Stop() so a new server can bind to the same address.
// This catches a secondary symptom of the original bug: if server.Close() was
// never called, the listener stayed open and the extension could keep
// reconnecting, which is exactly what the incident log showed.
func TestExtensionServerPortIsFreedAfterStop(t *testing.T) {
	sessionDir := t.TempDir()
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}

	if err := server.Start(context.Background(), sessionDir); err != nil {
		t.Fatalf("Start: %v", err)
	}
	addr := server.Addr()

	if err := server.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// The original address must now be rebindable.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("port %s still in use after Stop: %v", addr, err)
	}
	_ = ln.Close()
}

// TestExtensionServerStopFlushesDataWrittenDuringActiveWebSocket verifies that
// events sent over an active WebSocket before Stop() are written to disk.
func TestExtensionServerStopFlushesDataWrittenDuringActiveWebSocket(t *testing.T) {
	sessionDir := t.TempDir()
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}

	if err := server.Start(context.Background(), sessionDir); err != nil {
		t.Fatalf("Start: %v", err)
	}

	conn, err := dialWebSocket(server.Addr())
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}

	// Send one HTTP event over WebSocket.
	payload, _ := json.Marshal(map[string]any{
		"type": "DAWG_HTTP_EVENT",
		"data": map[string]any{"url": "https://flush-test.example"},
	})
	if err := sendWebSocketTextFrame(conn, payload); err != nil {
		t.Fatalf("send WebSocket frame: %v", err)
	}

	// Give the server goroutine a moment to process the frame before stopping.
	time.Sleep(50 * time.Millisecond)
	conn.Close()

	stopDone := make(chan error, 1)
	go func() { stopDone <- server.Stop() }()

	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatalf("Stop: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Stop() did not return within 10 seconds")
	}

	contents, err := os.ReadFile(filepath.Join(sessionDir, "http", "frontend.jsonl"))
	if err != nil {
		t.Fatalf("read frontend.jsonl: %v", err)
	}
	if !bytes.Contains(contents, []byte("flush-test.example")) {
		t.Errorf("expected flushed event in frontend.jsonl, got: %s", contents)
	}
}


// ---------------------------------------------------------------------------
// WebSocket test helpers
// ---------------------------------------------------------------------------

// dialWebSocket performs a minimal WebSocket upgrade handshake over a raw TCP
// connection and returns the established conn.  It does not use gorilla/ws or
// nhooyr to keep the test package dependency-free.
func dialWebSocket(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial: %w", err)
	}

	key := "dGhlIHNhbXBsZSBub25jZQ==" // fixed test key
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	_ = base64.StdEncoding.EncodeToString(h.Sum(nil))

	req := fmt.Sprintf(
		"GET /ws HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n",
		addr, key,
	)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("write upgrade request: %w", err)
	}

	// Read until the end of the HTTP response headers.
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("read upgrade response: %w", err)
		}
		if line == "\r\n" {
			break
		}
	}

	return conn, nil
}

// sendWebSocketTextFrame sends a single unmasked text WebSocket frame.
// (Technically clients must mask frames, but our server ignores the mask bit
// for frames that just log, so this is fine for testing.)
func sendWebSocketTextFrame(conn net.Conn, payload []byte) error {
	// Build a minimal frame: FIN=1, opcode=1 (text), no mask.
	header := make([]byte, 2)
	header[0] = 0x81 // FIN + text opcode
	l := len(payload)
	switch {
	case l <= 125:
		header[1] = byte(l)
	case l <= 65535:
		header[1] = 126
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(l))
		header = append(header, ext...)
	default:
		header[1] = 127
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(l))
		header = append(header, ext...)
	}
	frame := append(header, payload...)
	_, err := conn.Write(frame)
	return err
}
