package applicationprocessor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
)

// FormulaToMetricMap is a map of service assigned formula to metric name
type FormulaToMetricMap struct {
	mapping map[FormulaKey]internal.MetricMetadata
}

// NewFormulaToMetricMap creates a new FormulaToMetricMap
func NewFormulaToMetricMap() *FormulaToMetricMap {
	return &FormulaToMetricMap{
		mapping: make(map[FormulaKey]internal.MetricMetadata),
	}
}

func (ftmp *FormulaToMetricMap) GetMetricName(service string, formula string) internal.MetricMetadata {
	key := FormulaKey{Service: service, Formula: formula}
	if metadata, exists := ftmp.mapping[key]; exists {
		return metadata
	}
	return internal.MetricMetadata{}
}

func (ftmp *FormulaToMetricMap) AddMetric(service string, formula string, metricName string, metricUnit string) {
	key := FormulaKey{Service: service, Formula: formula}
	ftmp.mapping[key] = internal.MetricMetadata{
		MetricName: metricName,
		MetricUnit: metricUnit,
	}
}

func (ftmp *FormulaToMetricMap) DeleteMetric(service string) {
	for key := range ftmp.mapping {
		if key.Service == service {
			delete(ftmp.mapping, key)
		}
	}
}

// FormulaKey is a key for the FormulaToMetricMap
type FormulaKey struct {
	Service string
	Formula string
}
