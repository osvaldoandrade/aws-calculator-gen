package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/awspc"
	"github.com/example/seidor-aws-cli/internal/incentives"
	"github.com/example/seidor-aws-cli/internal/render"
)

// MapCommand exposes MAP incentive helpers.
type MapCommand struct{}

func (MapCommand) Name() string { return "map" }

func (MapCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "map",
		Short: "MAP incentive helpers",
	}

	var outDir string
	wizardCmd := &cobra.Command{
		Use:   "wizard",
		Short: "interactive MAP wizard",
		Long:  "Collect opportunity details, create AWS Pricing Calculator estimate and generate MAP artifacts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var customer string
			if err := survey.AskOne(&survey.Input{Message: "Customer name:"}, &customer, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
			var arrStr string
			arrOptions := []string{"100000", "150000", "300000"}
			if err := survey.AskOne(&survey.Select{Message: "Expected ARR (USD):", Options: arrOptions}, &arrStr); err != nil {
				return err
			}
			arr := map[string]float64{"100000": 100000, "150000": 150000, "300000": 300000}[arrStr]
			var services []string
			svcOpts := []string{"EC2", "EBS", "S3", "RDS", "Lambda", "EKS", "CloudFront"}
			if err := survey.AskOne(&survey.MultiSelect{Message: "AWS services involved:", Options: svcOpts}, &services); err != nil {
				return err
			}
			var title string
			if err := survey.AskOne(&survey.Input{Message: "Estimate title:"}, &title, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			plan := incentives.ComputeMAPFunding(arr)
			pterm.Info.Printf("Tier %s with cap %.2f\n", plan.Tier, plan.CapAmount)
			client := awspc.StubClient{}
			id, err := client.CreateWorkloadEstimate(cmd.Context(), title)
			if err != nil {
				return err
			}
			link := fmt.Sprintf("https://console.aws.amazon.com/costmanagement/home#/pricing-calculator/workload-estimates/%s", id)
			pterm.Success.Printf("AWS estimate created: %s\n", link)

			spin, _ := pterm.DefaultSpinner.Start("Generating MAP artifacts")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				spin.Fail("failed")
				return fmt.Errorf("create output: %w", err)
			}
			if err := render.MAPMarkdown(filepath.Join(outDir, "MAP.md"), plan, customer); err != nil {
				spin.Fail("markdown")
				return err
			}
			if err := render.MAPXLSX(filepath.Join(outDir, "MAP.xlsx"), plan, customer); err != nil {
				spin.Fail("xlsx")
				return err
			}
			spin.Success("artifacts written")
			pterm.Success.Println("MAP artifacts generated in", outDir)
			return nil
		},
	}
	wizardCmd.Flags().StringVar(&outDir, "out", "./out", "output directory")
	cmd.AddCommand(wizardCmd)
	return cmd
}
