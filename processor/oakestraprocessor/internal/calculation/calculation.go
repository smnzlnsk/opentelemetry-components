package calculation

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Calculation struct {
	Service           string
	AtomicCalculation map[string]map[string]*MetricDatapoint
	Metrics           map[string]pmetric.Metric // for debugging
}

func NewCalculation(filter internal.Filter) *Calculation {
	ca := &Calculation{
		AtomicCalculation: make(map[string]map[string]*MetricDatapoint),
		Metrics:           make(map[string]pmetric.Metric), // for debugging
	}
	for state, active := range filter.StateFilter {
		if _, exists := ca.AtomicCalculation[state]; !exists && active {
			ca.AtomicCalculation[state] = make(map[string]*MetricDatapoint)
		}
		for metric, ok := range filter.MetricFilter {
			if ok {
				ca.AtomicCalculation[state][metric] = &MetricDatapoint{}
			}
		}
	}

	// for debugging
	for metric, _ := range filter.MetricFilter {
		ca.Metrics[metric] = pmetric.NewMetric()
	}
	return ca
}

func (c *Calculation) SetValue(state string, metric string, v *MetricDatapoint) {
	c.AtomicCalculation[state][metric] = v
}
