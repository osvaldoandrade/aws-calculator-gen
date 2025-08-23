## SEIDOR CLOUD

Command‑line utility to generate AWS Pricing Calculator estimates and MAP incentive artifacts.

## Features

- Create official AWS Pricing Calculator estimates
- Produce MAP funding summaries in Markdown, Excel and text formats
- Interactive wizard for quick data collection
- Falls back to Bill estimates when the AWS Pricing Calculator can't create a workload estimate

## Installation

Requires Go 1.24 or newer.

```bash
go install github.com/example/seidor-cloud@latest
```

## Quick example

Run the interactive MAP wizard and write artifacts to `./out`:

```bash
seidor-cloud map wizard --out ./out
```

## Development

Use `make` targets for common tasks:

```bash
make build   # compile the project
make test    # run unit tests
```

