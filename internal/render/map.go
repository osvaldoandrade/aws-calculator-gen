package render

import (
	"os"
	"text/template"

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
