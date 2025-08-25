package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"

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

	// ==== UI spinners por fases ====

	// 1) Estimating
	if c.startSpinner != nil {
		s1, _ := c.startSpinner("⏳ Estimating the right solution...")
		s1.Success("OK")
		fmt.Printf("\n")
	}

	// 2) Opening calculator / Adding services / Generating link (mesmo spinner, rótulo evolui)
	var result calc.Result
	if c.startSpinner != nil {
		spin, _ := c.startSpinner("⏳ Opening AWS public calculator...")

		resCh := make(chan calc.Result, 1)
		errCh := make(chan error, 1)

		go func() {
			res, runErr := c.runOrchestrator(ctx, orch)
			if runErr != nil {
				errCh <- runErr
				return
			}
			resCh <- res
		}()

		step := 0
		ticker := time.NewTicker(7 * time.Second)
		defer ticker.Stop()

	loop:
		for {
			select {
			case <-ticker.C:
				switch step {
				case 0:
					spin.UpdateText("⏳ Adding services...")
					step = 1
				case 1:
					spin.UpdateText("⏳ Generating share link...")
					step = 2
				default:
				}
			case err = <-errCh:
				spin.Fail("error: " + err.Error())
				return err
			case result = <-resCh:
				spin.Success("OK")
				break loop
			}
		}
	} else {
		result, err = c.runOrchestrator(ctx, orch)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\n")

	// ===== Workplan =====
	const (
		hourlyRateBRL = 305.0
		usdToBrl      = 5.5
		assessmentPct = 0.05
		maxBudgetUSD  = 75000.0
		hoursPerDay   = 8.0
	)

	// Budget a partir do ARR
	budgetUSD := arr * assessmentPct
	if budgetUSD > maxBudgetUSD {
		budgetUSD = maxBudgetUSD
	}
	budgetBRL := budgetUSD * usdToBrl

	// Horas totais que o budget cobre
	totalHoursBudget := budgetBRL / hourlyRateBRL

	type bucket struct {
		Key    string
		Name   string
		Weight float64
		People int
	}
	// Alocação: 20% / 50% / 30%
	buckets := []bucket{
		{Key: "business_case", Name: "Caso de negócios inicial", Weight: 0.20, People: 1},
		{Key: "discovery", Name: "Descoberta inicial", Weight: 0.50, People: 2},
		{Key: "strategy", Name: "Análise de estratégia", Weight: 0.30, People: 2},
	}

	activities := make([]map[string]any, 0, len(buckets))
	var planHours float64
	var planCostBRL float64

	for _, b := range buckets {
		hoursAlloc := totalHoursBudget * b.Weight
		if hoursAlloc < 0 {
			hoursAlloc = 0
		}
		days := math.Ceil(hoursAlloc / (float64(b.People) * hoursPerDay))
		if hoursAlloc > 0 && days < 1 {
			days = 1
		}
		roundedHours := days * float64(b.People) * hoursPerDay
		costBRL := roundedHours * hourlyRateBRL

		planHours += roundedHours
		planCostBRL += costBRL

		activities = append(activities, map[string]any{
			"key":     b.Key,
			"name":    b.Name,
			"people":  b.People,
			"days":    int(days),
			"hours":   int(roundedHours),
			"costBRL": costBRL,
			"share":   b.Weight,
		})
	}

	workplan := map[string]any{
		"hourlyRateBRL":    hourlyRateBRL,
		"usdToBrl":         usdToBrl,
		"budgetUSD":        budgetUSD,
		"budgetBRL":        budgetBRL,
		"totalHoursBudget": totalHoursBudget,
		"activities":       activities,
		"totals": map[string]any{
			"hoursPlanned":   planHours,
			"costBRLPlanned": planCostBRL,
			"withinBudget":   planCostBRL <= budgetBRL,
		},
		// Atalho pedido no item (6): custo total do assessment em BRL
		"assessmentTotalCostBRL": planCostBRL,
	}

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
		"workplan":      workplan,
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
