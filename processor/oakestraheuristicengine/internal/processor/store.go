package processor

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
)

// store implements interfaces.TreeStore
type store struct {
	processors map[string]interfaces.Processor
}

func NewStore() interfaces.ProcessorStore {
	return &store{
		processors: make(map[string]interfaces.Processor),
	}
}

func (s *store) Get(identifier string) interfaces.Processor {
	return s.processors[identifier]
}

func (s *store) Add(processor interfaces.Processor) error {
	if _, ok := s.processors[processor.Identifier()]; ok {
		return fmt.Errorf("processor with identifier %s already exists", processor.Identifier())
	}
	s.processors[processor.Identifier()] = processor
	return nil
}

func (s *store) GetAll() map[string]interfaces.Processor {
	return s.processors
}
