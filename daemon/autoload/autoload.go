package autoload

import (
	"github.com/bingoohuang/ngg/daemon"
	"github.com/bingoohuang/ngg/ss"
)

func init() {
	if yes, _ := ss.GetenvBool("DAEMON", false); yes {
		daemon.Option{}.Daemonize()
	}
}
