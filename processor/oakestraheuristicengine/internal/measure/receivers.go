package measure

type MeasureReceiver interface {
	ExecuteMeasure(measures []Measure) error
}

type measureReceiver struct {
	name string
}

func NewMeasureReceiver(name string) MeasureReceiver {
	return &measureReceiver{
		name: name,
	}
}

func (m *measureReceiver) ExecuteMeasure(measures []Measure) error {
	// TODO: implement measure execution
	return nil
}
