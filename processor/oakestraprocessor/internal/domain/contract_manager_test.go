package domain

import (
	"testing"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/calculation"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractManager(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		cm := NewContractManager()
		assert.NotNil(t, cm)
		assert.Equal(t, 0, cm.Length())
	})

	t.Run("add and get contract", func(t *testing.T) {
		tests := []struct {
			name           string
			contract       calculation.Contract
			expectedLength int
			shouldError    bool
		}{
			{
				name: "add valid contract",
				contract: calculation.Contract{
					Service: "test-service",
					Formula: "test-formula",
				},
				expectedLength: 1,
				shouldError:    false,
			},
			{
				name: "add duplicate contract",
				contract: calculation.Contract{
					Service: "test-service",
					Formula: "test-formula",
				},
				expectedLength: 1,
				shouldError:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cm := NewContractManager()

				// Add contract
				err := cm.AddContract(tt.contract)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLength, cm.Length())

				// Get contract
				key := contract.Key{
					Service: tt.contract.Service,
					Formula: tt.contract.Formula,
				}
				retrievedContract, err := cm.GetContract(key)
				require.NoError(t, err)
				assert.Equal(t, tt.contract, retrievedContract)
			})
		}

		t.Run("get non-existent contract", func(t *testing.T) {
			cm := NewContractManager()
			key := contract.Key{
				Service: "non-existent",
				Formula: "non-existent",
			}
			_, err := cm.GetContract(key)
			assert.Error(t, err)
		})
	})

	t.Run("delete contract", func(t *testing.T) {
		tests := []struct {
			name           string
			contract       calculation.Contract
			expectedLength int
		}{
			{
				name: "delete existing contract",
				contract: calculation.Contract{
					Service: "test-service",
					Formula: "test-formula",
				},
				expectedLength: 0,
			},
			{
				name: "delete non-existent contract",
				contract: calculation.Contract{
					Service: "non-existent",
					Formula: "non-existent",
				},
				expectedLength: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cm := NewContractManager()

				// Add initial contract
				initialContract := calculation.Contract{
					Service: "test-service",
					Formula: "test-formula",
				}
				err := cm.AddContract(initialContract)
				require.NoError(t, err)

				// Delete contract
				err = cm.DeleteContract(tt.contract)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLength, cm.Length())

				// Verify contract is deleted
				key := contract.Key{
					Service: tt.contract.Service,
					Formula: tt.contract.Formula,
				}
				_, err = cm.GetContract(key)
				assert.Error(t, err)
			})
		}
	})

	t.Run("get all contracts", func(t *testing.T) {
		tests := []struct {
			name          string
			contracts     []calculation.Contract
			expectedCount int
		}{
			{
				name: "multiple contracts",
				contracts: []calculation.Contract{
					{Service: "service1", Formula: "formula1"},
					{Service: "service2", Formula: "formula2"},
				},
				expectedCount: 2,
			},
			{
				name:          "no contracts",
				contracts:     []calculation.Contract{},
				expectedCount: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cm := NewContractManager()

				// Add contracts
				for _, contract := range tt.contracts {
					err := cm.AddContract(contract)
					require.NoError(t, err)
				}

				// Get all contracts
				contracts := cm.GetAllContracts()
				assert.Equal(t, tt.expectedCount, len(contracts))

				// Verify contracts
				for _, c := range tt.contracts {
					key := contract.Key{
						Service: c.Service,
						Formula: c.Formula,
					}
					assert.Equal(t, c, contracts[key])
				}
			})
		}
	})

	t.Run("get default contracts", func(t *testing.T) {
		tests := []struct {
			name          string
			contracts     []calculation.Contract
			expectedCount int
		}{
			{
				name: "multiple default contracts",
				contracts: []calculation.Contract{
					{Service: "default", Formula: "formula1"},
					{Service: "default", Formula: "formula2"},
					{Service: "other-service", Formula: "formula3"},
				},
				expectedCount: 2,
			},
			{
				name: "no default contracts",
				contracts: []calculation.Contract{
					{Service: "service1", Formula: "formula1"},
					{Service: "service2", Formula: "formula2"},
				},
				expectedCount: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cm := NewContractManager()

				// Add contracts
				for _, contract := range tt.contracts {
					err := cm.AddContract(contract)
					require.NoError(t, err)
				}

				// Get default contracts
				defaultContracts := cm.GetDefaultContracts()
				assert.Equal(t, tt.expectedCount, len(defaultContracts))

				// Verify default contracts
				for _, contract := range tt.contracts {
					if contract.Service == "default" {
						assert.Equal(t, contract, defaultContracts[contract.Formula])
					}
				}
			})
		}
	})

	t.Run("service registration", func(t *testing.T) {
		tests := []struct {
			name           string
			contract       calculation.Contract
			serviceToCheck string
			expectedResult bool
		}{
			{
				name: "check registered service",
				contract: calculation.Contract{
					Service: "test-service",
					Formula: "test-formula",
				},
				serviceToCheck: "test-service",
				expectedResult: true,
			},
			{
				name: "check non-registered service",
				contract: calculation.Contract{
					Service: "test-service",
					Formula: "test-formula",
				},
				serviceToCheck: "non-existent-service",
				expectedResult: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cm := NewContractManager()

				// Add contract
				err := cm.AddContract(tt.contract)
				require.NoError(t, err)

				// Check service registration
				result := cm.IsServiceRegistered(tt.serviceToCheck)
				assert.Equal(t, tt.expectedResult, result)
			})
		}
	})
}
