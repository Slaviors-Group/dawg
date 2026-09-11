package capture

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// TestRapidStartStopNoRaceOrPanic exercises the most common real-world race:
// the UI and a keyboard shortcut both fire Start or Stop within milliseconds
// of each other. Run with -race to confirm no data races.
func TestRapidStartStopNoRaceOrPanic(t *testing.T) {
	const cycles = 50

	for i := range cycles {
		dir := t.TempDir()
		session := newTestSession(t, dir, nil)

		// First start must succeed.
		if err := session.Start(context.Background()); err != nil {
			t.Fatalf("cycle %d: start: %v", i, err)
		}

		// Concurrent duplicate start: must return ErrCaptureAlreadyRunning, not panic.
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := session.Start(context.Background())
			if err == nil || !errors.Is(err, dawgtypes.ErrCaptureAlreadyRunning) {
				t.Errorf("cycle %d: concurrent start: expected ErrCaptureAlreadyRunning, got %v", i, err)
			}
		}()
		wg.Wait()

		// Stop must succeed.
		if err := session.Stop(); err != nil {
			t.Fatalf("cycle %d: stop: %v", i, err)
		}

		// Second Stop must return ErrCaptureNotRunning cleanly (not panic).
		if err := session.Stop(); !errors.Is(err, dawgtypes.ErrCaptureNotRunning) {
			t.Fatalf("cycle %d: double-stop: expected ErrCaptureNotRunning, got %v", i, err)
		}
	}
}

// TestRapidConcurrentStopNoPanic fires Stop() from two goroutines simultaneously
// to verify the mutex prevents a panic even under concurrent load.
func TestRapidConcurrentStopNoPanic(t *testing.T) {
	const workers = 10

	for range workers {
		dir := t.TempDir()
		session := newTestSession(t, dir, nil)
		if err := session.Start(context.Background()); err != nil {
			t.Fatalf("start: %v", err)
		}

		var wg sync.WaitGroup
		errs := make([]error, workers)
		for j := range workers {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				errs[idx] = session.Stop()
			}(j)
		}
		wg.Wait()

		// Exactly one goroutine should get nil; the rest get ErrCaptureNotRunning.
		var successes int
		for _, err := range errs {
			if err == nil {
				successes++
			} else if !errors.Is(err, dawgtypes.ErrCaptureNotRunning) {
				t.Errorf("unexpected stop error: %v", err)
			}
		}
		if successes != 1 {
			t.Errorf("expected exactly 1 successful stop, got %d", successes)
		}
	}
}
