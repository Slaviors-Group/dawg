//go:build windows

package procutil

import (
	"os/exec"
	"syscall"
)

const createNewProcessGroup = 0x00000200

// Detach gives the capture daemon an independent Windows process group and no
// console window, allowing the short-lived launcher process to exit safely.
func Detach(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow | createNewProcessGroup
}
