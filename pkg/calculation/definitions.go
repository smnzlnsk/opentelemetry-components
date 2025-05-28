package calculation

import "fmt"

// CalculationResultKey is a key for calculation results
type ResultKey struct {
	Service string
	Formula string
	State   string
}

// CalculationResults maps result keys to calculated values
type Results map[ResultKey]float64

// CalculationParameters map[metric(age){state}]metricValue
type Parameters map[string]interface{}

type Argument struct {
	Metric string
	State  string
	Age    int
}

// Contract represents a contract for calculating metrics based on a formula
type Contract struct {
	Processor string // The processor this contract belongs to
	Formula   string // The formula to evaluate
	Service   string // The service this contract belongs to
	State     string // The output state of the contract
	// Metrics   map[string]bool       // Metrics derived from formula for later metric filtering
	Arguments []Argument // Metrics and states derived from formula for later metric filtering
}

func (c Contract) String() string {
	return fmt.Sprintf(
		"\nProcessor: %v\n"+
			"Formula: %s\n"+
			"Service: %s\n"+
			"State: %s\n"+
			"Arguments: %v\n",
		c.Processor, c.Formula, c.Service, c.State, c.Arguments)
}
