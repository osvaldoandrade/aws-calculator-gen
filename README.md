# seidor-tools

`seidor-tools` is an interactive CLI utility for creating public AWS Pricing Calculator estimates for EC2 workloads.  The current implementation provides a single subcommand:

```
seidor-tools map
```

The command asks for basic opportunity information and attempts to create an estimate using browser automation.  The project layout follows a simple command factory architecture and uses [pterm](https://github.com/pterm/pterm) for the text UI.

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
