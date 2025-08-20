package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Command represents a cobra command provider.
type Command interface {
	Name() string
	Command() *cobra.Command
}

// Factory builds registered commands.
type Factory struct {
	cmds map[string]Command
}

// NewFactory returns a factory.
func NewFactory() *Factory {
	return &Factory{cmds: make(map[string]Command)}
}

// Register a command.
func (f *Factory) Register(c Command) {
	f.cmds[c.Name()] = c
}

// Build retrieves cobra command by name.
func (f *Factory) Build(name string) (*cobra.Command, error) {
	if c, ok := f.cmds[name]; ok {
		return c.Command(), nil
	}
	return nil, fmt.Errorf("command %s not found", name)
}
