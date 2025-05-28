package domain

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Knetic/govaluate"
	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/pkg/calculation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/contract"
	"github.com/smnzlnsk/opentelemetry-components/pkg/formulae"
	datapoint "github.com/smnzlnsk/opentelemetry-components/pkg/metric/datapoint"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type workItem struct {
	service     string
	contractKey contract.Key
}

type ContractState struct {
	sync.RWMutex
	once                sync.Once
	processorName       string
	logger              *zap.Logger
	Contracts           ContractManager
	Filters             MetricFilter
	Datapoints          DatapointManager
	compiledExpressions map[contract.Key]*govaluate.EvaluableExpression
	service             ContractService
}

func NewContractState(name string, logger *zap.Logger, services Services, dm DatapointManager) (*ContractState, error) {
	if services == nil {
		return nil, fmt.Errorf("services cannot be nil")
	}

	if dm == nil {
		return nil, fmt.Errorf("datapoint manager cannot be nil")
	}

	return &ContractState{
		processorName:       name,
		logger:              logger,
		Contracts:           NewContractManager(),
		Filters:             NewFilter(),
		Datapoints:          dm,
		compiledExpressions: make(map[contract.Key]*govaluate.EvaluableExpression),
		service:             services.GetContractService(),
	}, nil
}

// NewCalculationContractsFromProto creates CalculationContracts from protobuf requests
func NewCalculationContractsFromProto(service string, reqs []*pb.CalculationRequest) []calculation.Contract {
	res := make([]calculation.Contract, 0)
	for _, req := range reqs {
		for _, state := range req.States {
			res = append(res, calculation.Contract{
				Formula:   req.Formula,
				Service:   service,
				State:     state,
				Arguments: formulae.GetCalculationArguments(formulae.Sanitize(req.Formula, state)),
			})
		}
	}
	return res
}

func (cs *ContractState) Sync() {
	cs.once.Do(func() {
		// TODO: Implement contract population logic
		// Here we should re-populate the contracts from the database
		// This is to ensure proper functionality in case of a restart or crash
		documents, err := cs.service.GetContractsForProcessor(context.Background(), cs.processorName)
		if err != nil {
			cs.logger.Error("Failed to get contracts for processor", zap.Error(err))
			return
		}

		for _, document := range documents {
			for _, ct := range document.Contracts {
				k := contract.Key{Service: ct.Service, Formula: ct.Formula}

				// Add contract to the state
				cs.Contracts.AddContract(ct)

				for _, argument := range ct.Arguments {
					// Add filters for the contract
					cs.Filters.AddMetricFilter(argument.Metric, argument.State)
				}

				// Compile the formula
				expr, err := govaluate.NewEvaluableExpression(ct.Formula)
				if err != nil {
					cs.logger.Error("Failed to compile formula", zap.Error(err))
					continue
				}

				// Add compiled formula to the state
				cs.compiledExpressions[k] = expr
			}
		}
	})
	cs.logger.Info("Contracts synced", zap.Int("count", cs.Contracts.Length()))
}

func (cs *ContractState) GetDefaultContracts() map[string]calculation.Contract {
	cs.RLock()
	defer cs.RUnlock()
	return cs.Contracts.GetDefaultContracts()
}

func (cs *ContractState) GenerateDefaultContract(formula string, states []string) error {
	if formula == "" {
		return fmt.Errorf("formula cannot be empty")
	}

	if states == nil {
		return fmt.Errorf("states cannot be nil")
	}
	// Loop through the states and generate a default contract for each state
	for _, state := range states {
		// Sanitize the formula to add default states where missing
		sanitizedFormula := formulae.Sanitize(formula, state)
		expr, err := govaluate.NewEvaluableExpression(sanitizedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", sanitizedFormula, err)
		}

		key := contract.Key{
			Service: "default",
			Formula: formula,
		}

		arguments := formulae.GetCalculationArguments(sanitizedFormula)
		contract := calculation.Contract{
			Processor: cs.processorName,
			Formula:   sanitizedFormula,
			Service:   "default",
			State:     state,
			Arguments: arguments,
		}

		cs.Contracts.AddContract(contract)
		cs.compiledExpressions[key] = expr

		// Add metrics to filters
		for _, argument := range arguments {
			if err := cs.Filters.AddMetricFilter(argument.Metric, argument.State); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *ContractState) RegisterService(service string, contracts []calculation.Contract, normalizationValue string) error {
	cs.Lock()
	defer cs.Unlock()

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
	if cs.Contracts.IsServiceRegistered(service) {
		return fmt.Errorf("service %s already registered", service)
	}

	// First register default contracts for this service
	for formula, ct := range cs.Contracts.GetDefaultContracts() {

		serviceKey := contract.Key{
			Service: service,
			Formula: formula,
		}

		serviceContract := ct
		serviceContract.Service = service
		serviceContract.Processor = cs.processorName

		normalisedFormula := fmt.Sprintf("(%s) / %f", formula, normValue)
		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", formula, err)
		}

		cs.Contracts.AddContract(serviceContract)
		cs.compiledExpressions[serviceKey] = expr

		// Update filters for default contract metrics with states
		arguments := formulae.GetCalculationArguments(formula)
		for _, argument := range arguments {
			if err := cs.Filters.AddMetricFilter(argument.Metric, argument.State); err != nil {
				return err
			}
		}
	}

	// Then register service-specific contracts
	for _, ct := range contracts {
		sanitizedFormula := formulae.Sanitize(ct.Formula, ct.State)
		key := contract.Key{Service: service, Formula: sanitizedFormula}

		// Set processor name for the contract and update formula
		ct.Formula = sanitizedFormula
		ct.Processor = cs.processorName
		ct.Service = service

		normalisedFormula := fmt.Sprintf("(%s) / %f", sanitizedFormula, normValue)

		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", ct.Formula, err)
		}

		cs.Contracts.AddContract(ct)
		cs.compiledExpressions[key] = expr

		// Update filters with metric states
		arguments := formulae.GetCalculationArguments(normalisedFormula)
		for _, argument := range arguments {
			if err := cs.Filters.AddMetricFilter(argument.Metric, argument.State); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *ContractState) DeleteService(service string) error {
	cs.Lock()
	defer cs.Unlock()

	if service == "default" {
		return fmt.Errorf("cannot delete default contracts")
	}

	// Clean up metric filters and contracts
	for key, contract := range cs.Contracts.GetServiceContracts() {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for _, argument := range contract.Arguments {
				err := cs.Filters.DeleteMetricFilter(argument.Metric, argument.State)
				if err != nil {
					cs.logger.Error("Failed to delete metric filter", zap.Error(err))
				}
			}
			cs.Contracts.DeleteContract(contract)
			delete(cs.compiledExpressions, key)
		}
	}

	// Clean up datapoints
	err := cs.Datapoints.DeleteDatapointForService(service)
	if err != nil {
		cs.logger.Error("Failed to delete datapoints for service", zap.Error(err))
	}

	return nil
}

func (cs *ContractState) RemoveContract(service string) error {
	if service == "default" {
		return fmt.Errorf("cannot remove default contracts")
	}

	cs.Lock()
	defer cs.Unlock()

	// Find and remove all contracts for the service
	for key, contract := range cs.Contracts.GetAllContracts() {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for _, argument := range contract.Arguments {
				err := cs.Filters.DeleteMetricFilter(argument.Metric, argument.State)
				if err != nil {
					cs.logger.Error("Failed to delete metric filter", zap.Error(err))
				}
			}
			cs.Contracts.DeleteContract(contract)
			delete(cs.compiledExpressions, key)
		}
	}
	return nil
}

func (cs *ContractState) SaveMetrics(metrics pmetric.Metrics) error {
	cs.Lock()
	defer cs.Unlock()

	err := cs.Datapoints.SaveMetrics(metrics)
	if err != nil {
		cs.logger.Error("Failed to populate datapoints", zap.Error(err))
		return err
	}

	return nil
}

func (cs *ContractState) Evaluate() calculation.Results {
	cs.RLock()
	defer cs.RUnlock()

	serviceContracts := make(map[string][]contract.Key)
	for key := range cs.Contracts.GetServiceContracts() {
		serviceContracts[key.Service] = append(serviceContracts[key.Service], key)
	}

	res := make(calculation.Results)
	var mu sync.Mutex
	var wg sync.WaitGroup

	workChan := make(chan workItem)
	numWorkers := 4

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				contract, err := cs.Contracts.GetContract(work.contractKey)
				if err != nil {
					cs.logger.Error("Failed to get contract", zap.Error(err))
					continue
				}
				expr := cs.compiledExpressions[work.contractKey]
				if expr == nil {
					continue
				}

				cs.logger.Debug("Evaluating", zap.Any("contract", contract))
				params := cs.GetParameters(contract)
				result, err := expr.Evaluate(params)
				if err != nil {
					cs.logger.Error("Failed to evaluate expression", zap.Error(err))
					continue
				}

				resultKey := calculation.ResultKey{
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

func (cs *ContractState) GetParameters(cc calculation.Contract) calculation.Parameters {
	res := make(map[string]interface{})

	for _, arg := range cc.Arguments {
		// Reset service name for system metrics
		serviceForLookup := cc.Service
		if IsSystemMetric(arg.Metric) {
			serviceForLookup = ""
		}

		key := datapoint.Key{
			Service: serviceForLookup,
			Metric:  arg.Metric,
			State:   arg.State,
		}

		dp, exists := cs.Datapoints.GetDatapoint(key, arg.Age)
		if exists {
			res[fmt.Sprintf("%s(%d){%s}", arg.Metric, arg.Age, arg.State)] = dp.Value
		}
	}
	return res
}
