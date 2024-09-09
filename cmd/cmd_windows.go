package cmd

import (
	"os/exec"
	"syscall"
)

func createBaseCommand(c *Cmd) *exec.Cmd {
	return exec.Command(`cmd.exe`, "/C", c.Command)
}

// WithUser allows the command to be run as a different
// user.
//
// Example:
//
//	token := syscall.Token(handle)
//	c := New("echo hello", token)
//	c.Run(context.TODO())
func WithUser(token syscall.Token) func(c *Cmd) {
	return func(c *Cmd) {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{
			Token: token,
		}
	}
}
