package domain

type Processor interface {
	Identifier() string
	Process(instanceNumber int, prev float64, params map[string]interface{}) (EvaluationEntry, error)
	Evaluator() Evaluator
}
