package factory

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal"

type MeasureHandler interface {
	ExecuteMeasure(request ...*internal.EvaluationResult) error
}

type MeasureFunc func(request ...*internal.EvaluationResult) error

func (mf MeasureFunc) ExecuteMeasure(request ...*internal.EvaluationResult) error {
	return mf(request...)
}

type measureHandler struct {
	name        string
	measureFunc MeasureFunc
}

func NewMeasureHandler(name string) MeasureHandler {
	return &measureHandler{
		name: name,
		measureFunc: func(request ...*internal.EvaluationResult) error {
			return nil
		},
	}
}

func (m *measureHandler) ExecuteMeasure(request ...*internal.EvaluationResult) error {
	return m.measureFunc(request...)
}

type Option func(*measureHandler)

func WithMeasureFunc(measureFunc MeasureFunc) Option {
	return func(m *measureHandler) {
		m.measureFunc = measureFunc
	}
}
