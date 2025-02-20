package processor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/wpt"
)

// heuristicProcessor implements interfaces.HeuristicProcessor
type heuristicProcessor struct {
	identifier        string
	decisionTreeStore interfaces.TreeStore
}

func NewHeuristicProcessor(identifier string, decisionTrees ...interfaces.DecisionTree) interfaces.HeuristicProcessor {
	decisionTreeStore := wpt.NewStore()
	for _, decisionTree := range decisionTrees {
		decisionTreeStore.Add(decisionTree.Identifier(), decisionTree)
	}
	return &heuristicProcessor{
		identifier:        identifier,
		decisionTreeStore: decisionTreeStore,
	}
}

func (h *heuristicProcessor) GetStore() interfaces.TreeStore {
	return h.decisionTreeStore
}

func (h *heuristicProcessor) Identifier() string {
	return h.identifier
}

func (h *heuristicProcessor) Execute(treeIdentifier string, params map[string]interface{}) (float64, error) {
	decisionTree := h.decisionTreeStore.Get(treeIdentifier)
	return decisionTree.Traverse(1, params)
}
