package database

// This is the defintion as it is stored in the database
// The entries are managed by the cluster service manager

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
