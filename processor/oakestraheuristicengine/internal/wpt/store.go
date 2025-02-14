package wpt

import "fmt"

type Store interface {
	Add(identifier string, decisionTree DecisionTree) error
	Get(identifier string) DecisionTree
}

type store struct {
	decisionTrees map[string]DecisionTree
}

func NewStore() Store {
	return &store{
		decisionTrees: make(map[string]DecisionTree),
	}
}

func (s *store) Get(identifier string) DecisionTree {
	return s.decisionTrees[identifier]
}

func (s *store) Add(identifier string, decisionTree DecisionTree) error {
	if _, ok := s.decisionTrees[identifier]; ok {
		return fmt.Errorf("decision tree with identifier %s already exists", identifier)
	}
	s.decisionTrees[identifier] = decisionTree
	return nil
}
