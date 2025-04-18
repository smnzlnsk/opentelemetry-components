package domain

type Processor interface {
	Identifier() string
	Process(appname string, params map[string]interface{}) Evaluation
	Evaluator() Evaluator
}
