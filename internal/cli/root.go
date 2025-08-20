package cli

import (
	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/commands"
)

// NewRoot returns the root command.
func NewRoot() *cobra.Command {
	factory := commands.NewFactory()
  
	factory.Register(commands.MapCommand{})

	root := &cobra.Command{
		Use:   "seidor-aws-cli",
		Short: "Automation for AWS MAP incentives",
		Long: `Generate AWS Pricing Calculator estimates and MAP incentive artifacts.

Example:
  seidor-aws-cli map wizard --out ./out`,
	}

	if c, err := factory.Build("map"); err == nil {



		root.AddCommand(c)
	}
	return root
}
