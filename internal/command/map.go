package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/pterm/pterm"

	"github.com/example/seidor-tools/internal/calc"
)

type MapCommand struct{}

func (c *MapCommand) Name() string { return "map" }

func (c *MapCommand) Run(ctx context.Context) error {
	pterm.DefaultSection.Println("Seidor Cloud")

	customer, _ := pterm.DefaultInteractiveTextInput.Show("Customer name")
	description, _ := pterm.DefaultInteractiveTextInput.Show("Deal description")

	region, _ := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"us-east-1", "us-east-2", "us-west-2", "eu-west-1", "sa-east-1"}).
		Show("Select region")

	arrStr, _ := pterm.DefaultInteractiveTextInput.Show("ARR USD/year")
	arr, err := strconv.ParseFloat(arrStr, 64)
	if err != nil {
		return fmt.Errorf("invalid ARR: %w", err)
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

	spinner, _ := pterm.DefaultSpinner.Start("Abrindo navegador…")
	result, err := orch.Run(ctx)
	if err != nil {
		spinner.Fail("erro: " + err.Error())
		return err
	}
	spinner.Success("concluído")

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

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
