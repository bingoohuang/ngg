//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || plan9 || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd plan9 solaris

package daemon

import (
	"syscall"

	godaemon "github.com/sevlyar/go-daemon"
)

type Extra struct {
	// Credential holds user and group identities to be assumed by a daemon-process.
	Credential *syscall.Credential
}

func (o Option) fulfile(c *godaemon.Context) {
	c.Credential = o.Credential
}
