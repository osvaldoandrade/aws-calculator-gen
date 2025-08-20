package incentives

import "github.com/example/seidor-aws-cli/pkg/types"

// MAP tiers and caps.
var tiers = []struct {
	limit float64
	tier  string
}{
	{205000, "205k"},
	{300000, "300k"},
	{600000, "600k"},
}

// ComputeMAPFunding returns funding plan based on ARR.
func ComputeMAPFunding(arr float64) types.FundingPlan {
	plan := types.FundingPlan{ARR: arr, CapPercent: 0.10}
	for _, t := range tiers {
		if arr <= t.limit {
			plan.Tier = t.tier
			plan.CapAmount = arr * plan.CapPercent
			return plan
		}
	}
	plan.Tier = ">600k"
	plan.CapAmount = arr * plan.CapPercent
	return plan
}
