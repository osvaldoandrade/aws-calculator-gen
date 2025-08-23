package commands

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/example/seidor-aws-cli/internal/awspc"
	"github.com/example/seidor-aws-cli/internal/incentives"
	"github.com/example/seidor-aws-cli/internal/llm"
	"github.com/example/seidor-aws-cli/internal/render"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
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
			var description string
			if err := survey.AskOne(&survey.Input{Message: "Deal description:"}, &description, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
			var region string
			regions := []string{"us-east-1", "us-west-2", "eu-west-1", "sa-east-1"}
			if err := survey.AskOne(&survey.Select{Message: "Region:", Options: regions}, &region); err != nil {
				return err
			}
			var template string
			templates := []string{"generic-reactive", "generic-lake"}
			if err := survey.AskOne(&survey.Select{Message: "Solution template:", Options: templates}, &template); err != nil {
				return err
			}
			var arrStr string
			if err := survey.AskOne(&survey.Input{Message: "ARR (USD):"}, &arrStr, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
			arr, err := strconv.ParseFloat(arrStr, 64)
			if err != nil {
				return err
			}
			title := fmt.Sprintf("%s-%s-%s", time.Now().Format("20060102"), slugify(customer), slugify(description))
			pterm.Info.Printf("Estimate title: %s\n", title)

			plan := incentives.ComputeMAPFunding(arr)
			pterm.Info.Printf("Tier %s with cap %.2f\n", plan.Tier, plan.CapAmount)

			llmClient := llm.NewOpenAIClientFromEnv()
			summary, err := llmClient.Generate(cmd.Context(), llm.Prompt{
				System: "You generate brief MAP summaries in Portuguese.",
				User:   fmt.Sprintf("Cliente %s, template %s, ARR %.2f", customer, template, arr),
			})
			if err != nil {
				pterm.Warning.Printf("LLM resumo falhou: %v\n", err)
			} else {
				pterm.Info.Println(summary)
			}
			client, err := awspc.New(cmd.Context())
			if err != nil {
				pterm.Warning.Printf("using stub AWS client: %v\n", err)
				client = awspc.StubClient{}
			}
			link, err := client.CreateWorkloadEstimate(cmd.Context(), title, region, template, arr)
			if err != nil {
				if strings.Contains(err.Error(), "ServiceQuotaExceededException") || strings.Contains(err.Error(), "AccessDenied") {
					pterm.Warning.Printf("workload estimate failed: %v\n", err)
					pterm.Warning.Println("retrying via BILL")
					link, err = client.CreateBillEstimate(cmd.Context(), title)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
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
			if err := render.MAPText(filepath.Join(outDir, "MAP.txt"), customer, description, region, arr, plan, link); err != nil {
				spin.Fail("txt")
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

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	r := strings.NewReplacer(" ", "-", "_", "-")
	s = r.Replace(s)
	re := regexp.MustCompile(`[^a-z0-9-]+`)
	s = re.ReplaceAllString(s, "")
	re = regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")

}
