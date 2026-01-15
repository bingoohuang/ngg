package lsdu

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bingoohuang/ngg/lsdu/internal"
)

// du -s -m *
// if initial path is a symbolic link to a dir it is followed
// however symbolic links under it are not followed (so not checking for circular links)
// (ReadDir uses Lstat)
// hard links are not checked (they will be accounted more than once)

// DirEntry describes a file or a directory and all the files beneath it
type DirEntry struct {
	internal.FileInfo
	path              string
	deepApparentSize  int64
	deepAllocatedSize int64
	FreeSpace         uint64
	TotalSpace        uint64
}

// Size returns the allocated size (for sorting)
func (fi DirEntry) Size() int64 {
	return fi.deepAllocatedSize
}

// DeepApparentSize returns the apparent size of all files beneath this entry
func (fi DirEntry) DeepApparentSize() int64 {
	return fi.deepApparentSize
}

// DeepAllocatedSize returns the allocated size of all files beneath this entry
func (fi DirEntry) DeepAllocatedSize() int64 {
	return fi.deepAllocatedSize
}

// Path returns the path for -d mode display
func (fi DirEntry) Path() string {
	return fi.path
}

func (fi DirEntry) String() string {
	return fmt.Sprintf("%v %d %s \"%s\"", fi.Mode(), fi.Size(), fi.ModTime().Format("2006-01-02"), fi.Name())
}

// StatDir returns the deep size info for a single path (for -d mode)
func StatDir(path string) (dirEntry DirEntry, apparent, allocated int64, err error) {
	_, apparent, allocated, err = ReadDirDeep(path)
	if err != nil {
		return DirEntry{}, 0, 0, err
	}
	// Get the file info for the path itself
	fileInfo, err := internal.Lstat(path)
	if err != nil {
		return DirEntry{}, 0, 0, err
	}
	freeSpace, totalSpace, _, _ := internal.FreeSpace(path)
	return DirEntry{
		FileInfo:          fileInfo,
		path:              path,
		deepApparentSize:  apparent,
		deepAllocatedSize: allocated,
		FreeSpace:         freeSpace,
		TotalSpace:        totalSpace,
	}, apparent, allocated, nil
}

func readDir(path string) ([]internal.FileInfo, error) {
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	fileInfos, err := internal.Readdir(dir, 0)
	if err != nil {
		return nil, err
	}
	return fileInfos, err
}

// ReadDirDeep reads a directory and the deep size of each entry
func ReadDirDeep(path string) (dirEntries []DirEntry, totalApparent, totalAllocated int64, err error) {
	// This test is only needed for the initial caller. We only call ourselves when the path is a directory.
	fileInfo, err := internal.Lstat(path)
	if err != nil {
		return nil, 0, 0, err
	}
	if !fileInfo.IsDir() {
		freeSpace, totalSpace, _, _ := internal.FreeSpace(path)
		return []DirEntry{{
			FileInfo:          fileInfo,
			path:              path,
			deepApparentSize:  fileInfo.ApparentSize(),
			deepAllocatedSize: fileInfo.AllocatedSize(),
			FreeSpace:         freeSpace,
			TotalSpace:        totalSpace,
		}}, fileInfo.ApparentSize(), fileInfo.AllocatedSize(), nil
	}

	fileInfos, err := readDir(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, err)
			freeSpace, totalSpace, _, _ := internal.FreeSpace(path)
			return []DirEntry{{
				FileInfo:          fileInfo,
				path:              path,
				deepApparentSize:  fileInfo.ApparentSize(),
				deepAllocatedSize: fileInfo.AllocatedSize(),
				FreeSpace:         freeSpace,
				TotalSpace:        totalSpace,
			}}, fileInfo.ApparentSize(), fileInfo.AllocatedSize(), nil
		}
		return nil, 0, 0, err
	}
	for _, fileInfo := range fileInfos {
		var apparent, allocated int64
		entryPath := filepath.Join(path, fileInfo.Name())
		if !fileInfo.IsDir() {
			apparent = fileInfo.ApparentSize()
			allocated = fileInfo.AllocatedSize()
		} else {
			_, apparent, allocated, err = ReadDirDeep(entryPath)
			if err != nil {
				return nil, 0, 0, err
			}
			apparent += fileInfo.ApparentSize()
			allocated += fileInfo.AllocatedSize()
		}
		totalApparent += apparent
		totalAllocated += allocated
		freeSpace, totalSpace, _, _ := internal.FreeSpace(entryPath)
		dirEntries = append(dirEntries, DirEntry{
			FileInfo:          fileInfo,
			path:              entryPath,
			deepApparentSize:  apparent,
			deepAllocatedSize: allocated,
			FreeSpace:         freeSpace,
			TotalSpace:        totalSpace,
		})
	}

	return dirEntries, totalApparent, totalAllocated, nil
}
