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
	Assess     MAPPhase
	Mobilize   MAPPhase
}

// MAPPhase groups MAP tasks for a phase such as Assess or Mobilize.
type MAPPhase struct {
	Name       string
	Tasks      []MAPTask
	TotalDays  float64
	TotalWeeks float64
}

// MAPTask describes a single workflow task in a MAP phase.
type MAPTask struct {
	Workflow      string
	Description   string
	InScope       bool
	EffortDays    float64
	EffortPercent float64
}
