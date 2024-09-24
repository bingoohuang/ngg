package time

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/spf13/cobra"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:   "time",
		Short: "convert unix time and human-readable format",
		RunE:  fc.run,
	}

	root.AddCommand(c, fc)
}

type subCmd struct {
}

func (f *subCmd) run(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		printTime("", time.Now())
		return nil
	}

	for i, arg := range args {
		if i > 0 {
			fmt.Println()
		}

		if regexp.MustCompile(`^\d+$`).MatchString(arg) {
			found := false
			for _, f := range []string{`20060102150405`, `200601021504`, `2006010215`, `20060102`} {
				t, err := time.ParseInLocation(f, arg, time.Local)
				if err == nil {
					printTime(arg, t)
					found = true
					break
				}
			}

			if !found {
				printUnixTime(arg)
			}
			continue
		}

		t, err := time.Parse(time.RFC3339, arg)
		if err == nil {
			printTime(arg, t)
			continue
		}
		formats := []string{
			`2006-01-02T15:04:05`,
			`2006-01-02T15:04`,
			`2006-01-02T15`,
			`2006-01-02 15:04:05`,
			`2006-01-02 15:04`,
			`2006-01-02 15`,
			`2006-01-02`,
		}
		for _, f := range formats {
			t, err = time.ParseInLocation(f, arg, time.Local)
			if err == nil {
				printTime(arg, t)
				break
			}
		}
	}

	return nil
}

func printUnixTime(arg string) {
	if arg != "" {
		fmt.Println(arg, "intercepted:")
	}
	v, _ := strconv.ParseInt(arg, 10, 64)

	if d := time.Unix(v, 0); d.Year() <= 9999 {
		fmt.Println("as unix:\t", d.Format(time.RFC3339))
	} else if d = time.UnixMilli(v); d.Year() <= 9999 {
		fmt.Println("as unix milli:\t", d.Format(time.RFC3339))
	} else if d = time.Unix(v/1e9, v%1e9); d.Year() <= 9999 {
		fmt.Println("as unix nano:\t", d.Format(time.RFC3339))
	}
}

func printTime(arg string, now time.Time) {
	if arg != "" {
		fmt.Println(arg, "intercepted:")
	}
	fmt.Println("now:\t\t", now.Format(time.RFC3339))
	fmt.Println("unix:\t\t", now.Unix())
	fmt.Println("unix milli:\t", now.UnixMilli())
	fmt.Println("unix micro:\t", now.UnixMicro())
	fmt.Println("unix nano:\t", now.UnixNano())
}

// List of supported time layouts.
var formats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.RFC1123Z,
	time.RFC1123,
	time.RFC850,
	time.RFC822Z,
	time.RFC822,
	time.Layout,
	time.RubyDate,
	time.UnixDate,
	time.ANSIC,
	time.StampNano,
	time.StampMicro,
	time.StampMilli,
	time.Stamp,
	time.Kitchen,
}

const (
	maxNanoseconds  = int64(math.MaxInt64)
	maxMicroseconds = maxNanoseconds / 1000
	maxMilliseconds = maxMicroseconds / 1000
	maxSeconds      = maxMilliseconds / 1000

	minNanoseconds  = int64(math.MinInt64)
	minMicroseconds = minNanoseconds / 1000
	minMilliseconds = minMicroseconds / 1000
	minSeconds      = minMilliseconds / 1000
)
