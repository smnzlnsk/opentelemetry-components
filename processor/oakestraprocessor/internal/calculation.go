package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Define workItem for the evaluation process
type workItem struct {
	service     string
	contractKey domain.ContractKey
}

type ContractState struct {
	sync.RWMutex
	once                sync.Once
	Contracts           map[domain.ContractKey]domain.CalculationContract
	Filters             *Filter
	Datapoints          map[domain.DatapointKey]domain.MetricDatapoint
	compiledExpressions map[domain.ContractKey]*govaluate.EvaluableExpression
}

func NewContractState() *ContractState {
	return &ContractState{
		Contracts:           make(map[domain.ContractKey]domain.CalculationContract),
		Filters:             newFilter(),
		Datapoints:          make(map[domain.DatapointKey]domain.MetricDatapoint),
		compiledExpressions: make(map[domain.ContractKey]*govaluate.EvaluableExpression),
	}
}

func (c *ContractState) Sync() {
	c.once.Do(func() {
		// We'll leave it as a placeholder for now since we don't have the implementation
		// TODO: Implement contract population logic if needed
	})
}

func (c *ContractState) GetDefaultContracts() map[string]domain.CalculationContract {
	c.RLock()
	defer c.RUnlock()

	res := make(map[string]domain.CalculationContract)
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

	key := domain.ContractKey{
		Service: "default",
		Formula: formula,
	}

	contract := domain.CalculationContract{
		Formula: formula,
		Service: "default",
		States:  states,
		Metrics: filterMetricsFromFormula(formula),
	}

	c.Contracts[key] = contract
	c.compiledExpressions[key] = expr

	return nil
}

func (c *ContractState) RegisterService(service string, contracts map[string]domain.CalculationContract, normalizationValue string) error {
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
			serviceKey := domain.ContractKey{
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
	for _, contract := range contracts {
		key := domain.ContractKey{Service: service, Formula: contract.Formula}

		normalisedFormula := fmt.Sprintf("(%s) / %f", contract.Formula, normValue)
		fmt.Printf("normalisedFormula: %s\n", normalisedFormula)
		expr, err := govaluate.NewEvaluableExpression(normalisedFormula)
		if err != nil {
			return fmt.Errorf("invalid formula %s: %w", contract.Formula, err)
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

				// Check if metric is registered
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

					key := domain.DatapointKey{
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

func (c *ContractState) Evaluate() domain.CalculationResults {
	c.RLock()
	defer c.RUnlock()

	// Skip default contracts in evaluation
	serviceContracts := make(map[string][]domain.ContractKey)
	for key := range c.Contracts {
		if key.Service != "default" {
			serviceContracts[key.Service] = append(serviceContracts[key.Service], key)
		}
	}

	res := make(domain.CalculationResults)
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

					resultKey := domain.CalculationResultKey{
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

func (c *ContractState) GetParameters(cc domain.CalculationContract) domain.CalculationParameters {
	res := make(map[string]map[string]interface{})

	for state := range cc.States {
		res[state] = make(map[string]interface{})
		for metric := range cc.Metrics {
			// Reset service name for system metrics
			serviceForLookup := cc.Service
			if strings.HasPrefix(metric, "system.") {
				serviceForLookup = ""
			}

			key := domain.DatapointKey{
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

// CreateMetricDatapoint creates a MetricDatapoint from an OpenTelemetry metric
func CreateMetricDatapoint(metric pmetric.Metric, idx int) domain.MetricDatapoint {
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
	md := domain.MetricDatapoint{
		Metadata: domain.MetricMetadata{
			MetricType: metric.Type().String(),
			MetricName: metric.Name(),
			MetricUnit: metric.Unit(),
			Attributes: metric.Metadata(),
		},
		Value: domain.Datapoint{
			ValueDataType: ndp.ValueType().String(),
			FloatValue:    value,
		},
	}
	return md
}

// NewCalculationContractsFromProto creates CalculationContracts from protobuf requests
func NewCalculationContractsFromProto(service string, reqs []*pb.CalculationRequest) map[string]domain.CalculationContract {
	res := make(map[string]domain.CalculationContract, len(reqs))
	for _, req := range reqs {
		states := make(map[string]bool, len(req.States))
		for _, state := range req.States {
			states[state] = true
		}

		res[req.Formula] = domain.CalculationContract{
			Formula: req.Formula,
			Service: service,
			States:  states,
			Metrics: filterMetricsFromFormula(req.Formula),
		}
	}
	return res
}
