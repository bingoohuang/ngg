//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!plan9,!solaris

package daemon

import godaemon "github.com/sevlyar/go-daemon"

type Extra struct {
}

func (o Option) fulfile(c *godaemon.Context) {
}
