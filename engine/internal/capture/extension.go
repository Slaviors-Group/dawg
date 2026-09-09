package capture

import (
	"bufio"
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
	"sync"
	"time"
)

// ExtensionServer receives streaming DOM events (rrweb), browser action traces,
// and frontend HTTP traffic from the DAWG Chrome Extension.
type ExtensionServer struct {
	ListenPort int
	ListenAddr string

	mu          sync.Mutex
	listener    net.Listener
	server      *http.Server
	rrwebFile   *os.File
	actionsFile *os.File
	httpFile    *os.File
	actualAddr  string
}

// Name implements SessionComponent.
func (s *ExtensionServer) Name() string {
	return "browser"
}

// Addr returns the actual listening address of the server.
func (s *ExtensionServer) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.actualAddr
}

// Start creates session files and launches the HTTP/WebSocket ingestion server.
func (s *ExtensionServer) Start(ctx context.Context, directory string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		return fmt.Errorf("capture: extension server already running")
	}

	for _, dir := range []string{"traces", "http", "actions"} {
		if err := os.MkdirAll(filepath.Join(directory, dir), 0o700); err != nil {
			return fmt.Errorf("capture: create output directory %s: %w", dir, err)
		}
	}

	var err error
	s.rrwebFile, err = os.OpenFile(filepath.Join(directory, "traces", "rrweb.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("capture: open rrweb.jsonl: %w", err)
	}

	s.actionsFile, err = os.OpenFile(filepath.Join(directory, "actions", "browser.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		_ = s.rrwebFile.Close()
		return fmt.Errorf("capture: open browser.jsonl: %w", err)
	}

	s.httpFile, err = os.OpenFile(filepath.Join(directory, "http", "frontend.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		_ = s.rrwebFile.Close()
		_ = s.actionsFile.Close()
		return fmt.Errorf("capture: open frontend.jsonl: %w", err)
	}

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
		_ = s.rrwebFile.Close()
		_ = s.actionsFile.Close()
		_ = s.httpFile.Close()
		return fmt.Errorf("capture: start extension listener on %s: %w", addr, err)
	}
	s.listener = listener
	s.actualAddr = listener.Addr().String()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/v1/stream/rrweb", s.handleStreamRRWeb)
	mux.HandleFunc("/api/v1/stream/actions", s.handleStreamActions)
	mux.HandleFunc("/api/v1/stream/http", s.handleStreamHTTP)
	mux.HandleFunc("/api/v1/stream/event", s.handleStreamEvent)

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		_ = s.server.Serve(listener)
	}()

	return nil
}

// Stop shuts down the extension ingestion server and closes output files.
func (s *ExtensionServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.server.Shutdown(shutdownCtx)

	var closeErr error
	if s.rrwebFile != nil {
		if err := s.rrwebFile.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		s.rrwebFile = nil
	}
	if s.actionsFile != nil {
		if err := s.actionsFile.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		s.actionsFile = nil
	}
	if s.httpFile != nil {
		if err := s.httpFile.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		s.httpFile = nil
	}

	s.server = nil
	s.listener = nil
	return closeErr
}

func (s *ExtensionServer) appendJSONL(file *os.File, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if file == nil {
		return fmt.Errorf("output file closed")
	}
	if len(data) == 0 {
		return nil
	}
	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	_, err := file.Write(data)
	return err
}

func (s *ExtensionServer) handleStreamRRWeb(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = s.appendJSONL(s.rrwebFile, body)
	w.WriteHeader(http.StatusOK)
}

func (s *ExtensionServer) handleStreamActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = s.appendJSONL(s.actionsFile, body)
	w.WriteHeader(http.StatusOK)
}

func (s *ExtensionServer) handleStreamHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = s.appendJSONL(s.httpFile, body)
	w.WriteHeader(http.StatusOK)
}

func (s *ExtensionServer) handleStreamEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.processEventPayload(body)
	w.WriteHeader(http.StatusOK)
}

type streamEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (s *ExtensionServer) processEventPayload(rawPayload []byte) {
	var env streamEnvelope
	if err := json.Unmarshal(rawPayload, &env); err != nil {
		return
	}
	switch env.Type {
	case "DAWG_RRWEB_EVENT":
		_ = s.appendJSONL(s.rrwebFile, env.Data)
	case "DAWG_ACTION_EVENT":
		_ = s.appendJSONL(s.actionsFile, env.Data)
	case "DAWG_HTTP_EVENT":
		_ = s.appendJSONL(s.httpFile, env.Data)
	}
}

func (s *ExtensionServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	secKey := r.Header.Get("Sec-WebSocket-Key")
	if secKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h := sha1.New()
	h.Write([]byte(secKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()

	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	if _, err := buf.WriteString(response); err != nil {
		return
	}
	_ = buf.Flush()

	s.readWebSocketFrames(buf, conn)
}

func (s *ExtensionServer) readWebSocketFrames(reader *bufio.ReadWriter, conn net.Conn) {
	for {
		header := make([]byte, 2)
		if _, err := io.ReadFull(reader, header); err != nil {
			return
		}

		fin := (header[0] & 0x80) != 0
		opcode := header[0] & 0x0F
		masked := (header[1] & 0x80) != 0
		payloadLen := uint64(header[1] & 0x7F)

		if opcode == 8 { // Close frame
			return
		}

		if payloadLen == 126 {
			extended := make([]byte, 2)
			if _, err := io.ReadFull(reader, extended); err != nil {
				return
			}
			payloadLen = uint64(binary.BigEndian.Uint16(extended))
		} else if payloadLen == 127 {
			extended := make([]byte, 8)
			if _, err := io.ReadFull(reader, extended); err != nil {
				return
			}
			payloadLen = binary.BigEndian.Uint64(extended)
		}

		var maskKey []byte
		if masked {
			maskKey = make([]byte, 4)
			if _, err := io.ReadFull(reader, maskKey); err != nil {
				return
			}
		}

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return
		}

		if masked {
			for i := uint64(0); i < payloadLen; i++ {
				payload[i] ^= maskKey[i%4]
			}
		}

		if fin && (opcode == 1 || opcode == 2) {
			s.processEventPayload(payload)
		}
	}
}

