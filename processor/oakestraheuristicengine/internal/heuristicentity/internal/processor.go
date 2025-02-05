package internal

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/internal/wpt"
)

type HeuristicProcessor interface {
	GetStore() wpt.Store
	Execute(identifier string, params map[string]interface{}) (float64, error)
}

type heuristicProcessor struct {
	decisionTreeStore wpt.Store
}

func NewHeuristicProcessor(decisionTrees ...wpt.DecisionTree) HeuristicProcessor {
	decisionTreeStore := wpt.NewStore()
	for _, decisionTree := range decisionTrees {
		decisionTreeStore.Add(decisionTree.Identifier(), decisionTree)
	}
	return &heuristicProcessor{
		decisionTreeStore: decisionTreeStore,
	}
}

func (h *heuristicProcessor) GetStore() wpt.Store {
	return h.decisionTreeStore
}

func (h *heuristicProcessor) Execute(identifier string, params map[string]interface{}) (float64, error) {
	decisionTree := h.decisionTreeStore.Get(identifier)
	return decisionTree.Traverse(1, params)
}
