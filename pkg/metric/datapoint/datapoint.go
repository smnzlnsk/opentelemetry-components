package metric

import (
	"fmt"
	"time"
)

type Key struct {
	Service string // empty for system metrics
	Metric  string
	State   string
}

func (k Key) String() string {
	return fmt.Sprintf("DatapointKey{Service: %s, Metric: %s, State: %s}", k.Service, k.Metric, k.State)
}

// Datapoint represents a single datapoint with its metadata
type Datapoint struct {
	Metadata Metadata `json:"metadata,omitempty"`
	Value    float64  `json:"value"`
}

// Metadata contains metadata about a metric datapoint
type Metadata struct {
	MetricType  string      `json:"type,omitempty"`
	MetricState string      `json:"state,omitempty"`
	MetricName  string      `json:"name,omitempty"`
	MetricUnit  string      `json:"unit,omitempty"`
	Attributes  interface{} `json:"attributes,omitempty"`
	Timestamp   time.Time   `json:"timestamp,omitempty"`
}
