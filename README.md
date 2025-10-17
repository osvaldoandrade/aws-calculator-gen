# aws-calculator-gen

`aws-calculator-gen` is an interactive CLI utility that automates the creation of public AWS Pricing Calculator estimates for EC2 workloads. The tool applies a greedy approach to assemble resources that meet a target Annual Recurring Revenue (ARR) so sales teams can quickly generate customer-facing calculators. The current implementation provides a single subcommand:

```
aws-calculator-gen map
```

The command asks for basic opportunity information and attempts to create an estimate using browser automation.  Parameters may be supplied interactively or via the `--params` flag:

```
aws-calculator-gen map --params customer=Acme description="Test deal" region=us-east-1 arr=1200
```

The project layout follows a simple command factory architecture and uses [pterm](https://github.com/pterm/pterm) for the text UI.

> **Note**
> The DOM automation logic is provided as a skeleton and does not fully drive the AWS Pricing Calculator.  It is intended as a starting point for further development.

## Building

```
make build
```

## Running

```
make run ARGS="map"
```

## Testing

```
make test
```

