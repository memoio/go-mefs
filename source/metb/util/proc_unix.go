// +build !windows

package metbutil

import (
	"os/exec"
	"syscall"
)

func setupOpt(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
