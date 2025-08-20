package incentives

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeMAPFunding(t *testing.T) {
	plan := ComputeMAPFunding(250000)
	require.Equal(t, "300k", plan.Tier)
	require.Equal(t, 250000*0.10, plan.CapAmount)
}

func TestComputeMAPFundingAbove(t *testing.T) {
	plan := ComputeMAPFunding(700000)
	require.Equal(t, ">600k", plan.Tier)
}
