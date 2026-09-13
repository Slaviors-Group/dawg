package capture

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

func TestSessionCreatesCompleteCaptureOutputAndStopsInReverseOrder(t *testing.T) {
	directory := t.TempDir()
	order := []string{}
	session := newTestSession(t, directory, []SessionComponent{
		&recordingComponent{name: "browser", order: &order},
		&recordingComponent{name: "proxy", order: &order},
	})
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("start session: %v", err)
	}
	for _, child := range []string{"traces", "http", "db", "logs", "env", "cassettes", "actions"} {
		info, err := os.Stat(filepath.Join(directory, child))
		if err != nil || !info.IsDir() {
			t.Fatalf("expected session directory %s: %v", child, err)
		}
	}
	if err := session.Stop(); err != nil {
		t.Fatalf("stop session: %v", err)
	}
	if want := []string{"start:browser", "start:proxy", "stop:proxy", "stop:browser"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("unexpected lifecycle order: got %#v, want %#v", order, want)
	}
	contents, err := os.ReadFile(filepath.Join(directory, "meta.json"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var metadata dawgtypes.CaptureMetadata
	if err := json.Unmarshal(contents, &metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if metadata.StartedAt.IsZero() || metadata.StoppedAt.IsZero() || !reflect.DeepEqual(metadata.CapturedComponents, []string{"browser", "proxy"}) {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
}

func TestSessionRejectsDuplicateStartAndStop(t *testing.T) {
	session := newTestSession(t, t.TempDir(), nil)
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := session.Start(context.Background()); !errors.Is(err, dawgtypes.ErrCaptureAlreadyRunning) || !IsSessionStateError(err) {
		t.Fatalf("expected duplicate start error, got %v", err)
	}
	if err := session.Stop(); err != nil {
		t.Fatalf("stop session: %v", err)
	}
	if err := session.Stop(); !errors.Is(err, dawgtypes.ErrCaptureNotRunning) || !IsSessionStateError(err) {
		t.Fatalf("expected duplicate stop error, got %v", err)
	}
}

func TestSessionStopsPreviouslyStartedComponentsOnFailure(t *testing.T) {
	directory := t.TempDir()
	order := []string{}
	session := newTestSession(t, directory, []SessionComponent{
		&recordingComponent{name: "browser", order: &order},
		&recordingComponent{name: "proxy", order: &order, startErr: errors.New("not available")},
	})
	err := session.Start(context.Background())
	if err == nil || !strings.Contains(err.Error(), "start component proxy") {
		t.Fatalf("expected component start error, got %v", err)
	}
	if want := []string{"start:browser", "start:proxy", "stop:browser"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("unexpected cleanup order: got %#v, want %#v", order, want)
	}
}

func newTestSession(t *testing.T, directory string, components []SessionComponent) *Session {
	t.Helper()
	session, err := NewSession(SessionOptions{
		Directory:  directory,
		Components: components,
		Metadata: dawgtypes.CaptureMetadata{
			SessionID: "session-1",
			TargetURL: "https://staging.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return session
}

type recordingComponent struct {
	name     string
	order    *[]string
	startErr error
}

func (component *recordingComponent) Name() string {
	return component.name
}

func (component *recordingComponent) Start(_ context.Context, _ string) error {
	*component.order = append(*component.order, "start:"+component.name)
	return component.startErr
}

func (component *recordingComponent) Stop() error {
	*component.order = append(*component.order, "stop:"+component.name)
	return nil
}
