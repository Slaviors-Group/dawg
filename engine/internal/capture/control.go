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
	SessionID        string `json:"sessionId"`
	SessionDirectory string `json:"sessionDirectory"`
	ResultFile       string `json:"resultFile,omitempty"`
	PID              int    `json:"pid"`
}

// StopRequestSource is implemented by capture components whose unexpected
// termination must stop the owning session before sanitization and packaging.
type StopRequestSource interface {
	StopRequested() <-chan struct{}
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
	forwardDone := make(chan struct{})
	defer close(forwardDone)
	for _, component := range session.options.Components {
		source, ok := component.(StopRequestSource)
		if !ok {
			continue
		}
		requests := source.StopRequested()
		go func() {
			select {
			case <-requests:
				requestStop()
			case <-forwardDone:
			case <-ctx.Done():
			}
		}()
	}
	server := &http.Server{Handler: controlHandler(token, requestStop)}
	serverDone := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
			err = nil
		}
		serverDone <- err
	}()
	state := ControlState{
		Address:          listener.Addr().String(),
		Token:            token,
		SessionID:        session.options.Metadata.SessionID,
		SessionDirectory: session.options.Directory,
		ResultFile:       session.options.ResultFile,
		PID:              os.Getpid(),
	}
	if err := writeControlState(controlPath, state); err != nil {
		_ = server.Close()
		<-serverDone
		_ = session.Stop()
		return err
	}
	defer os.Remove(controlPath)

	var serveErr error
	serverExited := false
	select {
	case <-ctx.Done():
	case <-stop:
	case serveErr = <-serverDone:
		serverExited = true
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if !serverExited {
		if err := server.Shutdown(shutdownCtx); err != nil {
			serveErr = errors.Join(serveErr, err)
			_ = server.Close()
		}
		serveErr = errors.Join(serveErr, <-serverDone)
	}
	return errors.Join(serveErr, session.Stop())
}

// StopControlledSession authenticates and requests shutdown of a local capture daemon.
func StopControlledSession(ctx context.Context, controlPath string) error {
	deadline := time.Now().Add(10 * time.Second)
	var state ControlState
	var err error

	for {
		state, err = ReadControlState(controlPath)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) || time.Now().After(deadline) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
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
	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if request.Header.Get(controlTokenHeader) != token {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
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
	if err := writeAtomicFile(path, append(contents, '\n')); err != nil {
		return fmt.Errorf("capture: write control state %s: %w", path, err)
	}
	return nil
}

// ReadControlState loads and validates the active daemon's control record.
func ReadControlState(path string) (ControlState, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return ControlState{}, fmt.Errorf("capture: read control state %s: %w", path, err)
	}
	var state ControlState
	if err := json.Unmarshal(contents, &state); err != nil {
		return ControlState{}, fmt.Errorf("capture: decode control state %s: %w", path, err)
	}
	if state.Address == "" || state.Token == "" {
		return ControlState{}, fmt.Errorf("capture: invalid control state %s", path)
	}
	return state, nil
}

// ControlledSessionActive authenticates against a control record instead of
// trusting the mere existence of a potentially stale file.
func ControlledSessionActive(ctx context.Context, state ControlState) bool {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+state.Address+"/health", nil)
	if err != nil {
		return false
	}
	request.Header.Set(controlTokenHeader, state.Token)
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func writeAtomicFile(path string, contents []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".dawg-state-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
