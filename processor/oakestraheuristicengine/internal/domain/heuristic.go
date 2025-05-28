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
	// Evaluate evaluates the heuristics
	// The arguments depend on the entity implementation
	// Therefore, the arguments are passed as a variadic argument
	Evaluate(arguments ...interface{}) error
	Start() error
	Shutdown() error
	// Set the notification interfaces on the heuristic entity
	SetAlert(NotificationInterface[any]) error
	SetRoute(NotificationInterface[any]) error
	SetSchedule(NotificationInterface[any]) error
	// Get the notification interfaces on the heuristic entity (for testing purposes)
	Alert() NotificationInterface[any]
	Route() NotificationInterface[any]
	Schedule() NotificationInterface[any]
}
