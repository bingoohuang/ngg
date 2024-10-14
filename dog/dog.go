package dog

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/shirou/gopsutil/v4/process"
)

type Dog struct {
	*Config

	states []*thresholdState
}

// removeFiles 删除指定目录 dir 下，符合 pattern 的文件
func removeFiles(dir, pattern string) (removeFiles []string) {
	if err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if ok, _ := filepath.Match(pattern, info.Name()); !ok {
			return nil
		}
		removeFiles = append(removeFiles, path)

		return nil
	}); err != nil {
		log.Printf("E! walkdir: %q error: %v", dir, err)
	}

	for _, f := range removeFiles {
		if e := os.Remove(f); e != nil {
			log.Printf("E! clean file: %s, error: %v", f, e)
		} else {
			log.Printf("clean file: %s", f)
		}
	}

	return
}

func New(options ...ConfigFn) *Dog {
	d := &Dog{
		Config: createConfig(options),
	}

	// 删除历史文件，例如
	// Dog.cpu.868.20241010174315.pprof
	// Dog.cpu.872.20241010173506.pprof
	removeFiles(d.Dir, "Dog.*.pprof")

	if d.RSSThreshold > 0 {
		d.states = append(d.states, newThresholdState(RSS, d.RSSThreshold, d.statRSS, d.Dir, d.Pid))
	}
	if d.CPUPercentThreshold > 0 {
		d.states = append(d.states, newThresholdState(CPU, d.CPUPercentThreshold, d.statCPU, d.Dir, d.Pid))
	}

	return d
}

type State struct {
	RSS        uint64
	VMS        uint64
	CPUPercent float64
}

func (w *Dog) Watch(ctx context.Context) error {
	pid := w.Pid

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("get process %d: %w", pid, err)
	}

	return tick.Tick(ctx, w.Interval, w.Jitter, func() error {
		w.stat(p)
		if reasons, yes := w.reachTimes(); yes {
			if w.Debug {
				log.Printf("godo reach times: %v", reasons)
			}

			w.Action.DoAction(w.Dir, w.Debug, reasons)
		}

		return nil
	})

}

type statFn func(p *process.Process, state *thresholdState) (debugMessage string)

func (w *Dog) stat(p *process.Process) {
	var debugMessages []string
	for _, state := range w.states {
		if msg := state.statFn(p, state); msg != "" {
			debugMessages = append(debugMessages, msg)
		}
	}

	if len(debugMessages) > 0 {
		log.Printf("%s", strings.Join(debugMessages, ", "))
	}
}

func (w *Dog) statRSS(p *process.Process, state *thresholdState) (debugMessage string) {
	// 获取内存信息
	if memInfo, err := p.MemoryInfo(); err == nil {
		rss := memInfo.RSS // 常驻集大小，即实际使用的物理内存
		state.setReached(w.Debug, rss)

		if w.Debug {
			debugMessage = fmt.Sprintf("current RSS: %s", ss.IBytes(rss))
		}
	} else if w.Debug {
		log.Printf("E! get memory %d error: %v", p.Pid, err)
	}

	return
}

func (w *Dog) statCPU(p *process.Process, state *thresholdState) (debugMessage string) {
	// 获取CPU使用情况，如果 cpuPercent 是 70%，这里值是 70
	if cpuPercent, err := p.CPUPercent(); err == nil {
		state.setReached(w.Debug, uint64(cpuPercent))
		if w.Debug {
			debugMessage = fmt.Sprintf("CPU: %f", cpuPercent)
		}
	} else if w.Debug {
		log.Printf("E! get cpu percent %d error: %v", p.Pid, err)
	}

	return
}

type ReasonItem struct {
	Type      ThresholdType `json:"type"`
	Reason    string        `json:"reason"`
	Values    []uint64      `json:"values"`
	Threshold any           `json:"threshold"`
	Profile   string        `json:"profile"`
}

func (w *Dog) reachTimes() (reasons []ReasonItem, reached bool) {
	for _, state := range w.states {
		if r := state.reached(w.Times, w.Debug); r.Reached {
			reasons = append(reasons, ReasonItem{
				Type:      state.Type,
				Reason:    fmt.Sprintf("连续 %d 次超标", w.Times),
				Values:    r.Values,
				Threshold: state.Threshold,
				Profile:   r.Profile,
			})
			reached = true
		}
	}

	return reasons, reached
}

type ThresholdType string

const (
	RSS ThresholdType = "RSS"
	CPU ThresholdType = "CPU"
)

type thresholdState struct {
	Type      ThresholdType
	Threshold uint64
	Values    []uint64

	statFn

	profile string
	prof    ss.Prof
	Dir     string
	Pid     int
}

func newThresholdState(typ ThresholdType, threshold uint64, fn statFn, dir string, pid int) *thresholdState {
	return &thresholdState{
		Type:      typ,
		Threshold: threshold,
		statFn:    fn,
		Dir:       dir,
		Pid:       pid,
	}
}

type reachResult struct {
	Profile string
	Values  []uint64
	Reached bool
}

func (t *thresholdState) reached(maxTimes int, debug bool) (r reachResult) {
	if debug && len(t.Values) > 0 {
		log.Printf("current %s thresholdState: %v", t.Type, t.Values)
	}

	if r.Reached = len(t.Values) >= maxTimes; r.Reached {
		r.Values = t.Values
		t.Values = nil

		if err := t.prof.Close(); err != nil && debug {
			log.Printf("E! close profile error: %v", err)
		}
		r.Profile = t.profile
		t.prof = nil
	}

	return
}

func (t *thresholdState) setReached(debug bool, value uint64) {
	if reached := value > t.Threshold; reached {
		if t.prof == nil {
			t.prof = ss.NoopProfile
			timestamp := time.Now().Format(`20060102150405`)
			switch t.Type {
			case CPU:
				name := filepath.Join(t.Dir, fmt.Sprintf("Dog.cpu.%d.%s.pprof", t.Pid, timestamp))
				p, err := ss.StartCPUProf(name)
				if err == nil {
					t.prof = p
					t.profile = name
				} else if debug {
					log.Printf("E! create cpu profile error: %v", err)
				}
			case RSS:
				name := filepath.Join(t.Dir, fmt.Sprintf("Dog.mem.%d.%s.pprof", t.Pid, timestamp))
				p, err := ss.StartMemProf(name)
				if err == nil {
					t.prof = p
					t.profile = name
				} else if debug {
					log.Printf("E! create mem profile error: %v", err)
				}
			}
		}
		t.Values = append(t.Values, value)
	} else {
		if t.prof != nil {
			if err := t.prof.Close(); err != nil && debug {
				log.Printf("E! profile close error: %v", err)
			}
			_ = os.Remove(t.profile)
			t.prof = nil
			t.profile = ""
		}
		if len(t.Values) > 0 {
			t.Values = t.Values[:0]
		}
	}
}
