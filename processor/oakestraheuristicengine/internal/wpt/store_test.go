package wpt

import (
	"testing"
)

func TestStore_Add(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantErr    bool
	}{
		{
			name:       "add new decision tree",
			identifier: "test-tree",
			wantErr:    false,
		},
		{
			name:       "add duplicate decision tree",
			identifier: "test-tree",
			wantErr:    true,
		},
	}

	s := NewStore()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTree := &mockDecisionTree{}

			// First addition
			err := s.Add(tt.identifier, mockTree)
			if (err != nil) != tt.wantErr {
				t.Errorf("store.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_Get(t *testing.T) {
	s := NewStore()
	mockTree := &mockDecisionTree{}
	identifier := "test-tree"

	// Add a tree to retrieve later
	err := s.Add(identifier, mockTree)
	if err != nil {
		t.Fatalf("Failed to add decision tree: %v", err)
	}

	tests := []struct {
		name       string
		identifier string
		want       DecisionTree
	}{
		{
			name:       "get existing decision tree",
			identifier: "test-tree",
			want:       mockTree,
		},
		{
			name:       "get non-existent decision tree",
			identifier: "non-existent",
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Get(tt.identifier)
			if got != tt.want {
				t.Errorf("store.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockDecisionTree is a simple mock implementation of DecisionTree interface
type mockDecisionTree struct{}

// Ensure mockDecisionTree implements DecisionTree interface
var _ DecisionTree = (*mockDecisionTree)(nil)

func (m *mockDecisionTree) Identifier() string {
	return "mock-tree"
}

func (m *mockDecisionTree) Traverse(factor float64, params map[string]interface{}) (float64, error) {
	return 0, nil
}
