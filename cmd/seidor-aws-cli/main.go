package main

import (
	"context"
	"log"

	"github.com/example/seidor-aws-cli/internal/cli"
)

func main() {
	root := cli.NewRoot()
	if err := root.ExecuteContext(context.Background()); err != nil {
		log.Fatal(err)
	}
}
