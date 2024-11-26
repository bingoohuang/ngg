package rotate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/metrics/pkg/util"
)

const yyyyMMdd = "yyyy-MM-dd"

type File struct {
	file     *os.File
	Filename string

	lastDay    string
	dir        string
	MaxBackups int
}

var Debug = false

// NewFile create a rotation option.
func NewFile(filename string, maxBackups int) (*File, error) {
	o := &File{
		Filename:   filename,
		MaxBackups: maxBackups,
		dir:        filepath.Dir(filename),
	}

	if err := os.MkdirAll(o.dir, 0o755); err != nil {
		return nil, err
	}

	if err := o.open(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *File) open() error {
	f, err := os.OpenFile(o.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("log file %s created error %w", o.Filename, err)
	}

	o.file = f

	if Debug {
		log.Printf("log file %s created", o.Filename)
	}

	return nil
}

func (o *File) rotateFiles(t time.Time) error {
	rotated, outMaxBackups := o.detectRotate(t)

	return o.doRotate(rotated, outMaxBackups)
}

func (o *File) doRotate(rotated string, outMaxBackups []string) error {
	if rotated != "" {
		if err := o.close(); err != nil {
			return err
		}

		if err := os.Rename(o.Filename, rotated); err != nil {
			return fmt.Errorf("rotate %s to %s error %w", o.Filename, rotated, err)
		}

		if Debug {
			log.Printf("%s rotated to %s", o.Filename, rotated)
		}

		if err := o.open(); err != nil {
			return err
		}
	}

	for _, old := range outMaxBackups {
		if err := os.Remove(old); err != nil {
			return fmt.Errorf("remove log file %s before max backup days %d error %v", old, o.MaxBackups, err)
		}

		if Debug {
			log.Printf("%s before max backup days %d removed", old, o.MaxBackups)
		}
	}

	return nil
}

func (o *File) close() error {
	if o.file == nil {
		return nil
	}

	err := o.file.Close()
	o.file = nil

	return err
}

func (o *File) detectRotate(t time.Time) (rotated string, outMaxBackups []string) {
	ts := util.FormatTime(t, yyyyMMdd)

	if o.lastDay == "" {
		o.lastDay = ts
	}

	if o.lastDay != ts {
		o.lastDay = ts

		yesterday := t.AddDate(0, 0, -1)
		rotated = o.Filename + "." + util.FormatTime(yesterday, yyyyMMdd)
	}

	if o.MaxBackups > 0 {
		day := t.AddDate(0, 0, -o.MaxBackups)
		_ = filepath.Walk(o.dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return err
			}

			if strings.HasPrefix(path, o.Filename+".") {
				fis := path[len(o.Filename+"."):]
				if backDay, err := util.ParseTime(fis, yyyyMMdd); err != nil {
					return nil // ignore this file
				} else if backDay.Before(day) {
					outMaxBackups = append(outMaxBackups, path)
				}
			}

			return nil
		})
	}

	return rotated, outMaxBackups
}

// Close closes the file.
func (o *File) Close() error {
	return o.close()
}

func (o *File) write(d []byte, flush bool) (int, error) {
	if err := o.rotateFiles(time.Now()); err != nil {
		return 0, err
	}

	n, err := o.file.Write(d)
	if err != nil {
		return n, err
	}

	if flush {
		err = o.file.Sync()
	}

	return n, err
}

// Write writes data to a file.
func (o *File) Write(d []byte) (int, error) {
	return o.write(d, false)
}

// Flush flushes the file.
func (o *File) Flush() error {
	return o.file.Sync()
}
