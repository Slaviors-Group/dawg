//go:build windows

// Package procutil provides small OS-process helpers shared across capture and replay.
package procutil

import (
	"os/exec"
	"syscall"
)

// createNoWindow is the Windows CREATE_NO_WINDOW process creation flag.
const createNoWindow = 0x08000000

// HideWindow configures cmd so that Windows does not allocate a visible console
// window for it. Without this, every background/sidecar process (mitmdump,
// node/Playwright, or a re-launched dawg daemon) pops up its own console host
// window (e.g. Windows Terminal) that stays open for the process lifetime.
func HideWindow(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
