# MAP Funding Summary

| Field       | Value |
|-------------|-------|
| Customer    | {{.Customer}} |
| ARR Tier    | {{.Plan.Tier}} |
| Funding Cap | {{printf "%.2f" .Plan.CapAmount}} |

## Assess

| Fluxo de trabalho | Descrição | No escopo | Esforço estimado para todos os recursos em dias (8 horas por dia) | % do esforço total |
|-------------------|-----------|-----------|--------------------------------------------------------------------|--------------------|
{{range .Plan.Assess.Tasks}}| {{.Workflow}} | {{.Description}} | {{if .InScope}}Sim{{else}}Não{{end}} | {{.EffortDays}} | {{.EffortPercent}}% |
{{end}}| **Esforço em dias** |  |  | {{.Plan.Assess.TotalDays}} | 100% |
| **Esforço em semanas** |  |  | {{.Plan.Assess.TotalWeeks}} | |

## Mobilize

| Fluxo de trabalho | Descrição | No escopo | Esforço estimado para todos os recursos em dias (8 horas por dia) | % do esforço total |
|-------------------|-----------|-----------|--------------------------------------------------------------------|--------------------|
{{range .Plan.Mobilize.Tasks}}| {{.Workflow}} | {{.Description}} | {{if .InScope}}Sim{{else}}Não{{end}} | {{.EffortDays}} | {{.EffortPercent}}% |
{{end}}| **Esforço em dias** |  |  | {{.Plan.Mobilize.TotalDays}} | 100% |
| **Esforço em semanas** |  |  | {{.Plan.Mobilize.TotalWeeks}} | |

