package domain

type Evaluator interface {
	Evaluate(initalValue float64, params map[string]interface{}) float64
}

type Evaluation struct {
	JobName string
	Entries []EvaluationEntry
}

type EvaluationEntry struct {
	InstanceNumber int     `json:"instance_number"`
	Priority       float64 `json:"priority"`
	IpType         string  `json:"IpType"`
}

type EvaluationResult struct {
	JobName string                            `json:"job_name"`
	Values  map[string]map[string]interface{} `json:"values,omitempty"`
	Results []EvaluationEntry                 `json:"results"`
}
