package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/seidor-tools/internal/command"
)

const usage = `seidor-tools is a CLI utility.

Usage:
  seidor-tools <command> [--params key=value ...]

Available commands:
  map    Create MAP estimate
`

// main is the entry point for the CLI.
func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Fprint(os.Stdout, usage)
		return
	}

	name := os.Args[1]
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		if name == "map" {
			fmt.Fprintln(os.Stdout, "Usage: seidor-tools map [--params key=value ...]")
			fmt.Fprintln(os.Stdout, "Creates an AWS Pricing Calculator estimate using MAP.")
			return
		}
	}

	params := map[string]string{}
	for i, arg := range os.Args {
		if arg == "--params" {
			p, err := command.ParseParams(os.Args[i+1:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			params = p
			break
		}
	}

	ctx := context.Background()
	cmd, err := command.Resolve(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := cmd.Run(ctx, params); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
