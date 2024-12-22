package internal

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"regexp"
	"strings"
)

type ContractState struct {
	Contracts        map[string]map[string]CalculationContract // currently active contracts; map[service][formula]CalculcationContract
	DefaultContracts map[string]CalculationContract
	Filters          *Filter // filter metrics needed by contracts
	// read: [metric][state]MetricDatapoint
	SystemDatapoints map[string]map[string]MetricDatapoint
	// read: [service][metric][state]MetricDatapoint
	ContainerDatapoints map[string]map[string]map[string]MetricDatapoint
}

func NewContractState() *ContractState {
	return &ContractState{
		Contracts:           make(map[string]map[string]CalculationContract),
		DefaultContracts:    make(map[string]CalculationContract),
		Filters:             newFilter(),
		SystemDatapoints:    make(map[string]map[string]MetricDatapoint),
		ContainerDatapoints: make(map[string]map[string]map[string]MetricDatapoint),
	}
}

func (c *ContractState) RegisterService(service string, contracts map[string]CalculationContract) error {
	if c.Contracts[service] != nil {
		return fmt.Errorf("service %s already registered", service)
	}
	c.ContainerDatapoints[service] = make(map[string]map[string]MetricDatapoint)
	c.Contracts[service] = contracts
	for formula, contract := range c.DefaultContracts {
		// add default contracts to service to be registered
		// override service name
		contract.Service = service
		c.Contracts[service] = make(map[string]CalculationContract)
		c.Contracts[service][formula] = contract
	}
	// add filters for all contracts
	for formula, contract := range c.Contracts[service] {
		fmt.Printf("adding contract for service %s with formula %s:\n", service, formula)
		neededMetrics := filterMetricsFromFormula(formula)
		fmt.Printf("adding metrics to filter: %v\n", neededMetrics)
		for metric, _ := range neededMetrics {

			err := c.Filters.AddMetricFilter(metric, contract.States)
			if err != nil {
				return err
			}
			c.ContainerDatapoints[service][metric] = make(map[string]MetricDatapoint)
		}
	}
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
			// fmt.Printf("Found container_id: %s\n", serviceName)
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
					// fmt.Printf("ommiting container metric: %s\n", mmetric.Name())
					continue
				}

				// guard clause, check that metric is set in filter and active
				metricFilter, ok := c.Filters.MetricFilters[mmetric.Name()]
				if !ok {
					continue
				}
				//fmt.Printf("found metric %s with stateFilter: %v\n", mmetric.Name(), metricFilter.StateFilter)

				if c.SystemDatapoints[mmetric.Name()] == nil {
					//fmt.Printf("intialising SystemDatapoints map for metric %s\n", mmetric.Name())
					c.SystemDatapoints[mmetric.Name()] = make(map[string]MetricDatapoint)
				}

				statesPresent := mmetric.Sum().DataPoints().Len() > 1
				// verify we have states present in the filter
				if statesPresent {
					if metricFilter.StateFilter == nil {
						//fmt.Printf("found more than one datapoint in metric: %s, while stateFilter is not set", mmetric.Name())
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
							//fmt.Printf("setting system metric %s and default state with value %f\n", mmetric.Name(), mdp.Value.FloatValue)
							c.SystemDatapoints[mmetric.Name()]["default"] = mdp
							continue
						} else {
							if v, ok := ndp.Attributes().Get("state"); ok && c.Filters.MetricFilters[mmetric.Name()].StateFilter[v.Str()] != 0 {
								//fmt.Printf("setting system metric %s with state %s and value %f\n", mmetric.Name(), v.Str(), mdp.Value.FloatValue)
								c.SystemDatapoints[mmetric.Name()][v.Str()] = mdp
								continue
							}
						}
					} else {
						// container metric
						if !statesPresent {
							//fmt.Printf("setting container metric %s and default state for service %s with value %f\n", mmetric.Name(), serviceName, mdp.Value.FloatValue)
							c.ContainerDatapoints[serviceName][mmetric.Name()]["default"] = mdp
							continue
						} else {
							if v, ok := ndp.Attributes().Get("state"); ok && c.Filters.MetricFilters[mmetric.Name()].StateFilter[v.Str()] != 0 {
								//fmt.Printf("setting container metric %s and state %s for service %s with value %f\n", mmetric.Name(), v.Str(), serviceName, mdp.Value.FloatValue)
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

func (c *ContractState) Evaluate() (res CalculationResults) {
	res = make(CalculationResults)
	for service, formulae := range c.Contracts {
		res[service] = make(map[string]map[string]float64)
		for formula, contract := range formulae {
			res[service][formula] = make(map[string]float64)
			expression, err := govaluate.NewEvaluableExpression(formula)
			if err != nil {
				continue
			}
			params := c.GetParameters(contract)
			for state, cp := range params {
				result, err := expression.Evaluate(cp)
				if err != nil {
					fmt.Println(err)
				}
				res[service][formula][state] = result.(float64)
			}
		}
	}
	return
}

func (c *ContractState) GetParameters(cc CalculationContract) (res CalculationParameters) {
	res = make(map[string]map[string]interface{})
	for state, _ := range cc.States {
		res[state] = make(map[string]interface{})
		for metric, _ := range cc.Metrics {
			if strings.HasPrefix(metric, "container.") {
				// fetch from containerMetrics
				res[state][metric] = c.ContainerDatapoints[cc.Service][metric][state].Value.FloatValue
			} else {
				res[state][metric] = c.SystemDatapoints[metric][state].Value.FloatValue
			}
		}
	}
	// fmt.Println("returning parameters:", res)
	return
}

// extract all necessary metrics from formula as a map to filter in the future
func filterMetricsFromFormula(formula string) map[string]bool {
	res := make(map[string]bool)
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindAllStringSubmatch(formula, -1)
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
