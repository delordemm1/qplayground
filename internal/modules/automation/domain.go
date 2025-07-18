package automation

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/playwright-community/playwright-go"
)

// RunContext holds shared resources and state for a single automation run.
// This will be passed to each plugin action.
type RunContext struct {
	PlaywrightBrowser playwright.Browser
	PlaywrightPage    playwright.Page
	StorageService    storage.StorageService
	Logger            *slog.Logger
}

// PluginAction defines the interface for any executable action provided by a plugin.
type PluginAction interface {
	// Execute performs the action.
	// actionConfig: The specific configuration for this action (from automation_actions.action_config_json).
	// runContext: Shared context for the entire automation run.
	Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error
}

// ActionFactory is a function type that creates a new instance of a PluginAction.
type ActionFactory func() PluginAction

// Global registry for plugin actions
var actionRegistry = make(map[string]ActionFactory)

// RegisterAction registers a new plugin action type with its factory function.
func RegisterAction(actionType string, factory ActionFactory) {
	actionRegistry[actionType] = factory
}

// GetAction retrieves a plugin action instance by type.
func GetAction(actionType string) (PluginAction, error) {
	factory, exists := actionRegistry[actionType]
	if !exists {
		return nil, fmt.Errorf("unknown action type: %s", actionType)
	}
	return factory(), nil
}

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

	// Order management
	GetStepByID(ctx context.Context, id string) (*AutomationStep, error)
	GetActionByID(ctx context.Context, id string) (*AutomationAction, error)
	GetStepByAutomationIDAndOrder(ctx context.Context, automationID string, order int) (*AutomationStep, error)
	GetActionByStepIDAndOrder(ctx context.Context, stepID string, order int) (*AutomationAction, error)
	GetMaxStepOrder(ctx context.Context, automationID string) (int, error)
	GetMaxActionOrder(ctx context.Context, stepID string) (int, error)
	ShiftStepOrders(ctx context.Context, automationID string, startOrder, endOrder int, increment bool) error
	ShiftActionOrders(ctx context.Context, stepID string, startOrder, endOrder int, increment bool) error
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

	// Order management helpers
	GetMaxStepOrder(ctx context.Context, automationID string) (int, error)
	GetMaxActionOrder(ctx context.Context, stepID string) (int, error)
	
	// Run cache management
	UpdateRunStatus(ctx context.Context, runID, status string) error
}
