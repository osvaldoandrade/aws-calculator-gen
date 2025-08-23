## SEIDOR CLOUD Uso

Este guia demonstra como gerar artefatos de incentivo MAP e um estimate no AWS Pricing Calculator usando o `seidor-cloud`.

## Execução rápida

```bash
seidor-cloud map wizard --out ./out
```

O assistente interativo solicitará:

1. Nome do cliente
2. Descrição da oportunidade
3. Região AWS
4. Template da solução
5. Valor de ARR

Com essas informações a ferramenta:

- Cria um estimate oficial no AWS Pricing Calculator
- Gera os arquivos `MAP.md`, `MAP.xlsx` e `MAP.txt` no diretório fornecido
- Inclui no `MAP.txt` o link direto para o estimate gerado
- Caso o limite mensal do AWS Pricing Calculator seja atingido, tenta gerar o estimate via BILL

O estimate no AWS Pricing Calculator receberá automaticamente o nome `yyyymmdd-nome-cliente-descricao-deal-resumida`.

Use os artefatos produzidos para submeter e acompanhar o incentivo MAP junto ao time da AWS.

