package wpt

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"

type builder struct {
	decision    string
	trueWeight  float64 // weight to apply when expression is true
	falseWeight float64 // weight to apply when expression is false
	left        *builder
	right       *builder
	parent      *builder
}

func NewBuilder(decision string, trueWeight, falseWeight float64) interfaces.TreeBuilder {
	return &builder{
		decision:    decision,
		trueWeight:  trueWeight,
		falseWeight: falseWeight,
	}
}

func (b *builder) Left(decision string, trueWeight, falseWeight float64) interfaces.TreeBuilder {
	childBuilder := NewBuilder(decision, trueWeight, falseWeight)
	childBuilder.(*builder).parent = b
	b.left = childBuilder.(*builder)
	return childBuilder
}

func (b *builder) Right(decision string, trueWeight, falseWeight float64) interfaces.TreeBuilder {
	childBuilder := NewBuilder(decision, trueWeight, falseWeight)
	childBuilder.(*builder).parent = b
	b.right = childBuilder.(*builder)
	return childBuilder
}

func (b *builder) Build() interfaces.Node {
	var leftNode, rightNode interfaces.Node
	if b.left != nil {
		leftNode = b.left.Build()
	}
	if b.right != nil {
		rightNode = b.right.Build()
	}
	return newDecisionNode(b.decision, b.trueWeight, b.falseWeight, leftNode, rightNode)
}

func (b *builder) BuildTree(identifier string) interfaces.DecisionTree {
	root := b
	for root.parent != nil {
		root = root.parent
	}
	return newDecisionTree(identifier, root.Build())
}
