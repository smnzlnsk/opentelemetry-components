package evaluators

import "math"

func exponentialDecay(n, k float64) float64 {
	return 1 - math.Exp(-k*n)
}
