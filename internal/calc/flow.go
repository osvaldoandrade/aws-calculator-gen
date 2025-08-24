package calc

import (
	"context"
	"time"
)

type Orchestrator struct {
	EstimateName string
	RegionCode   string
	TargetMRR    float64
	Headful      bool
	Tolerance    float64
	Timeout      time.Duration
	MaxRetries   int
}

type Result struct {
	ShareURL      string
	RegionLabel   string
	InstanceType  string
	Count         int
	AchievedMRR   float64
	RelativeError float64
}

// Run executes the pricing calculator automation. The current implementation
// is a stub that returns placeholder values.
func (o *Orchestrator) Run(ctx context.Context) (Result, error) {
	return Result{
		ShareURL:      "https://calculator.aws/#/estimate/mock",
		RegionLabel:   regionLabelFromCode(o.RegionCode),
		InstanceType:  "m7i.large",
		Count:         1,
		AchievedMRR:   o.TargetMRR,
		RelativeError: 0,
	}, nil
}
