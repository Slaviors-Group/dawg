//go:build !windows

package replay

import (
	"os/exec"
)

// ReplayJob is a no-op on non-Windows platforms.
type ReplayJob struct{}

// NewReplayJob returns a no-op job object.
func NewReplayJob() (*ReplayJob, error) {
	return &ReplayJob{}, nil
}

// AssignProcess is a no-op on non-Windows platforms.
func (j *ReplayJob) AssignProcess(cmd *exec.Cmd) error {
	return nil
}

// Close is a no-op on non-Windows platforms.
func (j *ReplayJob) Close() error {
	return nil
}
