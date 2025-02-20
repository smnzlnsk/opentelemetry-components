package alert

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"

type alertNotifierBuilder struct {
	host     string
	port     int
	endpoint string
}

func NewAlertNotifierBuilder() interfaces.NotificationInterfaceBuilder {
	return &alertNotifierBuilder{}
}

func (b *alertNotifierBuilder) WithHost(host string) interfaces.NotificationInterfaceBuilder {
	b.host = host
	return b
}

func (b *alertNotifierBuilder) WithPort(port int) interfaces.NotificationInterfaceBuilder {
	b.port = port
	return b
}

func (b *alertNotifierBuilder) WithEndpoint(endpoint string) interfaces.NotificationInterfaceBuilder {
	b.endpoint = endpoint
	return b
}

func (b *alertNotifierBuilder) Build() interfaces.NotificationInterface {
	return &alertNotifier{
		host:     b.host,
		port:     b.port,
		endpoint: b.endpoint,
	}
}
