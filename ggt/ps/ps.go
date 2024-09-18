package ps

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gum"
	"github.com/bingoohuang/ngg/ss"
	"github.com/glebarez/sqlite"
	"github.com/klauspost/cpuid/v2"
	ps "github.com/mitchellh/go-ps"
	"github.com/samber/lo"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	Register(root.Cmd)
}

type Cmd struct {
	*root.RootCmd `kong:"-"`

	Pid         int           `short:"p" help:"pid"`
	Watch       time.Duration `short:"w" help:"watch interval"`
	Children    bool          `short:"C" help:"including children processes"`
	Disk        bool          `short:"d" help:"only disk"`
	Cpu         bool          `short:"u" help:"only cpu"`
	Open        bool          `short:"o" help:"show open files of process"`
	Connections bool          `short:"c" help:"show TCP connections of process"`
	Env         bool          `short:"e" help:"show environment variables of process"`
}

func Register(rootCmd *root.RootCmd) {
	c := &cobra.Command{
		Use:   "ps",
		Short: "process info",
	}

	fc := &Cmd{RootCmd: rootCmd}
	c.RunE = fc.run
	ss.PanicErr(root.InitFlags(fc, c.Flags()))
	rootCmd.AddCommand(c)
}

func (r *Cmd) run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if r.Pid == 0 && len(args) > 0 {
		if argPid, err := ss.Parse[int](args[0]); err == nil {
			r.Pid = argPid
		} else {
			r.Pid = chooseProcess(ctx, args)
		}
		if r.Pid == 0 {
			fmt.Printf("unable to find process by %v\n", args)
			os.Exit(1)
		}
	}

	if r.Pid > 0 && r.Watch > 0 {
		return watchProcess(r.Pid, r.Watch, r.Children)
	}

	if r.Pid > 0 {
		return r.printProcessInfo(context.TODO())
	}

	if r.Cpu {
		return showCpuInfo()
	}

	if r.Disk {
		return showDisks()
	}

	return printSystemInfo()
}

func chooseProcess(ctx context.Context, args []string) int {
	pss, _ := process.ProcessesWithContext(ctx)
	var selectedProcesses []*process.Process
	for _, p := range pss {
		if p.Pid == int32(os.Getpid()) {
			continue
		}

		cmdLine, _ := p.Cmdline()
		if !strings.HasPrefix(cmdLine, "/") {
			// 忽略 tail/less 等命令行
			continue
		}
		if ss.HasPrefix(cmdLine, "/bin/bash", "bin/sh") {
			continue
		}

		if ss.ContainsFold(cmdLine, args...) {
			selectedProcesses = append(selectedProcesses, p)
		}
	}
	switch len(selectedProcesses) {
	case 0:
		return 0
	case 1:
		return int(selectedProcesses[0].Pid)
	default:
		var chooseItems []string
		for _, p := range selectedProcesses {
			chooseItems = append(chooseItems, fmt.Sprintf("%d\t%s", p.Pid, ss.Pick1(p.Cmdline())))
		}
		result, _ := gum.Choose(chooseItems, 1)
		if len(result) > 0 {
			pid, _ := ss.Parse[int](ss.Fields(result[0], 2)[0])
			return pid
		}
	}

	return 0
}

var printf = func(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

// VirtualMemoryStat usage statistics. Total, Available and Used contain numbers of bytes
// for human consumption.
//
// The other fields in this struct contain kernel specific values.
type VirtualMemoryStat struct {
	// Total amount of RAM on this system
	Total string `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	Available string `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	Used string `json:"used"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free string `json:"free"`
}

func printSystemInfo() error {
	printf("host:\t%s", Pick1(json.Marshal(Pick1(host.Info()))))
	printf("load:\t%s", Pick1(json.Marshal(Pick1(load.Avg()))))
	memoryStat := Pick1(mem.VirtualMemory())

	humanMemoryStat := &VirtualMemoryStat{
		Total:     humanIBytes(memoryStat.Total),
		Available: humanIBytes(memoryStat.Available),
		Used:      humanIBytes(memoryStat.Used),
		Free:      humanIBytes(memoryStat.Free),
	}

	printf("memory:\t%s", Pick1(json.Marshal(memoryStat)))
	printf("memory(human):\t%s", Pick1(json.Marshal(humanMemoryStat)))
	printf("mac addr:\t%v", Pick1Err(GetMacAddresses()))

	return showCpuInfo()
}

func showCpuInfo() error {
	cpu := cpuid.CPU
	// Print basic CPU information:
	fmt.Println("CPU BrandName:", cpu.BrandName)
	fmt.Println("CPU PhysicalCores:", cpu.PhysicalCores)
	fmt.Println("CPU ThreadsPerCore:", cpu.ThreadsPerCore)
	fmt.Println("CPU LogicalCores:", cpu.LogicalCores)
	fmt.Println("CPU Family", cpu.Family, "Model:", cpu.Model, "Vendor ID:", cpu.VendorID)
	fmt.Println("CPU Features:", strings.Join(cpu.FeatureSet(), ","))
	fmt.Println("CPU Cacheline bytes:", cpu.CacheLine)
	fmt.Println("CPU L1 Data Cache:", cpu.Cache.L1D, "bytes")
	fmt.Println("CPU L1 Instruction Cache:", cpu.Cache.L1I, "bytes")
	fmt.Println("CPU L2 Cache:", cpu.Cache.L2, "bytes")
	fmt.Println("CPU L3 Cache:", cpu.Cache.L3, "bytes")
	fmt.Println("CPU Frequency", cpu.Hz, "hz")

	return nil
}

func showDisks() error {
	infos, _ := GetDiskInfos()
	for _, info := range infos {
		printf("diskInfo:\t%s", Pick1(json.Marshal(info)))
		fmt.Println()
	}

	fmt.Println()

	partitions, _ := disk.Partitions(true)
	sort.Slice(partitions, func(i, j int) bool {
		a, _ := disk.Usage(partitions[i].Mountpoint)
		b, _ := disk.Usage(partitions[j].Mountpoint)
		return a.Total > b.Total
	})

	for i, p := range partitions {
		printf("#%02d disk %s usage:\t%s  partition:\t%s", i+1, p.Mountpoint,
			Pick1(json.Marshal(Pick1(disk.Usage(p.Mountpoint)))), Pick1(json.Marshal(p)))
	}

	return nil
}

var LogLevel = func() logger.LogLevel {
	env := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(env) {
	case "info":
		return logger.Info
	case "silent":
		return logger.Silent
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}()

func openDB(dbName string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(LogLevel),
	})
	if err != nil {
		log.Fatalf("gorm open %s error: %v", dbName, err)
	}

	// 迁移 schema
	if err = db.AutoMigrate(&ProcessTick{}); err != nil {
		log.Printf("AutoMigrate error: %v", err)
	}

	return db
}

func watchProcess(pid int, watchInterval time.Duration, includingChildren bool) error {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("process.NewProcess %d: %w", pid, err)
	}

	db := openDB(fmt.Sprintf("ps-%d.db", pid))
	defer func() {
		if d, _ := db.DB(); d != nil {
			d.Close()
		}
	}()

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// 创建一个用于接收信号的通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动一个 goroutine 监听信号
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %s", sig)
		// 取消上下文
		cancel()
	}()

	for {
		select {
		case <-ticker.C:
			tickProcess(db, p, includingChildren)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type ProcessTick struct {
	Timestamp string `gorm:"primarykey"`

	RSS        uint64
	CPU        float64
	NumFD      int32
	NumThreads int32

	ChildrenRSS        uint64
	ChildrenCPU        float64
	ChildrenNumFD      int32
	ChildrenNumThreads int32
	Children           int
}

func tickProcess(db *gorm.DB, p *process.Process, includingChildren bool) {
	memInfo, err := p.MemoryInfo()
	if err != nil {
		log.Printf("get memory info error: %v", err)
		return
	}

	var t ProcessTick
	t.Timestamp = time.Now().Format(time.RFC3339)
	t.RSS = memInfo.RSS
	t.CPU = Pick1(p.CPUPercent())
	t.NumFD = Pick1(p.NumFDs())
	t.NumThreads = Pick1(p.NumThreads())

	if includingChildren {
		children, _ := p.Children()
		t.Children = len(children)
		for _, cp := range children {
			childMem, _ := cp.MemoryInfo()
			if childMem != nil {
				t.ChildrenRSS += childMem.RSS
			}
			t.ChildrenCPU += Pick1(cp.CPUPercent())
			t.ChildrenNumFD += Pick1(cp.NumFDs())
			t.ChildrenNumThreads += Pick1(cp.NumThreads())
		}
	}

	jt, _ := json.Marshal(t)
	log.Printf("Tick: %s", jt)

	db.Save(t)
}

func (r *Cmd) printProcessInfo(ctx context.Context) error {
	pid := r.Pid
	p, err := ps.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if p == nil {
		return fmt.Errorf("process %d not found", pid)
	}
	printf("Pid:\t%d", pid)
	printf("Ppid:\t%v", GetParentPIDs(pid))
	printf("Executable:\t%s", p.Executable())

	p2, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("process.NewProcess %d: %w", pid, err)
	}

	cmdLine := Pick1(p2.Cmdline())
	printf("Cmdline:\t%s", cmdLine)
	cmdlineSlice := Pick1(p2.CmdlineSlice())
	if !(len(cmdlineSlice) == 1 && cmdlineSlice[0] == cmdLine) {
		printf("CmdlineSlice:\t%v", cmdlineSlice)
	}
	printf("Username:\t%v", Pick1Err(p2.Username()))

	gids, err := p2.Gids()
	gids = lo.Uniq(gids)
	printf("Gids:\t%v", Pick1Err(gids, err))
	gidNames := lo.Map(gids, func(item uint32, index int) string {
		name, err := user.LookupGroupId(fmt.Sprintf("%d", item))
		if err != nil {
			return "error=>" + err.Error()
		}
		return name.Name
	})
	printf("GidNames:\t%v", Pick1Err(gidNames, err))

	groups, err := p2.Groups()
	groups = lo.Uniq(groups)
	printf("Groups:\t%v", Pick1Err(groups, err))
	groupNames := lo.Map(groups, func(item uint32, index int) string {
		name, err := user.LookupGroupId(fmt.Sprintf("%d", item))
		if err != nil {
			return "error=>" + err.Error()
		}
		return name.Name
	})
	printf("GroupsNames:\t%v", groupNames)

	printf("Cwd:\t%v", Pick1Err(p2.Cwd()))
	printf("Exe:\t%v", Pick1Err(p2.Exe()))
	printf("CPUPercent:\t%f", Pick1(p2.CPUPercent()))
	createTime := time.UnixMilli(Pick1(p2.CreateTime())).In(time.Local)
	printf("CreateTime:\t%s, elapsed %s",
		createTime.In(time.Local).Format("2006-01-02 15:04:05"),
		time.Since(createTime))
	printf("Background:\t%t", Pick1(p2.Background()))
	printf("Name:\t%s", Pick1(p2.Name()))
	printf("Status:\t%v", Pick1(p2.Status()))

	if r.Env {
		index := 0
		envs := Pick1(p2.Environ())
		width := len(strconv.Itoa(len(envs)))
		for _, env := range envs {
			if env != "" {
				index++
				printf("Environ:\t#%0*d %v", width, index, env)
			}
		}
	}

	printf("IsRunning:\t%v", Pick1(p2.IsRunning()))
	printf("MemoryInfo:\t%+v", ToMemoryInfoStat(Pick1(p2.MemoryInfo())))
	printf("NumFDs:\t%d", Pick1(p2.NumFDs()))

	if r.Open {
		openFiles, err := p2.OpenFilesWithContext(ctx)
		if err != nil {
			printf("OpenFiles: \t%s", err)
		} else {
			width := len(strconv.Itoa(len(openFiles)))
			for i, fd := range openFiles {
				fi := ""
				if stat, err := os.Stat(fd.Path); err == nil {
					fi = fmt.Sprintf(", size: %s, modified: %s", ss.IBytes(uint64(stat.Size())), stat.ModTime().Format(time.RFC3339))
				}

				printf("OpenFile: \t%0*d %s(fd: %d%s)", width, i+1, fd.Path, fd.Fd, fi)
			}
		}
	}

	printf("NumThreads:\t%v", Pick1(p2.NumThreads()))
	children := lo.Map(Pick1(p2.Children()),
		func(item *process.Process, index int) int32 {
			return item.Pid
		},
	)
	printf("Children:\t%v (Total %d)", children, len(children))

	if r.Connections {
		index := 0
		conns := Pick1(p2.Connections())
		width := len(strconv.Itoa(len(conns)))
		for _, c := range conns {
			if c.Status != "NONE" {
				index++
				fmt.Printf("Local/Remote:\t#%0*d %v:%v / %v:%v (%v)\n", width, index,
					c.Laddr.IP, c.Laddr.Port, c.Raddr.IP, c.Raddr.Port, c.Status)
			}
		}
	}

	return nil
}

// Pick1Err means pick the first element.
func Pick1Err[T any](arg1 T, err error) any {
	if err != nil {
		return "error=>" + err.Error()
	}
	return arg1
}

// Pick1 means pick the first element.
func Pick1[T1 any](arg1 T1, _ ...any) T1 {
	return arg1
}

type MemoryInfoStat struct {
	RSS    string `json:"rss"`    // bytes
	VMS    string `json:"vms"`    // bytes
	HWM    string `json:"hwm"`    // bytes
	Data   string `json:"data"`   // bytes
	Stack  string `json:"stack"`  // bytes
	Locked string `json:"locked"` // bytes
	Swap   string `json:"swap"`   // bytes
}

func ToMemoryInfoStat(p *process.MemoryInfoStat) MemoryInfoStat {
	return MemoryInfoStat{
		RSS:    humanIBytes(p.RSS),
		VMS:    humanIBytes(p.VMS),
		HWM:    humanIBytes(p.HWM),
		Data:   humanIBytes(p.Data),
		Stack:  humanIBytes(p.Stack),
		Locked: humanIBytes(p.Locked),
		Swap:   humanIBytes(p.Swap),
	}
}

func humanIBytes(v uint64) string {
	if v == 0 {
		return "0"
	}

	s := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, ss.IBytes(v))

	return fmt.Sprintf("%d/%s", v, s)
}

// RawDiskInfo 磁盘信息
type RawDiskInfo struct {
	Name       string `json:"name"`
	TypeName   string `json:"typeName"`
	Type       int    `json:"type"`
	SectorNum  int    `json:"sectorNum"`  // 扇区数
	SectorSize int    `json:"sectorSize"` // 扇区大小（Bytes）
}

// GetDiskInfos 获取磁盘信息
func GetDiskInfos() ([]RawDiskInfo, error) {
	fInfos, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, err
	}

	var disks []RawDiskInfo
	for _, it := range fInfos {
		name := it.Name()
		if strings.HasPrefix(name, "loop") || strings.Contains(name, "ram") {
			continue
		}

		rotational := filepath.Join("/sys/block", name, "queue/rotational")
		size := filepath.Join("/sys/block", name, "size")
		sectorSize := filepath.Join("/sys/block", name, "queue/hw_sector_size")

		buf, err := os.ReadFile(rotational)
		if err != nil || len(buf) == 0 {
			continue
		}

		buf1, err := os.ReadFile(size)
		if err != nil {
			continue
		}

		buf2, err := os.ReadFile(sectorSize)
		if err != nil {
			continue
		}

		diskInfo := RawDiskInfo{
			Name:       name,
			Type:       mustInt(string(buf[0])),
			SectorNum:  mustInt(string(buf1)),
			SectorSize: mustInt(string(buf2)),
		}

		if diskInfo.Type == int(DiskTypeSSD) && strings.HasPrefix(name, "mmcblk") {
			diskInfo.Type = int(DiskTypeEMMC)
		}
		diskInfo.TypeName = DiskTypeNames[diskInfo.Type]

		disks = append(disks, diskInfo)
	}

	return disks, nil
}

type DiskType int32

const (
	DiskTypeSSD DiskType = iota
	DiskTypeHHD
	DiskTypeEMMC
)

var DiskTypeNames = map[int]string{
	0: "SSD",
	1: "HHD",
	2: "EMMC",
}

func mustInt(s string) int {
	s = strings.TrimSpace(s)
	i, _ := strconv.Atoi(s)
	return i
}

func GetMacAddresses() (macAddrs []string, err error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("get net interfaces: %w", err)
	}

	for _, i := range netInterfaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}

		macAddr := i.HardwareAddr
		if len(i.HardwareAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr.String())
	}
	return macAddrs, nil
}

func GetParentPIDs(pid int) []int {
	var parentPIDs []int
	for {
		p, err := ps.FindProcess(pid)
		if err != nil {
			break
		}
		// 获取父级 PID
		ppid := p.PPid()
		if ppid == 0 {
			break
		}

		parentPIDs = append(parentPIDs, ppid)
		pid = ppid
	}

	return parentPIDs
}
