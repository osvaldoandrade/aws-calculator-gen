package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/seidor-tools/internal/command"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "expected subcommand 'map'")
		os.Exit(1)
	}
	ctx := context.Background()
	cmd, err := command.Resolve(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := cmd.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
