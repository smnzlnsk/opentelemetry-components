package applicationprocessor

import (
	datapoint "github.com/smnzlnsk/opentelemetry-components/internal/shared/metric/datapoint"
)

// FormulaToMetricMap is a map of service assigned formula to metric name
type FormulaToMetricMap struct {
	mapping map[FormulaKey]datapoint.Metadata
}

// NewFormulaToMetricMap creates a new FormulaToMetricMap
func NewFormulaToMetricMap() *FormulaToMetricMap {
	return &FormulaToMetricMap{
		mapping: make(map[FormulaKey]datapoint.Metadata),
	}
}

func (ftmp *FormulaToMetricMap) GetMetricName(service string, formula string) datapoint.Metadata {
	key := FormulaKey{Service: service, Formula: formula}
	if metadata, exists := ftmp.mapping[key]; exists {
		return metadata
	}
	return datapoint.Metadata{}
}

func (ftmp *FormulaToMetricMap) AddMetric(service string, formula string, metricName string, metricUnit string) {
	key := FormulaKey{Service: service, Formula: formula}
	ftmp.mapping[key] = datapoint.Metadata{
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
