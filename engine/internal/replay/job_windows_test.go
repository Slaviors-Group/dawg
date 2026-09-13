//go:build windows

package replay

import (
	"testing"
)

func TestNewReplayJob(t *testing.T) {
	job, err := NewReplayJob()
	if err != nil {
		t.Fatalf("expected to create job object without error, got %v", err)
	}
	if job == nil {
		t.Fatalf("expected job object to not be nil")
	}

	// Make sure we can close it safely
	err = job.Close()
	if err != nil {
		t.Fatalf("expected to close job object without error, got %v", err)
	}
}
