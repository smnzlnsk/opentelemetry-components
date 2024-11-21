package internal

import (
	"github.com/Knetic/govaluate"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Calculation struct {
	Service           string
	Formula           string
	AtomicCalculation map[string]CalculationParameters
	Metrics           map[string]pmetric.Metric // for debugging
}

type CalculationParameters map[string]*MetricDatapoint

func NewCalculation(formula string, filter Filter) *Calculation {
	ca := &Calculation{
		Formula:           formula,
		AtomicCalculation: make(map[string]CalculationParameters),
		Metrics:           make(map[string]pmetric.Metric), // for debugging
	}
	for state, active := range filter.StateFilter {
		// add default
		ca.AtomicCalculation["default"] = make(map[string]*MetricDatapoint)
		// add custom
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

func (c *Calculation) EvaluateFormula() map[string]interface{} {
	expr, err := govaluate.NewEvaluableExpression(c.Formula)
	if err != nil {
		return nil
	}
	res := make(map[string]interface{})
	for state, metric := range c.AtomicCalculation {
		params := metric.parse()
		res[state], err = expr.Evaluate(params)
	}
	return res
}

func (cp CalculationParameters) parse() map[string]interface{} {
	res := make(map[string]interface{})
	for x, y := range cp {
		res[x] = y.Value.FloatValue
	}
	return res
}
