package internal

// Returns boolean of condition: map m contains key k
func mapContains(m map[string]any, k string) bool {
	_, ok := m[k]
	return ok
}
