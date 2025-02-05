package wpt

type Builder struct {
	decision    string
	trueWeight  float64 // weight to apply when expression is true
	falseWeight float64 // weight to apply when expression is false
	left        *Builder
	right       *Builder
	parent      *Builder
}

func NewBuilder(decision string, trueWeight, falseWeight float64) *Builder {
	return &Builder{
		decision:    decision,
		trueWeight:  trueWeight,
		falseWeight: falseWeight,
	}
}

func (b *Builder) Left(decision string, trueWeight, falseWeight float64) *Builder {
	childBuilder := NewBuilder(decision, trueWeight, falseWeight)
	childBuilder.parent = b
	b.left = childBuilder
	return childBuilder
}

func (b *Builder) Right(decision string, trueWeight, falseWeight float64) *Builder {
	childBuilder := NewBuilder(decision, trueWeight, falseWeight)
	childBuilder.parent = b
	b.right = childBuilder
	return childBuilder
}

func (b *Builder) Build() Node {
	var leftNode, rightNode Node
	if b.left != nil {
		leftNode = b.left.Build()
	}
	if b.right != nil {
		rightNode = b.right.Build()
	}
	return newDecisionNode(b.decision, b.trueWeight, b.falseWeight, leftNode, rightNode)
}

func (b *Builder) BuildTree(identifier string) DecisionTree {
	root := b
	for root.parent != nil {
		root = root.parent
	}
	return newDecisionTree(identifier, root.Build())
}
