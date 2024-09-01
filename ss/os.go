package ss

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var Home string

func init() {
	Home, _ = os.UserHomeDir()
}

func ExpandHome(s string) string {
	if strings.HasPrefix(s, "~") {
		return filepath.Join(Home, s[1:])
	}

	return s
}

// ExpandFilename first expand ~, then expand symbol link
func ExpandFilename(file string) (string, error) {
	filename := ExpandHome(file)
	fi, err := os.Lstat(filename)
	if err != nil {
		if _, err2 := os.Stat(filename); os.IsNotExist(err2) {
			return filename, err2
		}

		return filename, fmt.Errorf("lstat %s: %w", filename, err)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		s, err := os.Readlink(filename)
		if err != nil {
			return filename, fmt.Errorf("readlink %s: %w", filename, err)
		}
		return s, nil
	}

	return filename, nil
}

// WriteTempFile writes the content to a temporary file.
func WriteTempFile(tempDir, pattern string, data []byte, errorPanic bool) (name string, err error) {
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	name, err = func() (string, error) {
		f, err := os.CreateTemp(tempDir, pattern)
		if err != nil {
			return "", err
		}
		defer f.Close()

		if _, err := f.Write(data); err != nil {
			return "", err
		}

		return f.Name(), nil
	}()

	if err != nil && errorPanic {
		panic(err)
	}

	return name, err
}

func Exists(name string) (bool, error) {
	if _, err := os.Stat(name); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false, err
	}
}

func OpenInBrowser(addr string, paths ...string) (string, error) {
	if strings.HasPrefix(addr, ":") {
		addr = "http://127.0.0.1" + addr
	}
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}

	addr, err := url.JoinPath(addr, paths...)
	if err != nil {
		return "", fmt.Errorf("JoinPath: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", addr).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", addr).Start()
	case "darwin":
		err = exec.Command("open", addr).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return "", fmt.Errorf("openbrowser: %w", err)
	}

	return addr, nil
}
