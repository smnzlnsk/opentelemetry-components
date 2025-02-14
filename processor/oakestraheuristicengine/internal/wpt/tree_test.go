package wpt

import (
	"testing"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() DecisionTree
		params   map[string]interface{}
		expected float64
	}{
		{
			name: "single node",
			build: func() DecisionTree {
				return NewBuilder("x > 5", 2.0, 0.5).BuildTree("test")
			},
			params: map[string]interface{}{
				"x": 10,
			},
			expected: 2.0, // Expression is true, so true weight (2.0) is applied
		},
		{
			name: "simple tree with comparison",
			build: func() DecisionTree {
				builder := NewBuilder("x > y", 2.0, 0.5)
				builder.Left("z == true", 3.0, 0.3)
				builder.Right("z == false", 4.0, 0.4)
				return builder.BuildTree("test")
			},
			params: map[string]interface{}{
				"x": 10,
				"y": 5,
				"z": true,
			},
			expected: 6.0, // x > y is true (2.0) and z == true is true (3.0), so 2.0 * 3.0
		},
		{
			name: "complex tree with boolean weights",
			build: func() DecisionTree {
				builder := NewBuilder("x > 0 && y < 10", 2.0, 0.5)
				leftNode := builder.Left("z == false", 3.0, 0.3)
				leftNode.Left("a == true", 4.0, 0.4)
				leftNode.Right("a == false", 5.0, 0.5)
				return builder.BuildTree("test")
			},
			params: map[string]interface{}{
				"x": 5,
				"y": 8,
				"z": false,
				"a": true,
			},
			expected: 24.0, // (2.0 * 3.0 * 4.0) for true, true, true path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := tt.build()
			result, err := tree.Traverse(1.0, tt.params)
			if err != nil {
				t.Errorf("%s: expected no error, got %v", tt.name, err)
			}
			if result != tt.expected {
				t.Errorf("%s: expected %f, got %f", tt.name, tt.expected, result)
			}
		})
	}
}
