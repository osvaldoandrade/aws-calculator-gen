package command

import (
	"context"
	"fmt"
)

type Command interface {
	Name() string
	Run(ctx context.Context) error
}

var registry = map[string]Command{}

func Register(cmd Command) {
	registry[cmd.Name()] = cmd
}

func Resolve(name string) (Command, error) {
	if c, ok := registry[name]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("unknown command %s", name)
}

func init() {
	Register(&MapCommand{})
}
