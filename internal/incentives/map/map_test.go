package mappkg

import (
	"context"
	"testing"

	"github.com/example/seidor-aws-cli/internal/llm"
	"github.com/example/seidor-aws-cli/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestBuildPlan(t *testing.T) {
	engine := NewEngine(llm.NewOpenAI(true))
	plan, err := engine.BuildPlan(context.Background(), types.ARR300k, 40000)
	require.NoError(t, err)
	require.Equal(t, float64(types.ARR300k)*0.10, plan.Cap)
	require.Equal(t, float64(types.ARR300k)*0.10, plan.Requested)
	require.NotEmpty(t, plan.Narrative)
}
