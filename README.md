# seidor-aws-cli

Lightweight command line tool to calculate AWS pricing and generate MAP incentive artifacts.

### Quick examples

```bash
# calculate cost from usage.yml and show breakdown
seidor-aws-cli pricing calc --input usage.yml

# create AWS Pricing Calculator estimate and get console link
seidor-aws-cli pricing calc --input usage.yml --aws-calc --title "My workload"

# run interactive MAP wizard and generate MAP.md and MAP.xlsx
seidor-aws-cli incentive map wizard --out ./out
```

## Building

```
make build
```

## Testing

```
make test
```
