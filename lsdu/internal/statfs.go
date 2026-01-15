//go:build linux || darwin

package internal

import "syscall"

// https://stackoverflow.com/questions/20108520/get-amount-of-free-disk-space-using-go
// see /usr/lib/go/src/syscall/ztypes_linux_amd64.go for Statfs_t.Block, Statfs_t.Bsize
// and /usr/lib/go/src/syscall/types_linux.go for Fsid
// freebsd uses different types (int64) and the fields are not available on openbsd, netbsd

// FreeSpace reports the disk space available to unprivileged user in bytes, the total disk space and fsid
func FreeSpace(path string) (freeSpace, totalSpace uint64, fsid syscall.Fsid, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, fsid, err
	}
	return stat.Bavail * uint64(stat.Bsize),
		stat.Blocks * uint64(stat.Bsize),
		stat.Fsid,
		nil
}
