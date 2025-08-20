package llm

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

type fakeAPI struct{}

func (f fakeAPI) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok"}}}}, nil
}

func TestDryRun(t *testing.T) {
	client := NewOpenAI(true)
	text, err := client.Generate(context.Background(), Prompt{System: "sys", User: "Hello {{name}}", Vars: map[string]string{"name": "Bob"}})
	require.NoError(t, err)
	require.Equal(t, "narrativa gerada (dry-run)", text)
}

func TestReal(t *testing.T) {
	client := &OpenAI{api: fakeAPI{}, dryRun: false}
	text, err := client.Generate(context.Background(), Prompt{System: "s", User: "hi", Vars: map[string]string{"x": "y"}})
	require.NoError(t, err)
	require.Equal(t, "ok", text)
}
