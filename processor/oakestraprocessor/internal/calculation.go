package internal

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
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

func (c *ContractState) RegisterService(service string, contracts map[string]CalculationContract) error {
	c.Lock()
	defer c.Unlock()

	if service == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if contracts == nil {
		return fmt.Errorf("contracts map cannot be nil")
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

			expr, err := govaluate.NewEvaluableExpression(key.Formula)
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

		expr, err := govaluate.NewEvaluableExpression(formula)
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
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		var serviceName string
		containerMetric := false
		rm := metrics.ResourceMetrics().At(i)
		rmAttr := rm.Resource().Attributes()

		// Check if metrics bundle is associated with a service
		sn, ok := rmAttr.Get("container_id")
		_, okk := rmAttr.Get("namespace")
		if ok && okk {
			containerMetric = true
			serviceName = sn.Str()

			// Check if service is registered
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

				if containerMetric && !strings.HasPrefix(mmetric.Name(), "container.") {
					continue
				}

				metricFilter, ok := c.Filters.MetricFilters[mmetric.Name()]
				if !ok {
					continue
				}

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
						Service: serviceName, // empty for system metrics
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

// CalculationResults map[service][formula][state]result
type CalculationResults map[string]map[string]map[string]float64

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

	res := make(CalculationResults, len(serviceContracts))
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		service string
		results map[string]map[string]float64
	}, len(serviceContracts))

	for service, contractKeys := range serviceContracts {
		wg.Add(1)
		go func(service string, contractKeys []ContractKey) {
			defer wg.Done()
			serviceResults := make(map[string]map[string]float64)

			for _, key := range contractKeys {
				contract := c.Contracts[key]
				expr := c.compiledExpressions[key]
				if expr == nil {
					continue
				}

				params := c.GetParameters(contract)
				stateResults := make(map[string]float64, len(params))

				for state, cp := range params {
					if result, err := expr.Evaluate(cp); err == nil {
						stateResults[state] = result.(float64)
					}
				}
				serviceResults[key.Formula] = stateResults
			}

			resultChan <- struct {
				service string
				results map[string]map[string]float64
			}{service, serviceResults}
		}(service, contractKeys)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		res[result.service] = result.results
	}

	return res
}

func (c *ContractState) GetParameters(cc CalculationContract) CalculationParameters {
	res := make(map[string]map[string]interface{})

	for state := range cc.States {
		res[state] = make(map[string]interface{})
		for metric := range cc.Metrics {
			key := DatapointKey{
				Service: cc.Service,
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
