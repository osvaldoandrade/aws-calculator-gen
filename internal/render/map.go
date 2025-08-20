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
	data := map[string]interface{}{"Customer": customer, "Tier": plan.Tier, "Cap": plan.CapAmount}
	return t.Execute(f, data)
}

// MAPXLSX renders a basic MAP.xlsx workbook.
func MAPXLSX(path string, plan types.FundingPlan, customer string) error {
	f := excelize.NewFile()
	sheets := []string{"Cover", "Request", "Workloads", "Budget", "Milestones"}
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
	_, err = fmt.Fprintf(f, "Customer: %s\nDescription: %s\nRegion: %s\nARR: %.2f\nTier: %s\nFunding Cap: %.2f\nAWS Calc: %s\n", customer, description, region, arr, plan.Tier, plan.CapAmount, link)
	return err
}
