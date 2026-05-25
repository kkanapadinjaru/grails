//go:build windows

package cmdutil

import (
	"os/exec"
	"syscall"
)

// HideWindow configures the command to not create a visible console window.
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
