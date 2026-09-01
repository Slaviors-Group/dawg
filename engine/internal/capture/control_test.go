package capture

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunControlledSessionStopsFromAuthenticatedControlRequest(t *testing.T) {
	directory := t.TempDir()
	controlPath := filepath.Join(directory, "control", "session.json")
	session := newTestSession(t, filepath.Join(directory, "session"), nil)
	result := make(chan error, 1)
	go func() { result <- RunControlledSession(context.Background(), session, controlPath) }()
	state := waitForControlState(t, controlPath)
	if state.Address == "" || len(state.Token) != 64 {
		t.Fatalf("unexpected control state: %#v", state)
	}
	if err := StopControlledSession(context.Background(), controlPath); err != nil {
		t.Fatalf("request controlled stop: %v", err)
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("controlled session: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for controlled session shutdown")
	}
	if _, err := os.Stat(controlPath); !os.IsNotExist(err) {
		t.Fatalf("control state was not removed: %v", err)
	}
}

func TestControlHandlerRejectsMissingToken(t *testing.T) {
	server := httptest.NewServer(controlHandler("secret", func() {}))
	defer server.Close()
	response, err := http.Post(server.URL+"/stop", "", nil)
	if err != nil {
		t.Fatalf("send unauthenticated stop: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected status: %s", response.Status)
	}
}

func waitForControlState(t *testing.T, path string) ControlState {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		contents, err := os.ReadFile(path)
		if err == nil {
			var state ControlState
			if err := json.Unmarshal(contents, &state); err != nil {
				t.Fatalf("decode control state: %v", err)
			}
			return state
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for control state %s", path)
	return ControlState{}
}
