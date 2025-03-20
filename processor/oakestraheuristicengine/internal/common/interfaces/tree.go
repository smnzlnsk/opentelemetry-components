package interfaces

type TreeBuilder interface {
	Left(decision string, trueWeight, falseWeight float64) TreeBuilder
	Right(decision string, trueWeight, falseWeight float64) TreeBuilder
	BuildNode() Node
	BuildTree(identifier string, initialValue float64) Evaluator
}

type Node interface {
	Evaluate(factor float64, params map[string]interface{}) float64
}
