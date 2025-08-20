package types

// Customer represents a simple customer record.
type Customer struct {
	Name string
}

// Workload describes a workload used for pricing or incentives.
type Workload struct {
	Name    string
	Service string
}

// FundingPlan represents MAP funding information.
type FundingPlan struct {
	ARR        float64
	Tier       string
	CapPercent float64
	CapAmount  float64
}
