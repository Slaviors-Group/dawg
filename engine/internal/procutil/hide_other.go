//go:build !windows

package procutil

import "os/exec"

// HideWindow is a no-op on non-Windows platforms, which have no console window
// allocation behavior to suppress.
func HideWindow(cmd *exec.Cmd) {}
