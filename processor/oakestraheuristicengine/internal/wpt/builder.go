package wpt

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"

type builder struct {
	expression  string
	trueWeight  float64 // weight to apply when expression is true
	falseWeight float64 // weight to apply when expression is false
	left        *builder
	right       *builder
	parent      *builder
}

func NewBuilder(expression string, trueWeight, falseWeight float64) interfaces.TreeBuilder {
	return &builder{
		expression:  expression,
		trueWeight:  trueWeight,
		falseWeight: falseWeight,
	}
}

func (b *builder) Left(expression string, trueWeight, falseWeight float64) interfaces.TreeBuilder {
	childBuilder := NewBuilder(expression, trueWeight, falseWeight).(*builder)
	childBuilder.parent = b
	b.left = childBuilder
	return childBuilder
}

func (b *builder) Right(expression string, trueWeight, falseWeight float64) interfaces.TreeBuilder {
	childBuilder := NewBuilder(expression, trueWeight, falseWeight).(*builder)
	childBuilder.parent = b
	b.right = childBuilder
	return childBuilder
}

func (b *builder) BuildNode() interfaces.Node {
	var leftNode, rightNode interfaces.Node
	if b.left != nil {
		leftNode = b.left.BuildNode()
	}
	if b.right != nil {
		rightNode = b.right.BuildNode()
	}
	return newDecisionNode(b.expression, b.trueWeight, b.falseWeight, leftNode, rightNode)
}

func (b *builder) BuildTree(identifier string, initialValue float64) interfaces.Evaluator {
	root := b
	for root.parent != nil {
		root = root.parent
	}
	return newDecisionTree(identifier, root.BuildNode())
}
