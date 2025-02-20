package route

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"

type routeNotifierBuilder struct {
	host     string
	port     int
	endpoint string
}

func NewRouteNotifierBuilder() interfaces.NotificationInterfaceBuilder {
	return &routeNotifierBuilder{}
}

func (r *routeNotifierBuilder) WithHost(host string) interfaces.NotificationInterfaceBuilder {
	r.host = host
	return r
}

func (r *routeNotifierBuilder) WithPort(port int) interfaces.NotificationInterfaceBuilder {
	r.port = port
	return r
}

func (r *routeNotifierBuilder) WithEndpoint(endpoint string) interfaces.NotificationInterfaceBuilder {
	r.endpoint = endpoint
	return r
}

func (r *routeNotifierBuilder) Build() interfaces.NotificationInterface {
	return &routeNotifier{
		host:     r.host,
		port:     r.port,
		endpoint: r.endpoint,
	}
}
