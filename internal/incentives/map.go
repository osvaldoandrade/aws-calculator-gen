package incentives

import "github.com/example/seidor-aws-cli/pkg/types"

// MAP tiers and caps.
var tiers = []struct {
	limit float64
	tier  string
}{
	{100000, "100k"},
	{150000, "150k"},
	{300000, "300k"},
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
	plan.Tier = ">300k"
	plan.CapAmount = arr * plan.CapPercent
	return plan
}
