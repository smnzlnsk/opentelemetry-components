package domain

import (
	"testing"

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
			contract       CalculationContract
			expectedLength int
			shouldError    bool
		}{
			{
				name: "add valid contract",
				contract: CalculationContract{
					Service: "test-service",
					Formula: "test-formula",
				},
				expectedLength: 1,
				shouldError:    false,
			},
			{
				name: "add duplicate contract",
				contract: CalculationContract{
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
				key := ContractKey{
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
			key := ContractKey{
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
			contract       CalculationContract
			expectedLength int
		}{
			{
				name: "delete existing contract",
				contract: CalculationContract{
					Service: "test-service",
					Formula: "test-formula",
				},
				expectedLength: 0,
			},
			{
				name: "delete non-existent contract",
				contract: CalculationContract{
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
				initialContract := CalculationContract{
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
				key := ContractKey{
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
			contracts     []CalculationContract
			expectedCount int
		}{
			{
				name: "multiple contracts",
				contracts: []CalculationContract{
					{Service: "service1", Formula: "formula1"},
					{Service: "service2", Formula: "formula2"},
				},
				expectedCount: 2,
			},
			{
				name:          "no contracts",
				contracts:     []CalculationContract{},
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
				for _, contract := range tt.contracts {
					key := ContractKey{
						Service: contract.Service,
						Formula: contract.Formula,
					}
					assert.Equal(t, contract, contracts[key])
				}
			})
		}
	})

	t.Run("get default contracts", func(t *testing.T) {
		tests := []struct {
			name          string
			contracts     []CalculationContract
			expectedCount int
		}{
			{
				name: "multiple default contracts",
				contracts: []CalculationContract{
					{Service: "default", Formula: "formula1"},
					{Service: "default", Formula: "formula2"},
					{Service: "other-service", Formula: "formula3"},
				},
				expectedCount: 2,
			},
			{
				name: "no default contracts",
				contracts: []CalculationContract{
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
			contract       CalculationContract
			serviceToCheck string
			expectedResult bool
		}{
			{
				name: "check registered service",
				contract: CalculationContract{
					Service: "test-service",
					Formula: "test-formula",
				},
				serviceToCheck: "test-service",
				expectedResult: true,
			},
			{
				name: "check non-registered service",
				contract: CalculationContract{
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
