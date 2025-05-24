package domain

type PolicyBuilder interface {
	WithName(name string) PolicyBuilder
	WithRouteInterface(measure NotificationInterface[any]) PolicyBuilder
	WithAlertInterface(measure NotificationInterface[any]) PolicyBuilder
	WithScheduleInterface(measure NotificationInterface[any]) PolicyBuilder
	WithHeuristicEntity(entity HeuristicEntity) PolicyBuilder
	NotificationInterfaceBuilder() NotificationInterfaceBuilder[any]
	Build() Policy
}

type Policy interface {
	Enforce(processorIdentifier string, arguments ...interface{}) error
	Name() string
	HeuristicEntity() HeuristicEntity
}
