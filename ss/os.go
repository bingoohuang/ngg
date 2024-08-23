package ss

import (
	"errors"
	"os"
	"path/filepath"
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
