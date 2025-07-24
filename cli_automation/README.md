# QPlayground CLI Automation

A standalone CLI tool for running web automation tests in headless environments, particularly suited for CI/CD pipelines and GitHub Actions.

## Features

- **Headless Browser Automation**: Powered by Playwright for reliable web testing
- **Multi-User Simulation**: Support for concurrent user simulation with parallel/sequential execution
- **Comprehensive Reporting**: Generates HTML, JSON, and CSV reports with performance metrics
- **Variable Resolution**: Static, dynamic (faker), and environment variable support
- **Local File Storage**: Saves screenshots and reports locally without external dependencies
- **Docker Ready**: Runs in containerized environments with all dependencies included
- **GitHub Actions Integration**: Ready-to-use workflow for CI/CD automation

## Quick Start

### Using Docker (Recommended)

1. **Build the Docker image:**
   ```bash
   cd cli_automation
   docker build -t cli-automation .
   ```

2. **Run an automation:**
   ```bash
   docker run --rm \
     -v /path/to/your/config.json:/app/config.json:ro \
     -v /path/to/output:/app/output \
     cli-automation \
     --config-path /app/config.json \
     --output-dir /app/output
   ```

### Local Development

1. **Install dependencies:**
   ```bash
   cd cli_automation
   go mod download
   ```

2. **Install Playwright browsers:**
   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
   ```

3. **Run the CLI:**
   ```bash
   go run cmd/main.go \
     --config-path sample_automation.json \
     --output-dir ./output
   ```

## Configuration Format

The CLI accepts automation configurations exported from the main QPlayground application. Here's the structure:

```json
{
  "automation": {
    "name": "My Test Automation",
    "description": "Description of what this automation does",
    "config": {
      "variables": [
        {
          "key": "testEmail",
          "type": "dynamic",
          "value": "{{faker.email}}",
          "description": "Random email for testing"
        }
      ],
      "multirun": {
        "enabled": true,
        "mode": "parallel",
        "count": 5,
        "delay": 1000
      },
      "timeout": 300,
      "retries": 0,
      "screenshots": {
        "enabled": true,
        "onError": true,
        "onSuccess": false,
        "path": "screenshots/{{timestamp}}-{{loopIndex}}.png"
      }
    }
  },
  "steps": [
    {
      "name": "Navigate to Website",
      "step_order": 1,
      "actions": [
        {
          "id": "action-1",
          "action_type": "playwright:goto",
          "action_config": {
            "url": "https://example.com"
          },
          "action_order": 1
        }
      ]
    }
  ]
}
```

## Supported Actions

### Navigation Actions
- `playwright:goto` - Navigate to URL
- `playwright:reload` - Reload page
- `playwright:go_back` - Browser back
- `playwright:go_forward` - Browser forward

### Interaction Actions
- `playwright:click` - Click elements
- `playwright:fill` - Fill input fields
- `playwright:type` - Type text
- `playwright:press` - Press keys
- `playwright:check` - Check checkboxes
- `playwright:uncheck` - Uncheck checkboxes
- `playwright:select_option` - Select dropdown options
- `playwright:hover` - Hover over elements

### Waiting Actions
- `playwright:wait_for_selector` - Wait for elements
- `playwright:wait_for_timeout` - Wait for duration
- `playwright:wait_for_load_state` - Wait for page load

### Information Actions
- `playwright:get_text` - Extract text content
- `playwright:get_attribute` - Get element attributes
- `playwright:screenshot` - Take screenshots

### Control Flow Actions
- `playwright:if_else` - Conditional logic
- `playwright:loop_until` - Loop with conditions
- `playwright:log` - Custom logging

### Utility Actions
- `playwright:evaluate` - Execute JavaScript
- `playwright:scroll` - Scroll page
- `playwright:set_viewport` - Set browser viewport

## Variable System

### Static Variables
```json
{
  "key": "baseUrl",
  "type": "static",
  "value": "https://example.com"
}
```

### Dynamic Variables (Faker)
```json
{
  "key": "userEmail",
  "type": "dynamic", 
  "value": "{{faker.email}}"
}
```

Available faker methods: `name`, `email`, `phone`, `address`, `company`, `username`, `password`, `uuid`, `number`, `date`

### Environment Variables
```json
{
  "key": "timestamp",
  "type": "environment",
  "value": "{{timestamp}}"
}
```

Available environment variables: `loopIndex`, `localLoopIndex`, `timestamp`, `runId`, `projectId`, `automationId`

## Multi-User Simulation

Configure concurrent user simulation:

```json
{
  "multirun": {
    "enabled": true,
    "mode": "parallel",    // or "sequential"
    "count": 10,           // number of concurrent users
    "delay": 1000          // delay between runs (ms)
  }
}
```

## Output Structure

```
output/
├── YYYYMMDD-HHMMSS-<runId>/
│   ├── reports/
│   │   ├── report.html      # Interactive HTML report
│   │   ├── report.json      # Raw data dump
│   │   └── logs.csv         # CSV export
│   └── screenshots/
│       ├── screenshot1.png
│       └── screenshot2.png
```

## GitHub Actions Integration

Add this workflow to `.github/workflows/automation.yml`:

```yaml
name: Web Automation Tests

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 */6 * * *'  # Run every 6 hours

jobs:
  automation:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Run automation
      run: |
        mkdir -p automation_output
        docker run --rm \
          -v ${{ github.workspace }}/automation_config.json:/app/config.json:ro \
          -v ${{ github.workspace }}/automation_output:/app/output \
          ghcr.io/${{ github.repository }}/cli-automation:latest \
          --config-path /app/config.json \
          --output-dir /app/output
    
    - name: Upload reports
      uses: actions/upload-artifact@v4
      with:
        name: automation-reports
        path: automation_output/
```

## Performance Metrics

The CLI automatically generates performance insights including:

- **Step Performance**: Average duration and failure rates per step
- **Concurrent User Analysis**: How steps perform under load
- **Run Latency**: Performance trends across multiple runs
- **Success/Failure Rates**: Overall automation reliability metrics

## Error Handling

- Detailed error logging with step and action context
- Screenshot capture on failures (when enabled)
- Graceful handling of timeouts and cancellations
- Comprehensive error reporting in all output formats

## License

This CLI tool is part of the QPlayground project and follows the same licensing terms.