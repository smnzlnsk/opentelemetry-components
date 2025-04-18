package route

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"

type routeNotifierBuilder struct {
	host     string
	port     int
	endpoint string
}

func NewRouteNotifierBuilder() domain.NotificationInterfaceBuilder {
	return &routeNotifierBuilder{}
}

func (r *routeNotifierBuilder) WithHost(host string) domain.NotificationInterfaceBuilder {
	r.host = host
	return r
}

func (r *routeNotifierBuilder) WithPort(port int) domain.NotificationInterfaceBuilder {
	r.port = port
	return r
}

func (r *routeNotifierBuilder) WithEndpoint(endpoint string) domain.NotificationInterfaceBuilder {
	r.endpoint = endpoint
	return r
}

func (r *routeNotifierBuilder) WithCapability(capability domain.NotificationInterfaceCapability) domain.NotificationInterfaceBuilder {
	return r
}

func (r *routeNotifierBuilder) Build() domain.NotificationInterface {
	return &routeNotifier{
		host:       r.host,
		port:       r.port,
		endpoint:   r.endpoint,
		capability: domain.NotificationInterfaceCapability_Route,
	}
}
