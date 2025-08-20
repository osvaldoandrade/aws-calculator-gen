package main

import (
	"context"
	"os"
	"github.com/pterm/pterm"
	"github.com/example/seidor-aws-cli/internal/cli"
)

func main() {
	root := cli.NewRoot()
	if err := root.ExecuteContext(context.Background()); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
