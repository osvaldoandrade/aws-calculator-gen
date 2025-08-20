package incentives


import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeMAPFunding(t *testing.T) {
	plan := ComputeMAPFunding(100000)
	require.Equal(t, "100k", plan.Tier)
	require.Equal(t, 100000*0.10, plan.CapAmount)
}

func TestComputeMAPFundingAbove(t *testing.T) {
	plan := ComputeMAPFunding(350000)
	require.Equal(t, ">300k", plan.Tier)
}
