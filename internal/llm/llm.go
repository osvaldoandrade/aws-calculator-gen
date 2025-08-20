package llm

import (
	"context"
	"fmt"
)

// Prompt represents input to the LLM.
type Prompt struct {
	System string
	User   string
	Lang   string
	Vars   map[string]string
}

// LLMClient defines generation interface.
type LLMClient interface {
	Generate(ctx context.Context, p Prompt) (string, error)
}

// OpenAIClient is a trivial implementation that returns deterministic text.
type OpenAIClient struct {
	DryRun bool
}

// Generate returns a simple formatted string. When DryRun is true the output is deterministic.
func (c OpenAIClient) Generate(ctx context.Context, p Prompt) (string, error) {
	if c.DryRun {
		return "dry run output", nil
	}
	return fmt.Sprintf("%s: %s", p.System, p.User), nil
}
