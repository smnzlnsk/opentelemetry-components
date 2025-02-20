package interfaces

type HeuristicProcessor interface {
	GetStore() TreeStore
	Identifier() string
	Execute(treeIdentifier string, params map[string]interface{}) (float64, error)
}
