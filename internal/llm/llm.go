package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type openaiAPI interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// Prompt defines generation parameters.
type Prompt struct {
	System string
	User   string
	Vars   map[string]string
	Lang   string
}

// LLMClient abstracts language models.
type LLMClient interface {
	Generate(ctx context.Context, p Prompt) (string, error)
}

// OpenAI implements LLMClient using OpenAI API.
type OpenAI struct {
	api    openaiAPI
	dryRun bool
}

// NewOpenAI creates a new client reading API key from env.
func NewOpenAI(dryRun bool) *OpenAI {
	key := os.Getenv("OPENAI_API_KEY")
	var api openaiAPI
	if key != "" {
		api = openai.NewClient(key)
	}
	return &OpenAI{api: api, dryRun: dryRun}
}

// Generate calls the model or returns deterministic text when dry-run.
func (o *OpenAI) Generate(ctx context.Context, p Prompt) (string, error) {
	if o.dryRun || o.api == nil {
		return "narrativa gerada (dry-run)", nil
	}
	msg := p.User
	for k, v := range p.Vars {
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{{%s}}", k), v)
	}
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: p.System},
			{Role: openai.ChatMessageRoleUser, Content: msg},
		},
	}
	var lastErr error
	for i := 0; i < 3; i++ {
		resp, err := o.api.CreateChatCompletion(ctx, req)
		if err == nil {
			return strings.TrimSpace(resp.Choices[0].Message.Content), nil
		}
		lastErr = err
		time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
	}
	return "", lastErr
}
