package command

import (
	"context"
	"testing"
)

type dummy struct{}

func (d *dummy) Name() string                                            { return "dummy" }
func (d *dummy) Run(ctx context.Context, params map[string]string) error { return nil }

func TestRegisterResolve(t *testing.T) {
	Register(&dummy{})
	cmd, err := Resolve("dummy")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if cmd.Name() != "dummy" {
		t.Fatalf("unexpected command: %s", cmd.Name())
	}
}
