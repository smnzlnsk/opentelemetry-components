package processor

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// store implements domain.ProcessorStore
type store struct {
	processors map[string]domain.Processor
}

var _ domain.ProcessorStore = &store{}

func NewStore() domain.ProcessorStore {
	return &store{
		processors: make(map[string]domain.Processor),
	}
}

func (s *store) Get(identifier string) domain.Processor {
	return s.processors[identifier]
}

func (s *store) Add(processor domain.Processor) error {
	if _, ok := s.processors[processor.Identifier()]; ok {
		return fmt.Errorf("processor with identifier %s already exists", processor.Identifier())
	}
	s.processors[processor.Identifier()] = processor
	return nil
}

func (s *store) GetAll() map[string]domain.Processor {
	return s.processors
}
