package internal

// FlattenMap flattens a map into a slice of values,
// effectively converting a map[K]V into a []V.
func FlattenMap[K comparable, V any](m map[K]V) []V {
	flattened := make([]V, 0, len(m))
	for _, v := range m {
		flattened = append(flattened, v)
	}
	return flattened
}
