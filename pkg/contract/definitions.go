package contract

import "github.com/smnzlnsk/opentelemetry-components/pkg/calculation"

type Key struct {
	Service string
	Formula string
}

// ContractDocument represents the MongoDB document structure
type Document struct {
	Service   string                 `bson:"service"`
	Contracts []calculation.Contract `bson:"contracts"`
}
