package interfaces

type TreeStore interface {
	Add(identifier string, decisionTree DecisionTree) error
	Get(identifier string) DecisionTree
}
