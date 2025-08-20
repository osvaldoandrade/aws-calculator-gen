package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
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

// OpenAIClient calls the real OpenAI API when an API key is configured.
type OpenAIClient struct {
	APIKey string
}

// NewOpenAIClientFromEnv creates an OpenAI client using environment variables.
// It checks both OPENAI_API_KEY and SECRET_KEY for compatibility.
func NewOpenAIClientFromEnv() OpenAIClient {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		key = os.Getenv("SECRET_KEY")
	}
	return OpenAIClient{APIKey: key}
}

// Generate returns text from the OpenAI API. When no key is set the output is deterministic.
func (c OpenAIClient) Generate(ctx context.Context, p Prompt) (string, error) {
	if c.APIKey == "" {
		return "dry run output", nil
	}
	client := openai.NewClient(c.APIKey)
	msg := p.User
	for k, v := range p.Vars {
		msg = strings.ReplaceAll(msg, "{{"+k+"}}", v)
	}
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: p.System}, {Role: openai.ChatMessageRoleUser, Content: msg}},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}
