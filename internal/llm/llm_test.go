package llm

import (
	"context"
	"os"
	"testing"
)

func TestOpenAIClientDryRun(t *testing.T) {
	c := OpenAIClient{}
	out, err := c.Generate(context.Background(), Prompt{System: "sys", User: "u"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "dry run output" {
		t.Fatalf("unexpected output %s", out)
	}
}

func TestOpenAIClientReal(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	c := OpenAIClient{APIKey: key}
	out, err := c.Generate(context.Background(), Prompt{System: "sys", User: "user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatalf("expected output")
	}
}
