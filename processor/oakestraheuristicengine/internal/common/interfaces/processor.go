package interfaces

type Processor interface {
	Identifier() string
	Process(params map[string]interface{}) float64
	Evaluator() Evaluator
}
