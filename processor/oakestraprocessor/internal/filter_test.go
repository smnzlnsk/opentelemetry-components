package internal

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Run("new filter initialization", func(t *testing.T) {
		f := newFilter()
		if f.MetricFilters == nil {
			t.Error("MetricFilters map was not initialized")
		}
		if len(f.MetricFilters) != 0 {
			t.Error("MetricFilters map should be empty on initialization")
		}
	})

	t.Run("add new metric filter", func(t *testing.T) {
		f := newFilter()
		states := map[string]bool{"running": true, "stopped": true}

		err := f.AddMetricFilter("cpu_usage", states)
		if err != nil {
			t.Errorf("Failed to add metric filter: %v", err)
		}

		// Verify metric was added
		mf, exists := f.MetricFilters["cpu_usage"]
		if !exists {
			t.Fatal("Metric filter was not added")
		}

		// Verify states were added
		for state := range states {
			count, exists := mf.StateFilter[state]
			if !exists {
				t.Errorf("State %s was not added to filter", state)
			}
			if count != 1 {
				t.Errorf("Expected state count to be 1, got %d", count)
			}
		}

		// Verify active contracts count
		if mf.activeContracts != 1 {
			t.Errorf("Expected activeContracts to be 1, got %d", mf.activeContracts)
		}
	})

	t.Run("update existing metric filter", func(t *testing.T) {
		f := newFilter()
		initialStates := map[string]bool{"running": true}
		additionalStates := map[string]bool{"running": true, "stopped": true}

		// Add initial filter
		err := f.AddMetricFilter("cpu_usage", initialStates)
		if err != nil {
			t.Fatalf("Failed to add initial metric filter: %v", err)
		}

		// Update with additional states
		err = f.AddMetricFilter("cpu_usage", additionalStates)
		if err != nil {
			t.Fatalf("Failed to update metric filter: %v", err)
		}

		mf := f.MetricFilters["cpu_usage"]

		// Check running state count (should be 2)
		if count := mf.StateFilter["running"]; count != 2 {
			t.Errorf("Expected running state count to be 2, got %d", count)
		}

		// Check stopped state count (should be 1)
		if count := mf.StateFilter["stopped"]; count != 1 {
			t.Errorf("Expected stopped state count to be 1, got %d", count)
		}
	})
}

func TestMetricFilterStruct(t *testing.T) {
	t.Run("new metric filter struct initialization", func(t *testing.T) {
		mfs := newMetricFilterStruct()

		if mfs.activeContracts != 1 {
			t.Errorf("Expected activeContracts to be 1, got %d", mfs.activeContracts)
		}

		if mfs.StateFilter == nil {
			t.Error("StateFilter map was not initialized")
		}

		if len(mfs.StateFilter) != 0 {
			t.Error("StateFilter map should be empty on initialization")
		}
	})

	t.Run("add states", func(t *testing.T) {
		mfs := newMetricFilterStruct()
		states := map[string]bool{
			"running": true,
			"stopped": true,
		}

		mfs.addStates(states)

		expectedCounts := map[string]int{
			"running": 1,
			"stopped": 1,
		}

		if !reflect.DeepEqual(mfs.StateFilter, expectedCounts) {
			t.Errorf("State counts don't match expected values. Got %v, want %v",
				mfs.StateFilter, expectedCounts)
		}
	})

	t.Run("remove states", func(t *testing.T) {
		mfs := newMetricFilterStruct()

		// Add states first
		states := map[string]bool{
			"running": true,
			"stopped": true,
		}
		mfs.addStates(states)

		// Remove one state
		mfs.removeStates(map[string]bool{"running": true})

		// Verify running state was removed
		if _, exists := mfs.StateFilter["running"]; exists {
			t.Error("running state should have been removed")
		}

		// Verify stopped state remains with count 1
		if count := mfs.StateFilter["stopped"]; count != 1 {
			t.Errorf("Expected stopped state count to be 1, got %d", count)
		}
	})

	t.Run("remove non-existent states", func(t *testing.T) {
		mfs := newMetricFilterStruct()

		// Try to remove state that doesn't exist
		mfs.removeStates(map[string]bool{"nonexistent": true})

		// Should not panic and StateFilter should remain empty
		if len(mfs.StateFilter) != 0 {
			t.Error("StateFilter should remain empty")
		}
	})

	t.Run("multiple state operations", func(t *testing.T) {
		mfs := newMetricFilterStruct()

		// Add states multiple times
		states := map[string]bool{"running": true}
		mfs.addStates(states)
		mfs.addStates(states)

		// Verify count is 2
		if count := mfs.StateFilter["running"]; count != 2 {
			t.Errorf("Expected running state count to be 2, got %d", count)
		}

		// Remove state once
		mfs.removeStates(states)

		// Verify count is 1
		if count := mfs.StateFilter["running"]; count != 1 {
			t.Errorf("Expected running state count to be 1, got %d", count)
		}

		// Remove state again
		mfs.removeStates(states)

		// Verify state was removed
		if _, exists := mfs.StateFilter["running"]; exists {
			t.Error("running state should have been removed")
		}
	})
}

func BenchmarkFilter(b *testing.B) {
	b.Run("add new metric filter", func(b *testing.B) {
		f := newFilter()
		states := map[string]bool{"running": true, "stopped": true}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metricName := fmt.Sprintf("cpu_usage_%d", i)
			_ = f.AddMetricFilter(metricName, states)
		}
	})

	b.Run("update existing metric filter", func(b *testing.B) {
		f := newFilter()
		states := map[string]bool{"running": true}
		_ = f.AddMetricFilter("cpu_usage", states)

		newStates := map[string]bool{"running": true, "stopped": true}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = f.AddMetricFilter("cpu_usage", newStates)
		}
	})
}

func BenchmarkMetricFilterStruct(b *testing.B) {
	b.Run("add states small", func(b *testing.B) {
		states := map[string]bool{
			"running": true,
			"stopped": true,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mfs := newMetricFilterStruct()
			mfs.addStates(states)
		}
	})

	b.Run("add states large", func(b *testing.B) {
		states := make(map[string]bool)
		for i := 0; i < 100; i++ {
			states[fmt.Sprintf("state_%d", i)] = true
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mfs := newMetricFilterStruct()
			mfs.addStates(states)
		}
	})

	b.Run("remove states", func(b *testing.B) {
		states := map[string]bool{
			"running": true,
			"stopped": true,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mfs := newMetricFilterStruct()
			mfs.addStates(states)
			mfs.removeStates(states)
		}
	})

	b.Run("concurrent operations", func(b *testing.B) {
		mfs := newMetricFilterStruct()
		states1 := map[string]bool{"running": true}
		states2 := map[string]bool{"stopped": true}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mfs.addStates(states1)
				mfs.addStates(states2)
				mfs.removeStates(states1)
				mfs.removeStates(states2)
			}
		})
	})
}

func BenchmarkFilterScenarios(b *testing.B) {
	scenarios := []struct {
		name          string
		metricCount   int
		statesPerCall int
	}{
		{"small_scale", 10, 2},
		{"medium_scale", 100, 5},
		{"large_scale", 1000, 10},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			f := newFilter()
			states := make(map[string]bool)
			for i := 0; i < scenario.statesPerCall; i++ {
				states[fmt.Sprintf("state_%d", i)] = true
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < scenario.metricCount; j++ {
					metricName := fmt.Sprintf("metric_%d", j)
					_ = f.AddMetricFilter(metricName, states)
				}
			}
		})
	}
}
