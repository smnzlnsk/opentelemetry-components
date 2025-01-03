package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContractState(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		cs := NewContractState()

		if cs.Contracts == nil {
			t.Error("Contracts map was not initialized")
		}
		if cs.ContainerDatapoints == nil {
			t.Error("ContainerDatapoints map was not initialized")
		}
		if cs.Filters == nil {
			t.Error("Filters was not initialized")
		}
	})

	t.Run("register service", func(t *testing.T) {
		t.Run("basic registration", func(t *testing.T) {
			cs := NewContractState()
			service := "test-service"
			contracts := map[string]CalculationContract{
				"[metric1] + [metric2]": {
					Formula: "[metric1] + [metric2]",
					Service: service,
					States:  map[string]bool{"running": true},
				},
			}

			err := cs.RegisterService(service, contracts)
			if err != nil {
				t.Fatalf("Failed to register service: %v", err)
			}

			// Verify service registration
			serviceContracts, exists := cs.Contracts[service]
			if !exists {
				t.Fatal("Service was not registered in Contracts map")
			}

			// Verify contract details
			contract, exists := serviceContracts["[metric1] + [metric2]"]
			if !exists {
				t.Fatal("Contract was not registered")
			}
			if contract.Service != service {
				t.Errorf("Expected service %s, got %s", service, contract.Service)
			}

			// Verify metric filters
			expectedMetrics := map[string]bool{
				"metric1": true,
				"metric2": true,
			}
			for metric := range expectedMetrics {
				if _, exists := cs.Filters.MetricFilters[metric]; !exists {
					t.Errorf("Expected metric %s not found in Filters", metric)
				}
			}
		})

		t.Run("duplicate service registration", func(t *testing.T) {
			cs := NewContractState()
			service := "test-service"
			contracts := map[string]CalculationContract{
				"[metric1]": {
					Formula: "[metric1]",
					Service: service,
				},
			}

			// First registration should succeed
			err := cs.RegisterService(service, contracts)
			if err != nil {
				t.Fatalf("First registration failed: %v", err)
			}

			// Second registration should fail
			err = cs.RegisterService(service, contracts)
			if err == nil {
				t.Error("Expected error on duplicate registration, got nil")
			}
		})
	})

	t.Run("delete service", func(t *testing.T) {
		t.Run("delete single metric service", func(t *testing.T) {
			cs := NewContractState()
			service := "service1"
			contracts := map[string]CalculationContract{
				"[metric1] > 0.5": {
					Formula: "[metric1] > 0.5",
					Service: service,
					States:  map[string]bool{"state1": true},
				},
			}

			// Setup initial state
			err := cs.RegisterService(service, contracts)
			if err != nil {
				t.Fatalf("Failed to register service: %v", err)
			}

			// Verify initial setup
			if _, exists := cs.Filters.MetricFilters["metric1"]; !exists {
				t.Fatal("metric1 filter should exist before deletion")
			}

			// Delete service
			err = cs.DeleteService(service)
			if err != nil {
				t.Fatalf("Failed to delete service: %v", err)
			}

			// Verify cleanup
			if _, exists := cs.Contracts[service]; exists {
				t.Error("Service should be deleted from Contracts")
			}
			if _, exists := cs.ContainerDatapoints[service]; exists {
				t.Error("Service should be deleted from ContainerDatapoints")
			}
			if _, exists := cs.Filters.MetricFilters["metric1"]; exists {
				t.Error("Metric filter should be deleted")
			}
		})

		t.Run("delete service with multiple metrics", func(t *testing.T) {
			cs := NewContractState()
			service := "service1"
			contracts := map[string]CalculationContract{
				"[metric1] + [metric2] > 1.0": {
					Formula: "[metric1] + [metric2] > 1.0",
					Service: service,
					States:  map[string]bool{"state1": true, "state2": true},
				},
			}

			// Setup and verify initial state
			err := cs.RegisterService(service, contracts)
			if err != nil {
				t.Fatalf("Failed to register service: %v", err)
			}

			// Verify initial metric filters
			for _, metric := range []string{"metric1", "metric2"} {
				mf, exists := cs.Filters.MetricFilters[metric]
				if !exists {
					t.Fatalf("Metric %s filter should exist before deletion", metric)
				}
				if mf.activeContracts != 1 {
					t.Errorf("Expected 1 active contract for %s, got %d", metric, mf.activeContracts)
				}
			}

			// Delete service
			err = cs.DeleteService(service)
			if err != nil {
				t.Fatalf("Failed to delete service: %v", err)
			}

			// Verify all cleanup
			if _, exists := cs.Contracts[service]; exists {
				t.Error("Service should be deleted from Contracts")
			}
			for _, metric := range []string{"metric1", "metric2"} {
				if _, exists := cs.Filters.MetricFilters[metric]; exists {
					t.Errorf("Metric %s filter should be deleted", metric)
				}
			}
		})
	})

	t.Run("default contracts", func(t *testing.T) {
		cs := NewContractState()
		defaultFormula := "[metric3] + [metric4]"
		service := "test-service"

		// Set up default contract
		err := cs.GenerateDefaultContract(defaultFormula, map[string]bool{"running": true})
		if err != nil {
			t.Fatalf("Failed to generate default contract: %v", err)
		}

		// Register service with its own contract
		contracts := map[string]CalculationContract{
			"[metric1] + [metric2]": {
				Formula: "[metric1] + [metric2]",
				Service: service,
				States:  map[string]bool{"running": true},
			},
		}

		err = cs.RegisterService(service, contracts)
		if err != nil {
			t.Fatalf("Failed to register service: %v", err)
		}

		// Verify both contracts exist
		serviceContracts, exists := cs.Contracts[service]
		if !exists {
			t.Fatal("Service contracts not found")
		}

		if len(serviceContracts) != 2 {
			t.Errorf("Expected 2 contracts (1 service + 1 default), got %d", len(serviceContracts))
		}

		// Verify specific contracts
		if _, exists := serviceContracts["[metric1] + [metric2]"]; !exists {
			t.Error("Service-specific contract not found")
		}
		if _, exists := serviceContracts["[metric3] + [metric4]"]; !exists {
			t.Error("Default contract not found")
		}

		// Verify all metric filters
		expectedMetrics := []string{"metric1", "metric2", "metric3", "metric4"}
		for _, metric := range expectedMetrics {
			if _, exists := cs.Filters.MetricFilters[metric]; !exists {
				t.Errorf("Expected metric %s not found in Filters", metric)
			}
		}
	})
}

func BenchmarkRegisterService(b *testing.B) {
	benchmarks := []struct {
		name          string
		service       string
		contractCount int
	}{
		{"small_service", "test-service-small", 1},
		{"medium_service", "test-service-medium", 10},
		{"large_service", "test-service-large", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			contracts := generateTestContracts(bm.contractCount)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cs := NewContractState()
				start := time.Now()
				_ = cs.RegisterService(bm.service, contracts)
				b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
			}
		})
	}
}

// Helper function to generate test contracts
func generateTestContracts(count int) map[string]CalculationContract {
	contracts := make(map[string]CalculationContract)
	for i := 0; i < count; i++ {
		formula := "[metric1] + [metric2]"
		contracts[formula] = CalculationContract{
			Formula: formula,
			States:  map[string]bool{"running": true},
			Metrics: map[string]bool{"metric1": true, "metric2": true},
		}
	}
	return contracts
}

func TestRegisterServiceWithDefaults(t *testing.T) {
	cs := NewContractState()

	// Set up a default contract
	defaultFormula := "[metric3] + [metric4]"
	err := cs.GenerateDefaultContract(defaultFormula, map[string]bool{"running": true})
	if err != nil {
		t.Fatalf("Failed to generate default contract: %v", err)
	}

	// Create service-specific contracts
	serviceContracts := map[string]CalculationContract{
		"[metric1] + [metric2]": {
			Formula: "[metric1] + [metric2]",
			Service: "test-service",
			States:  map[string]bool{"running": true},
		},
	}

	// Register service
	err = cs.RegisterService("test-service", serviceContracts)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Verify both contracts exist
	if len(cs.Contracts["test-service"]) != 2 {
		t.Errorf("Expected 2 contracts (1 service + 1 default), got %d",
			len(cs.Contracts["test-service"]))
	}

	// Verify both formulas exist
	if _, exists := cs.Contracts["test-service"]["[metric1] + [metric2]"]; !exists {
		t.Error("Service-specific contract not found")
	}
	if _, exists := cs.Contracts["test-service"]["[metric3] + [metric4]"]; !exists {
		t.Error("Default contract not found")
	}
}

func TestDeleteService(t *testing.T) {
	tests := []struct {
		name            string
		initialService  string
		initialFormula  string
		initialStates   map[string]bool
		expectedMetrics map[string]bool
	}{
		{
			name:           "delete service with single metric",
			initialService: "service1",
			initialFormula: "[metric1] > 0.5",
			initialStates: map[string]bool{
				"state1": true,
			},
			expectedMetrics: map[string]bool{
				"metric1": false, // should not exist after deletion
			},
		},
		{
			name:           "delete service with multiple metrics",
			initialService: "service1",
			initialFormula: "[metric1] + [metric2] > 1.0",
			initialStates: map[string]bool{
				"state1": true,
				"state2": true,
			},
			expectedMetrics: map[string]bool{
				"metric1": false,
				"metric2": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewContractState()

			// Setup initial state
			contracts := map[string]CalculationContract{
				tt.initialFormula: newCalculcationContract(
					tt.initialFormula,
					tt.initialService,
					tt.initialStates,
					filterMetricsFromFormula(tt.initialFormula),
				),
			}

			// Verify initial setup
			err := cs.RegisterService(tt.initialService, contracts)
			require.NoError(t, err)

			// Verify metrics were properly registered
			for metric := range tt.expectedMetrics {
				_, exists := cs.Filters.MetricFilters[metric]
				require.True(t, exists, "metric filter should exist before deletion: %s", metric)

				mf := cs.Filters.MetricFilters[metric]
				require.Equal(t, 1, mf.activeContracts, "should have one active contract")

				// Verify states were properly registered
				for state := range tt.initialStates {
					count, exists := mf.StateFilter[state]
					require.True(t, exists, "state should exist: %s", state)
					require.Equal(t, 1, count, "state should have count of 1")
				}
			}

			// Delete service
			err = cs.DeleteService(tt.initialService)
			require.NoError(t, err)

			// Verify service deletion
			_, exists := cs.Contracts[tt.initialService]
			require.False(t, exists, "service should be deleted")

			// Verify metric filter cleanup
			for metric, shouldExist := range tt.expectedMetrics {
				_, exists := cs.Filters.MetricFilters[metric]
				require.Equal(t, shouldExist, exists, "metric filter existence mismatch for %s", metric)
			}

			// Verify container datapoints were cleaned up
			_, exists = cs.ContainerDatapoints[tt.initialService]
			require.False(t, exists, "container datapoints should be deleted")
		})
	}
}
