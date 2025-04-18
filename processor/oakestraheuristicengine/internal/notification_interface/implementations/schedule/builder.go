package schedule

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type scheduleNotifierBuilder struct {
	host     string
	port     int
	endpoint string
}

func NewScheduleNotifierBuilder() domain.NotificationInterfaceBuilder {
	return &scheduleNotifierBuilder{}
}

func (s *scheduleNotifierBuilder) WithHost(host string) domain.NotificationInterfaceBuilder {
	s.host = host
	return s
}

func (s *scheduleNotifierBuilder) WithPort(port int) domain.NotificationInterfaceBuilder {
	s.port = port
	return s
}

func (s *scheduleNotifierBuilder) WithEndpoint(endpoint string) domain.NotificationInterfaceBuilder {
	s.endpoint = endpoint
	return s
}

func (s *scheduleNotifierBuilder) WithCapability(capability domain.NotificationInterfaceCapability) domain.NotificationInterfaceBuilder {
	return s
}

func (s *scheduleNotifierBuilder) Build() domain.NotificationInterface {
	return &scheduleNotifier{
		host:       s.host,
		port:       s.port,
		endpoint:   s.endpoint,
		capability: domain.NotificationInterfaceCapability_Schedule,
	}
}
