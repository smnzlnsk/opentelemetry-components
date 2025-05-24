package domain

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Knetic/govaluate"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type ContractState struct {
	sync.RWMutex
	once                sync.Once
	processorName       string
	logger              *zap.Logger
	Contracts           ContractManager
	Filters             MetricFilter
	Datapoints          DatapointManager
	compiledExpressions map[ContractKey]*govaluate.EvaluableExpression
	service             ContractService
}

func NewContractState(name string, logger *zap.Logger, services Services, dm DatapointManager) *ContractState {
	return &ContractState{
		processorName:       name,
		logger:              logger,
		Contracts:           NewContractManager(),
		Filters:             NewFilter(),
		Datapoints:          dm,
		compiledExpressions: make(map[ContractKey]*govaluate.EvaluableExpression),
		service:             services.GetContractService(),
	}
}

func (c *ContractState) Sync() {
	c.once.Do(func() {
		// TODO: Implement contract population logic
		// Here we should re-populate the contracts from the database
		// This is to ensure proper functionality in case of a restart or crash
		documents, err := c.service.GetContractsForProcessor(context.Background(), c.processorName)
		if err != nil {
			c.logger.Error("Failed to get contracts for processor", zap.Error(err))
			return
		}

		for _, document := range documents {
			for _, contract := range document.Contracts {
				k := ContractKey{Service: contract.Service, Formula: contract.Formula}

				// Add contract to the state
				c.Contracts.AddContract(contract)

				for _, argument := range contract.Arguments {
					// Add filters for the contract
					c.Filters.AddMetricFilter(argument.Metric, argument.State)
				}

				// Compile the formula
				expr, err := govaluate.NewEvaluableExpression(contract.Formula)
				if err != nil {
					c.logger.Error("Failed to compile formula", zap.Error(err))
					continue
				}

				// Add compiled formula to the state
				c.compiledExpressions[k] = expr
			}
		}
	})
	c.logger.Info("Contracts synced", zap.Int("count", c.Contracts.Length()))
}

func (c *ContractState) GetDefaultContracts() map[string]CalculationContract {
	c.RLock()
	defer c.RUnlock()
	return c.Contracts.GetDefaultContracts()
}

func (c *ContractState) GenerateDefaultContract(formula string, states []string) error {
	if formula == "" {
		return fmt.Errorf("formula cannot be empty")
	}

	if states == nil {
		return fmt.Errorf("states cannot be nil")
	}
	// Loop through the states and generate a default contract for each state
	for _, state := range states {
		// Sanitize the formula to add default states where missing
		sanitizedFormula := sanitizeFormula(formula, state)
		expr, err := govaluate.NewEvaluableExpression(sanitizedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", sanitizedFormula, err)
		}

		key := ContractKey{
			Service: "default",
			Formula: formula,
		}

		arguments := getCalculationArguments(sanitizedFormula)
		contract := CalculationContract{
			Processor: c.processorName,
			Formula:   sanitizedFormula,
			Service:   "default",
			State:     state,
			Arguments: arguments,
		}

		c.Contracts.AddContract(contract)
		c.compiledExpressions[key] = expr

		// Add metrics to filters
		for _, argument := range arguments {
			if err := c.Filters.AddMetricFilter(argument.Metric, argument.State); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ContractState) RegisterService(service string, contracts []CalculationContract, normalizationValue string) error {
	c.Lock()
	defer c.Unlock()

	if service == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if contracts == nil {
		return fmt.Errorf("contracts map cannot be nil")
	}

	// verify normalization value is a parsable number
	normValue, err := strconv.ParseFloat(normalizationValue, 64)
	if err != nil {
		return fmt.Errorf("invalid normalization value %s: %w", normalizationValue, err)
	}

	// Check if service already exists
	if c.Contracts.IsServiceRegistered(service) {
		return fmt.Errorf("service %s already registered", service)
	}

	// First register default contracts for this service
	for formula, contract := range c.Contracts.GetDefaultContracts() {

		serviceKey := ContractKey{
			Service: service,
			Formula: formula,
		}

		serviceContract := contract
		serviceContract.Service = service
		serviceContract.Processor = c.processorName

		normalisedFormula := fmt.Sprintf("(%s) / %f", formula, normValue)
		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", formula, err)
		}

		c.Contracts.AddContract(serviceContract)
		c.compiledExpressions[serviceKey] = expr

		// Update filters for default contract metrics with states
		arguments := getCalculationArguments(formula)
		for _, argument := range arguments {
			if err := c.Filters.AddMetricFilter(argument.Metric, argument.State); err != nil {
				return err
			}
		}
	}

	// Then register service-specific contracts
	for _, contract := range contracts {
		sanitizedFormula := sanitizeFormula(contract.Formula, contract.State)
		key := ContractKey{Service: service, Formula: sanitizedFormula}

		// Set processor name for the contract and update formula
		contract.Formula = sanitizedFormula
		contract.Processor = c.processorName
		contract.Service = service

		normalisedFormula := fmt.Sprintf("(%s) / %f", sanitizedFormula, normValue)

		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", contract.Formula, err)
		}

		c.Contracts.AddContract(contract)
		c.compiledExpressions[key] = expr

		// Update filters with metric states
		arguments := getCalculationArguments(normalisedFormula)
		for _, argument := range arguments {
			if err := c.Filters.AddMetricFilter(argument.Metric, argument.State); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ContractState) DeleteService(service string) error {
	c.Lock()
	defer c.Unlock()

	if service == "default" {
		return fmt.Errorf("cannot delete default contracts")
	}

	// Clean up metric filters and contracts
	for key, contract := range c.Contracts.GetServiceContracts() {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for _, argument := range contract.Arguments {
				err := c.Filters.DeleteMetricFilter(argument.Metric, argument.State)
				if err != nil {
					c.logger.Error("Failed to delete metric filter", zap.Error(err))
				}
			}
			c.Contracts.DeleteContract(contract)
			delete(c.compiledExpressions, key)
		}
	}

	// Clean up datapoints
	err := c.Datapoints.DeleteDatapointForService(service)
	if err != nil {
		c.logger.Error("Failed to delete datapoints for service", zap.Error(err))
	}

	return nil
}

func (c *ContractState) RemoveContract(service string) error {
	if service == "default" {
		return fmt.Errorf("cannot remove default contracts")
	}

	c.Lock()
	defer c.Unlock()

	// Find and remove all contracts for the service
	for key, contract := range c.Contracts.GetAllContracts() {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for _, argument := range contract.Arguments {
				err := c.Filters.DeleteMetricFilter(argument.Metric, argument.State)
				if err != nil {
					c.logger.Error("Failed to delete metric filter", zap.Error(err))
				}
			}
			c.Contracts.DeleteContract(contract)
			delete(c.compiledExpressions, key)
		}
	}
	return nil
}

func (c *ContractState) SaveMetrics(metrics pmetric.Metrics) error {
	c.Lock()
	defer c.Unlock()

	err := c.Datapoints.SaveMetrics(metrics)
	if err != nil {
		c.logger.Error("Failed to populate datapoints", zap.Error(err))
		return err
	}

	return nil
}

func (c *ContractState) Evaluate() CalculationResults {
	c.RLock()
	defer c.RUnlock()

	serviceContracts := make(map[string][]ContractKey)
	for key := range c.Contracts.GetServiceContracts() {
		serviceContracts[key.Service] = append(serviceContracts[key.Service], key)
	}

	res := make(CalculationResults)
	var mu sync.Mutex
	var wg sync.WaitGroup

	workChan := make(chan workItem)
	numWorkers := 4

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				contract, err := c.Contracts.GetContract(work.contractKey)
				if err != nil {
					c.logger.Error("Failed to get contract", zap.Error(err))
					continue
				}
				expr := c.compiledExpressions[work.contractKey]
				if expr == nil {
					continue
				}

				c.logger.Debug("Evaluating", zap.Any("contract", contract))
				params := c.GetParameters(contract)
				result, err := expr.Evaluate(params)
				if err != nil {
					c.logger.Error("Failed to evaluate expression", zap.Error(err))
					continue
				}

				resultKey := CalculationResultKey{
					Service: work.service,
					Formula: work.contractKey.Formula,
					State:   contract.State,
				}

				mu.Lock()
				res[resultKey] = result.(float64)
				mu.Unlock()
			}
		}()
	}

	go func() {
		for service, contractKeys := range serviceContracts {
			for _, key := range contractKeys {
				workChan <- workItem{
					service:     service,
					contractKey: key,
				}
			}
		}
		close(workChan)
	}()

	wg.Wait()
	return res
}

func (c *ContractState) GetParameters(cc CalculationContract) CalculationParameters {
	res := make(map[string]interface{})

	for _, arg := range cc.Arguments {
		// Reset service name for system metrics
		serviceForLookup := cc.Service
		if IsSystemMetric(arg.Metric) {
			serviceForLookup = ""
		}

		key := DatapointKey{
			Service: serviceForLookup,
			Metric:  arg.Metric,
			State:   arg.State,
		}

		dp, exists := c.Datapoints.GetDatapoint(key, arg.Age)
		if exists {
			res[fmt.Sprintf("%s(%d){%s}", arg.Metric, arg.Age, arg.State)] = dp.Value
		}
	}
	return res
}
