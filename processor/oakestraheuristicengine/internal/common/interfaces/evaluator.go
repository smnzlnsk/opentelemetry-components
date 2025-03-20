package interfaces

type Evaluator interface {
	Evaluate(initalValue float64, params map[string]interface{}) float64
}
