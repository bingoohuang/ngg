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
