# QPlayground

A powerful playground for automating testing of web applications via Playwright with a user-friendly interface for creating and managing test cases, plus a comprehensive API for automating tests.

## 🚀 Features

### Core Automation Platform
- **Visual Automation Builder**: Create complex automation workflows through an intuitive web interface
- **Multi-User Simulation**: Support for concurrent user simulation with parallel/sequential execution
- **Real-time Monitoring**: Live progress tracking with Server-Sent Events (SSE)
- **Comprehensive Reporting**: Generate HTML, JSON, and CSV reports with performance metrics
- **Variable System**: Static, dynamic (faker), and runtime variable support
- **Conditional Logic**: Advanced if/else and loop constructs for complex workflows
- **API Integration**: Built-in HTTP client for API testing and data extraction

### Supported Action Types

#### Playwright Actions
- **Navigation**: `goto`, `reload`, `go_back`, `go_forward`
- **Interaction**: `click`, `fill`, `type`, `press`, `hover`
- **Form Controls**: `check`, `uncheck`, `select_option`
- **Waiting**: `wait_for_selector`, `wait_for_timeout`, `wait_for_load_state`
- **Data Extraction**: `get_text`, `get_attribute`
- **Screenshots**: `screenshot` with R2 storage integration
- **JavaScript**: `evaluate` for custom browser scripts
- **Viewport**: `set_viewport`, `scroll`
- **Control Flow**: `if_else`, `loop_until`
- **Logging**: `log` with variable support

#### API Actions
- **HTTP Methods**: `api:get`, `api:post`, `api:put`, `api:patch`, `api:delete`
- **Authentication**: Bearer, Basic, API Key, Custom auth support
- **Data Extraction**: After-hooks for extracting data from API responses
- **Conditional Logic**: `api:if_else` based on runtime variables
- **Runtime Loops**: `api:runtime_loop_until` for polling scenarios
- **Logging**: `api:log` with runtime variable interpolation

#### Storage Actions
- **R2 Integration**: `r2:upload`, `r2:delete` for Cloudflare R2 storage

### Advanced Features
- **Runtime Variables**: Extract and use data from API responses and page interactions
- **Multi-Run Configuration**: Execute automations with multiple concurrent users
- **Step Conditions**: Skip or run steps based on loop index or random conditions
- **Notification System**: Slack, email, and webhook notifications
- **Export/Import**: Export automation configurations for sharing or CI/CD
- **Performance Analytics**: Detailed performance metrics and visualizations

## 🏗️ Architecture

QPlayground follows a clean architecture pattern with clear separation of concerns:

```
├── cmd/                    # Application entry points
│   ├── app/               # Main web application
│   └── migrate/           # Database migration tool
├── internal/              # Internal application code
│   ├── controller/        # HTTP handlers and routing
│   ├── modules/          # Business logic modules
│   │   ├── auth/         # Authentication & authorization
│   │   ├── automation/   # Automation engine
│   │   ├── notification/ # Notification services
│   │   ├── organization/ # Organization management
│   │   ├── project/      # Project management
│   │   └── storage/      # File storage abstraction
│   ├── platform/         # Platform utilities
│   └── plugins/          # Action plugins
│       ├── playwright/   # Browser automation
│       ├── api/          # HTTP API actions
│       └── r2/           # Storage actions
├── cli_automation/       # Standalone CLI tool
├── resources/            # Frontend assets
│   └── src/             # Svelte 5 application
└── static/              # Static assets
```

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.25
- **Framework**: Chi router with Inertia.js
- **Database**: PostgreSQL with Goose migrations
- **Cache**: Redis for run state management
- **Storage**: Cloudflare R2 (S3-compatible)
- **Browser Automation**: Playwright
- **Session Management**: SCS (Secure Cookie Store)

### Frontend
- **Framework**: Svelte 5 with Runes
- **Build Tool**: Vite
- **Styling**: Tailwind CSS v4
- **UI Components**: Flowbite Svelte
- **Charts**: Chart.js with Svelte wrapper
- **Notifications**: Svelte French Toast

### Infrastructure
- **Containerization**: Docker with multi-stage builds
- **Development**: Air for hot reloading
- **Package Management**: Bun for frontend dependencies

## 🚀 Quick Start

### Prerequisites
- Go 1.25+
- Node.js 23.11.0+
- PostgreSQL 13+
- Redis 6+
- Bun package manager

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-org/qplayground.git
   cd qplayground
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**:
   ```bash
   # Backend dependencies
   go mod download
   
   # Frontend dependencies
   bun install
   ```

4. **Set up the database**:
   ```bash
   # Run migrations
   make migrate-up
   ```

5. **Build frontend assets**:
   ```bash
   bun run build
   ```

6. **Start the application**:
   ```bash
   # Development mode with hot reloading
   air
   
   # Or production mode
   go run cmd/app/main.go
   ```

7. **Access the application**:
   - Web Interface: http://localhost:8084
   - API Documentation: Available through the web interface

### Docker Setup

1. **Build and run with Docker**:
   ```bash
   docker build -t qplayground .
   docker run -p 8084:8084 qplayground
   ```

2. **Using Docker Compose** (recommended):
   ```bash
   # Create docker-compose.yml with PostgreSQL and Redis
   docker-compose up -d
   ```

## 📖 Usage Guide

### Creating Your First Automation

1. **Create a Project**:
   - Navigate to the Projects page
   - Click "New Project" and provide a name and description

2. **Create an Automation**:
   - Open your project and click "New Automation"
   - Configure variables, multi-run settings, and notifications
   - Set up screenshot and timeout preferences

3. **Add Steps and Actions**:
   - Click "New Step" to add workflow steps
   - Add actions to each step (navigation, interaction, API calls, etc.)
   - Configure action parameters and variable usage

4. **Run Your Automation**:
   - Click "Run Automation" to execute
   - Monitor progress in real-time
   - View detailed reports and performance metrics

### Variable System

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

#### Runtime Variables
Extract data from API responses or page interactions:
```json
{
  "after_hooks": [
    {
      "path": "data.user.id",
      "save_as": "user_id",
      "scope": "local"
    }
  ]
}
```

Use in subsequent actions:
```json
{
  "url": "https://api.example.com/users/{{runtime.user_id}}"
}
```

### Multi-User Simulation

Configure concurrent user simulation:
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

### Conditional Logic

#### Playwright Conditions
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
    ]
  }
}
```

#### API Conditions
```json
{
  "action_type": "api:if_else",
  "action_config": {
    "variable_path": "runtime.api_response.status",
    "condition_type": "equals",
    "expected_value": "success",
    "if_actions": [
      {
        "action_type": "api:log",
        "action_config": {
          "message": "API call successful: {{runtime.api_response.message}}"
        }
      }
    ]
  }
}
```

### Loop Actions

#### Playwright Loops
```json
{
  "action_type": "playwright:loop_until",
  "action_config": {
    "selector": ".loading-spinner",
    "condition_type": "is_hidden",
    "max_loops": 20,
    "timeout_ms": 30000,
    "loop_actions": [
      {
        "action_type": "playwright:wait_for_timeout",
        "action_config": {"timeout_ms": 1000}
      }
    ]
  }
}
```

#### API Runtime Loops
```json
{
  "action_type": "api:runtime_loop_until",
  "action_config": {
    "variable_path": "runtime.job_status.completed",
    "condition_type": "equals",
    "expected_value": true,
    "max_loops": 30,
    "timeout_ms": 60000,
    "loop_actions": [
      {
        "action_type": "api:get",
        "action_config": {
          "url": "https://api.example.com/job/{{runtime.job_id}}/status"
        }
      }
    ]
  }
}
```

## 🔧 Configuration

### Environment Variables

```bash
# Database
DATABASE_URL=postgresql://username:password@localhost:5432/qplayground

# Redis
REDIS_URL=redis://localhost:6379

# Application
APP_URL=http://localhost:8084
LOG_LEVEL=debug

# Authentication
GOOGLE_OAUTH_CLIENT_ID=your_google_client_id
GOOGLE_OAUTH_CLIENT_SECRET=your_google_client_secret

# SMTP Configuration
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_EMAIL=your_email@gmail.com

# Cloudflare R2 Configuration
CLOUDFLARE_ACCOUNT_ID=your_cloudflare_account_id
R2_ACCESS_KEY_ID=your_r2_access_key_id
R2_SECRET_ACCESS_KEY=your_r2_secret_access_key
R2_BUCKET_NAME=your_bucket_name
R2_PUBLIC_URL=https://your-bucket.your-account.r2.cloudflarestorage.com

# Automation Configuration
MAX_CONCURRENT_RUNS=5
```

### Database Migrations

```bash
# Create a new migration
make migrate-create name="add_new_feature"

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-status
```

## 📊 Monitoring and Reporting

### Real-time Monitoring
- Live progress updates via Server-Sent Events
- Step-by-step execution tracking
- Error reporting and debugging information
- Performance metrics collection

### Report Generation
- **HTML Reports**: Interactive reports with charts and visualizations
- **JSON Reports**: Raw data for programmatic analysis
- **CSV Reports**: Tabular data for spreadsheet analysis
- **Performance Analytics**: Step timing, failure rates, and user journey analysis

### Notification Channels
- **Slack**: Webhook-based notifications with rich formatting
- **Email**: SMTP-based email notifications (coming soon)
- **Webhooks**: Generic webhook notifications for custom integrations

## 🔌 Plugin System

QPlayground uses a plugin-based architecture for actions:

### Creating Custom Plugins

1. **Define the action interface**:
   ```go
   type CustomAction struct{}
   
   func (a *CustomAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
       // Implementation here
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

3. **Add frontend configuration** (optional):
   - Create a Svelte component for configuration
   - Add to `actionConfigMap.ts`

## 🧪 Testing

### Running Tests
```bash
# Run unit tests
make test

# Run with coverage
go test -race -cover ./...

# Vet code
make vet
```

### Test Automation Examples
See the `cli_automation/sample_automation.json` file for example automation configurations.

## 🚀 Deployment

### Production Deployment

1. **Build the application**:
   ```bash
   # Build frontend
   bun run build
   
   # Build backend
   go build -o qplayground cmd/app/main.go
   ```

2. **Deploy with Docker**:
   ```bash
   docker build -t qplayground .
   docker run -d -p 8084:8084 qplayground
   ```

3. **Environment Setup**:
   - Configure production database
   - Set up Redis instance
   - Configure Cloudflare R2 storage
   - Set up SMTP for notifications

### Scaling Considerations
- Use Redis for session storage and run state management
- Configure `MAX_CONCURRENT_RUNS` based on server capacity
- Set up database connection pooling
- Use CDN for static assets

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go best practices and conventions
- Use Svelte 5 runes for reactive state management
- Write comprehensive tests for new features
- Update documentation for API changes
- Follow the existing code organization patterns

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

- **Documentation**: Available in the `/docs` directory
- **Issues**: Report bugs and feature requests on GitHub
- **Discussions**: Join community discussions on GitHub Discussions
- **Email**: Contact support at support@qplayground.com

## 🗺️ Roadmap

### Upcoming Features
- [ ] Visual workflow designer with drag-and-drop interface
- [ ] Integration with popular CI/CD platforms
- [ ] Advanced scheduling and cron-based automation
- [ ] Team collaboration features
- [ ] API rate limiting and quotas
- [ ] Advanced analytics and insights
- [ ] Mobile app for monitoring
- [ ] Marketplace for community plugins

### Recent Updates
- ✅ API action plugins with runtime variable support
- ✅ Advanced conditional logic and loops
- ✅ Real-time monitoring with SSE
- ✅ Comprehensive reporting system
- ✅ Multi-user simulation capabilities
- ✅ Cloudflare R2 storage integration

---

**QPlayground** - Your Stage for Creativity & Automation 🎭