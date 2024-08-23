package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/bingoohuang/ngg/q"
	godaemon "github.com/sevlyar/go-daemon"
)

const MarkParentPID = "_GO_DAEMON_PID"

// Option is options for Daemonize function.
type Option struct {
	LogFileName   string
	ParentProcess func(child *os.Process)
	// Credential holds user and group identities to be assumed by a daemon-process.
	Credential *syscall.Credential
}

// Daemonize set the current process daemonized
func (o Option) Daemonize() {
	// goland 启动时，不进入后台模式
	if strings.Contains(os.Args[0], "/Caches/JetBrains") {
		return
	}

	workDir, err := os.Getwd()
	if err != nil {
		q.D("Getwd error", err)
	}

	ctx := &godaemon.Context{
		WorkDir:     workDir,
		LogFileName: o.LogFileName,
		Env: []string{
			fmt.Sprintf("%s=%d", MarkParentPID, os.Getpid()),
		},
		Credential: o.Credential,
	}

	child, err := ctx.Reborn()
	if err != nil {
		q.D("reborn error", err)
	} else if child != nil {
		// 有孩子，是父进程
		if o.ParentProcess != nil {
			o.ParentProcess(child)
		} else {
			os.Exit(0)
		}
	}
	// 子进程，继续
}

// GetParentPID returns the parent pid if forked by godaemon.
func GetParentPID() (pid int) {
	value := os.Getenv(MarkParentPID)
	pid, _ = strconv.Atoi(value)
	return pid
}

// ClearRebornEnv clear the reborn env.
func ClearRebornEnv() error {
	err := os.Unsetenv(godaemon.MARK_NAME)
	return err
}
