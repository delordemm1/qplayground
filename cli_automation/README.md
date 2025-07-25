# QPlayground CLI Automation

A standalone CLI tool for running web automation tests in headless environments, particularly suited for CI/CD pipelines and GitHub Actions. This tool provides powerful browser automation and API testing capabilities without requiring the full QPlayground web interface.

## 🚀 Features

### Core Capabilities
- **Headless Browser Automation**: Powered by Playwright for reliable web testing
- **API Testing**: Comprehensive HTTP client with authentication and data extraction
- **Multi-User Simulation**: Support for concurrent user simulation with parallel/sequential execution
- **Comprehensive Reporting**: Generates HTML, JSON, and CSV reports with performance metrics
- **Variable Resolution**: Static, dynamic (faker), and runtime variable support
- **Local File Storage**: Saves screenshots and reports locally without external dependencies
- **Docker Ready**: Runs in containerized environments with all dependencies included
- **GitHub Actions Integration**: Ready-to-use workflow for CI/CD automation

### Supported Action Types

#### Playwright Actions
- **Navigation**: `playwright:goto`, `playwright:reload`, `playwright:go_back`, `playwright:go_forward`
- **Interaction**: `playwright:click`, `playwright:fill`, `playwright:type`, `playwright:press`, `playwright:hover`
- **Form Controls**: `playwright:check`, `playwright:uncheck`, `playwright:select_option`
- **Waiting**: `playwright:wait_for_selector`, `playwright:wait_for_timeout`, `playwright:wait_for_load_state`
- **Data Extraction**: `playwright:get_text`, `playwright:get_attribute`
- **Screenshots**: `playwright:screenshot` with local storage
- **JavaScript**: `playwright:evaluate` for custom browser scripts
- **Viewport**: `playwright:set_viewport`, `playwright:scroll`
- **Control Flow**: `playwright:if_else`, `playwright:loop_until`
- **Logging**: `playwright:log` with variable support

#### API Actions
- **HTTP Methods**: `api:get`, `api:post`, `api:put`, `api:patch`, `api:delete`
- **Authentication**: Bearer, Basic, API Key, Custom auth support
- **Data Extraction**: After-hooks for extracting data from API responses
- **Conditional Logic**: `api:if_else` based on runtime variables
- **Runtime Loops**: `api:runtime_loop_until` for polling scenarios
- **Logging**: `api:log` with runtime variable interpolation

#### Storage Actions
- **Local Storage**: `r2:upload`, `r2:delete` for file operations (adapted for local storage)

## 🛠️ Installation

### Using Docker (Recommended)

1. **Build the Docker image**:
   ```bash
   cd cli_automation
   docker build -t qplayground-cli .
   ```

2. **Run an automation**:
   ```bash
   docker run --rm \
     -v /path/to/your/config.json:/app/config.json:ro \
     -v /path/to/output:/app/output \
     qplayground-cli \
     --config-path /app/config.json \
     --output-dir /app/output
   ```

### Local Development

1. **Install dependencies**:
   ```bash
   cd cli_automation
   go mod download
   ```

2. **Install Playwright browsers**:
   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
   ```

3. **Run the CLI**:
   ```bash
   go run cmd/main.go \
     --config-path sample_automation.json \
     --output-dir ./output
   ```

### Binary Installation

1. **Build the binary**:
   ```bash
   cd cli_automation
   go build -o qplayground-cli cmd/main.go
   ```

2. **Run the binary**:
   ```bash
   ./qplayground-cli \
     --config-path config.json \
     --output-dir ./output
   ```

## 📋 Configuration Format

The CLI accepts automation configurations exported from the main QPlayground application or created manually. Here's the structure:

### Basic Configuration

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
        },
        {
          "key": "baseUrl",
          "type": "static",
          "value": "https://example.com",
          "description": "Base URL for testing"
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
      },
      "notifications": []
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
            "url": "{{baseUrl}}/login"
          },
          "action_order": 1
        }
      ]
    }
  ]
}
```

### Advanced Configuration with API Testing

```json
{
  "automation": {
    "name": "API and Browser Testing",
    "description": "Combined API and browser automation",
    "config": {
      "variables": [
        {
          "key": "apiKey",
          "type": "static",
          "value": "your-api-key-here"
        }
      ],
      "multirun": {
        "enabled": true,
        "mode": "sequential",
        "count": 3,
        "delay": 2000
      },
      "timeout": 600,
      "retries": 1,
      "screenshots": {
        "enabled": true,
        "onError": true,
        "onSuccess": true
      }
    }
  },
  "steps": [
    {
      "name": "API Authentication",
      "step_order": 1,
      "actions": [
        {
          "id": "auth-1",
          "action_type": "api:post",
          "action_config": {
            "url": "https://api.example.com/auth/login",
            "headers": {
              "Content-Type": "application/json"
            },
            "body": "{\"apiKey\": \"{{apiKey}}\"}",
            "after_hooks": [
              {
                "path": "data.accessToken",
                "save_as": "access_token",
                "scope": "global"
              }
            ]
          },
          "action_order": 1
        }
      ]
    },
    {
      "name": "Browser Testing with API Data",
      "step_order": 2,
      "actions": [
        {
          "id": "browser-1",
          "action_type": "playwright:goto",
          "action_config": {
            "url": "https://example.com/dashboard"
          },
          "action_order": 1
        },
        {
          "id": "conditional-1",
          "action_type": "api:if_else",
          "action_config": {
            "variable_path": "runtime.access_token",
            "condition_type": "is_not_null",
            "if_actions": [
              {
                "action_type": "api:log",
                "action_config": {
                  "message": "Successfully authenticated with token: {{runtime.access_token}}",
                  "level": "info"
                }
              }
            ],
            "else_actions": [
              {
                "action_type": "api:log",
                "action_config": {
                  "message": "Authentication failed - no access token received",
                  "level": "error"
                }
              }
            ]
          },
          "action_order": 2
        }
      ]
    }
  ]
}
```

## 🔧 Variable System

### Variable Types

#### Static Variables
```json
{
  "key": "baseUrl",
  "type": "static",
  "value": "https://example.com"
}
```

#### Dynamic Variables (Faker)
```json
{
  "key": "userEmail",
  "type": "dynamic", 
  "value": "{{faker.email}}"
}
```

Available faker methods: `name`, `firstName`, `lastName`, `email`, `phone`, `address`, `company`, `username`, `password`, `uuid`, `number`, `date`

#### Environment Variables
```json
{
  "key": "timestamp",
  "type": "environment",
  "value": "{{timestamp}}"
}
```

Available environment variables: `loopIndex`, `localLoopIndex`, `timestamp`, `runId`, `projectId`, `automationId`

#### Runtime Variables
Extract data from API responses:
```json
{
  "after_hooks": [
    {
      "path": "data.user.id",
      "save_as": "user_id",
      "scope": "local"
    },
    {
      "path": "data.session.token",
      "save_as": "access_token",
      "scope": "global"
    }
  ]
}
```

Use in subsequent actions:
```json
{
  "url": "https://api.example.com/users/{{runtime.user_id}}",
  "headers": {
    "Authorization": "Bearer {{runtime.access_token}}"
  }
}
```

## 🔄 Multi-User Simulation

Configure concurrent user simulation:

### Parallel Execution
```json
{
  "multirun": {
    "enabled": true,
    "mode": "parallel",
    "count": 10,
    "delay": 1000
  }
}
```

### Sequential Execution
```json
{
  "multirun": {
    "enabled": true,
    "mode": "sequential",
    "count": 5,
    "delay": 2000
  }
}
```

## 🎯 Action Examples

### Navigation Actions
```json
{
  "action_type": "playwright:goto",
  "action_config": {
    "url": "https://example.com",
    "timeout": 30000,
    "wait_until": "networkidle"
  }
}
```

### Interaction Actions
```json
{
  "action_type": "playwright:fill",
  "action_config": {
    "selector": "input[name='email']",
    "value": "{{testEmail}}"
  }
}
```

### API Actions
```json
{
  "action_type": "api:get",
  "action_config": {
    "url": "https://api.example.com/users",
    "headers": {
      "Authorization": "Bearer {{runtime.access_token}}"
    },
    "after_hooks": [
      {
        "path": "data[0].id",
        "save_as": "first_user_id",
        "scope": "local"
      }
    ]
  }
}
```

### Conditional Logic
```json
{
  "action_type": "playwright:if_else",
  "action_config": {
    "selector": "#submit-button",
    "condition_type": "is_enabled",
    "if_actions": [
      {
        "action_type": "playwright:click",
        "action_config": {"selector": "#submit-button"}
      }
    ],
    "else_actions": [
      {
        "action_type": "playwright:log",
        "action_config": {
          "message": "Submit button is disabled",
          "level": "warn"
        }
      }
    ]
  }
}
```

### Loop Actions
```json
{
  "action_type": "api:runtime_loop_until",
  "action_config": {
    "variable_path": "runtime.job_status.completed",
    "condition_type": "equals",
    "expected_value": true,
    "max_loops": 30,
    "timeout_ms": 60000,
    "fail_on_force_stop": false,
    "loop_actions": [
      {
        "action_type": "api:get",
        "action_config": {
          "url": "https://api.example.com/job/{{runtime.job_id}}/status",
          "after_hooks": [
            {
              "path": "data",
              "save_as": "job_status",
              "scope": "local"
            }
          ]
        }
      },
      {
        "action_type": "api:log",
        "action_config": {
          "message": "Job status: {{runtime.job_status.status}} - Progress: {{runtime.job_status.progress}}%"
        }
      },
      {
        "action_type": "playwright:wait_for_timeout",
        "action_config": {"timeout_ms": 5000}
      }
    ]
  }
}
```

## 📊 Output Structure

The CLI generates organized output in the specified directory:

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

### Report Contents

#### HTML Report
- Interactive dashboard with step-by-step breakdown
- Performance metrics and charts
- Screenshot galleries
- Error details and stack traces
- User journey visualization

#### JSON Report
- Complete automation execution data
- Structured logs and events
- Performance metrics
- Variable states and transitions

#### CSV Report
- Tabular data for spreadsheet analysis
- Step and action timing
- Success/failure rates
- Output file references

## 🔧 Command Line Options

```bash
qplayground-cli [OPTIONS]

Options:
  --config-path string    Path to the automation configuration JSON file (required)
  --output-dir string     Directory to save reports and screenshots (required)
  --help                  Show help information
```

### Examples

```bash
# Basic usage
./qplayground-cli \
  --config-path automation.json \
  --output-dir ./results

# Docker usage
docker run --rm \
  -v $(pwd)/config.json:/app/config.json:ro \
  -v $(pwd)/output:/app/output \
  qplayground-cli \
  --config-path /app/config.json \
  --output-dir /app/output

# With custom configuration
./qplayground-cli \
  --config-path ./configs/production-test.json \
  --output-dir ./reports/$(date +%Y%m%d)
```

## 🐳 Docker Usage

### Building the Image

```bash
cd cli_automation
docker build -t qplayground-cli .
```

### Running Automations

```bash
# Single automation
docker run --rm \
  -v /path/to/config.json:/app/config.json:ro \
  -v /path/to/output:/app/output \
  qplayground-cli \
  --config-path /app/config.json \
  --output-dir /app/output

# Batch processing
for config in configs/*.json; do
  docker run --rm \
    -v $(pwd)/$config:/app/config.json:ro \
    -v $(pwd)/output:/app/output \
    qplayground-cli \
    --config-path /app/config.json \
    --output-dir /app/output
done
```

## 🔄 GitHub Actions Integration

Add this workflow to `.github/workflows/automation.yml`:

```yaml
name: Web Automation Tests

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 */6 * * *'  # Run every 6 hours
  workflow_dispatch:       # Manual trigger

jobs:
  automation:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        config: [
          'configs/smoke-test.json',
          'configs/regression-test.json',
          'configs/performance-test.json'
        ]
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Create output directory
      run: mkdir -p automation_output
    
    - name: Run automation
      run: |
        docker run --rm \
          -v ${{ github.workspace }}/${{ matrix.config }}:/app/config.json:ro \
          -v ${{ github.workspace }}/automation_output:/app/output \
          ghcr.io/${{ github.repository }}/qplayground-cli:latest \
          --config-path /app/config.json \
          --output-dir /app/output
    
    - name: Upload reports
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: automation-reports-${{ matrix.config }}
        path: automation_output/
        retention-days: 30
    
    - name: Comment PR with results
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v7
      with:
        script: |
          const fs = require('fs');
          const path = 'automation_output/reports/report.json';
          if (fs.existsSync(path)) {
            const report = JSON.parse(fs.readFileSync(path, 'utf8'));
            const comment = `## Automation Results
            
            **Status**: ${report.run.status}
            **Duration**: ${report.run.endTime - report.run.startTime}ms
            **Steps**: ${report.automation.steps?.length || 0}
            **Success Rate**: ${report.metrics?.overallFailureRate ? (100 - report.metrics.overallFailureRate).toFixed(1) : 'N/A'}%
            `;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
          }
```

## 📈 Performance Metrics

The CLI automatically generates performance insights including:

### Step Performance
- **Average Duration**: Mean execution time per step
- **Failure Rates**: Percentage of failed executions
- **Concurrent User Analysis**: How steps perform under load
- **P50/P95 Percentiles**: Performance distribution analysis

### Run Latency
- **User Journey Mapping**: Individual user execution paths
- **Bottleneck Identification**: Slowest steps and actions
- **Scalability Analysis**: Performance trends across multiple runs
- **Resource Utilization**: Memory and CPU usage patterns

### Success/Failure Analysis
- **Overall Reliability**: Automation success rates
- **Error Categorization**: Common failure patterns
- **Recovery Metrics**: Retry success rates
- **Trend Analysis**: Performance over time

## 🛠️ Troubleshooting

### Common Issues

#### Playwright Browser Issues
```bash
# Install browsers manually
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps

# Check browser installation
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install-deps
```

#### Docker Permission Issues
```bash
# Fix volume permissions
chmod -R 755 /path/to/output
chown -R $(id -u):$(id -g) /path/to/output
```

#### Memory Issues
```bash
# Increase Docker memory limit
docker run --memory=2g --rm qplayground-cli ...

# Monitor memory usage
docker stats qplayground-cli
```

### Debug Mode

Enable verbose logging:
```bash
# Set log level to debug
export LOG_LEVEL=debug

# Run with debug output
./qplayground-cli \
  --config-path config.json \
  --output-dir ./output 2>&1 | tee debug.log
```

### Configuration Validation

Validate your configuration before running:
```bash
# Check JSON syntax
jq . config.json

# Validate required fields
jq '.automation.name, .automation.config, .steps' config.json
```

## 🔗 Integration Examples

### CI/CD Pipeline Integration

#### Jenkins
```groovy
pipeline {
    agent any
    stages {
        stage('Run Automation') {
            steps {
                script {
                    docker.image('qplayground-cli').inside {
                        sh '''
                            qplayground-cli \
                                --config-path /workspace/automation.json \
                                --output-dir /workspace/reports
                        '''
                    }
                }
                publishHTML([
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: 'reports',
                    reportFiles: 'report.html',
                    reportName: 'Automation Report'
                ])
            }
        }
    }
}
```

#### GitLab CI
```yaml
automation_test:
  image: qplayground-cli:latest
  script:
    - qplayground-cli --config-path automation.json --output-dir reports
  artifacts:
    reports:
      junit: reports/junit.xml
    paths:
      - reports/
    expire_in: 1 week
```

### Monitoring Integration

#### Prometheus Metrics
```bash
# Export metrics from JSON report
jq -r '.metrics | to_entries[] | "\(.key) \(.value)"' report.json > metrics.prom
```

#### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "QPlayground Automation Metrics",
    "panels": [
      {
        "title": "Success Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "automation_success_rate"
          }
        ]
      }
    ]
  }
}
```

## 📚 Advanced Usage

### Custom Authentication Flows

```json
{
  "steps": [
    {
      "name": "OAuth Authentication",
      "step_order": 1,
      "actions": [
        {
          "action_type": "api:post",
          "action_config": {
            "url": "https://auth.example.com/oauth/token",
            "headers": {
              "Content-Type": "application/x-www-form-urlencoded"
            },
            "body": "grant_type=client_credentials&client_id={{clientId}}&client_secret={{clientSecret}}",
            "after_hooks": [
              {
                "path": "access_token",
                "save_as": "oauth_token",
                "scope": "global"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### Complex Data Extraction

```json
{
  "action_type": "api:get",
  "action_config": {
    "url": "https://api.example.com/users",
    "after_hooks": [
      {
        "path": "data[0].id",
        "save_as": "first_user_id",
        "scope": "local"
      },
      {
        "path": "data.length",
        "save_as": "total_users",
        "scope": "local"
      },
      {
        "path": "pagination.next_page",
        "save_as": "next_page_url",
        "scope": "local"
      }
    ]
  }
}
```

### Error Handling Patterns

```json
{
  "action_type": "api:if_else",
  "action_config": {
    "variable_path": "runtime.api_response.error",
    "condition_type": "is_not_null",
    "if_actions": [
      {
        "action_type": "api:log",
        "action_config": {
          "message": "API Error: {{runtime.api_response.error.message}} (Code: {{runtime.api_response.error.code}})",
          "level": "error"
        }
      },
      {
        "action_type": "playwright:screenshot",
        "action_config": {
          "full_page": true,
          "format": "png"
        }
      }
    ],
    "final_actions": [
      {
        "action_type": "api:log",
        "action_config": {
          "message": "Continuing automation despite error..."
        }
      }
    ]
  }
}
```

## 🔧 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/your-org/qplayground.git
cd qplayground/cli_automation

# Install dependencies
go mod download

# Build the binary
go build -o qplayground-cli cmd/main.go

# Run tests
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Adding New Action Types

1. **Define the action struct**:
   ```go
   type CustomAction struct{}
   
   func (a *CustomAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
       // Implementation
       return nil
   }
   ```

2. **Register the action**:
   ```go
   func init() {
       automation.RegisterAction("custom:action", func() automation.PluginAction { 
           return &CustomAction{} 
       })
   }
   ```

3. **Add configuration validation** (optional)
4. **Update documentation**

## 📄 License

This CLI tool is part of the QPlayground project and follows the same licensing terms.

---

**QPlayground CLI** - Powerful automation testing for CI/CD pipelines 🚀