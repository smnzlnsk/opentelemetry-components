package domain

import (
	"fmt"
	"sync"
)

type ContractManager interface {
	AddContract(contract CalculationContract) error
	GetContract(key ContractKey) (CalculationContract, error)
	DeleteContract(contract CalculationContract) error
	Length() int
	GetAllContracts() map[ContractKey]CalculationContract
	GetDefaultContracts() map[string]CalculationContract
	IsServiceRegistered(service string) bool
}

type contractManager struct {
	rwMutex   sync.RWMutex
	contracts map[ContractKey]CalculationContract
}

func NewContractManager() ContractManager {
	return &contractManager{
		rwMutex:   sync.RWMutex{},
		contracts: make(map[ContractKey]CalculationContract),
	}
}

func (c *contractManager) Length() int {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return len(c.contracts)
}

func (c *contractManager) AddContract(contract CalculationContract) error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	key := ContractKey{
		Service: contract.Service,
		Formula: contract.Formula,
	}
	c.contracts[key] = contract
	return nil
}

func (c *contractManager) GetContract(key ContractKey) (CalculationContract, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	contract, ok := c.contracts[key]
	if !ok {
		return CalculationContract{}, fmt.Errorf("contract not found")
	}
	return contract, nil
}

func (c *contractManager) GetAllContracts() map[ContractKey]CalculationContract {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return c.contracts
}

// GetDefaultContracts returns all contracts for the default service
// The map key is the formula
func (c *contractManager) GetDefaultContracts() map[string]CalculationContract {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	res := make(map[string]CalculationContract)
	for key, contract := range c.contracts {
		if key.Service == "default" {
			res[key.Formula] = contract
		}
	}
	return res
}

func (c *contractManager) IsServiceRegistered(service string) bool {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	for key := range c.contracts {
		if key.Service == service {
			return true
		}
	}
	return false
}

func (c *contractManager) DeleteContract(contract CalculationContract) error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	key := ContractKey{
		Service: contract.Service,
		Formula: contract.Formula,
	}
	delete(c.contracts, key)
	return nil
}
