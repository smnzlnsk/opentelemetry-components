package domain

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// Define workItem for the evaluation process
type workItem struct {
	service     string
	contractKey ContractKey
}

// CalculationResultKey is a key for calculation results
type CalculationResultKey struct {
	Service string
	Formula string
	State   string
}

// CalculationParameters map[state][metric]metricValue
type CalculationParameters map[string]map[string]interface{}

// CalculationResults maps result keys to calculated values
type CalculationResults map[CalculationResultKey]float64

// MetricDatapoint represents a single datapoint with its metadata
type MetricDatapoint struct {
	Metadata MetricMetadata
	Value    Datapoint
}

// MetricMetadata contains metadata about a metric
type MetricMetadata struct {
	MetricType string // Changed from pmetric.MetricType to avoid dependency
	MetricName string
	MetricUnit string
	Attributes interface{} // Changed from pcommon.Map to avoid dependency
}

// Datapoint represents a single data point value
type Datapoint struct {
	ValueDataType string // Changed from pmetric.NumberDataPointValueType to avoid dependency
	FloatValue    float64
}

// GetServicesMap converts calculation results to a nested map structure
func (cr CalculationResults) GetServicesMap() map[string]map[string]map[string]float64 {
	result := make(map[string]map[string]map[string]float64)

	for key, value := range cr {
		// Initialize nested maps if they don't exist
		if _, ok := result[key.Service]; !ok {
			result[key.Service] = make(map[string]map[string]float64)
		}
		if _, ok := result[key.Service][key.Formula]; !ok {
			result[key.Service][key.Formula] = make(map[string]float64)
		}

		result[key.Service][key.Formula][key.State] = value
	}

	return result
}

// GetServiceNames returns a slice of unique service names
func (cr CalculationResults) GetServiceNames() []string {
	services := make(map[string]struct{})
	for key := range cr {
		services[key.Service] = struct{}{}
	}

	result := make([]string, 0, len(services))
	for service := range services {
		result = append(result, service)
	}
	return result
}

// GetResultsForService returns all results for a given service
func (cr CalculationResults) GetResultsForService(service string) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)

	for key, value := range cr {
		if key.Service == service {
			if _, ok := result[key.Formula]; !ok {
				result[key.Formula] = make(map[string]float64)
			}
			result[key.Formula][key.State] = value
		}
	}

	return result
}

// Normalize scales all calculation results by the given normalization limit
func (cr CalculationResults) Normalize(serviceNormalizationLimit float64) {
	for key, value := range cr {
		cr[key] = value / serviceNormalizationLimit
	}
}

type ContractState struct {
	sync.RWMutex
	once                sync.Once
	processorName       string
	logger              *zap.Logger
	Contracts           map[ContractKey]CalculationContract
	Filters             MetricFilter
	Datapoints          map[DatapointKey]map[int]MetricDatapoint
	compiledExpressions map[ContractKey]*govaluate.EvaluableExpression
	indexTracker        map[DatapointKey]int
	service             ContractService
}

func NewContractState(name string, logger *zap.Logger, service ContractService) *ContractState {
	return &ContractState{
		processorName:       name,
		logger:              logger,
		Contracts:           make(map[ContractKey]CalculationContract),
		Filters:             NewFilter(),
		Datapoints:          make(map[DatapointKey]map[int]MetricDatapoint),
		compiledExpressions: make(map[ContractKey]*govaluate.EvaluableExpression),
		indexTracker:        make(map[DatapointKey]int),
		service:             service,
	}
}

func (c *ContractState) Sync() {
	c.once.Do(func() {
		// TODO: Implement contract population logic
		// Here we should re-populate the contracts from the database
		// This is to ensure proper functionality in case of a restart or crash
		contracts, err := c.service.GetContractsForProcessor(context.Background(), c.processorName)
		if err != nil {
			c.logger.Error("Failed to get contracts for processor", zap.Error(err))
			return
		}

		for _, documents := range contracts {
			for _, contract := range documents.Contracts {
				k := ContractKey{Service: contract.Service, Formula: contract.Formula}

				// add contract to the state
				c.Contracts[k] = contract

				// add filters for the contract
				for metric := range contract.Metrics {
					c.Filters.AddMetricFilter(metric, contract.States)
				}

				// compile the formula
				expr, err := govaluate.NewEvaluableExpression(contract.Formula)
				if err != nil {
					c.logger.Error("Failed to compile formula", zap.Error(err))
					continue
				}

				// add compiled formula to the state
				c.compiledExpressions[k] = expr
			}
		}
	})
	c.logger.Info("Contracts synced", zap.Int("count", len(c.Contracts)))
}

func (c *ContractState) GetDefaultContracts() map[string]CalculationContract {
	c.RLock()
	defer c.RUnlock()

	res := make(map[string]CalculationContract)
	for key, contract := range c.Contracts {
		if key.Service == "default" {
			res[key.Formula] = contract
		}
	}
	return res
}

func (c *ContractState) GenerateDefaultContract(formula string, states map[string]bool) error {
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		return fmt.Errorf("invalid formula %s: %w", formula, err)
	}

	key := ContractKey{
		Service: "default",
		Formula: formula,
	}

	contract := CalculationContract{
		Processor: c.processorName,
		Formula:   formula,
		Service:   "default",
		States:    states,
		Metrics:   filterMetricsFromFormula(formula),
	}

	c.Contracts[key] = contract
	c.compiledExpressions[key] = expr

	return nil
}

func (c *ContractState) RegisterService(service string, contracts map[string]CalculationContract, normalizationValue string) error {
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
	for key := range c.Contracts {
		if key.Service == service {
			return fmt.Errorf("service %s already registered", service)
		}
	}

	// First register default contracts for this service
	for key, contract := range c.Contracts {
		if key.Service != "default" {
			continue
		}

		serviceKey := ContractKey{
			Service: service,
			Formula: key.Formula,
		}

		serviceContract := contract
		serviceContract.Service = service
		serviceContract.Processor = c.processorName
		normalisedFormula := fmt.Sprintf("(%s) / %f", key.Formula, normValue)
		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", key.Formula, err)
		}

		c.Contracts[serviceKey] = serviceContract
		c.compiledExpressions[serviceKey] = expr

		// Update filters for default contract metrics
		for metric := range serviceContract.Metrics {
			if err := c.Filters.AddMetricFilter(metric, serviceContract.States); err != nil {
				return err
			}
		}
	}

	// Then register service-specific contracts
	for formula, contract := range contracts {
		key := ContractKey{Service: service, Formula: formula}

		// Set processor name for the contract
		contract.Processor = c.processorName
		contract.Service = service

		normalisedFormula := fmt.Sprintf("(%s) / %f", formula, normValue)
		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", formula, err)
		}

		c.Contracts[key] = contract
		c.compiledExpressions[key] = expr

		// Update filters
		for metric := range contract.Metrics {
			if err := c.Filters.AddMetricFilter(metric, contract.States); err != nil {
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
	for key, contract := range c.Contracts {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for metric := range contract.Metrics {
				c.Filters.DeleteMetricFilter(metric, contract.States)
			}
			delete(c.Contracts, key)
			delete(c.compiledExpressions, key)
		}
	}

	// Clean up datapoints
	for key := range c.Datapoints {
		if key.Service == service {
			delete(c.Datapoints, key)
		}
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
	for key, contract := range c.Contracts {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for metric := range contract.Metrics {
				c.Filters.DeleteMetricFilter(metric, contract.States)
			}
			delete(c.Contracts, key)
			delete(c.compiledExpressions, key)
		}
	}
	return nil
}

func (c *ContractState) PopulateData(metrics pmetric.Metrics) error {
	c.Lock()
	defer c.Unlock()

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		rmAttr := rm.Resource().Attributes()

		// Extract service name from attributes
		var serviceName string
		containerMetric := false

		// Check for service name in different attribute keys
		_, ok := rmAttr.Get("service.name")
		cid, ok2 := rmAttr.Get("container_id")
		if ok && ok2 {
			serviceName = cid.Str()
			containerMetric = true
		}

		// Skip if no service name found and it's not a system metric
		if serviceName == "" && containerMetric {
			continue
		}

		// Check if service is registered (only for non-system metrics)
		if serviceName != "" {
			serviceExists := false
			for key := range c.Contracts {
				if key.Service == "default" {
					continue
				}
				if key.Service == serviceName {
					serviceExists = true
					break
				}
			}
			if !serviceExists {
				continue
			}
		}

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			smetric := rm.ScopeMetrics().At(j)
			for k := 0; k < smetric.Metrics().Len(); k++ {
				mmetric := smetric.Metrics().At(k)

				// Skip non-container metrics for container services
				if containerMetric && !strings.HasPrefix(mmetric.Name(), "container.") {
					continue
				}

				// Check if metric is registered
				metricFilter, ok := c.Filters.MetricFiltersMap()[mmetric.Name()]
				if !ok {
					continue
				}

				// Check if metric has state information
				statesPresent := mmetric.Sum().DataPoints().Len() > 1
				if statesPresent && metricFilter.StateFilter == nil {
					continue
				}

				for x := 0; x < mmetric.Sum().DataPoints().Len(); x++ {
					ndp := mmetric.Sum().DataPoints().At(x)
					mdp := CreateMetricDatapoint(mmetric, x)

					state := "default"
					if statesPresent {
						if v, ok := ndp.Attributes().Get("state"); ok {
							if metricFilter.StateFilter[v.Str()] != 0 {
								state = v.Str()
							} else {
								continue
							}
						}
					}

					key := DatapointKey{
						Service: serviceName,
						Metric:  mmetric.Name(),
						State:   state,
					}

					// Check if the key's slice exists, if not create it
					if _, exists := c.Datapoints[key]; !exists {
						c.Datapoints[key] = make(map[int]MetricDatapoint)
						c.indexTracker[key] = 0
					}

					// Add the new datapoint to the current index position
					idx := c.indexTracker[key]
					c.logger.Debug("PopulateData", zap.Any("idx", idx), zap.Any("mdp", mdp), zap.Any("key", key))
					c.Datapoints[key][idx] = mdp
					// update index tracker for metric
					c.indexTracker[key] = (idx + 1) % 5 // wrap around to limit map size to 5
				}
			}
		}
	}

	return nil
}

func (c *ContractState) Evaluate() CalculationResults {
	c.RLock()
	defer c.RUnlock()

	// Skip default contracts in evaluation
	serviceContracts := make(map[string][]ContractKey)
	for key := range c.Contracts {
		if key.Service != "default" {
			serviceContracts[key.Service] = append(serviceContracts[key.Service], key)
		}
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
				contract := c.Contracts[work.contractKey]
				expr := c.compiledExpressions[work.contractKey]
				if expr == nil {
					continue
				}

				c.logger.Debug("Evaluating", zap.Any("contract", contract))
				params := c.GetParameters(contract)
				for state, cp := range params {
					result, err := expr.Evaluate(cp)
					if err != nil {
						fmt.Printf("error evaluating expression %v\n", err)
						continue
					}

					resultKey := CalculationResultKey{
						Service: work.service,
						Formula: work.contractKey.Formula,
						State:   state,
					}

					mu.Lock()
					res[resultKey] = result.(float64)
					mu.Unlock()
				}
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
	res := make(map[string]map[string]interface{})

	for state := range cc.States {
		age := 0
		res[state] = make(map[string]interface{})
		for metric := range cc.Metrics {
			dpMetricName := metric

			// Reset service name for system metrics

			serviceForLookup := cc.Service
			if strings.HasPrefix(metric, "system.") {
				serviceForLookup = ""
			}

			if strings.Contains(metric, "|") {
				parts := strings.Split(metric, "|")
				metric = parts[0]
				age, _ = strconv.Atoi(parts[1])
				// if the metric has an index, we need to format it back as it is in the formula
				dpMetricName = fmt.Sprintf("%s|%d", metric, age)
			}

			key := DatapointKey{
				Service: serviceForLookup,
				Metric:  metric,
				State:   state,
			}
			dp, exists := c.getDatapoint(key, age)
			if exists {
				res[state][dpMetricName] = dp.Value.FloatValue
			} else {
				// if the datapoint does not exist, we should fallback to the default state
				key.State = "default"
				dp, exists = c.getDatapoint(key, age)
				if exists {
					c.logger.Debug("GetParameters", zap.Any("key", key), zap.Any("dp", dp), zap.Bool("exists", exists))
					res[state][dpMetricName] = dp.Value.FloatValue
				}
			}
		}
	}
	return res
}

// getDatapoint returns the datapoint for a given key and age
// age 0 is the most recent datapoint
func (c *ContractState) getDatapoint(key DatapointKey, age int) (MetricDatapoint, bool) {
	if dps, exists := c.Datapoints[key]; exists {
		currentIdx := c.indexTracker[key]
		lookupIdx := (currentIdx - 1 - age + 5) % 5
		dp, exists := dps[lookupIdx]
		c.logger.Debug("getDatapoint", zap.Any("key", key), zap.Int("lookupIdx", lookupIdx), zap.Any("dp", dp), zap.Bool("exists", exists))
		return dp, exists
	}
	return MetricDatapoint{}, false
}

// extract all necessary metrics from formula as a map to filter in the future
var metricRegex = regexp.MustCompile(`\[(.*?)\]`)

func filterMetricsFromFormula(formula string) map[string]bool {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	res := make(map[string]bool, len(matches))

	for _, metricName := range matches {
		if strings.HasPrefix(metricName[1], "|") {
			parts := strings.Split(metricName[1], "|")
			res[parts[0]] = true
		} else {
			if len(metricName) > 1 {
				res[metricName[1]] = true
			}
		}
	}
	return res
}

// CreateMetricDatapoint creates a MetricDatapoint from an OpenTelemetry metric
func CreateMetricDatapoint(metric pmetric.Metric, idx int) MetricDatapoint {
	ndp := metric.Sum().DataPoints().At(idx)
	var value float64
	switch ndp.ValueType() {
	case pmetric.NumberDataPointValueTypeDouble:
		value = ndp.DoubleValue()
	case pmetric.NumberDataPointValueTypeInt:
		value = float64(ndp.IntValue())
	case pmetric.NumberDataPointValueTypeEmpty:
		value = 0
	}

	// Convert to domain types
	md := MetricDatapoint{
		Metadata: MetricMetadata{
			MetricType: metric.Type().String(),
			MetricName: metric.Name(),
			MetricUnit: metric.Unit(),
			Attributes: metric.Metadata(),
		},
		Value: Datapoint{
			ValueDataType: ndp.ValueType().String(),
			FloatValue:    value,
		},
	}
	return md
}

// NewCalculationContractsFromProto creates CalculationContracts from protobuf requests
func NewCalculationContractsFromProto(service string, reqs []*pb.CalculationRequest) map[string]CalculationContract {
	res := make(map[string]CalculationContract, len(reqs))
	for _, req := range reqs {
		states := make(map[string]bool, len(req.States))
		for _, state := range req.States {
			states[state] = true
		}

		res[req.Formula] = CalculationContract{

			Formula: req.Formula,
			Service: service,
			States:  states,
			Metrics: filterMetricsFromFormula(req.Formula),
		}
	}
	return res
}
