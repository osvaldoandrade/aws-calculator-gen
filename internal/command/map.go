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

	"github.com/example/aws-calculator-gen/internal/calc"
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
	pterm.DefaultSection.Println("AWS Calculator Generator")

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
		Headful:      true,
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

	// 2) Opening calculator / Adding services / Generating link
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
		hourlyRateBRL = 500.0 // pedido: valor/hora BRL 500,00
		usdToBrl      = 5.5
		assessmentPct = 0.05
		maxBudgetUSD  = 75000.0
		hoursPerDay   = 8.0
	)

	// Budget a partir do ARR (limitado a 75k)
	budgetUSD := arr * assessmentPct
	if budgetUSD > maxBudgetUSD {
		budgetUSD = maxBudgetUSD
	}
	budgetBRL := budgetUSD * usdToBrl

	// Horas totais que o budget cobre para 1 pessoa (informativo)
	totalHoursBudget := budgetBRL / hourlyRateBRL

	// Fases (% do esforço total)
	type bucket struct {
		Key    string
		Name   string
		Weight float64 // 30% / 40% / 30% conforme planilha
	}
	buckets := []bucket{
		{Key: "business_case", Name: "Caso de negócios inicial", Weight: 0.30},
		{Key: "discovery", Name: "Descoberta inicial", Weight: 0.40},
		{Key: "strategy", Name: "Análise de estratégia", Weight: 0.30},
	}

	// Estratégia para dias:
	// - Primeiro estimamos os *dias totais do projeto* para 1 pessoa: round(totalHoursBudget/8).
	// - Em seguida, alocamos os dias por fase segundo os pesos 30/40/30.
	// - O número TOTAL de pessoas é calculado para encaixar o budget: ceil(budget / (sumDays*8*rate)).
	totalDays := int(math.Round(totalHoursBudget / hoursPerDay))
	if totalDays < 1 {
		totalDays = 1
	}

	// Aloca dias por fase respeitando soma == totalDays
	phaseDays := make([]int, len(buckets))
	remaining := totalDays
	for i := range buckets {
		if i == len(buckets)-1 {
			phaseDays[i] = remaining
			break
		}
		d := int(math.Ceil(float64(totalDays) * buckets[i].Weight))
		if d < 0 {
			d = 0
		}
		phaseDays[i] = d
		remaining -= d
		if remaining < 0 {
			remaining = 0
		}
	}

	// Recalcula soma (só por segurança)
	sumDays := 0
	for _, d := range phaseDays {
		sumDays += d
	}
	if sumDays == 0 {
		sumDays = 1
	}

	// Número total de pessoas (ceil) para usar o teto do budget
	totalPeople := int(math.Ceil(budgetBRL / (float64(sumDays) * hoursPerDay * hourlyRateBRL)))
	if totalPeople < 1 {
		totalPeople = 1
	}

	// Pessoas por fase (proporcional ao peso; fracionário, como na planilha)
	phasePeople := make([]float64, len(buckets))
	for i := range buckets {
		phasePeople[i] = float64(totalPeople) * buckets[i].Weight
	}

	// Monta atividades
	activities := make([]map[string]any, 0, len(buckets))
	totalHours := 0
	for i, b := range buckets {
		days := phaseDays[i]
		hours := days * int(hoursPerDay)
		totalHours += hours

		activities = append(activities, map[string]any{
			"key":            b.Key,
			"name":           b.Name,
			"share":          b.Weight,
			"days":           days,
			"hours":          hours,          // esforço em horas (8h/dia) — não multiplica por pessoas
			"peopleFraction": phasePeople[i], // pessoas (fracionário) por fase
		})
	}

	// Custo total do assessment: forçamos o teto (pedido)
	assessmentTotalCostBRL := budgetBRL

	workplan := map[string]any{
		"hourlyRateBRL":    hourlyRateBRL,
		"usdToBrl":         usdToBrl,
		"budgetUSD":        budgetUSD,
		"budgetBRL":        budgetBRL,
		"totalHoursBudget": totalHoursBudget,
		"activities":       activities,
		"totals": map[string]any{
			"daysTotal":    sumDays,
			"hoursTotal":   totalHours,
			"peopleTotal":  totalPeople,
			"withinBudget": true,
		},
		// custo total do projeto em moeda local (BRL)
		"assessmentTotalCostBRL": assessmentTotalCostBRL,
	}

	data := map[string]any{
		"tool":          "aws-calculator-gen",
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
		// campo pedido: número total de pessoas do projeto (ceil)
		"number_of_people": totalPeople,
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
