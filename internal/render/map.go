package render

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/xuri/excelize/v2"

	"github.com/example/seidor-aws-cli/internal/templates"
	"github.com/example/seidor-aws-cli/pkg/types"
)

// MAPMarkdown renders MAP.md using funding plan data.
func MAPMarkdown(path string, plan types.FundingPlan, customer string) error {
	t, err := template.New("map").Parse(templates.MAPTemplate())
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data := struct {
		Customer string
		Plan     types.FundingPlan
	}{Customer: customer, Plan: plan}
	return t.Execute(f, data)
}

// MAPXLSX renders a basic MAP.xlsx workbook.
func MAPXLSX(path string, plan types.FundingPlan, customer string) error {
	f := excelize.NewFile()
	sheets := []string{"Cover", "Request", "Assess", "Mobilize", "Workloads", "Budget", "Milestones"}
	for i, s := range sheets {
		if i == 0 {
			f.SetSheetName("Sheet1", s)
		} else {
			f.NewSheet(s)
		}
	}
	f.SetCellValue("Cover", "A1", "Customer")
	f.SetCellValue("Cover", "B1", customer)
	f.SetCellValue("Request", "A1", "ARR Tier")
	f.SetCellValue("Request", "B1", plan.Tier)
	f.SetCellValue("Request", "A2", "Funding Cap")
	f.SetCellValue("Request", "B2", plan.CapAmount)
	// Assess sheet
	headers := []string{"Fluxo de trabalho", "Descrição", "No escopo", "Esforço (dias)", "% do esforço"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Assess", cell, h)
		f.SetCellValue("Mobilize", cell, h)
	}
	for r, tsk := range plan.Assess.Tasks {
		f.SetCellValue("Assess", fmt.Sprintf("A%d", r+2), tsk.Workflow)
		f.SetCellValue("Assess", fmt.Sprintf("B%d", r+2), tsk.Description)
		f.SetCellValue("Assess", fmt.Sprintf("C%d", r+2), tsk.InScope)
		f.SetCellValue("Assess", fmt.Sprintf("D%d", r+2), tsk.EffortDays)
		f.SetCellValue("Assess", fmt.Sprintf("E%d", r+2), tsk.EffortPercent)
	}
	lastAssess := len(plan.Assess.Tasks) + 2
	f.SetCellValue("Assess", fmt.Sprintf("D%d", lastAssess), plan.Assess.TotalDays)
	f.SetCellValue("Assess", fmt.Sprintf("E%d", lastAssess), 100)
	f.SetCellValue("Assess", fmt.Sprintf("D%d", lastAssess+1), plan.Assess.TotalWeeks)

	for r, tsk := range plan.Mobilize.Tasks {
		f.SetCellValue("Mobilize", fmt.Sprintf("A%d", r+2), tsk.Workflow)
		f.SetCellValue("Mobilize", fmt.Sprintf("B%d", r+2), tsk.Description)
		f.SetCellValue("Mobilize", fmt.Sprintf("C%d", r+2), tsk.InScope)
		f.SetCellValue("Mobilize", fmt.Sprintf("D%d", r+2), tsk.EffortDays)
		f.SetCellValue("Mobilize", fmt.Sprintf("E%d", r+2), tsk.EffortPercent)
	}
	lastMob := len(plan.Mobilize.Tasks) + 2
	f.SetCellValue("Mobilize", fmt.Sprintf("D%d", lastMob), plan.Mobilize.TotalDays)
	f.SetCellValue("Mobilize", fmt.Sprintf("E%d", lastMob), 100)
	f.SetCellValue("Mobilize", fmt.Sprintf("D%d", lastMob+1), plan.Mobilize.TotalWeeks)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return f.SaveAs(path)
}

// MAPText writes a text summary with relevant fields and AWS link.
func MAPText(path, customer, description, region string, arr float64, plan types.FundingPlan, link string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = fmt.Fprintf(f, "Customer: %s\nDescription: %s\nRegion: %s\nARR: %.2f\nTier: %s\nFunding Cap: %.2f\nAWS Calc: %s\n", customer, description, region, arr, plan.Tier, plan.CapAmount, link); err != nil {
		return err
	}
	// Assess details
	if _, err = fmt.Fprintln(f, "\nAssess:"); err != nil {
		return err
	}
	for _, tsk := range plan.Assess.Tasks {
		if _, err = fmt.Fprintf(f, "- %s: %s (dias: %.0f, %%: %.0f)\n", tsk.Workflow, tsk.Description, tsk.EffortDays, tsk.EffortPercent); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintf(f, "Total dias: %.0f (semanas: %.1f)\n", plan.Assess.TotalDays, plan.Assess.TotalWeeks); err != nil {
		return err
	}
	// Mobilize details
	if _, err = fmt.Fprintln(f, "\nMobilize:"); err != nil {
		return err
	}
	for _, tsk := range plan.Mobilize.Tasks {
		if _, err = fmt.Fprintf(f, "- %s: %s (dias: %.0f, %%: %.0f)\n", tsk.Workflow, tsk.Description, tsk.EffortDays, tsk.EffortPercent); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(f, "Total dias: %.0f (semanas: %.1f)\n", plan.Mobilize.TotalDays, plan.Mobilize.TotalWeeks)
	return err
}
