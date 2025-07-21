# QPlayground CLI Tool

A standalone CLI tool for running exported automation configurations from QPlayground.

## Usage

```bash
# Run an exported automation config
go run cmd/qplayground/main.go -config /path/to/automation_config.json

# Or build and run
go build -o qplayground cmd/qplayground/*.go
./qplayground -config /path/to/automation_config.json
```

## Features

- Runs automation configurations exported from the QPlayground web application
- Supports all Playwright actions (goto, click, fill, type, screenshot, etc.)
- Variable resolution (static, dynamic/faker, environment)
- Multi-run support (sequential/parallel execution)
- Local file storage for screenshots and output files
- Structured logging to both terminal and JSON log files

## Output Structure

```
cmd/qplayground/
├── logs/
│   └── run-<timestamp>/
│       ├── automation.log      # Structured JSON logs
│       └── final_logs.json     # Complete execution logs
└── files/
    └── run-<timestamp>/
        ├── screenshot_1.png    # Screenshots
        └── other_files...      # Other output files
```

## Dependencies

- Go 1.25+
- Playwright for Go
- gofakeit for fake data generation

## Installation

1. Install Playwright browsers:
```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install
```

2. Run the CLI tool:
```bash
go run cmd/qplayground/main.go -config your_config.json
```

## Configuration Format

The CLI tool expects a JSON configuration file exported from the QPlayground web application. The format includes:

- Automation metadata (name, description)
- Configuration (variables, multi-run settings, timeouts, etc.)
- Steps with ordered actions
- Action configurations with resolved parameters

## Notes

- R2 storage fields in screenshot actions are ignored - all files are saved locally
- Notifications are logged but not sent (no external service dependencies)
- The tool is completely independent of the main QPlayground application