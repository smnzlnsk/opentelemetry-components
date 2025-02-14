package processor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/wpt"
)

type HeuristicProcessor interface {
	GetStore() wpt.Store
	Identifier() string
	Execute(treeIdentifier string, params map[string]interface{}) (float64, error)
}

type heuristicProcessor struct {
	identifier        string
	decisionTreeStore wpt.Store
}

func NewHeuristicProcessor(identifier string, decisionTrees ...wpt.DecisionTree) HeuristicProcessor {
	decisionTreeStore := wpt.NewStore()
	for _, decisionTree := range decisionTrees {
		decisionTreeStore.Add(decisionTree.Identifier(), decisionTree)
	}
	return &heuristicProcessor{
		identifier:        identifier,
		decisionTreeStore: decisionTreeStore,
	}
}

func (h *heuristicProcessor) GetStore() wpt.Store {
	return h.decisionTreeStore
}

func (h *heuristicProcessor) Identifier() string {
	return h.identifier
}

func (h *heuristicProcessor) Execute(treeIdentifier string, params map[string]interface{}) (float64, error) {
	decisionTree := h.decisionTreeStore.Get(treeIdentifier)
	return decisionTree.Traverse(1, params)
}
