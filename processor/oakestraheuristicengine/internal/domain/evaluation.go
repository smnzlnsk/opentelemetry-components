package domain

type Evaluator interface {
	Evaluate(initalValue float64, params map[string]interface{}) float64
}

type Evaluation struct {
	JobName string
	Entries []EvaluationEntry
}

type EvaluationEntry struct {
	InstanceNumber int
	Priority       float64
}
