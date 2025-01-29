package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type ContractKey struct {
	Service string
	Formula string
}

type DatapointKey struct {
	Service string // empty for system metrics
	Metric  string
	State   string
}

type ContractState struct {
	sync.RWMutex
	Contracts           map[ContractKey]CalculationContract
	Filters             *Filter
	Datapoints          map[DatapointKey]MetricDatapoint
	compiledExpressions map[ContractKey]*govaluate.EvaluableExpression
}

func NewContractState() *ContractState {
	return &ContractState{
		Contracts:           make(map[ContractKey]CalculationContract),
		Filters:             newFilter(),
		Datapoints:          make(map[DatapointKey]MetricDatapoint),
		compiledExpressions: make(map[ContractKey]*govaluate.EvaluableExpression),
	}
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
		Formula: formula,
		Service: "default",
		States:  states,
		Metrics: filterMetricsFromFormula(formula),
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
		if key.Service == "default" {
			serviceKey := ContractKey{
				Service: service,
				Formula: key.Formula,
			}

			serviceContract := contract
			serviceContract.Service = service
			normalisedFormula := fmt.Sprintf("(%s) / %f", key.Formula, normValue)
			fmt.Printf("normalisedFormula: %s\n", normalisedFormula)
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
	}

	// Then register service-specific contracts
	for formula, contract := range contracts {
		key := ContractKey{Service: service, Formula: formula}

		normalisedFormula := fmt.Sprintf("(%s) / %f", formula, normValue)
		fmt.Printf("normalisedFormula: %s\n", normalisedFormula)
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

				metricFilter, ok := c.Filters.MetricFilters[mmetric.Name()]
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
					c.Datapoints[key] = mdp
				}
			}
		}
	}

	return nil
}

// CalculationParameters map[state][metric]metricValue
type CalculationParameters map[string]map[string]interface{}

// Define a new key structure for flattened results
type CalculationResultKey struct {
	Service string
	Formula string
	State   string
}

// Change CalculationResults to use the flattened structure
type CalculationResults map[CalculationResultKey]float64

type workItem struct {
	service     string
	contractKey ContractKey
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
		res[state] = make(map[string]interface{})
		for metric := range cc.Metrics {
			// Reset service name for system metrics
			serviceForLookup := cc.Service
			if strings.HasPrefix(metric, "system.") {
				serviceForLookup = ""
			}

			key := DatapointKey{
				Service: serviceForLookup,
				Metric:  metric,
				State:   state,
			}
			if dp, exists := c.Datapoints[key]; exists {
				res[state][metric] = dp.Value.FloatValue
			}
		}
	}
	return res
}

// extract all necessary metrics from formula as a map to filter in the future
var metricRegex = regexp.MustCompile(`\[(.*?)\]`)

func filterMetricsFromFormula(formula string) map[string]bool {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	res := make(map[string]bool, len(matches))

	for _, metricName := range matches {
		if len(metricName) > 1 {
			res[metricName[1]] = true
		}
	}
	return res
}

type CalculationContract struct {
	Formula string
	Service string
	States  map[string]bool // can be empty, if no state has to be considered
	Metrics map[string]bool // derived from formula for later metric filtering
}

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

type MetricDatapoint struct {
	Metadata MetricMetadata
	Value    Datapoint
}

type MetricMetadata struct {
	MetricType pmetric.MetricType
	MetricName string
	MetricUnit string
	Attributes pcommon.Map
}

type Datapoint struct {
	ValueDataType pmetric.NumberDataPointValueType
	FloatValue    float64
}

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
	md := MetricDatapoint{
		Metadata: MetricMetadata{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			MetricUnit: metric.Unit(),
			Attributes: metric.Metadata(),
		},
		Value: Datapoint{
			ValueDataType: ndp.ValueType(),
			FloatValue:    value,
		},
	}
	return md
}

// Helper methods for CalculationResults
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

// Add a method to normalize calculation results
func (cr CalculationResults) Normalize(serviceNormalizationLimit float64) {
	for key, value := range cr {
		cr[key] = value / serviceNormalizationLimit // Normalize each result
	}
}
