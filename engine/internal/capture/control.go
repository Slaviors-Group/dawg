package capture

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const controlTokenHeader = "X-Dawg-Control-Token"

// ErrSessionAlreadyStopped is returned by StopControlledSession when the
// capture daemon is no longer running (connection refused). The caller may
// treat this as an abnormal-but-recoverable end of session.
var ErrSessionAlreadyStopped = errors.New("capture: session is not running")

// ControlState is the restricted local handoff from a capture daemon to capture stop.
type ControlState struct {
	Address          string `json:"address"`
	Token            string `json:"token"`
	SessionDirectory string `json:"sessionDirectory"`
}

// RunControlledSession starts a session and blocks until its context ends or a valid local stop request arrives.
func RunControlledSession(ctx context.Context, session *Session, controlPath string) error {
	if session == nil || controlPath == "" {
		return fmt.Errorf("capture: session and control path are required")
	}
	if err := session.Start(ctx); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_ = session.Stop()
		return fmt.Errorf("capture: start control listener: %w", err)
	}
	defer listener.Close()
	token, err := newControlToken()
	if err != nil {
		_ = session.Stop()
		return err
	}
	stop := make(chan struct{})
	var stopOnce sync.Once
	requestStop := func() { stopOnce.Do(func() { close(stop) }) }
	server := &http.Server{Handler: controlHandler(token, requestStop)}
	go server.Serve(listener)
	state := ControlState{Address: listener.Addr().String(), Token: token, SessionDirectory: session.options.Directory}
	if err := writeControlState(controlPath, state); err != nil {
		_ = server.Close()
		_ = session.Stop()
		return err
	}
	defer os.Remove(controlPath)

	select {
	case <-ctx.Done():
	case <-stop:
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
	}
	return session.Stop()
}

// StopControlledSession authenticates and requests shutdown of a local capture daemon.
func StopControlledSession(ctx context.Context, controlPath string) error {
	deadline := time.Now().Add(10 * time.Second)
	var contents []byte
	var err error

	for {
		contents, err = os.ReadFile(controlPath)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) || time.Now().After(deadline) {
			return fmt.Errorf("capture: read control state %s: %w", controlPath, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	var state ControlState
	if err := json.Unmarshal(contents, &state); err != nil {
		return fmt.Errorf("capture: decode control state %s: %w", controlPath, err)
	}
	if state.Address == "" || state.Token == "" {
		return fmt.Errorf("capture: invalid control state %s", controlPath)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+state.Address+"/stop", nil)
	if err != nil {
		return fmt.Errorf("capture: create stop request: %w", err)
	}
	request.Header.Set(controlTokenHeader, state.Token)
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		// ECONNREFUSED means the daemon process has already exited. Return a
		// typed sentinel so callers can distinguish this from a genuine failure
		// and still attempt to recover the packaging result.
		if errors.Is(err, syscall.ECONNREFUSED) {
			return ErrSessionAlreadyStopped
		}
		return fmt.Errorf("capture: request session stop: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("capture: stop request rejected with status %s", response.Status)
	}
	return nil
}

func controlHandler(token string, requestStop func()) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/stop", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if request.Header.Get(controlTokenHeader) != token {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
		requestStop()
	})
	return mux
}

func newControlToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("capture: generate control token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func writeControlState(path string, state ControlState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("capture: create control state directory: %w", err)
	}
	contents, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("capture: serialize control state: %w", err)
	}
	if err := os.WriteFile(path, append(contents, '\n'), 0o600); err != nil {
		return fmt.Errorf("capture: write control state %s: %w", path, err)
	}
	return nil
}
