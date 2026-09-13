package capture

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	wsReadTimeout    = 60 * time.Second
	wsWriteTimeout   = 2 * time.Second
	wsDrainTimeout   = 5 * time.Second
	maxStreamBytes   = 16 << 20
	webSocketGUID    = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	wsOpcodeContinue = byte(0x0)
	wsOpcodeText     = byte(0x1)
	wsOpcodeBinary   = byte(0x2)
	wsOpcodeClose    = byte(0x8)
	wsOpcodePing     = byte(0x9)
	wsOpcodePong     = byte(0xA)
)

type webSocketClient struct {
	conn    net.Conn
	writer  *bufio.ReadWriter
	writeMu sync.Mutex
}

type streamTarget uint8

const (
	streamRRWeb streamTarget = iota
	streamActions
	streamHTTP
)

func (client *webSocketClient) writeFrame(opcode byte, payload []byte) error {
	client.writeMu.Lock()
	defer client.writeMu.Unlock()

	if err := client.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout)); err != nil {
		return err
	}
	header := []byte{0x80 | opcode}
	switch length := len(payload); {
	case length <= 125:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126, byte(length>>8), byte(length))
	default:
		header = append(header, 127)
		var extended [8]byte
		binary.BigEndian.PutUint64(extended[:], uint64(length))
		header = append(header, extended[:]...)
	}
	if _, err := client.writer.Write(header); err != nil {
		return err
	}
	if _, err := client.writer.Write(payload); err != nil {
		return err
	}
	return client.writer.Flush()
}

func (client *webSocketClient) writeMessage(messageType string, data any) error {
	payload, err := json.Marshal(map[string]any{"type": messageType, "data": data})
	if err != nil {
		return err
	}
	return client.writeFrame(wsOpcodeText, payload)
}

// ExtensionServer receives streaming DOM events, browser action traces, and
// frontend HTTP traffic from the DAWG browser extension.
type ExtensionServer struct {
	ListenPort       int
	ListenAddr       string
	TargetURL        string
	RequireHandshake bool
	StartupTimeout   time.Duration

	mu             sync.Mutex
	listener       net.Listener
	server         *http.Server
	serveDone      chan struct{}
	stopDone       chan struct{}
	stopErr        error
	stopping       bool
	connections    map[*webSocketClient]struct{}
	handlers       sync.WaitGroup
	stopRequests   chan struct{}
	clientDrained  chan struct{}
	sessionStarted chan struct{}
	startupErrors  chan error
	controller     *webSocketClient
	sessionReady   bool
	sessionToken   string
	rrwebBuf       *bufio.Writer
	actionsBuf     *bufio.Writer
	httpBuf        *bufio.Writer
	rrwebFile      *os.File
	actionsFile    *os.File
	httpFile       *os.File
	actualAddr     string
	active         atomic.Int32
}

func (s *ExtensionServer) Name() string { return "browser" }

func (s *ExtensionServer) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.actualAddr
}

// StopRequested returns a signal that fires when the extension sends its final
// DAWG_SESSION_STOP envelope. The channel is buffered so a stop received just
// after Start cannot be lost before the session coordinator begins waiting.
func (s *ExtensionServer) StopRequested() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopRequests == nil {
		s.stopRequests = make(chan struct{}, 1)
	}
	return s.stopRequests
}

func (s *ExtensionServer) Start(ctx context.Context, directory string) error {
	s.mu.Lock()

	if s.server != nil || s.stopping {
		s.mu.Unlock()
		return fmt.Errorf("capture: extension server already running")
	}
	if s.RequireHandshake && strings.TrimSpace(s.TargetURL) == "" {
		s.mu.Unlock()
		return fmt.Errorf("capture: target URL is required for extension recording")
	}
	if s.stopRequests == nil {
		s.stopRequests = make(chan struct{}, 1)
	} else {
		drainSignal(s.stopRequests)
	}
	s.clientDrained = make(chan struct{}, 1)
	s.sessionStarted = make(chan struct{}, 1)
	s.startupErrors = make(chan error, 1)
	s.stopDone = make(chan struct{})
	s.stopErr = nil
	s.controller = nil
	s.sessionReady = false
	s.sessionToken = ""
	if s.RequireHandshake {
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("capture: generate extension session token: %w", err)
		}
		s.sessionToken = base64.RawURLEncoding.EncodeToString(tokenBytes)
	}
	s.connections = make(map[*webSocketClient]struct{})
	s.handlers = sync.WaitGroup{}

	for _, child := range []string{"traces", "http", "actions"} {
		if err := os.MkdirAll(filepath.Join(directory, child), 0o700); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("capture: create output directory %s: %w", child, err)
		}
	}

	var err error
	s.rrwebFile, err = os.OpenFile(filepath.Join(directory, "traces", "rrweb.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("capture: open rrweb.jsonl: %w", err)
	}
	s.actionsFile, err = os.OpenFile(filepath.Join(directory, "actions", "browser.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		_ = s.rrwebFile.Close()
		s.rrwebFile = nil
		s.mu.Unlock()
		return fmt.Errorf("capture: open browser.jsonl: %w", err)
	}
	s.httpFile, err = os.OpenFile(filepath.Join(directory, "http", "frontend.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		_ = s.rrwebFile.Close()
		_ = s.actionsFile.Close()
		s.rrwebFile = nil
		s.actionsFile = nil
		s.mu.Unlock()
		return fmt.Errorf("capture: open frontend.jsonl: %w", err)
	}
	s.rrwebBuf = bufio.NewWriter(s.rrwebFile)
	s.actionsBuf = bufio.NewWriter(s.actionsFile)
	s.httpBuf = bufio.NewWriter(s.httpFile)

	addr := s.ListenAddr
	if addr == "" {
		port := s.ListenPort
		if port <= 0 {
			port = 8082
		}
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		_ = s.closeOutputsLocked()
		s.mu.Unlock()
		return fmt.Errorf("capture: start extension listener on %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.Handle("/api/v1/stream/rrweb", extensionCORS(http.HandlerFunc(s.handleStreamRRWeb)))
	mux.Handle("/api/v1/stream/actions", extensionCORS(http.HandlerFunc(s.handleStreamActions)))
	mux.Handle("/api/v1/stream/http", extensionCORS(http.HandlerFunc(s.handleStreamHTTP)))
	mux.Handle("/api/v1/stream/event", extensionCORS(http.HandlerFunc(s.handleStreamEvent)))

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	serveDone := make(chan struct{})
	s.listener = listener
	s.server = server
	s.serveDone = serveDone
	s.actualAddr = listener.Addr().String()
	s.active.Store(1)

	go func() {
		defer close(serveDone)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			log.Printf("[warn] ExtensionServer: serve failed: %v", err)
		}
	}()

	actualAddr := s.actualAddr
	requireHandshake := s.RequireHandshake
	startupTimeout := s.StartupTimeout
	sessionStarted := s.sessionStarted
	startupErrors := s.startupErrors
	s.mu.Unlock()

	log.Printf("[debug] ExtensionServer.Start: listening on %s", actualAddr)
	if !requireHandshake {
		return nil
	}
	if startupTimeout <= 0 {
		startupTimeout = 12 * time.Second
	}
	timer := time.NewTimer(startupTimeout)
	defer timer.Stop()

	var startErr error
	select {
	case <-sessionStarted:
		log.Printf("[debug] ExtensionServer.Start: extension is recording target %s", s.TargetURL)
		return nil
	case err := <-startupErrors:
		startErr = fmt.Errorf("capture: browser extension could not start recording: %w", err)
	case <-ctx.Done():
		startErr = fmt.Errorf("capture: wait for browser extension: %w", ctx.Err())
	case <-timer.C:
		startErr = fmt.Errorf("capture: browser extension did not connect and start within %s; install or reload the DAWG extension and keep Chrome/Edge open", startupTimeout)
	}
	if stopErr := s.Stop(); stopErr != nil {
		startErr = errors.Join(startErr, stopErr)
	}
	return startErr
}

// Stop asks connected extensions to stop their producers, waits briefly for
// their ordered DAWG_SESSION_STOP acknowledgement, then closes every listener
// and hijacked connection before flushing capture files.
func (s *ExtensionServer) Stop() error {
	s.mu.Lock()
	if s.server == nil {
		s.mu.Unlock()
		return nil
	}
	if s.stopping {
		done := s.stopDone
		s.mu.Unlock()
		if done != nil {
			<-done
		}
		s.mu.Lock()
		err := s.stopErr
		s.mu.Unlock()
		return err
	}

	s.stopping = true
	server := s.server
	listener := s.listener
	serveDone := s.serveDone
	stopDone := s.stopDone
	clientDrained := s.clientDrained
	clients := s.snapshotCaptureClientsLocked()
	s.mu.Unlock()

	log.Printf("[debug] ExtensionServer.Stop: draining %d WebSocket client(s)", len(clients))
	if len(clients) > 0 && !signalReady(clientDrained) {
		for _, client := range clients {
			_ = client.writeMessage("DAWG_COMMAND_STOP", nil)
		}
		timer := time.NewTimer(wsDrainTimeout)
		select {
		case <-clientDrained:
		case <-timer.C:
			log.Printf("[warn] ExtensionServer.Stop: extension drain timed out after %s", wsDrainTimeout)
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}

	s.mu.Lock()
	s.active.Store(0)
	clients = s.snapshotClientsLocked()
	s.mu.Unlock()

	var stopErr error
	if listener != nil {
		// Serve and Shutdown may win the close race. Either result means the
		// listener is no longer accepting connections, so Close is idempotent
		// lifecycle cleanup rather than an error to surface to callers.
		_ = listener.Close()
	}
	for _, client := range clients {
		_ = client.writeFrame(wsOpcodeClose, []byte{0x03, 0xE8})
		_ = client.conn.Close()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
		stopErr = errors.Join(stopErr, err)
		_ = server.Close()
	}
	cancel()
	if serveDone != nil {
		<-serveDone
	}
	s.handlers.Wait()

	s.mu.Lock()
	stopErr = errors.Join(stopErr, s.closeOutputsLocked())
	s.server = nil
	s.listener = nil
	s.serveDone = nil
	s.connections = nil
	s.controller = nil
	s.sessionReady = false
	s.sessionToken = ""
	s.actualAddr = ""
	s.stopping = false
	s.stopErr = stopErr
	if stopDone != nil {
		close(stopDone)
	}
	s.mu.Unlock()

	log.Printf("[debug] ExtensionServer.Stop: shutdown complete")
	return stopErr
}

func (s *ExtensionServer) snapshotClientsLocked() []*webSocketClient {
	clients := make([]*webSocketClient, 0, len(s.connections))
	for client := range s.connections {
		clients = append(clients, client)
	}
	return clients
}

func (s *ExtensionServer) snapshotCaptureClientsLocked() []*webSocketClient {
	if s.controller != nil {
		return []*webSocketClient{s.controller}
	}
	if !s.RequireHandshake {
		return s.snapshotClientsLocked()
	}
	return nil
}

func (s *ExtensionServer) closeOutputsLocked() error {
	var closeErr error
	for _, pair := range []struct {
		buf  *bufio.Writer
		file *os.File
	}{
		{s.rrwebBuf, s.rrwebFile},
		{s.actionsBuf, s.actionsFile},
		{s.httpBuf, s.httpFile},
	} {
		if pair.buf != nil {
			closeErr = errors.Join(closeErr, pair.buf.Flush())
		}
		if pair.file != nil {
			closeErr = errors.Join(closeErr, pair.file.Sync())
			closeErr = errors.Join(closeErr, pair.file.Close())
		}
	}
	s.rrwebBuf = nil
	s.actionsBuf = nil
	s.httpBuf = nil
	s.rrwebFile = nil
	s.actionsFile = nil
	s.httpFile = nil
	return closeErr
}

func (s *ExtensionServer) appendJSONL(target streamTarget, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var buf *bufio.Writer
	switch target {
	case streamRRWeb:
		buf = s.rrwebBuf
	case streamActions:
		buf = s.actionsBuf
	case streamHTTP:
		buf = s.httpBuf
	default:
		return fmt.Errorf("capture: unknown extension stream")
	}
	if s.active.Load() == 0 || buf == nil {
		return fmt.Errorf("capture: extension output is closed")
	}
	if len(data) == 0 {
		return nil
	}
	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	_, err := buf.Write(data)
	return err
}

func (s *ExtensionServer) handleStreamRRWeb(w http.ResponseWriter, r *http.Request) {
	s.handleRawStream(w, r, streamRRWeb)
}

func (s *ExtensionServer) handleStreamActions(w http.ResponseWriter, r *http.Request) {
	s.handleRawStream(w, r, streamActions)
}

func (s *ExtensionServer) handleStreamHTTP(w http.ResponseWriter, r *http.Request) {
	s.handleRawStream(w, r, streamHTTP)
}

func (s *ExtensionServer) handleRawStream(w http.ResponseWriter, r *http.Request, target streamTarget) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := readLimitedBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.appendJSONL(target, body); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *ExtensionServer) handleStreamEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := readLimitedBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.processEventPayload(body, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func extensionCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !strings.HasPrefix(origin, "chrome-extension://") {
				http.Error(w, "capture: request origin is not a browser extension", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func readLimitedBody(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxStreamBytes+1))
	if err != nil {
		return nil, fmt.Errorf("capture: read extension event: %w", err)
	}
	if len(body) > maxStreamBytes {
		return nil, fmt.Errorf("capture: extension event exceeds %d bytes", maxStreamBytes)
	}
	return body, nil
}

type streamEnvelope struct {
	Type         string          `json:"type"`
	Data         json.RawMessage `json:"data"`
	SessionToken string          `json:"sessionToken"`
}

func (s *ExtensionServer) processEventPayload(rawPayload []byte, client *webSocketClient) error {
	var envelope streamEnvelope
	if err := json.Unmarshal(rawPayload, &envelope); err != nil {
		return fmt.Errorf("capture: decode extension event: %w", err)
	}
	switch envelope.Type {
	case "DAWG_EXTENSION_READY":
		if client == nil {
			return fmt.Errorf("capture: extension ready handshake requires WebSocket transport")
		}
		var ready struct {
			IsRecording  bool   `json:"isRecording"`
			SessionToken string `json:"sessionToken"`
		}
		_ = json.Unmarshal(envelope.Data, &ready)
		s.mu.Lock()
		if s.sessionReady && ready.IsRecording && ready.SessionToken == s.sessionToken {
			s.controller = client
		}
		if !s.sessionReady && s.controller == nil {
			s.controller = client
		}
		shouldStart := s.RequireHandshake && !s.sessionReady && s.controller == client
		targetURL := s.TargetURL
		sessionToken := s.sessionToken
		s.mu.Unlock()
		if shouldStart {
			log.Printf("[debug] ExtensionServer: commanding extension to record %s", targetURL)
			return client.writeMessage("DAWG_COMMAND_START", map[string]string{"targetUrl": targetURL, "sessionToken": sessionToken})
		}
		return nil
	case "DAWG_RRWEB_EVENT":
		if !s.validSessionEnvelope(envelope, client) {
			return nil
		}
		return s.appendJSONL(streamRRWeb, envelope.Data)
	case "DAWG_ACTION_EVENT":
		if !s.validSessionEnvelope(envelope, client) {
			return nil
		}
		return s.appendJSONL(streamActions, envelope.Data)
	case "DAWG_HTTP_EVENT":
		if !s.validSessionEnvelope(envelope, client) {
			return nil
		}
		return s.appendJSONL(streamHTTP, envelope.Data)
	case "DAWG_SESSION_START":
		if !s.validSessionEnvelope(envelope, client) {
			return nil
		}
		var details struct {
			TargetURL string `json:"targetUrl"`
		}
		_ = json.Unmarshal(envelope.Data, &details)
		if s.RequireHandshake && details.TargetURL != s.TargetURL {
			if client != nil {
				_ = client.writeMessage("DAWG_COMMAND_STOP", nil)
			}
			s.signalStartupError(fmt.Errorf("extension acknowledged target %q instead of %q", details.TargetURL, s.TargetURL))
			return nil
		}
		log.Printf("[debug] ExtensionServer: session start handshake received")
		s.mu.Lock()
		if client != nil {
			s.controller = client
		}
		s.sessionReady = true
		started := s.sessionStarted
		s.mu.Unlock()
		nonBlockingSignal(started)
		return nil
	case "DAWG_SESSION_ERROR":
		if !s.validSessionEnvelope(envelope, client) {
			return nil
		}
		var details struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(envelope.Data, &details); err != nil || strings.TrimSpace(details.Message) == "" {
			details.Message = "unknown extension error"
		}
		s.signalStartupError(errors.New(details.Message))
		s.signalClientDrained()
		return nil
	case "DAWG_SESSION_STOP":
		if !s.validSessionEnvelope(envelope, client) {
			return nil
		}
		log.Printf("[debug] ExtensionServer: session stop handshake received")
		if client != nil {
			_ = client.writeMessage("DAWG_SESSION_STOP_ACK", nil)
		}
		s.signalClientDrained()
		s.signalStopRequested()
		return nil
	case "DAWG_KEEPALIVE":
		return nil
	default:
		return fmt.Errorf("capture: unknown extension event type %q", envelope.Type)
	}
}

func (s *ExtensionServer) validSessionEnvelope(envelope streamEnvelope, client *webSocketClient) bool {
	if !s.RequireHandshake {
		return true
	}
	s.mu.Lock()
	validToken := envelope.SessionToken != "" && envelope.SessionToken == s.sessionToken
	validClient := client == nil || s.controller == nil || client == s.controller
	s.mu.Unlock()
	if !validToken || !validClient {
		log.Printf("[warn] ExtensionServer: ignored %s from an unclaimed extension session", envelope.Type)
		return false
	}
	return true
}

func (s *ExtensionServer) signalStartupError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	startupErrors := s.startupErrors
	s.mu.Unlock()
	select {
	case startupErrors <- err:
	default:
	}
}

func (s *ExtensionServer) signalClientDrained() {
	s.mu.Lock()
	channel := s.clientDrained
	s.mu.Unlock()
	nonBlockingSignal(channel)
}

func (s *ExtensionServer) signalStopRequested() {
	s.mu.Lock()
	channel := s.stopRequests
	s.mu.Unlock()
	nonBlockingSignal(channel)
}

func nonBlockingSignal(channel chan struct{}) {
	if channel == nil {
		return
	}
	select {
	case channel <- struct{}{}:
	default:
	}
}

func signalReady(channel <-chan struct{}) bool {
	if channel == nil {
		return false
	}
	select {
	case <-channel:
		return true
	default:
		return false
	}
}

func drainSignal(channel chan struct{}) {
	for {
		select {
		case <-channel:
		default:
			return
		}
	}
}

func (s *ExtensionServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && !strings.HasPrefix(origin, "chrome-extension://") {
		http.Error(w, "capture: WebSocket origin is not a browser extension", http.StatusForbidden)
		return
	}
	secKey := r.Header.Get("Sec-WebSocket-Key")
	if secKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return
	}
	client := &webSocketClient{conn: conn, writer: buf}
	defer conn.Close()

	hash := sha1.Sum([]byte(secKey + webSocketGUID))
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + base64.StdEncoding.EncodeToString(hash[:]) + "\r\n\r\n"
	if _, err := buf.WriteString(response); err != nil {
		return
	}
	if err := buf.Flush(); err != nil {
		return
	}

	s.mu.Lock()
	if s.active.Load() == 0 || s.stopping {
		s.mu.Unlock()
		return
	}
	s.connections[client] = struct{}{}
	s.handlers.Add(1)
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.connections, client)
		if s.controller == client {
			s.controller = nil
		}
		var drained chan struct{}
		if s.stopping && len(s.connections) == 0 {
			drained = s.clientDrained
		}
		s.mu.Unlock()
		nonBlockingSignal(drained)
		s.handlers.Done()
	}()

	log.Printf("[debug] ExtensionServer: WebSocket client connected from %s", r.RemoteAddr)
	s.readWebSocketFrames(buf, client)
	log.Printf("[debug] ExtensionServer: WebSocket client disconnected from %s", r.RemoteAddr)
}

func (s *ExtensionServer) readWebSocketFrames(reader *bufio.ReadWriter, client *webSocketClient) {
	var fragmented []byte
	var fragmentedOpcode byte
	for {
		if err := client.conn.SetReadDeadline(time.Now().Add(wsReadTimeout)); err != nil {
			return
		}
		fin, opcode, payload, err := readClientFrame(reader)
		if err != nil {
			return
		}

		switch opcode {
		case wsOpcodeClose:
			_ = client.writeFrame(wsOpcodeClose, payload)
			return
		case wsOpcodePing:
			_ = client.writeFrame(wsOpcodePong, payload)
			continue
		case wsOpcodePong:
			continue
		case wsOpcodeText, wsOpcodeBinary:
			if fragmentedOpcode != 0 {
				return
			}
			if fin {
				if err := s.processEventPayload(payload, client); err != nil {
					log.Printf("[warn] ExtensionServer: rejected event: %v", err)
				}
				continue
			}
			fragmentedOpcode = opcode
			fragmented = append(fragmented[:0], payload...)
		case wsOpcodeContinue:
			if fragmentedOpcode == 0 || len(fragmented)+len(payload) > maxStreamBytes {
				return
			}
			fragmented = append(fragmented, payload...)
			if fin {
				if err := s.processEventPayload(fragmented, client); err != nil {
					log.Printf("[warn] ExtensionServer: rejected fragmented event: %v", err)
				}
				fragmented = nil
				fragmentedOpcode = 0
			}
		default:
			return
		}
	}
}

func readClientFrame(reader io.Reader) (bool, byte, []byte, error) {
	var header [2]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return false, 0, nil, err
	}
	fin := header[0]&0x80 != 0
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	if !masked {
		return false, 0, nil, fmt.Errorf("capture: unmasked WebSocket client frame")
	}
	payloadLength := uint64(header[1] & 0x7F)
	switch payloadLength {
	case 126:
		var extended [2]byte
		if _, err := io.ReadFull(reader, extended[:]); err != nil {
			return false, 0, nil, err
		}
		payloadLength = uint64(binary.BigEndian.Uint16(extended[:]))
	case 127:
		var extended [8]byte
		if _, err := io.ReadFull(reader, extended[:]); err != nil {
			return false, 0, nil, err
		}
		payloadLength = binary.BigEndian.Uint64(extended[:])
	}
	if payloadLength > maxStreamBytes {
		return false, 0, nil, fmt.Errorf("capture: WebSocket payload exceeds %d bytes", maxStreamBytes)
	}
	if opcode >= wsOpcodeClose && (!fin || payloadLength > 125) {
		return false, 0, nil, fmt.Errorf("capture: invalid WebSocket control frame")
	}

	var mask [4]byte
	if _, err := io.ReadFull(reader, mask[:]); err != nil {
		return false, 0, nil, err
	}
	payload := make([]byte, int(payloadLength))
	if _, err := io.ReadFull(reader, payload); err != nil {
		return false, 0, nil, err
	}
	for index := range payload {
		payload[index] ^= mask[index%len(mask)]
	}
	return fin, opcode, payload, nil
}
