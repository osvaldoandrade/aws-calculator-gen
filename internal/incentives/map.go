package incentives

import "github.com/example/seidor-aws-cli/pkg/types"

// MAP tiers and caps.
var tiers = []struct {
	limit float64
	tier  string
}{
	{205000, "205k"},
	{300000, "300k"},
	{600000, "600k"},
}

// ComputeMAPFunding returns funding plan based on ARR.
func ComputeMAPFunding(arr float64) types.FundingPlan {
	plan := types.FundingPlan{ARR: arr, CapPercent: 0.10}
	for _, t := range tiers {
		if arr <= t.limit {
			plan.Tier = t.tier
			break
		}
	}
	if plan.Tier == "" {
		plan.Tier = ">600k"
	}
	plan.CapAmount = arr * plan.CapPercent
	plan.Assess = assessPhase()
	plan.Mobilize = mobilizePhase()
	return plan
}

func assessPhase() types.MAPPhase {
	tasks := []types.MAPTask{
		{
			Workflow:      "Caso de negócios inicial",
			Description:   "Cria um relatório ou uma apresentação com o custo total de propriedade (TCO)/retorno sobre o investimento (ROI) comparando o estado atual com o da AWS, bem como as vantagens comerciais, os motivadores da mudança para a AWS e o custo de serviços profissionais.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 33,
		},
		{
			Workflow:      "Descoberta inicial",
			Description:   "Realiza preferencialmente uma descoberta automatizada baseada em ferramentas dos detalhes das workloads para viabilizar o dimensionamento da solução da AWS. Também pode ser uma planilha ou baseada em CMDB.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 33,
		},
		{
			Workflow:      "Análise de estratégia (7Rs)",
			Description:   "Define as estratégias de migração e modernização de cada workload. Os 7Rs são: redefinir a hospedagem, realocar, refatorar, redefinir a plataforma, recomprar, reter e retirar. O padrão escolhido deve estar alinhado com os motivadores do cliente e com a descoberta.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 33,
		},
	}
	total := 0.0
	for _, t := range tasks {
		total += t.EffortDays
	}
	return types.MAPPhase{
		Name:       "Assess",
		Tasks:      tasks,
		TotalDays:  total,
		TotalWeeks: total / 5,
	}
}

func mobilizePhase() types.MAPPhase {
	tasks := []types.MAPTask{
		{
			Workflow:      "Caso de negócios detalhado",
			Description:   "Cria um caso de negócios detalhado baseado em ferramentas, incluindo modelagem de custos, motivadores e portfólio. Isso poderá ser ignorado se for realizado com o nível de detalhes necessário durante a fase avaliar.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Descoberta detalhada",
			Description:   "Realiza uma descoberta detalhada baseada em ferramentas para entender o consumo de recursos, o mapeamento de dependências e a análise de funções da aplicação. Isso poderá ser ignorado se for realizado com o nível de detalhes necessário durante a fase avaliar.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Planejamento e governança",
			Description:   "Inclui planejamento da migração e marcação da modernização, escopos, agendamento, recursos, gerenciamento de problemas e riscos e comunicação com as partes interessadas.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Landing Zone",
			Description:   "Implementa um ambiente da AWS bem arquitetado e de várias contas para receber as workloads que serão migradas ou modernizadas. Isso fornece uma referência para gerenciamento de identidade e acesso, governança, segurança de dados, design de rede e registro em log.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Segurança e conformidade",
			Description:   "Implementa os requisitos coletados com o cliente para gerenciamento de identidade e acesso, registro em log, segurança de infraestrutura, proteção de dados e resposta a incidentes. Executa uma validação de segurança após a implementação dos controles e adota frameworks específicos do setor (SOC, FedRAMP, PCI, HIPPA etc.).",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Modelo operacional",
			Description:   "Define os proprietários e implementa as mudanças necessárias no modelo operacional de nuvem para ferramentas, pessoas e processos necessários.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Pessoas: habilidades, cultura, mudança e liderança",
			Description:   "Aborda os planos necessários de capacitação da equipe para aproveitar os benefícios da nuvem. Pode envolver o design de um Centro de Excelência da Nuvem (CCoE) para centralizar recursos humanos e processos.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
		{
			Workflow:      "Experiência de migração e modernização",
			Description:   "Identifica e executa as workloads iniciais e as ondas de migração ou modernização a serem realizadas para acelerar o sucesso da entrega da migração inicial de acordo com o plano de validação dos outros fluxos de trabalho implementados da fase mobilizar.",
			InScope:       true,
			EffortDays:    1,
			EffortPercent: 13,
		},
	}
	total := 0.0
	for _, t := range tasks {
		total += t.EffortDays
	}
	return types.MAPPhase{
		Name:       "Mobilize",
		Tasks:      tasks,
		TotalDays:  total,
		TotalWeeks: total / 5,
	}
}
