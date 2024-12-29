package internal

import (
	"testing"
	"time"
)

func TestRegisterService(t *testing.T) {
	tests := []struct {
		name           string
		service        string
		contracts      map[string]CalculationContract
		expectedError  bool
		expectedFilter map[string]bool // expected metrics in filter
	}{
		{
			name:    "basic registration",
			service: "test-service",
			contracts: map[string]CalculationContract{
				"[metric1] + [metric2]": {
					Formula: "[metric1] + [metric2]",
					Service: "test-service",
					States:  map[string]bool{"running": true},
				},
			},
			expectedError: false,
			expectedFilter: map[string]bool{
				"metric1": true,
				"metric2": true,
			},
		},
		{
			name:    "duplicate service registration",
			service: "test-service",
			contracts: map[string]CalculationContract{
				"[metric1]": {
					Formula: "[metric1]",
					Service: "test-service",
				},
			},
			expectedError: true,
		},
	}

	cs := NewContractState()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cs.RegisterService(tt.service, tt.contracts)

			// Check error expectation
			if (err != nil) != tt.expectedError {
				t.Errorf("RegisterService() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if err != nil {
				return
			}

			// Verify service registration
			if _, exists := cs.Contracts[tt.service]; !exists {
				t.Errorf("Service %s was not registered in Contracts map", tt.service)
			}

			// Verify container datapoints initialization
			if _, exists := cs.ContainerDatapoints[tt.service]; !exists {
				t.Errorf("Service %s was not initialized in ContainerDatapoints map", tt.service)
			}

			// Verify metrics filter registration
			for metric := range tt.expectedFilter {
				if _, exists := cs.Filters.MetricFilters[metric]; !exists {
					t.Errorf("Expected metric %s not found in Filters", metric)
				}
			}
		})
	}
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
