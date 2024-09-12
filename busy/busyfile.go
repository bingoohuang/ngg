package busy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
)

const DefaultCheckBusyInterval = 10 * time.Second

func Watch(ctx context.Context, dir string, debug bool, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tickBusy(ctx, dir, debug)
		}
	}
}

const DogBusy = "Dog.busy"

func tickBusy(ctx context.Context, dir string, debug bool) {
	var file File
	name := filepath.Join(dir, DogBusy)
	if err := ReadDeleteFile(name, debug, &file); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("E! readFile: %s, error: %v", name, err)
		}
		return
	}

	log.Printf("Got Dog.busy: %+v", file)

	if file.Mem != "" {
		go controlMem(ctx, file.Mem)
	}

	if file.Cpu > 0 {
		if file.Cores == 0 {
			file.Cores = int(math.Ceil(float64(file.Cpu) / 100))
		}
		go ControlCPULoad(ctx, file.Cores, file.Cpu/file.Cores, file.LockOsThread)
	}

	if file.Pprof.Duration > 0 {
		pid := os.Getpid()
		timestamp := time.Now().Format(`20060102150405`)
		name := filepath.Join(dir, fmt.Sprintf("Dog.mem.%d.%s.pprof", pid, timestamp))
		ppMem, err := ss.StartMemProf(name)
		if err != nil {
			log.Printf("E! create file mem profile error: %v", err)
		}
		name = filepath.Join(dir, fmt.Sprintf("Dog.cpu.%d.%s.pprof", pid, timestamp))
		ppCpu, err := ss.StartCPUProf(name)
		if err != nil {
			log.Printf("E! create file cpu profile error: %v", err)
		}

		go func() {
			time.Sleep(file.Pprof.Duration)
			ss.Close(ppMem, ppCpu)
		}()
	}
}

func controlMem(ctx context.Context, fileMem string) {
	maxMem, err := ss.ParseBytes(fileMem)
	if err != nil {
		log.Printf("humanizeBytes error: %v", err)
		return
	}
	if err := ControlMem(ctx, maxMem); err != nil {
		log.Printf("control mem to %s error: %v", fileMem, err)
	}
}

type File struct {
	Mem          string   `json:"mem,omitempty"`          // 最大内存
	Cores        int      `json:"cores,omitempty"`        // cpu 使用核数
	Cpu          int      `json:"cpu,omitempty"`          // cpu 每核百分比, 0-100
	LockOsThread bool     `json:"lockOsThread,omitempty"` // lockOsThread: 是否在 CPU 耗用时锁定 OS 线程
	Pprof        tick.Dur `json:"pprof,omitempty"`        // 指定时间后，生成 pprof 文件
}

func ReadDeleteFile(filename string, debug bool, v any) error {
	stat, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		_ = removeFile(filename, stat)
		return fmt.Errorf("read file %s: %w", filename, err)
	}

	if debug {
		log.Printf("read file %s: %q", filename, data)
	}

	if err := json.Unmarshal(data, v); err != nil {
		_ = removeFile(filename, stat)
		return fmt.Errorf("json unmarshal for %s: %w", filename, err)
	}

	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("remove file %s: %w", filename, err)
	}

	return nil
}

func removeFile(filename string, stat os.FileInfo) error {
	if time.Since(stat.ModTime()) > 10*time.Second {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("remove file %s: %w", filename, err)
		}
	}
	return nil
}
