# seidor-aws-cli Uso

Este documento explica rapidamente como usar o `seidor-aws-cli` para gerar artefatos de incentivo MAP e criar um estimate no AWS Pricing Calculator.

```
seidor-aws-cli map wizard --out ./out
```

O comando acima pergunta pelo cliente, descrição do deal, região e ARR. Com essas informações cria um estimate oficial no AWS Pricing Calculator e gera os arquivos `MAP.md`, `MAP.xlsx` e `MAP.txt` no diretório especificado. O arquivo `MAP.txt` contém também o link do estimate criado.

