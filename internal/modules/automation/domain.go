package automation

import (
	"context"
	"time"
)

// Automation represents an automation workflow
type Automation struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	ConfigJSON  string // JSON string containing variables, run settings, templates
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AutomationStep represents a step within an automation
type AutomationStep struct {
	ID           string
	AutomationID string
	Name         string
	StepOrder    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AutomationAction represents an action within a step
type AutomationAction struct {
	ID               string
	StepID           string
	ActionType       string // e.g., "playwright:goto", "playwright:click"
	ActionConfigJSON string // JSON string containing action-specific parameters
	ActionOrder      int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AutomationRun represents an execution of an automation
type AutomationRun struct {
	ID              string
	AutomationID    string
	Status          string // pending, running, completed, failed, cancelled
	StartTime       *time.Time
	EndTime         *time.Time
	LogsJSON        string // JSON string containing execution logs
	OutputFilesJSON string // JSON string containing file paths/URLs
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AutomationRepository defines the interface for automation data operations
type AutomationRepository interface {
	// Automation CRUD
	CreateAutomation(ctx context.Context, automation *Automation) error
	GetAutomationByID(ctx context.Context, id string) (*Automation, error)
	GetAutomationsByProjectID(ctx context.Context, projectID string) ([]*Automation, error)
	UpdateAutomation(ctx context.Context, automation *Automation) error
	DeleteAutomation(ctx context.Context, id string) error

	// Step CRUD
	CreateStep(ctx context.Context, step *AutomationStep) error
	GetStepsByAutomationID(ctx context.Context, automationID string) ([]*AutomationStep, error)
	UpdateStep(ctx context.Context, step *AutomationStep) error
	DeleteStep(ctx context.Context, id string) error

	// Action CRUD
	CreateAction(ctx context.Context, action *AutomationAction) error
	GetActionsByStepID(ctx context.Context, stepID string) ([]*AutomationAction, error)
	UpdateAction(ctx context.Context, action *AutomationAction) error
	DeleteAction(ctx context.Context, id string) error

	// Run CRUD
	CreateRun(ctx context.Context, run *AutomationRun) error
	GetRunByID(ctx context.Context, id string) (*AutomationRun, error)
	GetRunsByAutomationID(ctx context.Context, automationID string) ([]*AutomationRun, error)
	UpdateRun(ctx context.Context, run *AutomationRun) error
}

// AutomationService defines the interface for automation business logic
type AutomationService interface {
	// Automation management
	CreateAutomation(ctx context.Context, projectID, name, description, configJSON string) (*Automation, error)
	GetAutomationsByProject(ctx context.Context, projectID string) ([]*Automation, error)
	GetAutomationByID(ctx context.Context, id string) (*Automation, error)
	UpdateAutomation(ctx context.Context, automation *Automation) error
	DeleteAutomation(ctx context.Context, id string) error

	// Step management
	CreateStep(ctx context.Context, automationID, name string, stepOrder int) (*AutomationStep, error)
	GetStepsByAutomation(ctx context.Context, automationID string) ([]*AutomationStep, error)
	UpdateStep(ctx context.Context, step *AutomationStep) error
	DeleteStep(ctx context.Context, id string) error

	// Action management
	CreateAction(ctx context.Context, stepID, actionType, actionConfigJSON string, actionOrder int) (*AutomationAction, error)
	GetActionsByStep(ctx context.Context, stepID string) ([]*AutomationAction, error)
	UpdateAction(ctx context.Context, action *AutomationAction) error
	DeleteAction(ctx context.Context, id string) error

	// Run management
	TriggerRun(ctx context.Context, automationID string) (*AutomationRun, error)
	GetRunsByAutomation(ctx context.Context, automationID string) ([]*AutomationRun, error)
	GetRunByID(ctx context.Context, id string) (*AutomationRun, error)
}