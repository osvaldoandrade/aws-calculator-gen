package mappkg

import (
	"context"
	"fmt"

	"github.com/example/seidor-aws-cli/internal/llm"
	"github.com/example/seidor-aws-cli/pkg/types"
)

// Engine builds funding plans for MAP incentive.
type Engine struct {
	LLM llm.LLMClient
}

// NewEngine creates Engine.
func NewEngine(client llm.LLMClient) *Engine {
	return &Engine{LLM: client}
}

// BuildPlan computes funding plan and narrative.
func (e *Engine) BuildPlan(ctx context.Context, arr types.ARRTier, requested float64) (types.FundingPlan, error) {
	cap := float64(arr) * 0.10
	plan := types.FundingPlan{ARR: arr, Cap: cap}
	if requested > cap {
		plan.Requested = cap
	} else {
		plan.Requested = requested
	}
	narr, err := e.LLM.Generate(ctx, llm.Prompt{System: "You are a concise AWS technical writer", User: "Justifique incentivo para ARR {{arr}}", Vars: map[string]string{"arr": fmt.Sprintf("%d", arr)}, Lang: "pt-BR"})
	if err != nil {
		return plan, err
	}
	plan.Narrative = narr
	return plan, nil
}
