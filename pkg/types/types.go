package types

// Customer represents a customer account.
type Customer struct {
	Name string `json:"name"`
}

// Workload represents a workload for pricing or incentives.
type Workload struct {
	Name    string             `json:"name"`
	Service string             `json:"service"`
	Region  string             `json:"region"`
	Usage   map[string]float64 `json:"usage"`
}

// ARRTier is the annual recurring revenue tier.
type ARRTier int

const (
	ARR205k ARRTier = 205000
	ARR300k ARRTier = 300000
	ARR600k ARRTier = 600000
)

// FundingPlan represents the funding for an incentive request.
type FundingPlan struct {
	ARR       ARRTier
	Cap       float64
	Requested float64
	Narrative string
}
