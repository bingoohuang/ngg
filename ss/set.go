package ss

func ToSet[K comparable](v []K) map[K]bool {
	m := make(map[K]bool)
	for _, s := range v {
		m[s] = true
	}
	return m
}
