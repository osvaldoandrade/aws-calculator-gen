package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pterm/pterm"

	"github.com/example/seidor-tools/internal/calc"
)

// MapCommand implements the "map" subcommand.
type MapCommand struct {
	out             io.Writer
	startSpinner    func(text string) (*pterm.SpinnerPrinter, error)
	runOrchestrator func(ctx context.Context, o calc.Orchestrator) (calc.Result, error)
}

// NewMapCommand returns a MapCommand with default dependencies.
func NewMapCommand() *MapCommand {
	return &MapCommand{
		out: os.Stdout,
		startSpinner: func(text string) (*pterm.SpinnerPrinter, error) {
			return pterm.DefaultSpinner.Start(text)
		},
		runOrchestrator: func(ctx context.Context, o calc.Orchestrator) (calc.Result, error) {
			return o.Run(ctx)
		},
	}
}

// Name returns the command name.
func (c *MapCommand) Name() string { return "map" }

// Run executes the map command.
// Required parameters are: customer, description, region and arr (annual recurring revenue).
// Parameters can be provided via --params or will be requested interactively.
func (c *MapCommand) Run(ctx context.Context, params map[string]string) error {
	pterm.DefaultSection.Println("Seidor Cloud")

	customer, err := getStringParam(params, "customer", "Customer name")
	if err != nil {
		return err
	}
	description, err := getStringParam(params, "description", "Deal description")
	if err != nil {
		return err
	}

	region, err := getStringParam(params, "region", "Select region")
	if err != nil {
		return err
	}
	if region == "" {
		region, _ = pterm.DefaultInteractiveSelect.
			WithOptions([]string{"us-east-1", "us-east-2", "us-west-2", "eu-west-1", "sa-east-1"}).
			Show("Select region")
	}

	arr, err := getFloatParam(params, "arr", "ARR USD/year")
	if err != nil {
		return err
	}

	targetMRR := arr / 12
	pterm.Info.Printf("Target MRR: %.2f USD\n", targetMRR)

	orch := calc.Orchestrator{
		EstimateName: fmt.Sprintf("MAP • %s", customer),
		RegionCode:   region,
		TargetMRR:    targetMRR,
		Headful:      false,
		Tolerance:    0.03,
		Timeout:      0,
		MaxRetries:   3,
	}

	var result calc.Result
	if c.startSpinner != nil {
		spinner, _ := c.startSpinner("Opening browser...")
		result, err = c.runOrchestrator(ctx, orch)
		if err != nil {
			spinner.Fail("error: " + err.Error())
			return err
		}
		spinner.Success("done")
	} else {
		result, err = c.runOrchestrator(ctx, orch)
		if err != nil {
			return err
		}
	}

	pterm.Success.Println(result.ShareURL)

	data := map[string]any{
		"tool":          "seidor-tools",
		"command":       "map",
		"customer":      customer,
		"description":   description,
		"estimateName":  orch.EstimateName,
		"shareUrl":      result.ShareURL,
		"region":        result.RegionLabel,
		"os":            "Linux",
		"arch":          "x86",
		"tenancy":       "Shared",
		"purchase":      "On-Demand",
		"instanceType":  result.InstanceType,
		"count":         result.Count,
		"targetMRR":     targetMRR,
		"achievedMRR":   result.AchievedMRR,
		"relativeError": result.RelativeError,
	}

	enc := json.NewEncoder(c.out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// getStringParam retrieves a string parameter or asks the user if missing.
func getStringParam(params map[string]string, key, prompt string) (string, error) {
	if v, ok := params[key]; ok {
		pterm.Info.Printf("%s: %s\n", key, v)
		return v, nil
	}
	if prompt == "Select region" {
		return "", nil
	}
	val, _ := pterm.DefaultInteractiveTextInput.Show(prompt)
	return val, nil
}

// getFloatParam retrieves a float parameter or asks the user if missing.
func getFloatParam(params map[string]string, key, prompt string) (float64, error) {
	if v, ok := params[key]; ok {
		pterm.Info.Printf("%s: %s\n", key, v)
		return strconv.ParseFloat(v, 64)
	}
	val, _ := pterm.DefaultInteractiveTextInput.Show(prompt)
	return strconv.ParseFloat(val, 64)
}
