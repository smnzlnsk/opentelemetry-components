package wpt

import (
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// node implements interfaces.Node
type node struct {
	left        domain.Node
	right       domain.Node
	decision    string
	trueWeight  float64
	falseWeight float64
}

// decisionTree implements interfaces.DecisionTree
type decisionTree struct {
	identifier string
	root       domain.Node
}

func newDecisionNode(decision string, trueWeight, falseWeight float64, left, right domain.Node) domain.Node {
	return &node{
		decision:    decision,
		trueWeight:  trueWeight,
		falseWeight: falseWeight,
		left:        left,
		right:       right,
	}
}

func newDecisionTree(identifier string, root domain.Node) domain.Evaluator {
	return &decisionTree{
		identifier: identifier,
		root:       root,
	}
}

func (d *decisionTree) Identifier() string {
	return d.identifier
}

func (d *decisionTree) Evaluate(factor float64, params map[string]interface{}) (float64, error) {
	return d.root.Evaluate(factor, params), nil
}

func (n *node) Evaluate(factor float64, params map[string]interface{}) float64 {
	isTrue, err := evaluateDecision(n.decision, params)
	if err != nil {
		return 0
	}
	newFactor := factor * (map[bool]float64{true: n.trueWeight, false: n.falseWeight})[isTrue]

	// If both children are nil, return the weighted factor
	if n.left == nil && n.right == nil {
		return newFactor
	}

	// Continue traversal based on the decision
	if isTrue {
		if n.left == nil {
			return newFactor
		}
		return n.left.Evaluate(newFactor, params)
	}

	if n.right == nil {
		return newFactor
	}
	return n.right.Evaluate(newFactor, params)
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

func (d *decisionTree) AlarmCondition() notification.Function {
	return nil
}

func (d *decisionTree) RouteCondition() notification.Function {
	return nil
}

func (d *decisionTree) ScheduleCondition() notification.Function {
	return nil
}
