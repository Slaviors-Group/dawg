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
	"io"
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

func TestExtensionServerAllowsExtensionPreflightAndRejectsWebPages(t *testing.T) {
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}
	if err := server.Start(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = server.Stop() })
	endpoint := "http://" + server.Addr() + "/api/v1/stream/event"

	preflight, err := http.NewRequest(http.MethodOptions, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	preflight.Header.Set("Origin", "chrome-extension://test-extension")
	response, err := http.DefaultClient.Do(preflight)
	if err != nil {
		t.Fatalf("extension preflight: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusNoContent || response.Header.Get("Access-Control-Allow-Origin") != "chrome-extension://test-extension" {
		t.Fatalf("unexpected preflight response: status=%d headers=%v", response.StatusCode, response.Header)
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(`{"type":"DAWG_KEEPALIVE"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", "https://untrusted.example")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("untrusted request: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("untrusted web page status = %d, want %d", response.StatusCode, http.StatusForbidden)
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

func TestExtensionServerWaitsForRecorderHandshake(t *testing.T) {
	sessionDir := t.TempDir()
	server := &capture.ExtensionServer{
		ListenAddr:       "127.0.0.1:0",
		TargetURL:        "http://localhost:3000",
		RequireHandshake: true,
		StartupTimeout:   2 * time.Second,
	}
	startResult := make(chan error, 1)
	go func() { startResult <- server.Start(context.Background(), sessionDir) }()

	addr := waitForExtensionAddress(t, server)
	conn, err := dialWebSocket(addr)
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.Close()

	ready, _ := json.Marshal(map[string]any{
		"type": "DAWG_EXTENSION_READY",
		"data": map[string]any{"isRecording": false},
	})
	if err := sendWebSocketTextFrame(conn, ready); err != nil {
		t.Fatalf("send extension ready: %v", err)
	}
	command, err := readWebSocketTextFrame(conn)
	if err != nil {
		t.Fatalf("read start command: %v", err)
	}
	if !bytes.Contains(command, []byte(`"type":"DAWG_COMMAND_START"`)) ||
		!bytes.Contains(command, []byte(`"targetUrl":"http://localhost:3000"`)) {
		t.Fatalf("unexpected start command: %s", command)
	}
	var startCommand struct {
		Data struct {
			SessionToken string `json:"sessionToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(command, &startCommand); err != nil || startCommand.Data.SessionToken == "" {
		t.Fatalf("start command omitted session token: %s (err=%v)", command, err)
	}
	sessionToken := startCommand.Data.SessionToken

	select {
	case err := <-startResult:
		t.Fatalf("Start returned before recorder acknowledgement: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	started, _ := json.Marshal(map[string]any{
		"type":         "DAWG_SESSION_START",
		"sessionToken": sessionToken,
		"data":         map[string]any{"tabId": 7, "targetUrl": "http://localhost:3000"},
	})
	if err := sendWebSocketTextFrame(conn, started); err != nil {
		t.Fatalf("send session start: %v", err)
	}
	select {
	case err := <-startResult:
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Start did not observe the recorder acknowledgement")
	}

	rogue, err := dialWebSocket(server.Addr())
	if err != nil {
		t.Fatalf("dial second extension profile: %v", err)
	}
	defer rogue.Close()
	rogueReady, _ := json.Marshal(map[string]any{
		"type": "DAWG_EXTENSION_READY",
		"data": map[string]any{"isRecording": false},
	})
	if err := sendWebSocketTextFrame(rogue, rogueReady); err != nil {
		t.Fatalf("send second profile ready: %v", err)
	}
	rogueEvent, _ := json.Marshal(map[string]any{
		"type":         "DAWG_RRWEB_EVENT",
		"sessionToken": sessionToken,
		"data":         map[string]any{"source": "unclaimed-profile"},
	})
	if err := sendWebSocketTextFrame(rogue, rogueEvent); err != nil {
		t.Fatalf("send second profile event: %v", err)
	}
	authorizedEvent, _ := json.Marshal(map[string]any{
		"type":         "DAWG_RRWEB_EVENT",
		"sessionToken": sessionToken,
		"data":         map[string]any{"source": "claimed-profile"},
	})
	if err := sendWebSocketTextFrame(conn, authorizedEvent); err != nil {
		t.Fatalf("send claimed profile event: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	clientResult := make(chan error, 1)
	go func() {
		message, err := readWebSocketTextFrame(conn)
		if err != nil {
			clientResult <- err
			return
		}
		if !bytes.Contains(message, []byte(`"type":"DAWG_COMMAND_STOP"`)) {
			clientResult <- fmt.Errorf("unexpected stop command: %s", message)
			return
		}
		stopped, _ := json.Marshal(map[string]any{"type": "DAWG_SESSION_STOP", "sessionToken": sessionToken, "data": map[string]any{}})
		clientResult <- sendWebSocketTextFrame(conn, stopped)
	}()
	if err := server.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := <-clientResult; err != nil {
		t.Fatalf("client stop handshake: %v", err)
	}
	contents, err := os.ReadFile(filepath.Join(sessionDir, "traces", "rrweb.jsonl"))
	if err != nil {
		t.Fatalf("read rrweb output: %v", err)
	}
	if !bytes.Contains(contents, []byte("claimed-profile")) || bytes.Contains(contents, []byte("unclaimed-profile")) {
		t.Fatalf("extension profile isolation failed: %s", contents)
	}
}

func TestExtensionServerSurfacesRecorderStartupError(t *testing.T) {
	server := &capture.ExtensionServer{
		ListenAddr:       "127.0.0.1:0",
		TargetURL:        "http://localhost:3000",
		RequireHandshake: true,
		StartupTimeout:   2 * time.Second,
	}
	startResult := make(chan error, 1)
	go func() { startResult <- server.Start(context.Background(), t.TempDir()) }()

	conn, err := dialWebSocket(waitForExtensionAddress(t, server))
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.Close()
	ready, _ := json.Marshal(map[string]any{"type": "DAWG_EXTENSION_READY", "data": map[string]any{}})
	if err := sendWebSocketTextFrame(conn, ready); err != nil {
		t.Fatalf("send extension ready: %v", err)
	}
	command, err := readWebSocketTextFrame(conn)
	if err != nil {
		t.Fatalf("read start command: %v", err)
	}
	var startCommand struct {
		Data struct {
			SessionToken string `json:"sessionToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(command, &startCommand); err != nil || startCommand.Data.SessionToken == "" {
		t.Fatalf("start command omitted token: %s (err=%v)", command, err)
	}
	failure, _ := json.Marshal(map[string]any{
		"type":         "DAWG_SESSION_ERROR",
		"sessionToken": startCommand.Data.SessionToken,
		"data":         map[string]any{"message": "target page refused extension injection"},
	})
	if err := sendWebSocketTextFrame(conn, failure); err != nil {
		t.Fatalf("send session error: %v", err)
	}

	select {
	case err := <-startResult:
		if err == nil || !strings.Contains(err.Error(), "target page refused extension injection") {
			t.Fatalf("unexpected startup result: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Start did not surface the extension startup error")
	}
	if server.Addr() != "" {
		t.Fatalf("failed Start left listener active at %s", server.Addr())
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

func TestExtensionSessionStopRequestsOwnerShutdownAndAcknowledges(t *testing.T) {
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}
	if err := server.Start(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = server.Stop() })

	stopRequested := server.StopRequested()
	conn, err := dialWebSocket(server.Addr())
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.Close()

	payload, _ := json.Marshal(map[string]any{
		"type": "DAWG_SESSION_STOP",
		"data": map[string]any{"stoppedAt": time.Now().UTC().Format(time.RFC3339Nano)},
	})
	if err := sendWebSocketTextFrame(conn, payload); err != nil {
		t.Fatalf("send stop frame: %v", err)
	}

	message, err := readWebSocketTextFrame(conn)
	if err != nil {
		t.Fatalf("read stop acknowledgement: %v", err)
	}
	if !bytes.Contains(message, []byte(`"type":"DAWG_SESSION_STOP_ACK"`)) {
		t.Fatalf("unexpected acknowledgement: %s", message)
	}
	select {
	case <-stopRequested:
	case <-time.After(time.Second):
		t.Fatal("SESSION_STOP did not request owner shutdown")
	}
}

func TestExtensionServerStopCommandsClientAndWaitsForDrain(t *testing.T) {
	server := &capture.ExtensionServer{ListenAddr: "127.0.0.1:0"}
	if err := server.Start(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	conn, err := dialWebSocket(server.Addr())
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.Close()

	clientResult := make(chan error, 1)
	go func() {
		message, err := readWebSocketTextFrame(conn)
		if err != nil {
			clientResult <- err
			return
		}
		if !bytes.Contains(message, []byte(`"type":"DAWG_COMMAND_STOP"`)) {
			clientResult <- fmt.Errorf("unexpected daemon command: %s", message)
			return
		}
		payload, _ := json.Marshal(map[string]any{"type": "DAWG_SESSION_STOP", "data": map[string]any{}})
		clientResult <- sendWebSocketTextFrame(conn, payload)
	}()

	startedAt := time.Now()
	if err := server.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed >= time.Second {
		t.Fatalf("Stop did not observe client drain; elapsed=%s", elapsed)
	}
	if err := <-clientResult; err != nil {
		t.Fatalf("client drain handshake: %v", err)
	}
}

// ---------------------------------------------------------------------------
// WebSocket test helpers
// ---------------------------------------------------------------------------

func waitForExtensionAddress(t *testing.T, server *capture.ExtensionServer) string {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if addr := server.Addr(); addr != "" {
			return addr
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("extension listener did not become ready")
	return ""
}

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

// sendWebSocketTextFrame sends a masked client text frame as required by RFC 6455.
func sendWebSocketTextFrame(conn net.Conn, payload []byte) error {
	mask := [4]byte{0x12, 0x34, 0x56, 0x78}
	header := make([]byte, 2)
	header[0] = 0x81 // FIN + text opcode
	l := len(payload)
	switch {
	case l <= 125:
		header[1] = 0x80 | byte(l)
	case l <= 65535:
		header[1] = 0x80 | 126
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(l))
		header = append(header, ext...)
	default:
		header[1] = 0x80 | 127
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(l))
		header = append(header, ext...)
	}
	header = append(header, mask[:]...)
	maskedPayload := make([]byte, len(payload))
	for index := range payload {
		maskedPayload[index] = payload[index] ^ mask[index%len(mask)]
	}
	frame := append(header, maskedPayload...)
	_, err := conn.Write(frame)
	return err
}

func readWebSocketTextFrame(conn net.Conn) ([]byte, error) {
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return nil, err
	}
	var header [2]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return nil, err
	}
	if header[0]&0x0f != 0x1 {
		return nil, fmt.Errorf("unexpected WebSocket opcode %d", header[0]&0x0f)
	}
	length := uint64(header[1] & 0x7f)
	switch length {
	case 126:
		var extended [2]byte
		if _, err := io.ReadFull(conn, extended[:]); err != nil {
			return nil, err
		}
		length = uint64(binary.BigEndian.Uint16(extended[:]))
	case 127:
		var extended [8]byte
		if _, err := io.ReadFull(conn, extended[:]); err != nil {
			return nil, err
		}
		length = binary.BigEndian.Uint64(extended[:])
	}
	payload := make([]byte, int(length))
	_, err := io.ReadFull(conn, payload)
	return payload, err
}
