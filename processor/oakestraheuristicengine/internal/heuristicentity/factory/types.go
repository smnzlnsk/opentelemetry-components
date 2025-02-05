package factory

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/interfaces"

type HeuristicType string

const (
	RoutingEntity HeuristicType = "routing"
)

type HeuristicEntityFactory interface {
	CreateHeuristicEntity(heuristicType HeuristicType) (interfaces.HeuristicEntity, error)
}
