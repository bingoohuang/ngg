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
	c := &cobra.Command{
		Use:   "time",
		Short: "convert unix time and human-readable format",
	}
	root.AddCommand(c, &subCmd{})
}

type subCmd struct {
}

func (f *subCmd) Run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		printTime("", time.Now())
		return nil
	}
NEXT:
	for i, arg := range args {
		if i > 0 {
			fmt.Println()
		}

		if regexp.MustCompile(`^\d+$`).MatchString(arg) {
			found := false
			for _, f := range []string{
				`20060102150405000000`,
				`20060102150405000`,
				`20060102150405`,
				`200601021504`,
				`2006010215`,
				`20060102`} {
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

		for _, f := range []string{
			time.RFC3339Nano,
			time.RFC3339,
		} {

			t, err := time.Parse(f, arg)
			if err == nil {
				printTime(arg, t)
				goto NEXT
			}
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
			t, err := time.ParseInLocation(f, arg, time.Local)
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
		fmt.Println("as unix:  ", d.Format(time.RFC3339Nano))
	} else if d = time.UnixMilli(v); d.Year() <= 9999 {
		fmt.Println("as milli: ", d.Format(time.RFC3339Nano))
	} else if d = time.Unix(v/1e9, v%1e9); d.Year() <= 9999 {
		fmt.Println("as nano:  ", d.Format(time.RFC3339Nano))
	}
}

func printTime(arg string, t time.Time) {
	if arg != "" {
		fmt.Println(arg, "intercepted:")
	}
	fmt.Println("now:   ", t.Format(time.RFC3339Nano))
	fmt.Println("unix:  ", t.Unix())
	fmt.Println("milli: ", t.UnixMilli())
	fmt.Println("micro: ", t.UnixMicro())
	fmt.Println("nano:  ", t.UnixNano())
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
