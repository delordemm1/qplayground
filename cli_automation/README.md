# QPlayground CLI Automation

A completely standalone CLI tool for running web automation tests in headless environments, particularly suited for CI/CD pipelines and GitHub Actions. This tool provides powerful browser automation and API testing capabilities without requiring the full QPlayground web interface, with comprehensive reporting and cloud storage support.

## 🚀 Features

### Core Capabilities
- **Headless Browser Automation**: Powered by Playwright for reliable web testing
- **API Testing**: Comprehensive HTTP client with authentication and data extraction
- **Multi-User Simulation**: Support for concurrent user simulation with parallel/sequential execution
- **Comprehensive Reporting**: Generates detailed HTML, JSON, and CSV reports with performance metrics
- **Cloud Storage Integration**: Support for Cloudflare R2 and GCP storage for reports and screenshots
- **GitHub Actions Ready**: Built-in matrix execution support with consolidated reporting
- **Variable Resolution**: Static, dynamic (faker), and runtime variable support
- **Slack Notifications**: Real-time notifications for automation completion and failures
- **Report Consolidation**: Automatically consolidates results from multiple runners
- **Local File Storage**: Saves screenshots and reports locally without external dependencies
- **Docker Ready**: Runs in containerized environments with all dependencies included
- **GitHub Actions Integration**: Ready-to-use workflow for CI/CD automation

### Supported Action Types

#### Playwright Actions
- **Screenshots**: `playwright:screenshot` with R2/GCP storage integration
- **Navigation**: `playwright:goto`, `playwright:reload`, `playwright:go_back`, `playwright:go_forward`
- **Interaction**: `playwright:click`, `playwright:fill`, `playwright:type`, `playwright:press`, `playwright:hover`
- **Form Controls**: `playwright:check`, `playwright:uncheck`, `playwright:select_option`
- **Waiting**: `playwright:wait_for_selector`, `playwright:wait_for_timeout`, `playwright:wait_for_load_state`
- **Data Extraction**: `playwright:get_text`, `playwright:get_attribute`
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

#### Global Actions
- **Grouping**: `global:group` for organizing related actions
- **Conditional Logic**: `global:if_else` with complex condition support
- **Looping**: `global:loop` with multiple exit conditions

#### Storage Actions
- **Cloud Storage**: `r2:upload`, `r2:delete` for Cloudflare R2 and GCP operations

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

## 🐳 Docker Usage

### Building the Image

```bash
cd cli_automation
docker build -t qplayground-cli .
```

### Running Automations

#### Basic Usage
```bash
# Single automation with local storage
./qplayground-cli \
  --config-path config.json \
  --output-dir ./output
```

#### With Cloud Storage
```bash
# Single automation
docker run --rm \
  -v /path/to/config.json:/app/config.json:ro \
  -v /path/to/output:/app/output \
  -e STORAGE_PROVIDER=r2 \
  -e R2_ACCESS_KEY_ID=your_key \
  -e R2_SECRET_ACCESS_KEY=your_secret \
  -e R2_BUCKET_NAME=your_bucket \
  -e R2_PUBLIC_URL=https://your-bucket.r2.dev \
  -e CLOUDFLARE_ACCOUNT_ID=your_account_id \
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

#### GitHub Actions Matrix Execution
```bash
# Run as part of GitHub Actions matrix (handled automatically)
./qplayground-cli \
  --config-path config.json \
  --output-dir ./output \
  --runner-index 0

# Consolidate reports from multiple runners
./qplayground-cli \
  --output-dir ./consolidated-output \
  --consolidate
```

## 🔄 GitHub Actions Integration

The included workflow supports:
- **Matrix Execution**: 10 parallel runners by default
- **Concurrent Runs**: Each runner executes multiple concurrent automations
- **Report Consolidation**: Automatically combines results from all runners
- **Cloud Storage**: Uploads reports and screenshots to R2 or GCP
- **Artifact Management**: Saves all results as GitHub Actions artifacts

### Workflow Configuration

```yaml
name: Load Test with Consolidated Reports

on:
  workflow_dispatch:
    inputs:
      concurrent_runs:
        description: 'Number of concurrent runs per runner'
        default: '2'
      storage_provider:
        description: 'Storage provider (local, r2, gcp)'
        default: 'local'

env:
  STORAGE_PROVIDER: ${{ github.event.inputs.storage_provider }}
  # Add your storage credentials as repository secrets

jobs:
  load-test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        runner_index: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
      fail-fast: false

    steps:
      - name: Run automation
        # ... (see full workflow file)

  consolidate-reports:
    needs: load-test
    runs-on: ubuntu-latest
    if: always()
    
    steps:
      - name: Consolidate all results
        # Combines results from all 10 runners
      - name: Generate final reports
        # Creates consolidated HTML, JSON, and CSV reports
      - name: Upload to cloud storage
        # Uploads consolidated results to R2 or GCP
```

## 🔧 Configuration

### Environment Variables

```bash
# Storage Provider Selection
STORAGE_PROVIDER=r2  # or 'gcp' or 'local'
CLI_STORAGE_PROVIDER=r2  # Override for CLI specifically

# Cloudflare R2 Configuration
CLOUDFLARE_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=your_bucket_name
R2_PUBLIC_URL=https://your-bucket.r2.dev

# GCP Storage Configuration
GCP_BUCKET_ACCESS_KEY=your_access_key
GCP_BUCKET_SECRET=your_secret_key
GCP_BUCKET_NAME=your_bucket_name
GCP_BUCKET_ENDPOINT_URL=https://storage.googleapis.com
GCP_BUCKET_PUBLIC_URL=https://storage.googleapis.com/your_bucket_name

# SMTP Configuration (for email notifications)
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_EMAIL=your_email@gmail.com

# GitHub Actions (automatically set)
GITHUB_RUN_ID=123456789
GITHUB_RUN_NUMBER=42
RUNNER_INDEX=0
```

### Automation Configuration

The CLI supports the full automation configuration format from the main application:

```json
{
  "automation": {
    "name": "CLI Load Test",
    "description": "Comprehensive load testing via CLI",
    "config": {
      "variables": [
        {
          "key": "testEmail",
          "type": "dynamic",
          "value": "{{faker.email}}"
        }
      ],
      "multirun": {
        "enabled": true,
        "mode": "parallel",
        "count": 5,
        "delay": 1000
      },
      "screenshots": {
        "enabled": true,
        "onError": true,
        "onSuccess": true
      },
      "notifications": [
        {
          "id": "slack-alerts",
          "type": "slack",
          "onComplete": true,
          "onError": true,
          "config": {
            "webhook_url": "https://hooks.slack.com/services/...",
            "username": "QPlayground CLI Bot"
          }
        }
      ]
    }
  },
  "steps": [
    {
      "name": "Navigation and Screenshots",
      "step_order": 1,
      "actions": [
        {
          "action_type": "playwright:goto",
          "action_config": {
            "url": "https://example.com"
          }
        },
        {
          "action_type": "playwright:screenshot",
          "action_config": {
            "full_page": true,
            "format": "png",
            "upload_to_r2": true,
            "r2_key": "screenshots/cli-test-{{loopIndex}}.png"
          }
        }
      ]
    }
  ]
}
```

## 📊 Output Structure

### Individual Runner Output

```
output/runner-0/
├── reports/
│   ├── detailed_report.html
│   ├── user_journey_report.html
│   ├── report.json
│   └── report.csv
└── screenshots/
    ├── cli-user-0-step1.png
    └── cli-user-1-step1.png
```

### Consolidated Output

```
consolidated-output/
├── reports/
│   ├── consolidated_detailed_report.html
│   ├── consolidated_user_journey_report.html
│   ├── consolidated_report.json
│   └── consolidated_report.csv
├── screenshots/
│   ├── runner-0/
│   ├── runner-1/
│   └── ...
├── raw-data/
│   ├── runner-0-report.json
│   ├── runner-1-report.json
│   └── ...
└── consolidation_summary.json
```

### Report Contents

#### HTML Reports
- **Detailed Report**: Comprehensive analysis with performance metrics, step breakdown, and action statistics
- **User Journey Report**: Individual user flow analysis with timeline visualization
- Performance metrics and charts
- Screenshot galleries
- Error details and stack traces
- Bootstrap-based responsive design

#### JSON Report
- Complete automation execution data
- Structured logs and events
- Performance metrics
- Variable states and transitions
- Consolidated data from all runners

#### CSV Report
- Tabular data for spreadsheet analysis
- Step and action timing
- Success/failure rates
- Output file references
- Cross-runner analysis data

## 🔧 Command Line Options

```bash
qplayground-cli [OPTIONS]

Options:
  --config-path string    Path to the automation configuration JSON file (required)
  --output-dir string     Directory to save reports and screenshots (required)
  --runner-index string   Index of this runner for GitHub Actions matrix (default: "0")
  --consolidate          Consolidate reports from multiple runners
  --help                  Show help information
```

### Examples

```bash
# Basic usage
./qplayground-cli \
  --config-path automation.json \
  --output-dir ./results

# GitHub Actions matrix runner
./qplayground-cli \
  --config-path automation.json \
  --output-dir ./results \
  --runner-index 3

# Consolidate multiple runner results
./qplayground-cli \
  --output-dir ./consolidated-results \
  --consolidate

# Docker usage
docker run --rm \
  -v $(pwd)/config.json:/app/config.json:ro \
  -v $(pwd)/output:/app/output \
  -e STORAGE_PROVIDER=r2 \
  qplayground-cli \
  --config-path /app/config.json \
  --output-dir /app/output

# With cloud storage
./qplayground-cli \
  --config-path config.json \
  --output-dir ./output
# (Storage configured via environment variables)
```

## 🌐 Cloud Storage Integration

### Cloudflare R2

```bash
export STORAGE_PROVIDER=r2
export CLOUDFLARE_ACCOUNT_ID=your_account_id
export R2_ACCESS_KEY_ID=your_access_key
export R2_SECRET_ACCESS_KEY=your_secret_key
export R2_BUCKET_NAME=qplayground-reports
export R2_PUBLIC_URL=https://qplayground-reports.r2.dev
```

### Google Cloud Storage

```bash
export STORAGE_PROVIDER=gcp
export GCP_BUCKET_ACCESS_KEY=your_access_key
export GCP_BUCKET_SECRET=your_secret_key
export GCP_BUCKET_NAME=qplayground-reports
export GCP_BUCKET_ENDPOINT_URL=https://storage.googleapis.com
export GCP_BUCKET_PUBLIC_URL=https://storage.googleapis.com/qplayground-reports
```

## 📢 Notifications

### Slack Integration

Configure Slack notifications in your automation config:

```json
{
  "notifications": [
    {
      "id": "team-alerts",
      "type": "slack",
      "onComplete": true,
      "onError": true,
      "config": {
        "webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
        "username": "QPlayground CLI Bot",
        "icon_emoji": ":robot_face:",
        "channel": "#automation-alerts"
      }
    }
  ]
}
```

## 📈 Performance Metrics

The CLI generates comprehensive performance insights:

### Step Performance
- **Average Duration**: Mean execution time per step
- **Failure Rates**: Percentage of failed executions
- **Concurrent User Analysis**: How steps perform under load
- **P50/P95 Percentiles**: Performance distribution analysis
- **Cross-Runner Comparison**: Performance across different runners

### Run Latency
- **User Journey Mapping**: Individual user execution paths
- **Bottleneck Identification**: Slowest steps and actions
- **Scalability Analysis**: Performance trends across multiple runs
- **Resource Utilization**: Memory and CPU usage patterns
- **Matrix Execution Analysis**: Performance across GitHub Actions matrix

### Success/Failure Analysis
- **Overall Reliability**: Automation success rates
- **Error Categorization**: Common failure patterns
- **Recovery Metrics**: Retry success rates
- **Trend Analysis**: Performance over time
- **Runner Reliability**: Success rates per runner

### Consolidated Metrics
- **Total Execution Time**: Across all runners and runs
- **Throughput Analysis**: Runs per minute/hour
- **Resource Efficiency**: Cost per successful run
- **Scalability Insights**: Performance vs. concurrency

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

#### Cloud Storage Issues
```bash
# Test R2 connectivity
export STORAGE_PROVIDER=r2
# ... set R2 credentials
./qplayground-cli --config-path config.json --output-dir ./test-output

# Test GCP connectivity
export STORAGE_PROVIDER=gcp
# ... set GCP credentials
./qplayground-cli --config-path config.json --output-dir ./test-output
```

### Debug Mode

Enable verbose logging:
```bash
# Set log level to debug
export LOG_LEVEL=debug

# Enable storage debug logging
export STORAGE_DEBUG=true

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

# Validate storage configuration
jq '.automation.config.screenshots, .automation.config.notifications' config.json

# Check for R2 upload configurations
jq '.steps[].actions[] | select(.action_config.upload_to_r2 == true)' config.json
```

## 🔗 Integration Examples

### CI/CD Pipeline Integration

#### Jenkins
```groovy
pipeline {
    agent any
    environment {
        STORAGE_PROVIDER = 'r2'
        // Add storage credentials
    }
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
                    reportFiles: 'detailed_report.html',
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
  variables:
    STORAGE_PROVIDER: "r2"
  script:
    - qplayground-cli --config-path automation.json --output-dir reports
  artifacts:
    reports:
      junit: reports/report.xml
    paths:
      - reports/
    expire_in: 1 week
```

### Monitoring Integration

#### Prometheus Metrics
```bash
# Export metrics from JSON report
jq -r '.metrics | to_entries[] | "\(.key) \(.value)"' consolidated_report.json > metrics.prom
```

#### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "QPlayground CLI Load Test Metrics",
    "panels": [
      {
        "title": "Overall Success Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "cli_automation_success_rate"
          }
        ]
      },
      {
        "title": "Runner Performance",
        "type": "graph",
        "targets": [
          {
            "expr": "cli_automation_duration_by_runner"
          }
        ]
      }
    ]
  }
}
```

## 🚀 Advanced Features

### Matrix Execution Strategy

The CLI supports sophisticated matrix execution:

- **Parallel Runners**: Up to 10 concurrent GitHub Actions runners
- **Concurrent Runs**: Each runner executes multiple automation instances
- **Load Distribution**: Automatic load balancing across runners
- **Failure Isolation**: Individual runner failures don't affect others
- **Result Aggregation**: Intelligent consolidation of all results

### Report Consolidation

The consolidation process:

1. **Collection**: Gathers reports from all runners
2. **Merging**: Combines performance metrics and logs
3. **Analysis**: Generates cross-runner insights
4. **Visualization**: Creates unified HTML reports
5. **Storage**: Uploads consolidated results to cloud storage

### Cloud-First Architecture

- **Storage Abstraction**: Seamless switching between storage providers
- **Automatic Uploads**: Reports and screenshots uploaded in real-time
- **Public URLs**: Direct links to cloud-hosted reports
- **Backup Strategy**: Local storage as fallback
- **Cost Optimization**: Efficient storage usage patterns

## 📚 Advanced Usage

### Multi-Environment Testing

```bash
# Production load test
STORAGE_PROVIDER=r2 ./qplayground-cli \
  --config-path configs/production.json \
  --output-dir ./prod-results

# Staging validation
STORAGE_PROVIDER=gcp ./qplayground-cli \
  --config-path configs/staging.json \
  --output-dir ./staging-results

# Development smoke test
STORAGE_PROVIDER=local ./qplayground-cli \
  --config-path configs/dev.json \
  --output-dir ./dev-results
```

### Custom Authentication Flows

```json
{
  "steps": [
    {
      "name": "API Authentication with Storage",
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
        },
        {
          "action_type": "playwright:screenshot",
          "action_config": {
            "upload_to_r2": true,
            "r2_key": "auth/oauth-success-{{loopIndex}}.png"
          }
        }
      ]
    }
  ]
}
```

### Performance Testing Scenarios

```json
{
  "automation": {
    "name": "High-Load Performance Test",
    "config": {
      "multirun": {
        "enabled": true,
        "mode": "parallel",
        "count": 50,
        "delay": 100
      },
      "notifications": [
        {
          "type": "slack",
          "onComplete": true,
          "onError": true,
          "config": {
            "webhook_url": "https://hooks.slack.com/...",
            "channel": "#performance-alerts"
          }
        }
      ]
    }
  }
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

# Set up environment
cp .env.example .env
# Edit .env with your configuration

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

### Adding New Storage Providers

1. **Implement the `ObjectStorage` interface**:
   ```go
   type MyStorage struct{}
   
   func (s *MyStorage) Upload(ctx context.Context, key string, data io.Reader, options *UploadOptions) error {
       // Implementation
   }
   ```

2. **Register in the main function**:
   ```go
   case "mystorage":
       objectStorage, err = storage.NewMyStorage()
   ```

3. **Add environment variables** in `platform/env.go`
4. **Update documentation** and examples

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

## 🎯 Use Cases

### CI/CD Pipeline Integration
- **Pre-deployment Testing**: Validate functionality before releases
- **Regression Testing**: Ensure new changes don't break existing features
- **Performance Monitoring**: Track application performance over time
- **Load Testing**: Simulate high user loads

### Quality Assurance
- **Cross-browser Testing**: Validate across different environments
- **User Journey Validation**: Test complete user workflows
- **API Integration Testing**: Validate backend integrations
- **Visual Regression Testing**: Detect UI changes

### Monitoring and Alerting
- **Health Checks**: Regular application health monitoring
- **SLA Monitoring**: Track service level agreements
- **Performance Baselines**: Establish performance benchmarks
- **Incident Response**: Automated testing during incidents

## 📄 License

This CLI tool is part of the QPlayground project and follows the same licensing terms.

---

**QPlayground CLI** - Powerful, scalable automation testing for modern CI/CD pipelines 🚀