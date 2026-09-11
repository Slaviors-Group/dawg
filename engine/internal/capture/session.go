// Package capture implements PRD §6.1 session recording and environment snapshots.
package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// SessionComponent owns one capture source lifecycle within a session.
type SessionComponent interface {
	Name() string
	Start(context.Context, string) error
	Stop() error
}

// SessionOptions describes the metadata and components captured in one session.
type SessionOptions struct {
	Directory  string
	ResultFile string
	Metadata   dawgtypes.CaptureMetadata
	Components []SessionComponent
}

// Session orchestrates capture source startup and shutdown for PRD §6.1.
// A session deliberately does not merge source streams: each source preserves its own ordering and timestamps.
type Session struct {
	options SessionOptions

	mu      sync.Mutex
	started []SessionComponent
	active  bool
}

// NewSession creates a capture session. Components retain their supplied order for startup.
func NewSession(options SessionOptions) (*Session, error) {
	if options.Directory == "" || options.Metadata.SessionID == "" || options.Metadata.TargetURL == "" {
		return nil, fmt.Errorf("capture: session directory, session ID, and target URL are required")
	}
	return &Session{options: options}, nil
}

// Start creates the canonical capture layout and starts components in declared order.
func (session *Session) Start(ctx context.Context) error {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.active {
		return dawgtypes.ErrCaptureAlreadyRunning
	}
	if err := createSessionLayout(session.options.Directory); err != nil {
		return err
	}
	session.options.Metadata.StartedAt = time.Now().UTC()
	session.options.Metadata.StoppedAt = time.Time{}
	session.options.Metadata.CapturedComponents = session.options.Metadata.CapturedComponents[:0]
	if err := writeCaptureMetadata(session.options.Directory, session.options.Metadata); err != nil {
		return err
	}

	for _, component := range session.options.Components {
		if component == nil {
			session.stopStartedLocked()
			return fmt.Errorf("capture: nil session component")
		}
		if err := component.Start(ctx, session.options.Directory); err != nil {
			session.stopStartedLocked()
			return fmt.Errorf("capture: start component %s: %w", component.Name(), err)
		}
		session.started = append(session.started, component)
		session.options.Metadata.CapturedComponents = append(session.options.Metadata.CapturedComponents, component.Name())
	}
	if err := writeCaptureMetadata(session.options.Directory, session.options.Metadata); err != nil {
		session.stopStartedLocked()
		return err
	}
	session.active = true
	return nil
}

// Stop stops components in reverse startup order and records capture completion.
func (session *Session) Stop() error {
	session.mu.Lock()
	defer session.mu.Unlock()
	if !session.active {
		return dawgtypes.ErrCaptureNotRunning
	}
	stopErr := session.stopStartedLocked()
	session.options.Metadata.StoppedAt = time.Now().UTC()
	metadataErr := writeCaptureMetadata(session.options.Directory, session.options.Metadata)
	session.active = false
	if stopErr != nil {
		return stopErr
	}
	return metadataErr
}

func (session *Session) stopStartedLocked() error {
	var stopErr error
	for index := len(session.started) - 1; index >= 0; index-- {
		if err := session.started[index].Stop(); err != nil && stopErr == nil {
			stopErr = fmt.Errorf("capture: stop component %s: %w", session.started[index].Name(), err)
		}
	}
	session.started = nil
	return stopErr
}

func createSessionLayout(directory string) error {
	for _, child := range []string{"traces", "http", "db", "logs", "env", "cassettes", "actions"} {
		if err := os.MkdirAll(filepath.Join(directory, child), 0o700); err != nil {
			return fmt.Errorf("capture: create session directory %s: %w", child, err)
		}
	}
	return nil
}

func writeCaptureMetadata(directory string, metadata dawgtypes.CaptureMetadata) error {
	contents, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("capture: serialize metadata: %w", err)
	}
	contents = append(contents, '\n')
	path := filepath.Join(directory, "meta.json")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("capture: write metadata %s: %w", path, err)
	}
	return nil
}

// IsSessionStateError reports whether a lifecycle error is an idempotency violation.
func IsSessionStateError(err error) bool {
	return errors.Is(err, dawgtypes.ErrCaptureAlreadyRunning) || errors.Is(err, dawgtypes.ErrCaptureNotRunning)
}
