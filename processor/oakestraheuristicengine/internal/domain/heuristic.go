package domain

type HeuristicType string

const (
	RoutingEntity HeuristicType = "routing"
)

type HeuristicEntityFactory interface {
	CreateHeuristicEntity(heuristicType HeuristicType, services Services) (HeuristicEntity, error)
}

type HeuristicEntity interface {
	Processors() map[string]Processor
	AddProcessor(processor Processor)
	Evaluate(processorIdentifier string, arguments ...interface{}) EvaluationResult
	Start() error
	Shutdown() error
}
