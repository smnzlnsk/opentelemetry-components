package job

import (
	"time"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
)

// TODO: remove redundancy between appName and job_name in JobData
type Request struct {
	AppName   string       `json:"appName"`
	ServiceIP string       `json:"serviceIp,omitempty"`
	IpType    string       `json:"IpType,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	JobData   database.Job `json:"jobData"`
}
