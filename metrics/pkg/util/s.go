package util

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
)

// Hostname stores hostname.
var Hostname = hostname() // nolint

func hostname() string {
	v, _ := os.Hostname()

	return v
}

// StripAny strips any Unicode code points in chars are within s.
func StripAny(s, chars string) string {
	filter := func(r rune) rune {
		if !strings.ContainsRune(chars, r) {
			return r
		}

		return -1
	}

	return strings.Map(filter, s)
}

// Esc escapes s to a human readable format.
func Esc(s string) string {
	j, _ := json.Marshal(s)
	return string(j)[1 : len(j)-1]
}

// JSONCompact compact the JSON encoding of data silently.
func JSONCompact(data interface{}) string {
	return ss.Pick1(JSONCompactE(data))
}

// JSONCompactE compact the JSON encoding of data.
func JSONCompactE(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, []byte(v)); err != nil {
			return "", err
		}

		return buffer.String(), nil
	case []byte:
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, v); err != nil {
			return "", err
		}

		return buffer.String(), nil
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		return string(b), nil
	}
}

// ConvertTimeLayout converts date time format in java style to go style.
func ConvertTimeLayout(layout string) string {
	l := layout
	l = strings.ReplaceAll(l, "yyyy", "2006")
	l = strings.ReplaceAll(l, "yy", "06")
	l = strings.ReplaceAll(l, "MM", "01")
	l = strings.ReplaceAll(l, "dd", "02")
	l = strings.ReplaceAll(l, "HH", "15")
	l = strings.ReplaceAll(l, "mm", "04")
	l = strings.ReplaceAll(l, "ss", "05")
	l = strings.ReplaceAll(l, "SSS", "000")

	return l
}

// ParseTime 解析日期转字符串.
func ParseTime(d string, layout string) (time.Time, error) {
	return time.Parse(ConvertTimeLayout(layout), d)
}

// FormatTime 日期转字符串.
func FormatTime(d time.Time, layout string) string {
	return d.Format(ConvertTimeLayout(layout))
}
