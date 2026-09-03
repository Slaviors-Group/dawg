//go:build windows

package replay

import (
	"fmt"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ReplayJob manages a Windows Job Object to ensure child processes are terminated
// when the parent process (DAWG CLI) exits or is killed.
type ReplayJob struct {
	handle windows.Handle
}

// NewReplayJob creates a new Windows Job Object configured to kill all child processes on close.
func NewReplayJob() (*ReplayJob, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create job object: %w", err)
	}

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	
	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		windows.CloseHandle(job)
		return nil, fmt.Errorf("set job object limit: %w", err)
	}

	return &ReplayJob{handle: job}, nil
}

// AssignProcess assigns an already-started command to the job object.
func (j *ReplayJob) AssignProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("command has not been started")
	}

	// Open the process handle to assign it to the job object
	handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return fmt.Errorf("open process %d: %w", cmd.Process.Pid, err)
	}
	defer windows.CloseHandle(handle)

	if err := windows.AssignProcessToJobObject(j.handle, handle); err != nil {
		return fmt.Errorf("assign process to job object: %w", err)
	}

	return nil
}

// Close closes the job object handle. If the job was created with KILL_ON_JOB_CLOSE,
// all associated processes will be terminated immediately.
func (j *ReplayJob) Close() error {
	if j.handle != 0 {
		return windows.CloseHandle(j.handle)
	}
	return nil
}
