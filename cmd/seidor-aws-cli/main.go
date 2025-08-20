package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/example/seidor-aws-cli/internal/awspc"
	mappkg "github.com/example/seidor-aws-cli/internal/incentives/map"
	"github.com/example/seidor-aws-cli/internal/llm"
	"github.com/example/seidor-aws-cli/internal/pricing"
	"github.com/example/seidor-aws-cli/internal/render"
	"github.com/example/seidor-aws-cli/pkg/types"
)

func main() {
	root := &cobra.Command{Use: "seidor-aws-cli"}
	root.AddCommand(pricingCmd())
	root.AddCommand(incentiveCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pricingCmd() *cobra.Command {
	var input string
	var awsCalc bool
	var title string
	cmd := &cobra.Command{Use: "pricing calc", RunE: func(cmd *cobra.Command, args []string) error {
		cat, err := pricing.LoadCatalog("assets/pricing-sample.yml")
		if err != nil {
			return err
		}
		calc := pricing.NewCalculator(cat)
		b, err := os.ReadFile(input)
		if err != nil {
			return err
		}
		u, err := pricing.ParseUsage(b)
		if err != nil {
			return err
		}
		total, breakdown := calc.Calculate(u)
		fmt.Fprintf(cmd.OutOrStdout(), "Total: $%.2f\n", total)
		for svc, v := range breakdown {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: $%.2f\n", svc, v)
		}
		if awsCalc {
			client, err := awspc.New(context.Background())
			if err != nil {
				return err
			}
			id, err := client.CreateEstimate(context.Background(), title)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "WorkloadEstimateId: %s\nURL: %s\n", id, awspc.URL(id))
		}
		return nil
	}}
	cmd.Flags().StringVar(&input, "input", "", "usage yaml input")
	cmd.Flags().BoolVar(&awsCalc, "aws-calc", false, "create estimate via AWS Pricing Calculator")
	cmd.Flags().StringVar(&title, "title", "Estimate", "estimate title")
	cmd.MarkFlagRequired("input")
	return cmd
}

func incentiveCmd() *cobra.Command {
	var out string
	var dry bool
	cmd := &cobra.Command{Use: "incentive"}
	wiz := &cobra.Command{Use: "map wizard", RunE: func(cmd *cobra.Command, args []string) error {
		var customer string
		survey.AskOne(&survey.Input{Message: "Customer:"}, &customer)
		var arrStr string
		survey.AskOne(&survey.Select{Message: "ARR Tier:", Options: []string{"205000", "300000", "600000"}}, &arrStr)
		var amount float64
		survey.AskOne(&survey.Input{Message: "Requested amount:"}, &amount)
		arr := types.ARR205k
		switch arrStr {
		case "205000":
			arr = types.ARR205k
		case "300000":
			arr = types.ARR300k
		case "600000":
			arr = types.ARR600k
		}
		spinner, _ := pterm.DefaultSpinner.Start("Generating plan")
		engine := mappkg.NewEngine(llm.NewOpenAI(dry))
		plan, err := engine.BuildPlan(context.Background(), arr, amount)
		if err != nil {
			spinner.Fail("error")
			return err
		}
		spinner.Success("done")
		md, err := render.MAPMarkdown(plan, customer)
		if err != nil {
			return err
		}
		os.MkdirAll(out, 0o755)
		if err := os.WriteFile(filepath.Join(out, "MAP.md"), []byte(md), 0o644); err != nil {
			return err
		}
		if err := render.MAPXLSX(filepath.Join(out, "MAP.xlsx"), plan, customer); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Artifacts written to %s\n", out)
		return nil
	}}
	wiz.Flags().StringVar(&out, "out", "out", "output directory")
	wiz.Flags().BoolVar(&dry, "dry-run", false, "dry run for LLM")
	cmd.AddCommand(wiz)
	return cmd
}
