package ss

import "strings"

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

func ContainsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
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
