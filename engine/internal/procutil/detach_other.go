//go:build !windows

package procutil

import (
	"os/exec"
	"syscall"
)

// Detach configures a long-lived daemon so it does not inherit the launcher's
// controlling terminal/session. Without Setsid, closing a terminal or PTY can
// deliver SIGHUP immediately after `capture` reports success.
func Detach(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
}
