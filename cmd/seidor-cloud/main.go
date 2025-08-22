package main

import (
	"context"
	"os"

	"github.com/example/seidor-aws-cli/internal/cli"
	"github.com/pterm/pterm"
)

func main() {
	// Clear the screen and render a title before starting the CLI.
	pterm.Print("\033[H\033[2J")
	pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString("Seidor Cloud")).Render()

	root := cli.NewRoot()
	if err := root.ExecuteContext(context.Background()); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
