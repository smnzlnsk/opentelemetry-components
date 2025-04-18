package domain

// CalculationContract represents a contract for calculating metrics based on a formula
type CalculationContract struct {
	Formula string          // The formula to evaluate
	Service string          // The service this contract belongs to
	States  map[string]bool // States to consider (can be empty if no state has to be considered)
	Metrics map[string]bool // Metrics derived from formula for later metric filtering
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
