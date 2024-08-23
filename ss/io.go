package ss

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
)

func Close[T io.Closer](closers ...T) error {
	var err error
	for _, closer := range closers {
		err = multierr.Append(err, closer.Close())
	}
	return err
}

func ReadAll(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ExpandAtFile returns argument s if it starts with @filename, the file contents will be replaced as the data.
func ExpandAtFile(s string) (string, error) {
	if !strings.HasPrefix(s, "@") {
		return s, nil
	}

	data, err := os.ReadFile(s[1:])
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CommonDir returns the common directory for a slice of directories.
func CommonDir(dirs []string) string {
	baseDir := ""

	for _, dir := range dirs {
		d := filepath.Dir(dir)

		if baseDir == "" {
			baseDir = d
		} else {
			for !strings.HasPrefix(d, baseDir) {
				baseDir = filepath.Dir(baseDir)
			}
		}

		if baseDir == "/" {
			break
		}
	}

	if baseDir == "" {
		baseDir = "/"
	}

	return baseDir
}
