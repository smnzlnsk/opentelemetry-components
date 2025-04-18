package alert

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"

type alertNotifierBuilder struct {
	host     string
	port     int
	endpoint string
}

func NewAlertNotifierBuilder() domain.NotificationInterfaceBuilder {
	return &alertNotifierBuilder{}
}

func (b *alertNotifierBuilder) WithHost(host string) domain.NotificationInterfaceBuilder {
	b.host = host
	return b
}

func (b *alertNotifierBuilder) WithPort(port int) domain.NotificationInterfaceBuilder {
	b.port = port
	return b
}

func (b *alertNotifierBuilder) WithEndpoint(endpoint string) domain.NotificationInterfaceBuilder {
	b.endpoint = endpoint
	return b
}

func (b *alertNotifierBuilder) WithCapability(capability domain.NotificationInterfaceCapability) domain.NotificationInterfaceBuilder {
	return b
}

func (b *alertNotifierBuilder) Build() domain.NotificationInterface {
	return &alertNotifier{
		host:       b.host,
		port:       b.port,
		endpoint:   b.endpoint,
		capability: domain.NotificationInterfaceCapability_Alert,
	}
}
