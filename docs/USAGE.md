# seidor-aws-cli Uso

Este documento explica rapidamente como usar o `seidor-aws-cli` para calcular custos a partir de um arquivo `usage.yml`.

```
seidor-aws-cli pricing calc --input usage.yml
```

O comando acima lê o uso e exibe o custo mensal estimado.

Para criar um estimate oficial no AWS Pricing Calculator e obter o link do console:

```
seidor-aws-cli pricing calc --input usage.yml --aws-calc --title "Meu workload"
```

Para gerar os artefatos de incentivo MAP (planilha e markdown) interativamente:

```
seidor-aws-cli incentive map wizard --out ./out
```
