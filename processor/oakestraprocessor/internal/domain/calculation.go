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
type CalculationParameters map[string]interface{}

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
	Contracts           ContractManager
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
		Contracts:           NewContractManager(),
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

				for _, argument := range contract.arguments {
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
			arguments: arguments,
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

		// Set processor name for the contract
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
			for _, argument := range contract.arguments {
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
	for key, contract := range c.Contracts.GetAllContracts() {
		if key.Service == service {
			// Remove filters for this contract's metrics
			for _, argument := range contract.arguments {
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
			if !c.Contracts.IsServiceRegistered(serviceName) {
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

					if _, exists := c.Datapoints[key]; !exists {
						c.Datapoints[key] = make(map[int]MetricDatapoint)
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

	for _, arg := range cc.arguments {
		// Reset service name for system metrics
		serviceForLookup := cc.Service
		if strings.HasPrefix(arg.Metric, "system.") {
			serviceForLookup = ""
		}

		key := DatapointKey{
			Service: serviceForLookup,
			Metric:  arg.Metric,
			State:   arg.State,
		}

		dp, exists := c.getDatapoint(key, arg.Age)
		if exists {
			res[fmt.Sprintf("%s(%d){%s}", arg.Metric, arg.Age, arg.State)] = dp.Value.FloatValue
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
		dp := dps[lookupIdx]
		c.logger.Debug("getDatapoint", zap.Any("key", key), zap.Int("lookupIdx", lookupIdx), zap.Any("dp", dp), zap.Bool("exists", exists))
		return dp, exists
	}
	return MetricDatapoint{}, false
}

// Extract all necessary metrics from formula as a map to filter in the future
// The regex is designed to capture metrics with optional state and age parameters
// The regex is supposed to match against the following patterns:
// [metric(age){state}]
// [metric(age)]
// [metric{state}]
// [metric]
// This results in 3 groups:
// 1. The metric name
// 2. The age (optional)
// 3. The state (optional)
var metricRegex = regexp.MustCompile(`\[([^\[\]{}()]+)(?:\(([^()]+)\))?(?:{([^{}]+)})?\]`)

func filterMetricsFromFormula(formula string) map[string]bool {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	res := make(map[string]bool, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			// Extract just the metric name (without state or age)
			metric := match[1]
			res[metric] = true
		}
	}
	return res
}

// Extract states from formula for each metric
// If a metric doesn't have an explicit state in the formula, use the defaultState
/*func extractStatesFromFormula(formula string, defaultState string) map[string]bool {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	metricStates := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		metric := match[1]

		// Initialize state map for this metric if not already present
		if _, exists := metricStates[metric]; !exists {
			metricStates[metric] = make(map[string]bool)
		}

		// If custom state is specified in formula, use it
		if len(match) > 3 && match[3] != "" {
			customState := match[3]
			metricStates[metric][customState] = true
		} else {
			// Otherwise use all default states
			for state := range defaultStates {
				metricStates[metric][state] = true
			}
		}
	}

	return metricStates
}*/

// Sanitize formula by adding default states where missing
func sanitizeFormula(formula string, defaultState string) string {
	return metricRegex.ReplaceAllStringFunc(formula, func(match string) string {
		submatches := metricRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		metric := submatches[1]
		hasAge := len(submatches) > 2 && submatches[2] != ""
		hasState := len(submatches) > 3 && submatches[3] != ""

		ageStr := ""
		if hasAge {
			ageStr = "(" + submatches[2] + ")"
		} else {
			ageStr = "(0)" // Add default age of 0 if no age is specified
		}

		if !hasState {
			return "[" + metric + ageStr + "{" + defaultState + "}" + "]"
		}

		// If it has a state but no age, add the default age
		if hasState && !hasAge {
			return "[" + metric + ageStr + "{" + submatches[3] + "}" + "]"
		}

		return match
	})
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
func NewCalculationContractsFromProto(service string, reqs []*pb.CalculationRequest) []CalculationContract {
	res := make([]CalculationContract, 0)
	for _, req := range reqs {
		for _, state := range req.States {
			res = append(res, CalculationContract{
				Formula:   req.Formula,
				Service:   service,
				State:     state,
				arguments: getCalculationArguments(sanitizeFormula(req.Formula, state)),
			})
		}
	}
	return res
}

func getCalculationArguments(formula string) []CalculationArgument {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	res := make([]CalculationArgument, 0, len(matches))

	for _, match := range matches {
		age, _ := strconv.Atoi(match[2])
		res = append(res, CalculationArgument{
			Metric: match[1],
			Age:    age,
			State:  match[3],
		})
	}
	return res
}
