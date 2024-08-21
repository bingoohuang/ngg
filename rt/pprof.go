package rt

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

type Prof interface {
	io.Closer
}

// StartMemProf 创建内存性能分析文件
func StartMemProf(name string) (Prof, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("create profile %s: %w", name, err)
	}

	return &profile{
		file:     f,
		profType: pprofMem,
	}, nil
}

// StartCPUProf 创建 CPU 性能分析文件
func StartCPUProf(name string) (Prof, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("create profile %s: %w", name, err)
	}

	// 启动 CPU 性能分析
	if err := pprof.StartCPUProfile(f); err != nil {
		return nil, fmt.Errorf("start CPU profile: %w", err)
	}

	return &profile{
		file:     f,
		profType: pprofCPU,
	}, nil
}

type profType int

const (
	_ profType = iota
	pprofCPU
	pprofMem
)

type profile struct {
	file *os.File
	profType
}

func (c *profile) Close() error {
	switch c.profType {
	case pprofCPU:
		pprof.StopCPUProfile()
		if err := c.file.Close(); err != nil {
			return fmt.Errorf("close CPU profile: %w", err)
		}
	case pprofMem:
		// 进行内存性能分析并写入文件
		runtime.GC() // 触发 GC，获取最新的内存分配信息
		if err := pprof.WriteHeapProfile(c.file); err != nil {
			return fmt.Errorf("write heap profile: %w", err)
		}
		if err := c.file.Close(); err != nil {
			return fmt.Errorf("close CPU profile: %w", err)
		}
	}

	return nil
}

type noopProfile struct{}

var NoopProfile Prof = &noopProfile{}

func (n noopProfile) ProfileName() string { return "" }
func (n noopProfile) Close() error        { return nil }
