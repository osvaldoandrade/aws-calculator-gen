package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/pricing"
)

// PricingCalcCommand implements `pricing calc`.
type PricingCalcCommand struct{}

func (PricingCalcCommand) Name() string { return "pricing" }

func (PricingCalcCommand) Command() *cobra.Command {
	var input string
	cmd := &cobra.Command{
		Use:   "pricing",
		Short: "pricing utilities",
	}

	calcCmd := &cobra.Command{
		Use:   "calc",
		Short: "calculate cost from usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(input)
			if err != nil {
				return err
			}
			usage, err := pricing.LoadUsage(file)
			if err != nil {
				return err
			}
			catFile, err := os.Open("assets/pricing-sample.yml")
			if err != nil {
				return err
			}
			cat, err := pricing.LoadCatalog(catFile)
			if err != nil {
				return err
			}
			breakdown, total := pricing.Calculate(cat, usage)
			for svc, v := range breakdown {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %.2f\n", svc, v)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Total: %.2f\n", total)
			return nil
		},
	}
	calcCmd.Flags().StringVar(&input, "input", "", "usage file")
	calcCmd.MarkFlagRequired("input")
	cmd.AddCommand(calcCmd)
	return cmd
}
