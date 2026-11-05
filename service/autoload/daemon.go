package autoload

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/service"
)

type daemon struct {
	productName string
	mu          sync.Mutex
	proc        *os.Process
	cancel      context.CancelFunc
	done        chan struct{}
}

func (d *daemon) Start(s service.Service) error {
	binPath, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(binPath)
	logDir, err := prepareLogDir(dir)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.done = make(chan struct{})
	go func() {
		defer close(d.done)
		d.work(ctx, binPath, dir, logDir, os.Args[1:], workerEnvironment(os.Environ()))
	}()
	log.Println("daemon start")
	return nil
}

func (d *daemon) Stop(s service.Service) error {
	log.Println("service stop")
	if d.cancel == nil {
		return nil
	}
	d.cancel()
	d.mu.Lock()
	if d.proc != nil {
		_ = d.proc.Kill()
	}
	d.mu.Unlock()
	<-d.done
	return nil
}

func (d *daemon) run() error {
	log.Println("daemon run start")
	defer log.Println("daemon run end")
	s, err := service.New(d, d.config("", nil))
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	// Let the service manager call Start/Stop. In particular, do not launch a
	// worker before Windows has connected to the service control dispatcher.
	if err := s.Run(); err != nil {
		return fmt.Errorf("run service: %w", err)
	}
	return nil
}

func prepareLogDir(dir string) (string, error) {
	logDir := filepath.Join(dir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("create log directory: %w", err)
	}
	return logDir, nil
}

func workerEnvironment(environ []string) []string {
	envs := make([]string, 0, len(environ))
	for _, env := range environ {
		key, _, _ := strings.Cut(env, "=")
		// Environment names are case insensitive on Windows.
		if !strings.EqualFold(key, "SRV") {
			envs = append(envs, env)
		}
	}
	return envs
}

func (d *daemon) work(ctx context.Context, binPath, dir, logDir string, args, envs []string) {
	for ctx.Err() == nil {
		tmpDump := filepath.Join(logDir, "dump.log.tmp")
		dumpFile := filepath.Join(logDir, "dump.log")
		f, err := os.Create(tmpDump)
		if err != nil {
			log.Printf("E! start worker: %v", err)
			return
		}
		cmd := exec.Command(binPath, args...)
		cmd.Dir, cmd.Env = dir, envs
		// Windows services need valid handles even when no console is attached.
		// exec.Cmd supplies a null stdin; both output streams go to the log.
		cmd.Stdout, cmd.Stderr = f, f
		started := time.Now()
		d.mu.Lock()
		if ctx.Err() != nil {
			d.mu.Unlock()
			_ = f.Close()
			return
		}
		err = cmd.Start()
		if err == nil {
			d.proc = cmd.Process
		}
		d.mu.Unlock()
		if err == nil {
			log.Println("start worker process")
			err = cmd.Wait()
		}
		d.mu.Lock()
		d.proc = nil
		d.mu.Unlock()
		_ = f.Close()
		if err != nil && ctx.Err() == nil {
			log.Printf("E! worker stopped: %v", err)
		}
		// Windows cannot rename over an existing destination.
		_ = os.Remove(dumpFile)
		if err := os.Rename(tmpDump, dumpFile); err != nil {
			log.Printf("E! rename dump: %v", err)
		}
		delay := time.Second
		if time.Since(started) < 10*time.Second {
			delay = 10 * time.Second
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (d *daemon) config(exeAbsPath string, args []string) *service.Config {
	return &service.Config{
		Name:             d.productName,
		DisplayName:      d.productName,
		Description:      d.productName,
		Executable:       exeAbsPath,
		Arguments:        args,
		WorkingDirectory: filepath.Dir(exeAbsPath),
		EnvVars:          map[string]string{"SRV": "daemon"},
	}
}

func (d *daemon) Control(ctrlComm string, exeAbsPath string, args []string) error {
	s, err := service.New(d, d.config(exeAbsPath, args))
	if err != nil {
		return err
	}
	return service.Control(s, ctrlComm)
}
