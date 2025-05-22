package domain

import "fmt"

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
	Processor string // The processor this contract belongs to
	Formula   string // The formula to evaluate
	Service   string // The service this contract belongs to
	State     string // The output state of the contract
	// Metrics   map[string]bool       // Metrics derived from formula for later metric filtering
	arguments []CalculationArgument // Metrics and states derived from formula for later metric filtering
}

type CalculationArgument struct {
	Metric string
	State  string
	Age    int
}

type DatapointKey struct {
	Service string // empty for system metrics
	Metric  string
	State   string
}

func (c CalculationContract) String() string {
	return fmt.Sprintf(
		"\nProcessor: %v\n"+
			"Formula: %s\n"+
			"Service: %s\n"+
			"State: %s\n"+
			"Arguments: %v\n",
		c.Processor, c.Formula, c.Service, c.State, c.arguments)
}
