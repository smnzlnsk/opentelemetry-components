package internal

import "go.opentelemetry.io/collector/pdata/pmetric"

type MetricFilter interface {
	FilterByName(string, pmetric.Metrics) pmetric.Metrics
	GetFilterMetrics() []string
}

type Filter struct {
	// used to extract a set of metrics for calculations
	MetricFilter map[string]bool
	// used for calculations - matching state's get calculated with system - container
	StateFilter map[string]bool
}
