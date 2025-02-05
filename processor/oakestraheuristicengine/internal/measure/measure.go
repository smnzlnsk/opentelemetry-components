package measure

type Measure interface {
	GetID() string
	GetValue() float64
}

type measure struct {
	ID     string
	Value  float64
	Action func() error
}

func NewMeasure(id string, value float64) Measure {
	return &measure{
		ID:    id,
		Value: value,
	}
}

func (m *measure) GetID() string {
	return m.ID
}

func (m *measure) GetValue() float64 {
	return m.Value
}
