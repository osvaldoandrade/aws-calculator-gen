package cli

import (
	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/commands"
)

// NewRoot returns the root command.
func NewRoot() *cobra.Command {
	factory := commands.NewFactory()
	factory.Register(commands.PricingCalcCommand{})
	factory.Register(commands.IncentiveCommand{})

	root := &cobra.Command{
		Use:   "seidor-aws-cli",
		Short: "Automation for AWS MAP incentives and pricing",
		Long: `Generate MAP artifacts, compute pricing and create AWS Pricing Calculator estimates.

Examples:
  seidor-aws-cli pricing calc --input usage.yml
  seidor-aws-cli pricing calc --input usage.yml --aws-calc --title "My workload"
  seidor-aws-cli incentive map wizard --out ./out`,
	}

	if c, err := factory.Build("pricing"); err == nil {
		root.AddCommand(c)
	}
	if c, err := factory.Build("incentive"); err == nil {
		root.AddCommand(c)
	}
	return root
}
