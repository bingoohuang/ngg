# ss

1. `func If[T any](condition bool, a, b T) T`
2. `func IfFunc[T any](condition bool, a, b func() T) T `
3. `func ToSet[K comparable](v []K) map[K]bool`
4. `func Parse[T Parseable](str string) (T, error)`
