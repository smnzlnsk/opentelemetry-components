package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
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

// CalculationParameters map[metric(age){state}]metricValue
type CalculationParameters map[string]interface{}

type DatapointKey struct {
	Service string // empty for system metrics
	Metric  string
	State   string
}

func (dk DatapointKey) String() string {
	return fmt.Sprintf("DatapointKey{Service: %s, Metric: %s, State: %s}", dk.Service, dk.Metric, dk.State)
}

// Datapoint represents a single datapoint with its metadata
type Datapoint struct {
	Metadata DatapointMetadata `json:"metadata,omitempty"`
	Value    float64           `json:"value"`
}

// MetricMetadata contains metadata about a metric
type DatapointMetadata struct {
	MetricType  string      `json:"type,omitempty"`
	MetricState string      `json:"state,omitempty"`
	MetricName  string      `json:"name,omitempty"`
	MetricUnit  string      `json:"unit,omitempty"`
	Attributes  interface{} `json:"attributes,omitempty"`
	Timestamp   time.Time   `json:"timestamp,omitempty"`
}

// CalculationResults maps result keys to calculated values
type CalculationResults map[CalculationResultKey]float64

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

// NewCalculationContractsFromProto creates CalculationContracts from protobuf requests
func NewCalculationContractsFromProto(service string, reqs []*pb.CalculationRequest) []CalculationContract {
	res := make([]CalculationContract, 0)
	for _, req := range reqs {
		for _, state := range req.States {
			res = append(res, CalculationContract{
				Formula:   req.Formula,
				Service:   service,
				State:     state,
				Arguments: getCalculationArguments(sanitizeFormula(req.Formula, state)),
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
