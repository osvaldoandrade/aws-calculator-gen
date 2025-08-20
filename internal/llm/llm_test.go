package llm

import (
	"context"
	"testing"
)

func TestOpenAIClientDryRun(t *testing.T) {
	c := OpenAIClient{DryRun: true}
	out, err := c.Generate(context.Background(), Prompt{System: "sys", User: "u"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "dry run output" {
		t.Fatalf("unexpected output %s", out)
	}
}

func TestOpenAIClientReal(t *testing.T) {
	c := OpenAIClient{DryRun: false}
	out, err := c.Generate(context.Background(), Prompt{System: "sys", User: "user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatalf("expected output")
	}
}
