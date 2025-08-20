package render

import (
	"strings"
	"text/template"

	"github.com/xuri/excelize/v2"

	"github.com/example/seidor-aws-cli/internal/templates"
	"github.com/example/seidor-aws-cli/pkg/types"
)

// MAPMarkdown renders MAP.md content.
func MAPMarkdown(plan types.FundingPlan, customer string) (string, error) {
	tmpl, err := template.New("map").Parse(templates.MAPMarkdown)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	data := map[string]any{"Customer": customer, "ARR": plan.ARR, "Requested": plan.Requested, "Narrative": plan.Narrative}
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// MAPXLSX creates spreadsheet at path.
func MAPXLSX(path string, plan types.FundingPlan, customer string) error {
	f := excelize.NewFile()
	sheets := []string{"Cover", "Request", "Workloads", "Budget", "Milestones"}
	for _, s := range sheets {
		if s != "Sheet1" {
			f.NewSheet(s)
		}
	}
	f.DeleteSheet("Sheet1")
	f.SetCellValue("Cover", "A1", "Customer")
	f.SetCellValue("Cover", "B1", customer)
	f.SetCellValue("Request", "A1", "ARR")
	f.SetCellValue("Request", "B1", plan.ARR)
	f.SetCellValue("Request", "A2", "Requested")
	f.SetCellValue("Request", "B2", plan.Requested)
	return f.SaveAs(path)
}
