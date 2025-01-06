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

type ContractState struct {
	sync.RWMutex
	Contracts        map[string]map[string]CalculationContract // currently active contracts; map[service][formula]CalculcationContract
	DefaultContracts map[string]CalculationContract
	Filters          *Filter // filter metrics needed by contracts
	// read: [metric][state]MetricDatapoint
	SystemDatapoints map[string]map[string]MetricDatapoint
	// read: [service][metric][state]MetricDatapoint
	ContainerDatapoints map[string]map[string]map[string]MetricDatapoint
	compiledExpressions map[string]map[string]*govaluate.EvaluableExpression // map[service][formula]*Expression
}

func NewContractState() *ContractState {
	return &ContractState{
		Contracts:           make(map[string]map[string]CalculationContract),
		DefaultContracts:    make(map[string]CalculationContract),
		Filters:             newFilter(),
		SystemDatapoints:    make(map[string]map[string]MetricDatapoint),
		ContainerDatapoints: make(map[string]map[string]map[string]MetricDatapoint),
		compiledExpressions: make(map[string]map[string]*govaluate.EvaluableExpression),
	}
}

func (c *ContractState) RegisterService(service string, contracts map[string]CalculationContract) error {
	c.Lock()
	defer c.Unlock()

	if _, exists := c.Contracts[service]; exists {
		return fmt.Errorf("service %s already registered", service)
	}

	// Pre-compile expressions for the service
	compiledExprs := make(map[string]*govaluate.EvaluableExpression)
	serviceContracts := make(map[string]CalculationContract, len(contracts)+len(c.DefaultContracts))
	c.ContainerDatapoints[service] = make(map[string]map[string]MetricDatapoint)

	// First compile default contracts
	for formula, contract := range c.DefaultContracts {
		expr, err := govaluate.NewEvaluableExpression(formula)
		if err != nil {
			return fmt.Errorf("invalid default formula %s: %w", formula, err)
		}
		compiledExprs[formula] = expr

		contractCopy := contract
		contractCopy.Service = service
		serviceContracts[formula] = contractCopy
	}

	// Then compile service-specific contracts
	for formula, contract := range contracts {
		expr, err := govaluate.NewEvaluableExpression(formula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", formula, err)
		}
		compiledExprs[formula] = expr
		serviceContracts[formula] = contract
	}

	c.compiledExpressions[service] = compiledExprs
	c.Contracts[service] = serviceContracts

	for formula, contract := range c.Contracts[service] {
		neededMetrics := filterMetricsFromFormula(formula)
		for metric := range neededMetrics {

			err := c.Filters.AddMetricFilter(metric, contract.States)
			if err != nil {
				return err
			}
			c.ContainerDatapoints[service][metric] = make(map[string]MetricDatapoint)
		}
	}
	return nil
}

func (c *ContractState) DeleteService(service string) error {
	c.Lock()
	defer c.Unlock()

	// Clean up metric filters for all contracts of this service
	if contracts, exists := c.Contracts[service]; exists {
		for formula, contract := range contracts {
			// Get metrics from formula
			neededMetrics := filterMetricsFromFormula(formula)
			// Delete filters for each metric
			for metric := range neededMetrics {
				c.Filters.DeleteMetricFilter(metric, contract.States)
			}
		}
	}

	// Delete service data
	delete(c.Contracts, service)
	delete(c.ContainerDatapoints, service)
	return nil
}

func (c *ContractState) GenerateDefaultContract(formula string, states map[string]bool) error {
	neededMetrics := filterMetricsFromFormula(formula)
	c.DefaultContracts[formula] = newCalculcationContract(formula, "", states, neededMetrics)
	return nil
}

func (c *ContractState) RemoveContract(service string) error {
	// remove all contracts associated with a service
	if service == "default" {
		return fmt.Errorf("cannot disable default calculations")
	}
	delete(c.Contracts, service)
	delete(c.ContainerDatapoints, service)
	return nil
}

func (c *ContractState) PopulateData(metrics pmetric.Metrics) error {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		var serviceName string
		containerMetric := false
		rm := metrics.ResourceMetrics().At(i)

		rmAttr := rm.Resource().Attributes()

		// check if the metrics bundle is associated with a service
		sn, ok := rmAttr.Get("container_id")
		_, okk := rmAttr.Get("namespace")
		if ok && okk {
			containerMetric = ok
			serviceName = sn.Str()
			if c.ContainerDatapoints[serviceName] == nil {
				c.ContainerDatapoints[serviceName] = make(map[string]map[string]MetricDatapoint)
			}
			if _, ok := c.Contracts[serviceName]; !ok {
				// if service is not registered, do nothing
				continue
			}
		}

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			smetric := rm.ScopeMetrics().At(j)

			for k := 0; k < smetric.Metrics().Len(); k++ {
				mmetric := smetric.Metrics().At(k)
				// if we expect metrics specific to a service, check metric name has correct prefix
				if containerMetric && !strings.HasPrefix(mmetric.Name(), "container.") {
					continue
				}

				// guard clause, check that metric is set in filter and active
				metricFilter, ok := c.Filters.MetricFilters[mmetric.Name()]
				if !ok {
					continue
				}

				if c.SystemDatapoints[mmetric.Name()] == nil {
					c.SystemDatapoints[mmetric.Name()] = make(map[string]MetricDatapoint)
				}

				statesPresent := mmetric.Sum().DataPoints().Len() > 1
				// verify we have states present in the filter
				if statesPresent {
					if metricFilter.StateFilter == nil {
						continue
					}
				}
				if c.ContainerDatapoints[serviceName][mmetric.Name()] == nil && containerMetric {
					c.ContainerDatapoints[serviceName][mmetric.Name()] = make(map[string]MetricDatapoint)
				}

				for x := 0; x < mmetric.Sum().DataPoints().Len(); x++ {
					ndp := mmetric.Sum().DataPoints().At(x)

					mdp := CreateMetricDatapoint(mmetric, x)

					if !containerMetric {
						// system metric
						if !statesPresent {
							c.SystemDatapoints[mmetric.Name()]["default"] = mdp
							continue
						} else {
							if v, ok := ndp.Attributes().Get("state"); ok && c.Filters.MetricFilters[mmetric.Name()].StateFilter[v.Str()] != 0 {
								c.SystemDatapoints[mmetric.Name()][v.Str()] = mdp
								continue
							}
						}
					} else {
						// container metric
						if !statesPresent {
							c.ContainerDatapoints[serviceName][mmetric.Name()]["default"] = mdp
							continue
						} else {
							if v, ok := ndp.Attributes().Get("state"); ok && c.Filters.MetricFilters[mmetric.Name()].StateFilter[v.Str()] != 0 {
								c.ContainerDatapoints[serviceName][mmetric.Name()][v.Str()] = mdp
								continue
							}
						}
					}
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

	res := make(CalculationResults, len(c.Contracts))
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		service string
		results map[string]map[string]float64
	}, len(c.Contracts))

	for service, formulae := range c.Contracts {
		wg.Add(1)
		go func(service string, formulae map[string]CalculationContract) {
			defer wg.Done()
			serviceResults := make(map[string]map[string]float64, len(formulae))

			compiledExprs := c.compiledExpressions[service]
			for formula, contract := range formulae {
				expr := compiledExprs[formula]
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
				serviceResults[formula] = stateResults
			}

			resultChan <- struct {
				service string
				results map[string]map[string]float64
			}{service, serviceResults}
		}(service, formulae)
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

func (c *ContractState) GetParameters(cc CalculationContract) (res CalculationParameters) {
	res = make(map[string]map[string]interface{})
	for state := range cc.States {
		res[state] = make(map[string]interface{})
		for metric := range cc.Metrics {
			if strings.HasPrefix(metric, "container.") {
				// fetch from containerMetrics
				res[state][metric] = c.ContainerDatapoints[cc.Service][metric][state].Value.FloatValue
			} else {
				res[state][metric] = c.SystemDatapoints[metric][state].Value.FloatValue
			}
		}
	}
	return
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

func newCalculcationContract(
	formula string,
	service string,
	states map[string]bool,
	metricFilter map[string]bool,
) CalculationContract {
	return CalculationContract{
		Formula: formula,
		Service: service,
		States:  states,
		Metrics: metricFilter,
	}
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
