package tick

import (
	"regexp"
	"time"
)

type javaFmtGoLayout struct {
	reg    *regexp.Regexp
	layout string
}

var timeFormatConvert = []javaFmtGoLayout{
	{reg: regexp.MustCompile(`(?i)yyyy`), layout: "2006"},
	{reg: regexp.MustCompile(`(?i)yy`), layout: "06"},
	{reg: regexp.MustCompile(`MM`), layout: "01"},
	{reg: regexp.MustCompile(`(?i)dd`), layout: "02"},
	{reg: regexp.MustCompile(`(?i)hh`), layout: "15"},
	{reg: regexp.MustCompile(`mm`), layout: "04"},
	{reg: regexp.MustCompile(`(?i)sss`), layout: "000"},
	{reg: regexp.MustCompile(`(?i)ss`), layout: "05"},
}

// ToLayout converts Java style layout to golang.
func ToLayout(s string) string {
	for _, f := range timeFormatConvert {
		s = f.reg.ReplaceAllString(s, f.layout)
	}

	return s
}

// FormatTime format time with Java style layout.
func FormatTime(t time.Time, format string) string {
	result := format
	for _, f := range timeFormatConvert {
		result = f.reg.ReplaceAllStringFunc(result, func(formatExpr string) string {
			return t.Format(f.layout)
		})
	}

	return result
}
