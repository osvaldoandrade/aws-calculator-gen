package cli

import (
	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/commands"
)

// NewRoot returns the root command.
func NewRoot() *cobra.Command {
	factory := commands.NewFactory()
	factory.Register(commands.PricingCalcCommand{})

	root := &cobra.Command{Use: "seidor-aws-cli"}
	if c, err := factory.Build("pricing"); err == nil {
		root.AddCommand(c)
	}
	return root
}
