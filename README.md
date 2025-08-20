# seidor-aws-cli

CLI that automates AWS MAP incentive requests and pricing estimates.

## Build

```bash
make build
```

## Test

```bash
make test
```

## IAM Permissions

The AWS user requires:

- `bcm-pricing-calculator:CreateWorkloadEstimate`
- `bcm-pricing-calculator:CreateWorkloadEstimateUsage`
- `bcm-pricing-calculator:GetWorkloadEstimate`
