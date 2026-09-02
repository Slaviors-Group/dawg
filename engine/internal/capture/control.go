package capture

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const controlTokenHeader = "X-Dawg-Control-Token"

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
	_ = server.Shutdown(context.Background())
	return session.Stop()
}

// StopControlledSession authenticates and requests shutdown of a local capture daemon.
func StopControlledSession(ctx context.Context, controlPath string) error {
	contents, err := os.ReadFile(controlPath)
	if err != nil {
		return fmt.Errorf("capture: read control state %s: %w", controlPath, err)
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
