package wpt

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
)

// store implements interfaces.TreeStore
type store struct {
	decisionTrees map[string]interfaces.DecisionTree
}

func NewStore() interfaces.TreeStore {
	return &store{
		decisionTrees: make(map[string]interfaces.DecisionTree),
	}
}

func (s *store) Get(identifier string) interfaces.DecisionTree {
	return s.decisionTrees[identifier]
}

func (s *store) Add(identifier string, decisionTree interfaces.DecisionTree) error {
	if _, ok := s.decisionTrees[identifier]; ok {
		return fmt.Errorf("decision tree with identifier %s already exists", identifier)
	}
	s.decisionTrees[identifier] = decisionTree
	return nil
}
