package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/bingoohuang/ngg/lsdu"
	"github.com/dustin/go-humanize"
	flag "github.com/spf13/pflag"
)

func main() {
	var sortBySize, dirOnly, si bool
	flag.BoolVarP(&sortBySize, "sort", "s", true, "sort by size")
	flag.BoolVarP(&dirOnly, "dir", "d", false, "list directories themselves, not their contents")
	flag.BoolVar(&si, "si", false, "use SI units (1000 instead of 1024)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [FILE|DIRECTORY]...\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nOutput Columns Explanation:\n")
		fmt.Fprintf(os.Stderr, "  #          : Sequence number.\n")
		fmt.Fprintf(os.Stderr, "  Mode       : File mode and permissions.\n")
		fmt.Fprintf(os.Stderr, "  Apparent   : Logical size (sum of file contents).\n")
		fmt.Fprintf(os.Stderr, "  Allocated  : Physical size (actual disk usage, blocks * 512).\n")
		fmt.Fprintf(os.Stderr, "  FreeSpace  : Free space on the filesystem.\n")
		fmt.Fprintf(os.Stderr, "  TotalSpace : Total capacity of the filesystem.\n")
		fmt.Fprintf(os.Stderr, "  ModTime    : Last modification time.\n")
		fmt.Fprintf(os.Stderr, "  Path       : File or directory path.\n")
	}
	flag.Parse()

	paths := []string{"."}
	if flag.NArg() > 0 {
		paths = flag.Args()
	}

	var grandTotalApparent, grandTotalAllocated int64
	var allEntries []lsdu.DirEntry
	for _, path := range paths {
		if dirOnly {
			entry, apparent, allocated, err := lsdu.StatDir(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			allEntries = append(allEntries, entry)
			grandTotalApparent += apparent
			grandTotalAllocated += allocated
		} else {
			fmt.Printf("%s:\n", path)
			dirEntries, totalApparent, totalAllocated, err := lsdu.ReadDirDeep(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if sortBySize {
				sort.Slice(dirEntries, func(i, j int) bool {
					return dirEntries[i].Size() < dirEntries[j].Size()
				})
			}
			Print(dirEntries, si)
			grandTotalApparent += totalApparent
			grandTotalAllocated += totalAllocated
		}
	}
	if dirOnly {
		if sortBySize {
			sort.Slice(allEntries, func(i, j int) bool {
				return allEntries[i].Size() < allEntries[j].Size()
			})
		}
		Print(allEntries, si)
	}
}

// Print displays directory entries in Markdown table format with alignment padding
func Print(fis []lsdu.DirEntry, si bool) {
	if len(fis) == 0 {
		return
	}

	type row struct {
		idx     string
		mode    string
		appa    string
		allo    string
		free    string
		total   string
		modTime string
		path    string
	}

	headers := row{"#", "Mode", "Apparent", "Allocated", "FreeSpace", "TotalSpace", "ModTime", "Path"}
	rows := make([]row, len(fis))

	wIdx := len(headers.idx)
	wMode := len(headers.mode)
	wAppa := len(headers.appa)
	wAllo := len(headers.allo)
	wFree := len(headers.free)
	wTotal := len(headers.total)
	wTime := len(headers.modTime)
	wPath := len(headers.path)

	var totalAppa, totalAllo int64
	for i, fi := range fis {
		totalAppa += fi.DeepApparentSize()
		totalAllo += fi.DeepAllocatedSize()
		r := row{
			idx:     fmt.Sprintf("%d", i+1),
			mode:    fmt.Sprintf("%v", fi.Mode()),
			appa:    fmt.Sprintf("%s(%d)", Humanize(fi.DeepApparentSize(), si), fi.DeepApparentSize()),
			allo:    fmt.Sprintf("%s(%d)", Humanize(fi.DeepAllocatedSize(), si), fi.DeepAllocatedSize()),
			free:    fmt.Sprintf("%s(%d)", Humanize(int64(fi.FreeSpace), si), fi.FreeSpace),
			total:   fmt.Sprintf("%s(%d)", Humanize(int64(fi.TotalSpace), si), fi.TotalSpace),
			modTime: fi.ModTime().Format("2006-01-02 15:04:05"),
			path:    maybeQuote(fi.Path()),
		}
		rows[i] = r

		if len(r.idx) > wIdx {
			wIdx = len(r.idx)
		}
		if len(r.mode) > wMode {
			wMode = len(r.mode)
		}
		if len(r.appa) > wAppa {
			wAppa = len(r.appa)
		}
		if len(r.allo) > wAllo {
			wAllo = len(r.allo)
		}
		if len(r.free) > wFree {
			wFree = len(r.free)
		}
		if len(r.total) > wTotal {
			wTotal = len(r.total)
		}
		if len(r.modTime) > wTime {
			wTime = len(r.modTime)
		}
		if len(r.path) > wPath {
			wPath = len(r.path)
		}
	}

	totalRow := row{
		idx:     "-",
		mode:    "total",
		appa:    fmt.Sprintf("%s(%d)", Humanize(totalAppa, si), totalAppa),
		allo:    fmt.Sprintf("%s(%d)", Humanize(totalAllo, si), totalAllo),
		free:    "-",
		total:   "-",
		modTime: "-",
		path:    "-",
	}
	if len(totalRow.mode) > wMode {
		wMode = len(totalRow.mode)
	}
	if len(totalRow.appa) > wAppa {
		wAppa = len(totalRow.appa)
	}
	if len(totalRow.allo) > wAllo {
		wAllo = len(totalRow.allo)
	}

	// Print header
	fmt.Printf("| %*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
		wIdx, headers.idx, wMode, headers.mode, wAppa, headers.appa, wAllo, headers.allo, wFree, headers.free, wTotal, headers.total, wTime, headers.modTime, wPath, headers.path)
	// Print separator
	fmt.Printf("| %s | %s | %s | %s | %s | %s | %s | %s |\n",
		strings.Repeat("-", wIdx), strings.Repeat("-", wMode), strings.Repeat("-", wAppa),
		strings.Repeat("-", wAllo), strings.Repeat("-", wFree), strings.Repeat("-", wTotal), strings.Repeat("-", wTime), strings.Repeat("-", wPath))

	// Print rows
	for _, r := range rows {
		fmt.Printf("| %*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
			wIdx, r.idx, wMode, r.mode, wAppa, r.appa, wAllo, r.allo, wFree, r.free, wTotal, r.total, wTime, r.modTime, wPath, r.path)
	}
	// Print total row
	if len(fis) > 1 {
		fmt.Printf("| %*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
			wIdx, totalRow.idx, wMode, totalRow.mode, wAppa, totalRow.appa, wAllo, totalRow.allo, wFree, totalRow.free, wTotal, totalRow.total, wTime, totalRow.modTime, wPath, totalRow.path)
	}
}

// maybeQuote quotes a string with single quotes if it contains spaces
func maybeQuote(s string) string {
	if containsSpace(s) {
		return fmt.Sprintf("'%s'", s)
	}
	return s
}

func containsSpace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// Humanize reports the size in power of 1024 units (KiB, MiB etc) or 1000 units (KB, MB etc) if si is true
func Humanize(size int64, si bool) string {
	if si {
		return humanize.Bytes(uint64(size))
	}

	return humanize.IBytes(uint64(size))
}
