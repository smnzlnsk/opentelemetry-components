package domain

import "time"

type Job struct {
	JobName             string                     `json:"job_name"`
	ServiceInstanceList []ServiceInstanceListEntry `json:"instance_list"`
}

type ServiceInstanceListEntry struct {
	InstanceNumber int     `json:"instance_number"`
	Priority       float64 `json:"priority"`
}

// TODO: remove redundancy between appName and job_name in JobData
type JobRequest struct {
	AppName   string    `json:"appName"`
	ServiceIP string    `json:"serviceIp"`
	IpType    string    `json:"IpType,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	JobData   Job       `json:"jobData"`
}
