package main

import (
	"fmt"
	"io"
	"io/fs"
	"strconv"

	"github.com/cespare/xxhash/v2"
)

func XxHash(body []byte) string {
	x := xxhash.New()
	_, _ = x.Write(body)
	h := x.Sum64()
	xh := strconv.FormatUint(h, 16)
	return xh
}

func ReadFsFile(fsys fs.FS, path string) ([]byte, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	return body, nil
}
