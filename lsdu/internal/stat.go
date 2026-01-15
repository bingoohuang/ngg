//go:build !windows

package internal

import (
	"os"
	"syscall"
	"time"
)

// If Go 2 drops syscall we have to use golang.org/x/sys/unix

// FileInfo contains the information of os.FileInfo and the allocated size of a file which may be different if it contains holes.
type FileInfo interface {
	os.FileInfo
	AllocatedSize() int64
	ApparentSize() int64
}

// We need to override Size() so lots of boilerplate

type fileInfo struct {
	osfi os.FileInfo
}

func (fi fileInfo) AllocatedSize() int64 {
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		panic("Type assertion of *syscall.Stat_t failed")
	}
	// blocks are always 512-byte units (man 2 stat)
	return sys.Blocks * 512
}

func (fi fileInfo) ApparentSize() int64 {
	return fi.osfi.Size()
}

func (fi fileInfo) Name() string {
	return fi.osfi.Name()
}

func (fi fileInfo) Size() int64 {
	return fi.AllocatedSize()
}

func (fi fileInfo) Mode() os.FileMode {
	return fi.osfi.Mode()
}

func (fi fileInfo) ModTime() time.Time {
	return fi.osfi.ModTime()
}

func (fi fileInfo) IsDir() bool {
	return fi.osfi.IsDir()
}

func (fi fileInfo) Sys() interface{} {
	return fi.osfi.Sys()
}

// We assume an implementation of /usr/lib/go/src/os/types_unix.go as of go 1.15
// man 3 fstatat for an example
// man 2 stat

// Stat returns the info of os.Stat and the real allocated size of the file
func Stat(path string) (FileInfo, error) {
	fi, err := os.Stat(path)
	return fileInfo{fi}, err
}

// Lstat returns the info of os.Lstat and the real allocated size of the file
func Lstat(path string) (FileInfo, error) {
	fi, err := os.Lstat(path)
	return fileInfo{fi}, err
}

// Readdir wraps os.Readdir providing allocated size information
// We cannot have du.Readdir as a method on os.File like os.Readdir
func Readdir(f *os.File, n int) ([]FileInfo, error) {
	osfis, err := f.Readdir(n)
	// os.Readdir may return a non-empty slice even if the error is non-nil when n<=0
	var fis []FileInfo
	for _, fi := range osfis {
		fis = append(fis, fileInfo{fi})
	}
	return fis, err
}
