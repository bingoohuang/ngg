package ss

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
