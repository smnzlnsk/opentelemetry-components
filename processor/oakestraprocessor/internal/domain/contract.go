package domain

type ContractKey struct {
	Service string
	Formula string
}

// ContractDocument represents the MongoDB document structure
type ContractDocument struct {
	Service   string                `bson:"service"`
	Contracts []CalculationContract `bson:"contracts"`
}

// CalculationContract represents a contract for calculating metrics based on a formula
type CalculationContract struct {
	Processor string          // The processor this contract belongs to
	Formula   string          // The formula to evaluate
	Service   string          // The service this contract belongs to
	States    map[string]bool // States to consider (can be empty if no state has to be considered)
	Metrics   map[string]bool // Metrics derived from formula for later metric filtering
}

type DatapointKey struct {
	Service string // empty for system metrics
	Metric  string
	State   string
	Index   int // index of the datapoint in the slice, if it's the latest datapoint, it's 0. If none is given, it defaults to 0
}
