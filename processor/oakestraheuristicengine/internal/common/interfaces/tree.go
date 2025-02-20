package interfaces

type TreeBuilder interface {
	Left(decision string, trueWeight, falseWeight float64) TreeBuilder
	Right(decision string, trueWeight, falseWeight float64) TreeBuilder
	Build() Node
	BuildTree(identifier string) DecisionTree
}

type DecisionTree interface {
	Identifier() string
	Traverse(float64, map[string]interface{}) (float64, error)
}

type Node interface {
	Evaluate(factor float64, params map[string]interface{}) (float64, error)
}
