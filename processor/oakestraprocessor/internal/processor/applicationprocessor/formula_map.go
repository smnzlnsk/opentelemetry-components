package applicationprocessor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
)

// FormulaToMetricMap is a map of service assigned formula to metric name
type FormulaToMetricMap struct {
	mapping map[FormulaKey]domain.DatapointMetadata
}

// NewFormulaToMetricMap creates a new FormulaToMetricMap
func NewFormulaToMetricMap() *FormulaToMetricMap {
	return &FormulaToMetricMap{
		mapping: make(map[FormulaKey]domain.DatapointMetadata),
	}
}

func (ftmp *FormulaToMetricMap) GetMetricName(service string, formula string) domain.DatapointMetadata {
	key := FormulaKey{Service: service, Formula: formula}
	if metadata, exists := ftmp.mapping[key]; exists {
		return metadata
	}
	return domain.DatapointMetadata{}
}

func (ftmp *FormulaToMetricMap) AddMetric(service string, formula string, metricName string, metricUnit string) {
	key := FormulaKey{Service: service, Formula: formula}
	ftmp.mapping[key] = domain.DatapointMetadata{
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
