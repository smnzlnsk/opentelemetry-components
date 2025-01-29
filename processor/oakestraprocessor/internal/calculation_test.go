package internal

import (
	"fmt"
	"sync"
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
		if cs.Datapoints == nil {
			t.Error("Datapoints map was not initialized")
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
					Metrics: map[string]bool{"metric1": true, "metric2": true},
				},
			}

			err := cs.RegisterService(service, contracts, "100")
			require.NoError(t, err)

			// Verify contract registration
			key := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
			contract, exists := cs.Contracts[key]
			require.True(t, exists, "Contract was not registered")
			require.Equal(t, service, contract.Service, "Service mismatch")

			// Verify metric filters
			expectedMetrics := map[string]bool{
				"metric1": true,
				"metric2": true,
			}
			for metric := range expectedMetrics {
				_, exists := cs.Filters.MetricFilters[metric]
				require.True(t, exists, "Expected metric %s not found in Filters", metric)
			}
		})

		t.Run("duplicate service registration", func(t *testing.T) {
			cs := NewContractState()
			service := "test-service"
			contracts := map[string]CalculationContract{
				"[metric1]": {
					Formula: "[metric1]",
					Service: service,
					Metrics: map[string]bool{"metric1": true},
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
		cs := NewContractState()
		defaultFormula := "[metric3] + [metric4]"
		service := "test-service"

		// Set up default contract
		err := cs.GenerateDefaultContract(defaultFormula, map[string]bool{"running": true})
		require.NoError(t, err)

		// Verify default contract registration
		defaultKey := ContractKey{Service: "default", Formula: defaultFormula}
		_, exists := cs.Contracts[defaultKey]
		require.True(t, exists, "Default contract not registered")

		// Register service with its own contract
		contracts := map[string]CalculationContract{
			"[metric1] + [metric2]": {
				Formula: "[metric1] + [metric2]",
				Service: service,
				States:  map[string]bool{"running": true},
				Metrics: map[string]bool{"metric1": true, "metric2": true},
			},
		}

		err = cs.RegisterService(service, contracts, "100")
		require.NoError(t, err)

		// Count contracts for the service
		serviceContractCount := 0
		for key := range cs.Contracts {
			if key.Service == service {
				serviceContractCount++
			}
		}
		require.Equal(t, 2, serviceContractCount, "Expected 2 contracts (1 service + 1 default)")

		// Verify specific contracts exist
		key1 := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
		key2 := ContractKey{Service: service, Formula: "[metric3] + [metric4]"}
		_, exists = cs.Contracts[key1]
		require.True(t, exists, "Service-specific contract not found")
		_, exists = cs.Contracts[key2]
		require.True(t, exists, "Default contract not found for service")

		// Verify all metric filters
		expectedMetrics := []string{"metric1", "metric2", "metric3", "metric4"}
		for _, metric := range expectedMetrics {
			_, exists := cs.Filters.MetricFilters[metric]
			require.True(t, exists, "Expected metric %s not found in Filters", metric)
		}
	})

	t.Run("delete service", func(t *testing.T) {
		t.Run("cannot delete default service", func(t *testing.T) {
			cs := NewContractState()
			err := cs.DeleteService("default")
			require.Error(t, err, "Should not be able to delete default service")
		})

		t.Run("delete service with contracts", func(t *testing.T) {
			cs := NewContractState()
			service := "service1"
			formula := "[metric1] + [metric2] > 1.0"
			contracts := map[string]CalculationContract{
				formula: {
					Formula: formula,
					Service: service,
					States:  map[string]bool{"state1": true, "state2": true},
					Metrics: map[string]bool{"metric1": true, "metric2": true},
				},
			}

			// Setup and verify initial state
			err := cs.RegisterService(service, contracts, "100")
			require.NoError(t, err)

			// Verify initial metric filters
			for metric := range contracts[formula].Metrics {
				_, exists := cs.Filters.MetricFilters[metric]
				require.True(t, exists, "Metric %s filter should exist before deletion", metric)
			}

			// Delete service
			err = cs.DeleteService(service)
			require.NoError(t, err)

			// Verify all cleanup
			key := ContractKey{Service: service, Formula: formula}
			_, exists := cs.Contracts[key]
			require.False(t, exists, "Contract should be deleted")

			for metric := range contracts[formula].Metrics {
				_, exists := cs.Filters.MetricFilters[metric]
				require.False(t, exists, "Metric %s filter should be deleted", metric)
			}

			// Verify datapoints cleanup
			for metric := range contracts[formula].Metrics {
				for state := range contracts[formula].States {
					dpKey := DatapointKey{
						Service: service,
						Metric:  metric,
						State:   state,
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
				cs := NewContractState()
				start := time.Now()
				_ = cs.RegisterService(bm.service, contracts, "100")
				b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
			}
		})
	}
}

// Helper function to generate test contracts
func generateTestContracts(service string, count int) map[string]CalculationContract {
	contracts := make(map[string]CalculationContract)
	for i := 0; i < count; i++ {
		formula := fmt.Sprintf("[metric%d] + [metric%d]", i*2+1, i*2+2)
		metrics := map[string]bool{
			fmt.Sprintf("metric%d", i*2+1): true,
			fmt.Sprintf("metric%d", i*2+2): true,
		}
		contracts[formula] = CalculationContract{
			Formula: formula,
			Service: service,
			States:  map[string]bool{"running": true},
			Metrics: metrics,
		}
	}
	return contracts
}

func TestRegisterServiceWithDefaults(t *testing.T) {
	cs := NewContractState()
	service := "test-service"

	// Set up a default contract
	defaultFormula := "[metric3] + [metric4]"
	err := cs.GenerateDefaultContract(defaultFormula, map[string]bool{"running": true})
	require.NoError(t, err)

	// Create service-specific contracts
	serviceContracts := map[string]CalculationContract{
		"[metric1] + [metric2]": {
			Formula: "[metric1] + [metric2]",
			Service: service,
			States:  map[string]bool{"running": true},
		},
	}

	// Register service
	err = cs.RegisterService(service, serviceContracts, "100")
	require.NoError(t, err)

	// Count contracts for the service
	serviceContractCount := 0
	for key := range cs.Contracts {
		if key.Service == service {
			serviceContractCount++
		}
	}
	require.Equal(t, 2, serviceContractCount, "Expected 2 contracts (1 service + 1 default)")

	// Verify both formulas exist
	key1 := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
	key2 := ContractKey{Service: service, Formula: "[metric3] + [metric4]"}
	_, exists := cs.Contracts[key1]
	require.True(t, exists, "Service-specific contract not found")
	_, exists = cs.Contracts[key2]
	require.True(t, exists, "Default contract not found")
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
				tt.initialFormula: {
					Formula: tt.initialFormula,
					Service: tt.initialService,
					States:  tt.initialStates,
					Metrics: filterMetricsFromFormula(tt.initialFormula),
				},
			}

			// Register service
			err := cs.RegisterService(tt.initialService, contracts, "100")
			require.NoError(t, err)

			// Verify initial setup
			key := ContractKey{Service: tt.initialService, Formula: tt.initialFormula}
			_, exists := cs.Contracts[key]
			require.True(t, exists, "Contract should exist before deletion")

			// Verify metrics were properly registered
			for metric := range tt.expectedMetrics {
				mf, exists := cs.Filters.MetricFilters[metric]
				require.True(t, exists, "metric filter should exist before deletion: %s", metric)
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
			_, exists = cs.Contracts[key]
			require.False(t, exists, "Contract should be deleted")

			// Verify metric filter cleanup
			for metric, shouldExist := range tt.expectedMetrics {
				_, exists := cs.Filters.MetricFilters[metric]
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
	cs := NewContractState()
	defaultFormula := "[metric3] + [metric4]"
	service := "test-service"

	// Set up and verify default contract
	err := cs.GenerateDefaultContract(defaultFormula, map[string]bool{"running": true})
	require.NoError(t, err)

	defaultKey := ContractKey{Service: "default", Formula: defaultFormula}
	defaultContract, exists := cs.Contracts[defaultKey]
	require.True(t, exists, "Default contract not registered")
	require.Equal(t, "default", defaultContract.Service)
	require.Equal(t, defaultFormula, defaultContract.Formula)
	require.Contains(t, defaultContract.States, "running")
	require.Contains(t, defaultContract.Metrics, "metric3")
	require.Contains(t, defaultContract.Metrics, "metric4")

	// Register service and verify default contract is copied
	contracts := map[string]CalculationContract{
		"[metric1] + [metric2]": {
			Formula: "[metric1] + [metric2]",
			Service: service,
			States:  map[string]bool{"running": true},
			Metrics: map[string]bool{"metric1": true, "metric2": true},
		},
	}

	err = cs.RegisterService(service, contracts, "100")
	require.NoError(t, err)

	// Verify service has both its own contract and the default contract
	serviceDefaultKey := ContractKey{Service: service, Formula: defaultFormula}
	serviceContract, exists := cs.Contracts[serviceDefaultKey]
	require.True(t, exists, "Default contract not copied to service")
	require.Equal(t, service, serviceContract.Service)
	require.Equal(t, defaultFormula, serviceContract.Formula)

	// Verify filters include metrics from both contracts
	expectedMetrics := []string{"metric1", "metric2", "metric3", "metric4"}
	for _, metric := range expectedMetrics {
		filter, exists := cs.Filters.MetricFilters[metric]
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
	_, exists = cs.Contracts[defaultKey]
	require.True(t, exists, "Default contract should still exist")
	_, exists = cs.Contracts[serviceDefaultKey]
	require.False(t, exists, "Service's copy of default contract should be deleted")
}

func TestServiceRegistration(t *testing.T) {
	t.Run("register with multiple states", func(t *testing.T) {
		cs := NewContractState()
		service := "test-service"
		contracts := map[string]CalculationContract{
			"[metric1] + [metric2]": {
				Formula: "[metric1] + [metric2]",
				Service: service,
				States:  map[string]bool{"running": true, "stopped": true},
				Metrics: map[string]bool{"metric1": true, "metric2": true},
			},
		}

		err := cs.RegisterService(service, contracts, "100")
		require.NoError(t, err)

		// Verify contract registration
		key := ContractKey{Service: service, Formula: "[metric1] + [metric2]"}
		contract, exists := cs.Contracts[key]
		require.True(t, exists)
		require.Equal(t, service, contract.Service)
		require.Equal(t, 2, len(contract.States))

		// Verify filters
		for metric := range contract.Metrics {
			filter, exists := cs.Filters.MetricFilters[metric]
			require.True(t, exists)
			require.Greater(t, filter.StateFilter["running"], 0)
			require.Greater(t, filter.StateFilter["stopped"], 0)
		}
	})

	t.Run("register with overlapping metrics", func(t *testing.T) {
		cs := NewContractState()
		service := "test-service"
		contracts := map[string]CalculationContract{
			"[metric1] + [metric2]": {
				Formula: "[metric1] + [metric2]",
				Service: service,
				States:  map[string]bool{"running": true},
				Metrics: map[string]bool{"metric1": true, "metric2": true},
			},
			"[metric2] + [metric3]": {
				Formula: "[metric2] + [metric3]",
				Service: service,
				States:  map[string]bool{"running": true},
				Metrics: map[string]bool{"metric2": true, "metric3": true},
			},
		}

		err := cs.RegisterService(service, contracts, "100")
		require.NoError(t, err)

		// Verify metric filters
		filter, exists := cs.Filters.MetricFilters["metric2"]
		require.True(t, exists)
		require.Equal(t, 2, filter.activeContracts, "Metric2 should be used by two contracts")
	})
}

func TestContractState_Comprehensive(t *testing.T) {
	t.Run("default contract handling", func(t *testing.T) {
		cs := NewContractState()

		t.Run("empty default contract", func(t *testing.T) {
			err := cs.GenerateDefaultContract("", map[string]bool{})
			require.Error(t, err, "Should not allow empty formula")
		})

		t.Run("invalid formula", func(t *testing.T) {
			err := cs.GenerateDefaultContract("[metric1] ++ [metric2]", map[string]bool{"running": true})
			require.Error(t, err, "Should not allow invalid formula")
		})

		t.Run("valid default contract", func(t *testing.T) {
			formula := "[metric1] + [metric2]"
			err := cs.GenerateDefaultContract(formula, map[string]bool{"running": true})
			require.NoError(t, err)

			key := ContractKey{Service: "default", Formula: formula}
			contract, exists := cs.Contracts[key]
			require.True(t, exists)
			require.Equal(t, "default", contract.Service)
			require.Contains(t, contract.Metrics, "metric1")
			require.Contains(t, contract.Metrics, "metric2")
		})

		t.Run("duplicate default contract", func(t *testing.T) {
			formula := "[metric3] + [metric4]"
			err := cs.GenerateDefaultContract(formula, map[string]bool{"running": true})
			require.NoError(t, err, "Should allow multiple default contracts")

			count := 0
			for key := range cs.Contracts {
				if key.Service == "default" {
					count++
				}
			}
			require.Equal(t, 2, count, "Should have two default contracts")
		})
	})

	t.Run("service registration edge cases", func(t *testing.T) {
		cs := NewContractState()

		t.Run("register empty service name", func(t *testing.T) {
			contracts := map[string]CalculationContract{
				"[metric1]": {Formula: "[metric1]"},
			}
			err := cs.RegisterService("", contracts, "100")
			require.Error(t, err, "Should not allow empty service name")
		})

		t.Run("register with nil contracts", func(t *testing.T) {
			err := cs.RegisterService("test", nil, "100")
			require.Error(t, err, "Should not allow nil contracts")
		})

		t.Run("register with empty contracts", func(t *testing.T) {
			err := cs.RegisterService("test", map[string]CalculationContract{}, "100")
			require.NoError(t, err, "Should allow empty contracts map")
		})

		t.Run("register with invalid formula", func(t *testing.T) {
			contracts := map[string]CalculationContract{
				"[metric1] ++ [metric2]": {
					Formula: "[metric1] ++ [metric2]",
					Service: "test",
				},
			}
			err := cs.RegisterService("test", contracts, "100")
			require.Error(t, err, "Should not allow invalid formula")
		})
	})

	t.Run("metric filter handling", func(t *testing.T) {
		cs := NewContractState()

		t.Run("overlapping states", func(t *testing.T) {
			// Register first contract
			err := cs.RegisterService("service1", map[string]CalculationContract{
				"[metric1]": {
					Formula: "[metric1]",
					Service: "service1",
					States:  map[string]bool{"running": true, "stopped": true},
					Metrics: map[string]bool{"metric1": true},
				},
			}, "100")
			require.NoError(t, err)

			// Register second contract with overlapping states
			err = cs.RegisterService("service2", map[string]CalculationContract{
				"[metric1]": {
					Formula: "[metric1]",
					Service: "service2",
					States:  map[string]bool{"running": true, "paused": true},
					Metrics: map[string]bool{"metric1": true},
				},
			}, "100")
			require.NoError(t, err)

			filter, exists := cs.Filters.MetricFilters["metric1"]
			require.True(t, exists)
			require.Equal(t, 2, filter.activeContracts)
			require.Greater(t, filter.StateFilter["running"], 0)
			require.Greater(t, filter.StateFilter["stopped"], 0)
			require.Greater(t, filter.StateFilter["paused"], 0)
		})

		t.Run("metric cleanup", func(t *testing.T) {
			cs := NewContractState()

			// Register service with multiple metrics
			err := cs.RegisterService("service1", map[string]CalculationContract{
				"[metric1] + [metric2]": {
					Formula: "[metric1] + [metric2]",
					Service: "service1",
					States:  map[string]bool{"running": true},
					Metrics: map[string]bool{"metric1": true, "metric2": true},
				},
			}, "100")
			require.NoError(t, err)

			// Delete service
			err = cs.DeleteService("service1")
			require.NoError(t, err)

			// Verify all metrics are cleaned up
			_, exists := cs.Filters.MetricFilters["metric1"]
			require.False(t, exists, "Metric1 should be removed")
			_, exists = cs.Filters.MetricFilters["metric2"]
			require.False(t, exists, "Metric2 should be removed")
		})
	})

	t.Run("datapoint handling", func(t *testing.T) {
		cs := NewContractState()

		t.Run("datapoint cleanup", func(t *testing.T) {
			// Register service
			service := "test-service"
			err := cs.RegisterService(service, map[string]CalculationContract{
				"[metric1]": {
					Formula: "[metric1]",
					Service: service,
					States:  map[string]bool{"running": true},
					Metrics: map[string]bool{"metric1": true},
				},
			}, "100")
			require.NoError(t, err)

			// Add datapoints
			dpKey := DatapointKey{
				Service: service,
				Metric:  "metric1",
				State:   "running",
			}
			cs.Datapoints[dpKey] = MetricDatapoint{}

			// Delete service
			err = cs.DeleteService(service)
			require.NoError(t, err)

			// Verify datapoint cleanup
			_, exists := cs.Datapoints[dpKey]
			require.False(t, exists, "Datapoint should be removed")
		})
	})

	t.Run("concurrent operations", func(t *testing.T) {
		cs := NewContractState()

		// Add default contract
		err := cs.GenerateDefaultContract("[metric1]", map[string]bool{"running": true})
		require.NoError(t, err)

		var wg sync.WaitGroup
		services := []string{"service1", "service2", "service3", "service4", "service5"}

		// Concurrently register services
		for _, service := range services {
			wg.Add(1)
			go func(svc string) {
				defer wg.Done()
				contracts := map[string]CalculationContract{
					"[metric2]": {
						Formula: "[metric2]",
						Service: svc,
						States:  map[string]bool{"running": true},
						Metrics: map[string]bool{"metric2": true},
					},
				}
				err := cs.RegisterService(svc, contracts, "100")
				require.NoError(t, err)
			}(service)
		}

		wg.Wait()

		// Verify all services were registered
		serviceCount := 0
		for key := range cs.Contracts {
			if key.Service != "default" {
				serviceCount++
			}
		}
		require.Equal(t, len(services)*2, serviceCount, "Each service should have two contracts (default + specific)")
	})
}
