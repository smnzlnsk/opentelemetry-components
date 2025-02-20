package schedule

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"

type scheduleNotifierBuilder struct {
	host     string
	port     int
	endpoint string
}

func NewScheduleNotifierBuilder() interfaces.NotificationInterfaceBuilder {
	return &scheduleNotifierBuilder{}
}

func (s *scheduleNotifierBuilder) WithHost(host string) interfaces.NotificationInterfaceBuilder {
	s.host = host
	return s
}

func (s *scheduleNotifierBuilder) WithPort(port int) interfaces.NotificationInterfaceBuilder {
	s.port = port
	return s
}

func (s *scheduleNotifierBuilder) WithEndpoint(endpoint string) interfaces.NotificationInterfaceBuilder {
	s.endpoint = endpoint
	return s
}

func (s *scheduleNotifierBuilder) Build() interfaces.NotificationInterface {
	return &scheduleNotifier{
		host:     s.host,
		port:     s.port,
		endpoint: s.endpoint,
	}
}
