package domain

import (
	"fmt"
	"sync"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/calculation"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/contract"
)

type ContractManager interface {
	AddContract(contract calculation.Contract) error
	GetContract(key contract.Key) (calculation.Contract, error)
	DeleteContract(contract calculation.Contract) error
	Length() int
	GetAllContracts() map[contract.Key]calculation.Contract
	GetDefaultContracts() map[string]calculation.Contract
	GetServiceContracts() map[contract.Key]calculation.Contract
	IsServiceRegistered(service string) bool
}

type contractManager struct {
	rwMutex   sync.RWMutex
	contracts map[contract.Key]calculation.Contract
}

func NewContractManager() ContractManager {
	return &contractManager{
		rwMutex:   sync.RWMutex{},
		contracts: make(map[contract.Key]calculation.Contract),
	}
}

func (c *contractManager) Length() int {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return len(c.contracts)
}

func (c *contractManager) AddContract(ct calculation.Contract) error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	key := contract.Key{
		Service: ct.Service,
		Formula: ct.Formula,
	}
	c.contracts[key] = ct
	return nil
}

func (c *contractManager) GetContract(key contract.Key) (calculation.Contract, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	contract, ok := c.contracts[key]
	if !ok {
		return calculation.Contract{}, fmt.Errorf("contract not found")
	}
	return contract, nil
}

func (c *contractManager) GetAllContracts() map[contract.Key]calculation.Contract {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return c.contracts
}

// GetDefaultContracts returns all contracts for the default service
// The map key is the formula
func (c *contractManager) GetDefaultContracts() map[string]calculation.Contract {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	res := make(map[string]calculation.Contract)
	for key, contract := range c.contracts {
		if key.Service == "default" {
			res[key.Formula] = contract
		}
	}
	return res
}

func (c *contractManager) GetServiceContracts() map[contract.Key]calculation.Contract {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	res := make(map[contract.Key]calculation.Contract)
	for key, contract := range c.contracts {
		if key.Service != "default" {
			res[key] = contract
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

func (c *contractManager) DeleteContract(ct calculation.Contract) error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	key := contract.Key{
		Service: ct.Service,
		Formula: ct.Formula,
	}
	delete(c.contracts, key)
	return nil
}
