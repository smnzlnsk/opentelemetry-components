package domain

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestContractState(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)

		if cs.Contracts == nil {
			t.Error("Contracts map was not initialized")
		}
		if cs.Datapoints == nil {
			t.Error("Datapoints map was not initialized")
		}
		if cs.Filters == nil {
			t.Error("Filters was not initialized")
		}
	})

	t.Run("register service", func(t *testing.T) {
		t.Run("basic registration", func(t *testing.T) {
			cs := NewContractState("test", zap.NewNop(), nil)
			service := "test-service"
			contracts := []CalculationContract{
				{
					Formula:   "[metric1] + [metric2]",
					Service:   service,
					State:     "running",
					arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
				},
			}

			err := cs.RegisterService(service, contracts, "100")
			require.NoError(t, err)

			// Verify contract registration
			key := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
			contract, err := cs.Contracts.GetContract(key)
			require.NoError(t, err)

			assert.Equal(t, "test", contract.Processor, "Contract Processor should be set to the ContractState's processorName")

			require.Equal(t, service, contract.Service, "Service mismatch")

			// Verify metric filters
			expectedMetrics := map[string]bool{
				"metric1": true,
				"metric2": true,
			}
			for metric := range expectedMetrics {
				_, exists := cs.Filters.MetricFiltersMap()[metric]
				require.True(t, exists, "Expected metric %s not found in Filters", metric)
			}
		})

		t.Run("duplicate service registration", func(t *testing.T) {
			cs := NewContractState("test", zap.NewNop(), nil)
			service := "test-service"
			contracts := []CalculationContract{
				{
					Formula:   "[metric1]",
					Service:   service,
					arguments: getCalculationArguments(sanitizeFormula("[metric1]", "running")),
				},
			}

			// First registration should succeed
			err := cs.RegisterService(service, contracts, "100")
			require.NoError(t, err)

			// Second registration should fail
			err = cs.RegisterService(service, contracts, "100")
			require.Error(t, err, "Expected error on duplicate registration")
		})
	})

	t.Run("default contracts", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)
		defaultFormula := "[metric3] + [metric4]"
		service := "test-service"

		// Set up default contract
		err := cs.GenerateDefaultContract(defaultFormula, []string{"running"})
		require.NoError(t, err)

		// Verify default contract registration
		defaultKey := ContractKey{Service: "default", Formula: defaultFormula}
		_, err = cs.Contracts.GetContract(defaultKey)
		require.NoError(t, err, "Default contract not registered")

		// Register service with its own contract
		contracts := []CalculationContract{
			{
				Formula:   "[metric1] + [metric2]",
				Service:   service,
				State:     "running",
				arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
			},
		}

		err = cs.RegisterService(service, contracts, "100")
		require.NoError(t, err)

		// Count contracts for the service
		serviceContractCount := 0
		for key := range cs.Contracts.GetAllContracts() {
			if key.Service == service {
				serviceContractCount++
			}
		}
		require.Equal(t, 2, serviceContractCount, "Expected 2 contracts (1 service + 1 default)")

		// Verify specific contracts exist
		key1 := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
		key2 := ContractKey{Service: service, Formula: defaultFormula}
		_, err = cs.Contracts.GetContract(key1)
		require.NoError(t, err, "Service-specific contract not found")
		_, err = cs.Contracts.GetContract(key2)
		require.NoError(t, err, "Default contract not found for service")

		// Verify all metric filters
		expectedMetrics := []string{"metric1", "metric2", "metric3", "metric4"}
		for _, metric := range expectedMetrics {
			_, exists := cs.Filters.MetricFiltersMap()[metric]
			require.True(t, exists, "Expected metric %s not found in Filters", metric)
		}
	})

	t.Run("delete service", func(t *testing.T) {
		t.Run("cannot delete default service", func(t *testing.T) {
			cs := NewContractState("test", zap.NewNop(), nil)
			err := cs.DeleteService("default")
			require.Error(t, err, "Should not be able to delete default service")
		})

		t.Run("delete service with contracts", func(t *testing.T) {
			cs := NewContractState("test", zap.NewNop(), nil)
			service := "service1"
			formula := "[metric1] + [metric2] > 1.0"
			contracts := []CalculationContract{
				{
					Formula:   formula,
					Service:   service,
					State:     "state1",
					arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
				},
			}

			// Setup and verify initial state
			err := cs.RegisterService(service, contracts, "100")
			require.NoError(t, err)

			// Verify initial metric filters
			for _, contract := range contracts {
				for _, argument := range contract.arguments {
					_, exists := cs.Filters.MetricFiltersMap()[argument.Metric]
					require.True(t, exists, "Metric %s filter should exist before deletion", argument.Metric)
				}
			}

			// Delete service
			err = cs.DeleteService(service)
			require.NoError(t, err)

			// Verify all cleanup
			key := ContractKey{Service: service, Formula: formula}
			_, err = cs.Contracts.GetContract(key)
			require.Error(t, err, "Contract should be deleted")

			for _, contract := range contracts {
				for _, argument := range contract.arguments {
					_, exists := cs.Filters.MetricFiltersMap()[argument.Metric]
					require.False(t, exists, "Metric %s filter should be deleted", argument.Metric)
				}
			}

			// Verify datapoints cleanup
			for _, contract := range contracts {
				for _, argument := range contract.arguments {
					dpKey := DatapointKey{
						Service: service,
						Metric:  argument.Metric,
						State:   argument.State,
					}
					_, exists := cs.Datapoints[dpKey]
					require.False(t, exists, "Datapoint should be deleted")
				}
			}
		})
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
			contracts := generateTestContracts(bm.service, bm.contractCount)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cs := NewContractState("test", zap.NewNop(), nil)
				start := time.Now()
				_ = cs.RegisterService(bm.service, contracts, "100")
				b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
			}
		})
	}
}

// Helper function to generate test contracts
func generateTestContracts(service string, count int) []CalculationContract {
	contracts := make([]CalculationContract, count)
	for i := 0; i < count; i++ {
		formula := fmt.Sprintf("[metric%d] + [metric%d]", i*2+1, i*2+2)
		arguments := []CalculationArgument{
			{Metric: fmt.Sprintf("metric%d", i*2+1), State: "", Age: 0},
			{Metric: fmt.Sprintf("metric%d", i*2+2), State: "", Age: 0},
		}
		contracts[i] = CalculationContract{
			Formula:   formula,
			Service:   service,
			State:     "running",
			arguments: arguments,
		}
	}
	return contracts
}

func TestRegisterServiceWithDefaults(t *testing.T) {
	cs := NewContractState("test", zap.NewNop(), nil)
	service := "test-service"

	// Set up a default contract
	defaultFormula := "[metric3] + [metric4]"
	err := cs.GenerateDefaultContract(defaultFormula, []string{"running"})
	require.NoError(t, err)

	// Create service-specific contracts
	serviceContracts := []CalculationContract{
		{
			Formula: "[metric1] + [metric2]",
			Service: service,
			State:   "running",
		},
	}

	// Register service
	err = cs.RegisterService(service, serviceContracts, "100")
	require.NoError(t, err)

	// Count contracts for the service
	serviceContractCount := 0
	for key := range cs.Contracts.GetAllContracts() {
		if key.Service == service {
			serviceContractCount++
		}
	}
	require.Equal(t, 2, serviceContractCount, "Expected 2 contracts (1 service + 1 default)")

	// Verify both formulas exist
	key1 := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
	key2 := ContractKey{Service: service, Formula: defaultFormula}
	_, err = cs.Contracts.GetContract(key1)
	require.NoError(t, err, "Service-specific contract not found")
	_, err = cs.Contracts.GetContract(key2)
	require.NoError(t, err, "Default contract not found")
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
			cs := NewContractState("test", zap.NewNop(), nil)

			// Setup initial state
			contracts := []CalculationContract{
				{
					Formula:   tt.initialFormula,
					Service:   tt.initialService,
					State:     "state1", // Using the first state
					arguments: getCalculationArguments(sanitizeFormula(tt.initialFormula, "running")),
				},
			}

			// Register service
			err := cs.RegisterService(tt.initialService, contracts, "100")
			require.NoError(t, err)

			// Verify initial setup
			key := ContractKey{Service: tt.initialService, Formula: tt.initialFormula}
			_, err = cs.Contracts.GetContract(key)
			require.NoError(t, err, "Contract should exist before deletion")

			// Verify metrics were properly registered
			for metric := range tt.expectedMetrics {
				mf, exists := cs.Filters.MetricFiltersMap()[metric]
				require.True(t, exists, "metric filter should exist before deletion: %s", metric)
				require.Equal(t, 1, mf.ActiveContracts(), "should have one active contract")

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
			_, err = cs.Contracts.GetContract(key)
			require.Error(t, err, "Contract should be deleted")

			// Verify metric filter cleanup
			for metric, shouldExist := range tt.expectedMetrics {
				_, exists := cs.Filters.MetricFiltersMap()[metric]
				require.Equal(t, shouldExist, exists, "metric filter existence mismatch for %s", metric)
			}

			// Verify datapoints were cleaned up
			for metric := range tt.expectedMetrics {
				for state := range tt.initialStates {
					dpKey := DatapointKey{
						Service: tt.initialService,
						Metric:  metric,
						State:   state,
					}
					_, exists := cs.Datapoints[dpKey]
					require.False(t, exists, "datapoint should be deleted")
				}
			}
		})
	}
}

func TestDefaultContractHandling(t *testing.T) {
	cs := NewContractState("test", zap.NewNop(), nil)
	defaultFormula := "[metric3] + [metric4]"
	service := "test-service"

	// Set up and verify default contract
	err := cs.GenerateDefaultContract(defaultFormula, []string{"running"})
	require.NoError(t, err)

	defaultKey := ContractKey{Service: "default", Formula: defaultFormula}
	defaultContract, err := cs.Contracts.GetContract(defaultKey)
	require.NoError(t, err, "Default contract not registered")
	require.Equal(t, "default", defaultContract.Service)
	require.Equal(t, sanitizeFormula(defaultFormula, "running"), defaultContract.Formula)
	require.Equal(t, "running", defaultContract.State)
	require.Contains(t, defaultContract.arguments, CalculationArgument{Metric: "metric3", State: "", Age: 0})
	require.Contains(t, defaultContract.arguments, CalculationArgument{Metric: "metric4", State: "", Age: 0})

	// Register service and verify default contract is copied
	contracts := []CalculationContract{
		{
			Formula:   "[metric1] + [metric2]",
			Service:   service,
			State:     "running",
			arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
		},
	}

	err = cs.RegisterService(service, contracts, "100")
	require.NoError(t, err)

	// Verify service has both its own contract and the default contract
	serviceDefaultKey := ContractKey{Service: service, Formula: defaultFormula}
	serviceContract, err := cs.Contracts.GetContract(serviceDefaultKey)
	require.NoError(t, err, "Default contract not copied to service")
	require.Equal(t, service, serviceContract.Service)
	require.Equal(t, sanitizeFormula(defaultFormula, "running"), serviceContract.Formula)

	// Verify filters include metrics from both contracts
	expectedMetrics := []string{"metric1", "metric2", "metric3", "metric4"}
	for _, metric := range expectedMetrics {
		filter, exists := cs.Filters.MetricFiltersMap()[metric]
		require.True(t, exists, "Metric %s not in filters", metric)
		require.Greater(t, filter.StateFilter["running"], 0, "State 'running' not set for metric %s", metric)
	}

	// Verify deletion protection
	err = cs.DeleteService("default")
	require.Error(t, err, "Should not be able to delete default service")

	// Verify service deletion cleans up properly
	err = cs.DeleteService(service)
	require.NoError(t, err)

	// Check that default contract still exists but service contracts are gone
	_, err = cs.Contracts.GetContract(defaultKey)
	require.NoError(t, err, "Default contract should still exist")
	_, err = cs.Contracts.GetContract(serviceDefaultKey)
	require.Error(t, err, "Service's copy of default contract should be deleted")
}

func TestServiceRegistration(t *testing.T) {
	t.Run("register with multiple states", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)
		service := "test-service"
		contracts := []CalculationContract{
			{
				Formula:   "[metric1] + [metric2]",
				Service:   service,
				State:     "running",
				arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
			},
		}

		err := cs.RegisterService(service, contracts, "100")
		require.NoError(t, err)

		// Verify contract registration
		key := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
		contract, err := cs.Contracts.GetContract(key)
		require.NoError(t, err)
		require.Equal(t, service, contract.Service)
		require.Equal(t, "running", contract.State)

		// Verify filters
		for _, argument := range contract.arguments {
			filter, exists := cs.Filters.MetricFiltersMap()[argument.Metric]
			require.True(t, exists)
			require.Greater(t, filter.StateFilter["running"], 0)
		}
	})

	t.Run("register with overlapping metrics", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)
		service := "test-service"
		contracts := []CalculationContract{
			{
				Formula:   "[metric1] + [metric2]",
				Service:   service,
				State:     "running",
				arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
			},
			{
				Formula:   "[metric2] + [metric3]",
				Service:   service,
				State:     "running",
				arguments: getCalculationArguments(sanitizeFormula("[metric2] + [metric3]", "running")),
			},
		}

		err := cs.RegisterService(service, contracts, "100")
		require.NoError(t, err)

		// Verify metric filters
		filter, exists := cs.Filters.MetricFiltersMap()["metric2"]
		require.True(t, exists)
		require.Equal(t, 2, filter.ActiveContracts(), "Metric2 should be used by two contracts")
	})
}

func TestContractState_Comprehensive(t *testing.T) {
	t.Run("default contract handling", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)

		t.Run("empty default contract", func(t *testing.T) {
			err := cs.GenerateDefaultContract("", []string{})
			require.Error(t, err, "Should not allow empty formula")
		})

		t.Run("invalid formula", func(t *testing.T) {
			err := cs.GenerateDefaultContract("[metric1] ++ [metric2]", []string{"running"})
			require.Error(t, err, "Should not allow invalid formula")
		})

		t.Run("valid default contract", func(t *testing.T) {
			formula := "[metric1] + [metric2]"
			err := cs.GenerateDefaultContract(formula, []string{"running"})
			require.NoError(t, err)

			key := ContractKey{Service: "default", Formula: sanitizeFormula(formula, "running")}
			contract, err := cs.Contracts.GetContract(key)
			require.NoError(t, err)
			require.Equal(t, "default", contract.Service)
			require.Equal(t, sanitizeFormula(formula, "running"), contract.Formula)
			require.Contains(t, contract.arguments, CalculationArgument{Metric: "metric1", State: "running", Age: 0})
			require.Contains(t, contract.arguments, CalculationArgument{Metric: "metric2", State: "running", Age: 0})
		})

		t.Run("duplicate default contract", func(t *testing.T) {
			formula := "[metric3] + [metric4]"
			err := cs.GenerateDefaultContract(formula, []string{"running"})
			require.NoError(t, err, "Should allow multiple default contracts")

			count := 0
			for key := range cs.Contracts.GetAllContracts() {
				if key.Service == "default" {
					count++
				}
			}
			require.Equal(t, 2, count, "Should have two default contracts")
		})
	})

	t.Run("service registration edge cases", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)

		t.Run("register empty service name", func(t *testing.T) {
			contracts := []CalculationContract{
				{Formula: "[metric1]"},
			}
			err := cs.RegisterService("", contracts, "100")
			require.Error(t, err, "Should not allow empty service name")
		})

		t.Run("register with nil contracts", func(t *testing.T) {
			err := cs.RegisterService("test", nil, "100")
			require.Error(t, err, "Should not allow nil contracts")
		})

		t.Run("register with empty contracts", func(t *testing.T) {
			err := cs.RegisterService("test", []CalculationContract{}, "100")
			require.NoError(t, err, "Should allow empty contracts map")
		})

		t.Run("register with invalid formula", func(t *testing.T) {
			contracts := []CalculationContract{
				{
					Formula: "[metric1] ++ [metric2]",
					Service: "test",
				},
			}
			err := cs.RegisterService("test", contracts, "100")
			require.Error(t, err, "Should not allow invalid formula")
		})
	})

	t.Run("metric filter handling", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)

		t.Run("overlapping states", func(t *testing.T) {
			// Register first contract
			err := cs.RegisterService("service1", []CalculationContract{
				{
					Formula: "[metric1]",
					Service: "service1",
					State:   "running",
				},
			}, "100")
			require.NoError(t, err)

			// Register second contract with overlapping states
			err = cs.RegisterService("service2", []CalculationContract{
				{
					Formula: "[metric1]",
					Service: "service2",
					State:   "running",
				},
			}, "100")
			require.NoError(t, err)

			filter, exists := cs.Filters.MetricFiltersMap()["metric1"]
			require.True(t, exists)
			require.Equal(t, 2, filter.ActiveContracts())
			require.Greater(t, filter.StateFilter["running"], 0)
		})

		t.Run("metric cleanup", func(t *testing.T) {
			cs := NewContractState("test", zap.NewNop(), nil)

			// Register service with multiple metrics
			err := cs.RegisterService("service1", []CalculationContract{
				{
					Formula:   "[metric1] + [metric2]",
					Service:   "service1",
					State:     "running",
					arguments: getCalculationArguments(sanitizeFormula("[metric1] + [metric2]", "running")),
				},
			}, "100")
			require.NoError(t, err)

			// Delete service
			err = cs.DeleteService("service1")
			require.NoError(t, err)

			// Verify all metrics are cleaned up
			_, exists := cs.Filters.MetricFiltersMap()["metric1"]
			require.False(t, exists, "metric1 should be removed")
			_, exists = cs.Filters.MetricFiltersMap()["metric2"]
			require.False(t, exists, "metric2 should be removed")
		})
	})

	t.Run("datapoint handling", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)

		t.Run("datapoint cleanup", func(t *testing.T) {
			// Register service
			service := "test-service"
			err := cs.RegisterService(service, []CalculationContract{
				{
					Formula:   "[metric1]",
					Service:   service,
					State:     "running",
					arguments: getCalculationArguments(sanitizeFormula("[metric1]", "running")),
				},
			}, "100")
			require.NoError(t, err)

			// Add datapoints
			dpKey := DatapointKey{
				Service: service,
				Metric:  "metric1",
				State:   "running",
			}
			cs.Datapoints[dpKey] = map[int]MetricDatapoint{}

			// Delete service
			err = cs.DeleteService(service)
			require.NoError(t, err)

			// Verify datapoint cleanup
			_, exists := cs.Datapoints[dpKey]
			require.False(t, exists, "Datapoint should be removed")
		})
	})

	t.Run("concurrent operations", func(t *testing.T) {
		cs := NewContractState("test", zap.NewNop(), nil)

		// Add default contract
		err := cs.GenerateDefaultContract("[metric1]", []string{"running"})
		require.NoError(t, err)

		var wg sync.WaitGroup
		services := []string{"service1", "service2", "service3", "service4", "service5"}

		// Concurrently register services
		for _, service := range services {
			wg.Add(1)
			go func(svc string) {
				defer wg.Done()
				contracts := []CalculationContract{
					{
						Formula:   "[metric2]",
						Service:   svc,
						State:     "running",
						arguments: getCalculationArguments(sanitizeFormula("[metric2]", "running")),
					},
				}
				err := cs.RegisterService(svc, contracts, "100")
				require.NoError(t, err)
			}(service)
		}

		wg.Wait()

		// Verify all services were registered
		serviceCount := 0
		for key := range cs.Contracts.GetAllContracts() {
			if key.Service != "default" {
				serviceCount++
			}
		}
		require.Equal(t, len(services)*2, serviceCount, "Each service should have two contracts (default + specific)")
	})
}

func TestFormulaParsingWithStates(t *testing.T) {
	t.Run("filter metrics from formula", func(t *testing.T) {
		tests := []struct {
			name            string
			formula         string
			expectedMetrics map[string]bool
		}{
			{
				name:    "simple formula",
				formula: "[metric1] + [metric2]",
				expectedMetrics: map[string]bool{
					"metric1": true,
					"metric2": true,
				},
			},
			{
				name:    "formula with states",
				formula: "[metric1{in}] + [metric2{out}]",
				expectedMetrics: map[string]bool{
					"metric1": true,
					"metric2": true,
				},
			},
			{
				name:    "formula with age",
				formula: "[metric1(2)] + [metric2]",
				expectedMetrics: map[string]bool{
					"metric1": true,
					"metric2": true,
				},
			},
			{
				name:    "formula with states and age",
				formula: "[metric1(2){in}] + [metric2{out}]",
				expectedMetrics: map[string]bool{
					"metric1": true,
					"metric2": true,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := filterMetricsFromFormula(tt.formula)
				assert.Equal(t, tt.expectedMetrics, result)
			})
		}
	})

	t.Run("extract states from formula", func(t *testing.T) {
		tests := []struct {
			name              string
			formula           string
			expectedArguments []CalculationArgument
		}{
			{
				name:    "simple formula with default state",
				formula: "[metric1] + [metric2]",
				expectedArguments: []CalculationArgument{
					{Metric: "metric1", State: "", Age: 0},
					{Metric: "metric2", State: "", Age: 0},
				},
			},
			{
				name:    "formula with custom states",
				formula: "[metric1{in}] + [metric2{out}]",
				expectedArguments: []CalculationArgument{
					{Metric: "metric1", State: "in", Age: 0},
					{Metric: "metric2", State: "out", Age: 0},
				},
			},
			{
				name:    "formula with age",
				formula: "[metric1(2)] + [metric2]",
				expectedArguments: []CalculationArgument{
					{Metric: "metric1", State: "", Age: 2},
					{Metric: "metric2", State: "", Age: 0},
				},
			},
			{
				name:    "formula with states and age",
				formula: "[metric1(2){in}] + [metric2{out}]",
				expectedArguments: []CalculationArgument{
					{Metric: "metric1", State: "in", Age: 2},
					{Metric: "metric2", State: "out", Age: 0},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := getCalculationArguments(tt.formula)
				require.Equal(t, len(tt.expectedArguments), len(result), "Argument count mismatch")

				// Create maps for easier comparison (since order might not matter)
				resultMap := make(map[string]CalculationArgument)
				for _, arg := range result {
					resultMap[arg.Metric] = arg
				}

				expectedMap := make(map[string]CalculationArgument)
				for _, arg := range tt.expectedArguments {
					expectedMap[arg.Metric] = arg
				}

				for metric, expectedArg := range expectedMap {
					actualArg, exists := resultMap[metric]
					require.True(t, exists, "Expected metric %s not found in results", metric)
					assert.Equal(t, expectedArg.State, actualArg.State, "State mismatch for metric %s", metric)
					assert.Equal(t, expectedArg.Age, actualArg.Age, "Age mismatch for metric %s", metric)
				}
			})
		}
	})

	t.Run("sanitize formula", func(t *testing.T) {
		tests := []struct {
			name            string
			formula         string
			defaultState    string
			expectedFormula string
		}{
			{
				name:            "simple formula",
				formula:         "[metric1] + [metric2]",
				defaultState:    "default",
				expectedFormula: "[metric1(0){default}] + [metric2(0){default}]",
			},
			{
				name:            "formula with some states",
				formula:         "[metric1{in}] + [metric2]",
				defaultState:    "out",
				expectedFormula: "[metric1(0){in}] + [metric2(0){out}]",
			},
			{
				name:            "formula with ages",
				formula:         "[metric1(2)] + [metric2]",
				defaultState:    "default",
				expectedFormula: "[metric1(2){default}] + [metric2(0){default}]",
			},
			{
				name:            "formula with states and ages",
				formula:         "[metric1(2){in}] + [metric2]",
				defaultState:    "out",
				expectedFormula: "[metric1(2){in}] + [metric2(0){out}]",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := sanitizeFormula(tt.formula, tt.defaultState)
				assert.Equal(t, tt.expectedFormula, result)
			})
		}
	})
}

func TestContractState_FullWorkflow(t *testing.T) {
	cs := NewContractState("test", zap.NewNop(), nil)
	cs.GenerateDefaultContract("[container.metric1(0){in}] + [container.metric1{out}] + [container.metric2]", []string{"in", "out"})
	cs.RegisterService("full.service.name.instance.0", []CalculationContract{}, "100")

	ms := []struct {
		metricName string
		datapoints map[string]float64
	}{
		{
			metricName: "container.metric1",
			datapoints: map[string]float64{
				"in":  10,
				"out": 20,
			},
		},
		{
			metricName: "container.metric2",
			datapoints: map[string]float64{
				"in":  30,
				"out": 40,
			},
		},
	}
	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	resourceMetrics.Resource().Attributes().PutStr("container_id", "full.service.name.instance.0")
	resourceMetrics.Resource().Attributes().PutStr("service.name", "test")
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()

	for _, m := range ms {
		metric := scopeMetrics.Metrics().AppendEmpty()
		metric.SetName(m.metricName)
		metric.SetEmptySum()
		dp := metric.Sum().DataPoints()
		for state, value := range m.datapoints {
			dp := dp.AppendEmpty()
			dp.SetTimestamp(pcommon.Timestamp(time.Now().UnixNano()))
			dp.Attributes().PutStr("state", state)
			dp.SetDoubleValue(value)
		}
	}
	cs.PopulateData(metrics)

	_ = cs.Evaluate()
}
