package ss

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bingoohuang/gg/pkg/iox"
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

// ReadFile reads a file content, if it's a .gz, decompress it.
func ReadFile(filename string) ([]byte, error) {
	f, err := ExpandFilename(filename)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %w", f, err)
	}

	return data, nil
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

func ExitIfErr(err error) {
	if err != nil {
		Exit(err.Error(), 1)
	}
}

func Exit(msg string, code int) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

// LinesChan read file into lines.
func LinesChan(filePath string, chSize int) (ch chan string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(f)
	s.Split(ScanLines)
	ch = make(chan string, chSize)
	go func() {
		defer iox.Close(f)
		defer close(ch)

		for s.Scan() {
			t := s.Text()
			t = strings.TrimSpace(t)
			if len(t) > 0 {
				ch <- t
			}
		}

		if err := s.Err(); err != nil {
			log.Printf("E! scan file %s lines  error: %v", filePath, err)
		}
	}()

	return ch, nil
}

// ScanLines is a split function for a Scanner that returns each line of
// text, with end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// Lines read file into lines.
func Lines(filePath string) (lines []string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines, s.Err()
}
