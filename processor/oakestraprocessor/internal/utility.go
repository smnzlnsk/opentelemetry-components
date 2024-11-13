package internal

// Returns boolean of condition: map m contains key k
func Map_contains(m map[string]any, k string) bool {
	_, ok := m[k]
	return ok
}
