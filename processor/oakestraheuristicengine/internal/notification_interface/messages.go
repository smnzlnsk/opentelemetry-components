package notification_interface

type RoutingUpdate struct {
	JobName string `json:"job_name"`
	ServiceIpPriority []PriorityEntry `json:"service_ip_priority"`
}

type PriorityEntry struct {
	IpType string `json:"IpType"`
	Priority int `json:"Priority"`
}
