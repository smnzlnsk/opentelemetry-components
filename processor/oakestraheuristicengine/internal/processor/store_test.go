package processor

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
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
			mockProcessor := &mockProcessor{}

			// First addition
			err := s.Add(mockProcessor)
			if (err != nil) != tt.wantErr {
				t.Errorf("store.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_Get(t *testing.T) {
	s := NewStore()
	mockProcessor := &mockProcessor{
		identifier: "test-processor",
	}

	// Add a tree to retrieve later
	err := s.Add(mockProcessor)
	if err != nil {
		t.Fatalf("Failed to add processor: %v", err)
	}

	tests := []struct {
		name       string
		identifier string
		want       domain.Processor
	}{
		{
			name:       "get existing processor",
			identifier: "test-processor",
			want:       mockProcessor,
		},
		{
			name:       "get non-existent processor",
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

// mockProcessor is a simple mock implementation of Processor interface
type mockProcessor struct {
	identifier string
}

func (m *mockProcessor) Identifier() string {
	return m.identifier
}

func (m *mockProcessor) Evaluator() domain.Evaluator {
	return nil
}

func (m *mockProcessor) Process(instanceNumber int, prev float64, params map[string]interface{}) (evaluation.Entry, error) {
	return evaluation.Entry{
		InstanceNumber: instanceNumber,
		Priority:       prev,
	}, nil
}

func (m *mockProcessor) GetNotificationCondition(capability domain.NotificationInterfaceCapability) *govaluate.EvaluableExpression {
	return nil
}
