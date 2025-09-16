package autoload

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ver"
)

func init() {
	srv := os.Getenv("SRV")
	switch srv {
	case "install", "i":
		install()
		os.Exit(0)
	case "uninstall", "u":
		uninstall()
		os.Exit(0)
	case "daemon", "d":
		daemonRun()
		os.Exit(0)
	}
}

func daemonRun() {
	execPath, _ := os.Executable()
	execName := filepath.Base(execPath)
	d := daemon{productName: execName}
	d.run()
}

func install() {
	log.Println("start. version: ", ver.Version())
	log.Println("install start")
	defer log.Println("install end")
	execPath, execName := parseExec()
	defaultInstallPath := filepath.Join(defaultInstallBasePath, execName)
	if err := os.MkdirAll(defaultInstallPath, 0775); err != nil {
		log.Printf("E! MkdirAll %s error: %v", defaultInstallPath, err)
		return
	}
	if err := os.Chdir(defaultInstallPath); err != nil {
		log.Printf("E! cd error: %v", err)
		return
	}

	// auto uninstall
	uninstall()
	// save config file

	targetPath := filepath.Join(defaultInstallPath, execName)
	d := daemon{productName: execName}
	// copy files

	if err := copyExec(execPath, targetPath); err != nil {
		log.Printf("E! Copy %s error: %v", targetPath, err)
		return
	}

	// install system service
	log.Println("targetPath:", targetPath)
	if err := d.Control("install", targetPath, os.Args[1:]); err == nil {
		log.Println("install system service ok.")
	}
	time.Sleep(time.Second * 2)
	if err := d.Control("start", targetPath, os.Args[1:]); err != nil {
		log.Printf("E! start openp2p service error: %v", err)
	} else {
		log.Println("start service ok.")
	}
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
	if p := strings.Index(execName, "go_build_"); p > 0 { // /usr/local/___1go_build_ccagent_cmd_ccagent;
		execName = execName[p+len("go_build_"):]
	}
	return execPath, execName
}

func uninstall() {
	log.Println("uninstall start")
	defer log.Println("uninstall end")

	_, execName := parseExec()
	d := daemon{productName: execName}

	if err := d.Control("stop", "", nil); err != nil { // service maybe not install
		return
	}
	if err := d.Control("uninstall", "", nil); err != nil {
		log.Printf("E! uninstall system service error: %v", err)
	} else {
		log.Println("uninstall system service ok.")
	}

	defaultInstallPath := filepath.Join(defaultInstallBasePath, execName)
	binPath := filepath.Join(defaultInstallPath, execName)
	_ = os.Remove(binPath)
}
