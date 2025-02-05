package wpt

import (
	"fmt"

	"github.com/Knetic/govaluate"
)

type DecisionTree interface {
	Identifier() string
	Traverse(float64, map[string]interface{}) (float64, error)
}

type Node interface {
	Evaluate(factor float64, params map[string]interface{}) (float64, error)
}

type node struct {
	left        Node
	right       Node
	decision    string
	trueWeight  float64
	falseWeight float64
}

type decisionTree struct {
	identifier string
	root       Node
}

func newDecisionNode(decision string, trueWeight, falseWeight float64, left, right Node) Node {
	return &node{
		decision:    decision,
		trueWeight:  trueWeight,
		falseWeight: falseWeight,
		left:        left,
		right:       right,
	}
}

func newDecisionTree(identifier string, root Node) DecisionTree {
	return &decisionTree{
		identifier: identifier,
		root:       root,
	}
}

func (d *decisionTree) Identifier() string {
	return d.identifier
}

func (n *node) Evaluate(factor float64, params map[string]interface{}) (float64, error) {
	// Evaluate the decision and apply the appropriate weight
	isTrue, err := evaluateDecision(n.decision, params)
	if err != nil {
		return 0, err
	}
	newFactor := factor * (map[bool]float64{true: n.trueWeight, false: n.falseWeight})[isTrue]

	// If both children are nil, return the weighted factor
	if n.left == nil && n.right == nil {
		return newFactor, nil
	}

	// Continue traversal based on the decision
	if isTrue {
		if n.left == nil {
			return newFactor, nil
		}
		return n.left.Evaluate(newFactor, params)
	}

	if n.right == nil {
		return newFactor, nil
	}
	return n.right.Evaluate(newFactor, params)
}

func (d *decisionTree) Traverse(initialFactor float64, params map[string]interface{}) (float64, error) {
	return d.root.Evaluate(initialFactor, params)
}

// Helper function to evaluate boolean expressions using govaluate
func evaluateDecision(decision string, params map[string]interface{}) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(decision)
	if err != nil {
		return false, err
	}

	result, err := expression.Evaluate(params)
	if err != nil {
		return false, err
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("result for '%s' is not boolean: %v", decision, result)
	}

	return boolResult, nil
}
