package domain

type ContractKey struct {
	Service string
	Formula string
}

type DatapointKey struct {
	Service string // empty for system metrics
	Metric  string
	State   string
}
