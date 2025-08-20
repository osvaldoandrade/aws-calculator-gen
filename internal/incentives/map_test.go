package incentives

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeMAPFunding(t *testing.T) {
	plan := ComputeMAPFunding(250000)
	require.Equal(t, "300k", plan.Tier)
	require.Equal(t, 250000*0.10, plan.CapAmount)
	require.Equal(t, 3.0, plan.Assess.TotalDays)
	require.Len(t, plan.Assess.Tasks, 3)
	require.Equal(t, 8.0, plan.Mobilize.TotalDays)
	require.Len(t, plan.Mobilize.Tasks, 8)
}

func TestComputeMAPFundingAbove(t *testing.T) {
	plan := ComputeMAPFunding(700000)
	require.Equal(t, ">600k", plan.Tier)
}
