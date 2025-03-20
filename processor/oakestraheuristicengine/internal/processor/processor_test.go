package processor

import (
	"testing"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/wpt"
	"github.com/stretchr/testify/assert"
)

func TestNewHeuristicProcessor(t *testing.T) {
	// Create mock decision trees
	tree1 := wpt.NewBuilder("x > 5", 2.0, 0.5).BuildTree("tree1", 1.0)

	// Initialize processor with mock trees
	processor := NewProcessor("test_processor", tree1)

	// Assert processor was created
	assert.NotNil(t, processor)

	// Assert trees were added to store
	storedTree1 := processor.Evaluator()
	assert.Equal(t, tree1, storedTree1)
}

func TestHeuristicProcessorWorkflow(t *testing.T) {
	// Create a decision tree that checks if x > 5
	// If true, returns 1.0, if false returns 2.0
	testTree := wpt.NewBuilder("x > 5", 1.0, 2.0).BuildTree("test_tree", 1.0)

	processor := NewProcessor("test_processor", testTree)

	// Test cases
	testCases := []struct {
		name          string
		identifier    string
		params        map[string]interface{}
		expectedScore float64
	}{
		{
			name:          "x greater than 5",
			identifier:    "test_tree",
			params:        map[string]interface{}{"x": 10.0},
			expectedScore: 1.0,
		},
		{
			name:          "x less than 5",
			identifier:    "test_tree",
			params:        map[string]interface{}{"x": 3.0},
			expectedScore: 2.0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := processor.Process(tc.params)
			assert.Equal(t, tc.expectedScore, score)
		})
	}
}
