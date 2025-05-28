package evaluation

type Evaluation struct {
	JobName string
	Entries []Entry
}

type Entry struct {
	InstanceNumber int     `json:"instance_number"`
	Priority       float64 `json:"priority"`
	IpType         string  `json:"IpType"`
}

type Result struct {
	JobName string                            `json:"job_name"`
	Values  map[string]map[string]interface{} `json:"values,omitempty"`
	Results []Entry                           `json:"results"`
}
