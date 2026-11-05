package autoload

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/service"
	"github.com/bingoohuang/ngg/ver"
)

// Native Windows service callbacks must run after package initialization.
var daemonPending bool

func init() {
	srv := os.Getenv("SRV")
	switch srv {
	case "install", "i":
		if err := install(); err != nil {
			log.Printf("E! install service: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "uninstall", "u":
		if err := uninstall(); err != nil {
			log.Printf("E! uninstall service: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "daemon", "d":
		daemonPending = true
		if runtime.GOOS != "windows" {
			Run()
		}
	}
}

// Run starts the daemon requested by SRV=daemon and blocks until it stops.
// Call Run at the beginning of main, before starting application work. On
// Windows this must happen after init: callbacks from native SCM threads wait
// for Go package initialization to finish. Run is a no-op in a normal worker.
// Daemon failures exit with status 1; a clean shutdown exits with status 0.
func Run() {
	if !daemonPending {
		return
	}
	daemonPending = false
	if err := daemonRun(); err != nil {
		log.Printf("E! run service: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func daemonRun() error {
	execPath, _ := os.Executable()
	execName := filepath.Base(execPath)
	execName = strings.TrimSuffix(execName, ".exe") // Windows 上去掉 .exe，避免 "fost.exe" 被当作目录名
	d := daemon{productName: execName}
	return d.run()
}

func install() error {
	log.Println("start. version: ", ver.Version())
	log.Println("install start")
	defer log.Println("install end")
	execPath, execName := parseExec()
	defaultInstallPath := filepath.Join(defaultInstallBasePath, execName)
	if err := os.MkdirAll(defaultInstallPath, 0775); err != nil {
		return fmt.Errorf("create install directory: %w", err)
	}
	if err := os.Chdir(defaultInstallPath); err != nil {
		return fmt.Errorf("change install directory: %w", err)
	}

	// Remove an existing service, including one that failed to start.
	if err := uninstall(); err != nil {
		return err
	}

	targetPath := installedExecutable(defaultInstallPath, execName)
	d := daemon{productName: execName}
	// copy files

	if err := copyExec(execPath, targetPath); err != nil {
		return err
	}

	// install system service
	log.Println("targetPath:", targetPath)
	if err := d.Control("install", targetPath, os.Args[1:]); err != nil {
		return fmt.Errorf("register service: %w", err)
	}
	log.Println("install system service ok.")
	time.Sleep(time.Second * 2)
	if err := d.Control("start", targetPath, os.Args[1:]); err != nil {
		return fmt.Errorf("start %s service: %w", execName, err)
	}
	log.Println("start service ok.")
	return nil
}

func copyExec(execPath, targetPath string) error {
	src, err := os.Open(execPath) // can not use args[0], on Windows call openp2p is ok(=openp2p.exe)
	if err != nil {
		return fmt.Errorf("os.OpenFile %s :%w", os.Args[0], err)
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		return fmt.Errorf("os.OpenFile %s: %w", targetPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}

func parseExec() (string, string) {
	execPath, _ := os.Executable()
	execName := filepath.Base(execPath)
	if p := strings.Index(execName, "go_build_"); p > 0 {
		execName = execName[p+len("go_build_"):]
	}
	// Windows 上去掉 .exe，避免 "fost.exe" 被当作目录名
	execName = strings.TrimSuffix(execName, ".exe")
	return execPath, execName
}

func uninstall() error {
	log.Println("uninstall start")
	defer log.Println("uninstall end")

	_, execName := parseExec()
	d := daemon{productName: execName}

	s, err := service.New(&d, d.config("", nil))
	if err != nil {
		return err
	}
	status, err := s.Status()
	if errors.Is(err, service.ErrNotInstalled) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query service: %w", err)
	}
	if status != service.StatusStopped {
		if err := s.Stop(); err != nil {
			return fmt.Errorf("stop service: %w", err)
		}
	}
	if err := s.Uninstall(); err != nil {
		return fmt.Errorf("uninstall service: %w", err)
	}
	log.Println("uninstall system service ok.")

	defaultInstallPath := filepath.Join(defaultInstallBasePath, execName)
	binPath := installedExecutable(defaultInstallPath, execName)
	_ = os.Remove(binPath)
	return nil
}

// Keep the executable extension separate from the service/directory name.
func installedExecutable(dir, name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}
