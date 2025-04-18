package domain

type HeuristicType string

const (
	RoutingEntity HeuristicType = "routing"
)

type HeuristicEntityFactory interface {
	CreateHeuristicEntity(heuristicType HeuristicType) (HeuristicEntity, error)
}

type HeuristicEntity interface {
	Processors() map[string]Processor
	AddProcessor(processor Processor)
	Evaluate(processorIdentifier string, appname string, values map[string]interface{}) Evaluation
	Start() error
	Shutdown() error
}
