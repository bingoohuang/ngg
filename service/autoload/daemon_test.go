package autoload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWindowsDaemonInitReturns(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows service callbacks require completed package initialization")
	}
	if os.Getenv("AUTOLOAD_INIT_HELPER") == "1" {
		if !daemonPending {
			t.Fatal("daemon should be deferred until Run is called from main")
		}
		return
	}
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-test.run=^TestWindowsDaemonInitReturns$")
	cmd.Env = append(workerEnvironment(os.Environ()), "SRV=daemon", "AUTOLOAD_INIT_HELPER=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("daemon blocked or failed during init: %v\n%s", err, output)
	}
}

func TestInstalledServiceConfig(t *testing.T) {
	dir := t.TempDir()
	path := installedExecutable(dir, "fost")
	want := "fost"
	if runtime.GOOS == "windows" {
		want += ".exe"
	}
	if filepath.Base(path) != want {
		t.Fatalf("executable = %q, want %q", path, want)
	}
	d := daemon{productName: "fost"}
	args := []string{"-S", "tcp://example.invalid:10070", "-L", "socks5://localhost:33780"}
	cfg := d.config(path, args)
	if cfg.Name != "fost" || cfg.Executable != path || cfg.WorkingDirectory != dir || cfg.EnvVars["SRV"] != "daemon" || !reflect.DeepEqual(cfg.Arguments, args) {
		t.Fatalf("bad service config: %+v", cfg)
	}
}

func TestWorkerEnvironment(t *testing.T) {
	got := workerEnvironment([]string{"PATH=test", "SRV=install", "srv=daemon", "Srv=d", "SRV_OTHER=keep", "TOKEN=a=b"})
	want := []string{"PATH=test", "SRV_OTHER=keep", "TOKEN=a=b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("environment = %q, want %q", got, want)
	}
}

func TestPrepareLogDirError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "log"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareLogDir(dir); err == nil {
		t.Fatal("expected error when log is a file")
	}
}

// Exercise a real worker without installing a service or requiring privileges.
func TestWorkerLifecycle(t *testing.T) {
	for _, mode := range []string{"running", "exited"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			logDir, err := prepareLogDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			bin, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("AUTOLOAD_TEST_WORKER", mode)
			t.Setenv("SRV", "daemon")
			ctx, cancel := context.WithCancel(context.Background())
			d := daemon{cancel: cancel, done: make(chan struct{})}
			go func() {
				defer close(d.done)
				d.work(ctx, bin, dir, logDir, []string{"-test.run=^TestWorkerHelper$"}, workerEnvironment(os.Environ()))
			}()
			t.Cleanup(func() { _ = d.Stop(nil) })
			waitFor(t, func() bool {
				_, err := os.Stat(filepath.Join(dir, "ready"))
				return err == nil
			})
			if mode == "exited" {
				waitFor(t, func() bool {
					_, err := os.Stat(filepath.Join(logDir, "dump.log"))
					return err == nil
				})
			}
			stopped := make(chan struct{})
			go func() { _ = d.Stop(nil); close(stopped) }()
			select {
			case <-stopped:
			case <-time.After(5 * time.Second):
				t.Fatal("stop failed to kill worker or interrupt restart delay")
			}
			data, err := os.ReadFile(filepath.Join(logDir, "dump.log"))
			if err != nil || !strings.Contains(string(data), "worker stdout") || !strings.Contains(string(data), "worker stderr") {
				t.Fatalf("worker output = %q, error = %v", data, err)
			}
		})
	}
}

func waitFor(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !ready() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for worker")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestWorkerHelper(t *testing.T) {
	mode := os.Getenv("AUTOLOAD_TEST_WORKER")
	if mode == "" {
		return
	}
	if os.Getenv("SRV") != "" {
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, "worker stdout")
	fmt.Fprintln(os.Stderr, "worker stderr")
	if err := os.WriteFile("ready", nil, 0600); err != nil {
		os.Exit(3)
	}
	if mode == "running" {
		for {
			time.Sleep(time.Second)
		}
	}
	os.Exit(0)
}
