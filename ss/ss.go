package ss

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

func If[T any](condition bool, a, b T) T {
	if condition {
		return a
	}

	return b
}

func IfFunc[T any](condition bool, a, b func() T) T {
	if condition {
		return a()
	}

	return b()
}

// Repeat returns a new string consisting of count copies of the string s with separator.
//
// It panics if count is negative or if
// the result of (len(s) * count) overflows.
func Repeat(s, separator string, count int) string {
	return strings.Repeat(separator+s, count)[len(separator):]
}

func Contains(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}

	return false
}

func ContainsFold(s string, subs ...string) bool {
	s = strings.ToUpper(s)
	for _, sub := range subs {
		if strings.Contains(s, strings.ToUpper(sub)) {
			return true
		}
	}

	return false
}

func Must[A any](a A, err error) A {
	if err != nil {
		panic(err)
	}

	return a
}

// IndexN 返回字符串 s 中第 n 个子字符串 sub 的索引。
// 如果 n 是正数，则从左到右查找第 n 个 sub 的位置。
// 如果 n 是负数，则从右到左查找第 -n 个 sub 的位置。
// 如果找不到，返回 -1
func IndexN(s, sep string, n int) int {
	switch {
	case n == 0:
		return -1
	case n > 0:
		index, sepLen := 0, len(sep)
		for i := 0; i < n; i++ {
			idx := strings.Index(s, sep)
			if idx == -1 {
				return -1
			}
			s = s[idx+sepLen:]
			index += idx
		}
		return index
	default: // n < 0
		n = -n
		idx := -1
		for i := 0; i < n; i++ {
			idx = strings.LastIndex(s, sep)
			if idx == -1 {
				return -1
			}
			s = s[:idx]
		}
		return idx
	}

}

func Or[T comparable](a, b T) T {
	var zero T
	if a == zero {
		return b
	}

	return a
}

func AnyOfFunc[T any](target T, f func(idx int, elem, target T) bool, anys ...T) bool {
	return IndexOfFunc(target, f, anys...) >= 0
}

func IndexOfFunc[T any](target T, f func(idx int, elem, target T) bool, anys ...T) int {
	for i, el := range anys {
		if f(i, el, target) {
			return i
		}
	}

	return -1
}

func AnyOf[T comparable](target T, anys ...T) bool {
	return IndexOf(target, anys...) >= 0
}

func IndexOf[T comparable](target T, anys ...T) int {
	for i, el := range anys {
		if el == target {
			return i
		}
	}

	return -1
}

func SplitFunc(idx int, sub, subSep string, cur []string) (splitSub string, ok, kontinue bool) {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return "", false, true
	}

	return sub, true, true
}

func Split(s, sep string) []string {
	return SplitN(s, sep, -1, SplitFunc)
}

func Split2(s, sep string) (s1, s2 string) {
	res := SplitN(s, sep, 2, SplitFunc)
	if len(res) >= 2 {
		return res[0], res[1]
	} else if len(res) >= 1 {
		return res[0], ""
	} else {
		return "", ""
	}
}

func SplitN(s, sep string, n int, f func(idx int, sub, subSep string, cur []string) (splitSub string, ok, kontinue bool)) []string {
	if n == 0 {
		return nil
	}
	if sep == "" {
		panic("sep is empty")
	}

	if n < 0 {
		n = strings.Count(s, sep) + 1
	}

	if n > len(s)+1 {
		n = len(s) + 1
	}
	a := make([]string, 0, n)
	n--
	i := 0
	sepSave := len(sep)
	for i < n {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}

		sub, yes, kontinue := f(i, s[:m], s[:m+sepSave], a)
		if yes {
			a = append(a, sub)
		}
		if !kontinue {
			return a
		}

		s = s[m+len(sep):]
		i++
	}
	sub, yes, _ := f(i, s, s, a)
	if yes {
		a = append(a, sub)
	}

	if i+1 <= len(a) {
		return a[:i+1]
	}

	return a
}

func SplitSeps(s string, seps string, n int) []string {
	var v []string

	ff := FieldsFunc(s, n, func(r rune) bool {
		return strings.ContainsRune(seps, r)
	})

	for _, f := range ff {
		if f = strings.TrimSpace(f); f != "" {
			v = append(v, f)
		}
	}

	return v
}

// SplitToMap 将字符串 s 分割成 map, 其中 key 和 value 之间的间隔符是 kvSep, kv 和 kv 之间的分隔符是 kkSep
func SplitToMap(s string, kkSep, kvSep string) map[string]string {
	var m map[string]string

	ss := strings.Split(s, kkSep)
	m = make(map[string]string)

	for _, pair := range ss {
		p := strings.TrimSpace(pair)
		if p == "" {
			continue
		}

		k, v := Split2(p, kvSep)
		m[k] = v
	}

	return m
}

func IsDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func HasPrefix(s string, prefix ...string) bool {
	for _, fix := range prefix {
		if strings.HasPrefix(s, fix) {
			return true
		}
	}

	return false
}

func HasSuffix(s string, suffix ...string) bool {
	for _, fix := range suffix {
		if strings.HasSuffix(s, fix) {
			return true
		}
	}

	return false
}

// DefaultEllipse is the default char for marking abbreviation
const DefaultEllipse = "…"

// Abbreviate 将 string/[]byte 缩略到 maxLen （不包含 ellipse）
func Abbreviate(s string, maxLen int, ellipse string) string {
	if maxLen == 0 {
		return s
	}

	if runeLength := utf8.RuneCountInString(s); runeLength > maxLen {
		var result []rune
		for len(result) < maxLen {
			r, size := utf8.DecodeRuneInString(s)
			result = append(result, r)
			s = s[size:]
		}
		return string(result) + ellipse
	}

	return s
}

func AbbreviateBytes(p []byte, maxLen int, ellipse string) []byte {
	if maxLen == 0 {
		return p
	}

	if len(p) > maxLen {
		bb := make([]byte, 0, maxLen+3)
		bb = append(bb, p[:maxLen]...)
		return append(bb, []byte(ellipse)...)
	}

	return p
}

func AbbreviateAny(s any, maxLen int, ellipse string) any {
	switch p := s.(type) {
	case string:
		return Abbreviate(p, maxLen, ellipse)
	case []byte:
		return AbbreviateBytes(p, maxLen, ellipse)
	default:
		return s
	}
}

func Json(v any) []byte {
	vv, _ := json.Marshal(v)
	return vv
}

// JSONPretty prettify the JSON encoding of data
func JSONPretty(data any) string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")

	_ = encoder.Encode(data)
	return buf.String()
}

const (
	singleQuote = '\''
	escape      = '\\'
)

var ErrSyntax = errors.New("invalid syntax")

// QuoteSingle returns a single-quoted Go string literal representing s. But, nothing else escapes.
func QuoteSingle(s string) string {
	out := []rune{singleQuote}
	for _, r := range s {
		switch r {
		case singleQuote:
			out = append(out, escape, r)
		default:
			out = append(out, r)
		}
	}
	out = append(out, singleQuote)
	return string(out)
}

// UnquoteSingle interprets s as a single-quoted Go string literal, returning the string value that s quotes.
func UnquoteSingle(s string) (string, error) {
	if len(s) < 2 {
		return "", ErrSyntax
	}
	if s[0] != singleQuote || s[len(s)-1] != singleQuote {
		return "", ErrSyntax
	}
	var out []rune
	escaped := false
	for _, r := range s[1 : len(s)-1] {
		switch r {
		case escape:
			escaped = !escaped
			if !escaped {
				out = append(out, escape, escape)
			}
		case singleQuote:
			if !escaped {
				return "", ErrSyntax
			}
			out = append(out, r)
			escaped = false
		default:
			out = append(out, r)
			escaped = false
		}
	}
	return string(out), nil
}

func Pick1[T any](a T, _ error) T {
	return a
}

// StructTag represents tag of the struct field
type StructTag struct {
	Raw  string
	Main string
	Opts map[string]string
}

// GetOpt gets opt's value by its name
func (t StructTag) GetOpt(optName string) string {
	if opt, ok := t.Opts[optName]; ok && opt != "" {
		return opt
	}

	return ""
}

// ParseStructTag decode tag values
func ParseStructTag(rawTag string) StructTag {
	opts := make(map[string]string)
	mainPart := ""

	re := regexp.MustCompile(`\s+(\w+)\s*=\s*(\w+)`)
	submatchIndex := re.FindAllStringSubmatchIndex(rawTag, -1)

	if submatchIndex == nil {
		mainPart = rawTag
	} else {
		for i, g := range submatchIndex {
			if i == 0 {
				mainPart = strings.TrimSpace(rawTag[:g[0]])
			}

			k := rawTag[g[2]:g[3]]
			v := rawTag[g[4]:g[5]]
			opts[k] = v
		}
	}

	return StructTag{Raw: rawTag, Main: mainPart, Opts: opts}
}

func JoinMap[K comparable, V any](m map[K]V, kkSep, kvSep string) string {
	ss := make([]string, 0, len(m))
	for k, v := range m {
		ss = append(ss, fmt.Sprintf("%v%s%v", k, kkSep, v))
	}

	return strings.Join(ss, kkSep)
}
