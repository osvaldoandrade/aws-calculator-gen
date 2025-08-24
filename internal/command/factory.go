package command

import (
	"context"
	"fmt"
)

// Command represents a CLI subcommand.
type Command interface {
	Name() string
	Run(ctx context.Context, params map[string]string) error
}

// registry holds registered commands by name.
var registry = map[string]Command{}

// Register adds a command to the registry.
func Register(cmd Command) {
	registry[cmd.Name()] = cmd
}

// Resolve returns a command from the registry by name.
func Resolve(name string) (Command, error) {
	if c, ok := registry[name]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("unknown command %s", name)
}

func init() {
	Register(NewMapCommand())
}
