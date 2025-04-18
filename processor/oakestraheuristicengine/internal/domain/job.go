package domain

type Job struct {
	JobName             string                     `json:"job_name"`
	ServiceInstanceList []ServiceInstanceListEntry `json:"instance_list"`
}

type ServiceInstanceListEntry struct {
	InstanceNumber int     `json:"instance_number"`
	Priority       float64 `json:"priority"`
}
