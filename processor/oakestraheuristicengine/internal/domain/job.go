package domain

import "time"

type Job struct {
	JobName             string                     `json:"job_name"`
	ServiceInstanceList []ServiceInstanceListEntry `json:"instance_list"`
	ServiceIPList       []ServiceIPListEntry       `json:"service_ip_list"`
}

type ServiceInstanceListEntry struct {
	InstanceNumber int `json:"instance_number"`
	Priority       int `json:"priority"`
}

type ServiceIPListEntry struct {
	IpType     string `json:"IpType"`
	Address    string `json:"Address"`
	Address_v6 string `json:"Address_v6"`
}

// TODO: remove redundancy between appName and job_name in JobData
type JobRequest struct {
	AppName   string    `json:"appName"`
	ServiceIP string    `json:"serviceIp,omitempty"`
	IpType    string    `json:"IpType,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	JobData   Job       `json:"jobData"`
}
