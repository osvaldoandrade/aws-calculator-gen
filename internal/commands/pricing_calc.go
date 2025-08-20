package commands

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/awspc"
	"github.com/example/seidor-aws-cli/internal/pricing"
)

// PricingCalcCommand implements `pricing calc`.
type PricingCalcCommand struct{}

// Name returns command name.
func (PricingCalcCommand) Name() string { return "pricing" }

// Command builds pricing command.
func (PricingCalcCommand) Command() *cobra.Command {
	var input, title string
	var awsCalc bool
	cmd := &cobra.Command{
		Use:     "pricing",
		Short:   "pricing utilities",
		Long:    "Calculate costs from usage and optionally create an AWS Pricing Calculator estimate.",
		Example: "seidor-aws-cli pricing calc --input usage.yml\nseidor-aws-cli pricing calc --input usage.yml --aws-calc --title 'My estimate'",
	}

	calcCmd := &cobra.Command{
		Use:   "calc",
		Short: "calculate cost from usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(input)
			if err != nil {
				return fmt.Errorf("open usage: %w", err)
			}
			usage, err := pricing.LoadUsage(file)
			if err != nil {
				return fmt.Errorf("load usage: %w", err)
			}
			catFile, err := os.Open("assets/pricing-sample.yml")
			if err != nil {
				return fmt.Errorf("open catalog: %w", err)
			}
			cat, err := pricing.LoadCatalog(catFile)
			if err != nil {
				return fmt.Errorf("load catalog: %w", err)
			}
			breakdown, total := pricing.Calculate(cat, usage)
			rows := [][]string{{"Service", "USD"}}
			for svc, v := range breakdown {
				rows = append(rows, []string{svc, fmt.Sprintf("%.2f", v)})
			}
			_ = pterm.DefaultTable.WithHasHeader().WithData(rows).Render()
			pterm.Info.Printf("Total: %.2f\n", total)
			if awsCalc {
				client := awspc.StubClient{}
				id, err := client.CreateWorkloadEstimate(cmd.Context(), title, usage.Region, total)
				if err != nil {
					return err
				}
				link := fmt.Sprintf("https://console.aws.amazon.com/costmanagement/home#/pricing-calculator/workload-estimates/%s", id)
				pterm.Success.Printf("AWS estimate created: %s\n", link)
			}
			return nil
		},
	}
	calcCmd.Flags().StringVar(&input, "input", "", "usage file")
	calcCmd.Flags().BoolVar(&awsCalc, "aws-calc", false, "create AWS Pricing Calculator estimate")
	calcCmd.Flags().StringVar(&title, "title", "Estimate", "estimate title")
	calcCmd.MarkFlagRequired("input")
	cmd.AddCommand(calcCmd)
	return cmd
}
