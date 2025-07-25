package automation

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/delordemm1/qplayground-cli/internal/storage"
	"github.com/playwright-community/playwright-go"
)

// RunEventType represents the type of event emitted during automation execution
type RunEventType string

const (
	RunEventTypeLog        RunEventType = "log"
	RunEventTypeError      RunEventType = "error"
	RunEventTypeOutputFile RunEventType = "output_file"
	RunEventTypeStep       RunEventType = "step"
)

// RunEvent represents an event emitted during automation execution
type RunEvent struct {
	Type           RunEventType           `json:"type"`
	Timestamp      time.Time              `json:"timestamp"`
	StepID         string                 `json:"step_id,omitempty"`
	StepName       string                 `json:"step_name,omitempty"`
	ActionID       string                 `json:"action_id,omitempty"`
	ActionName     string                 `json:"action_name,omitempty"`
	ParentActionID string                 `json:"parent_action_id,omitempty"`
	ActionType     string                 `json:"action_type,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Error          string                 `json:"error,omitempty"`
	OutputFile     string                 `json:"output_file,omitempty"`
	Duration       int64                  `json:"duration_ms,omitempty"`
	LoopIndex      int                    `json:"loop_index,omitempty"`
	LocalLoopIndex int                    `json:"local_loop_index,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
}

// RunContext holds shared resources and state for a single automation run.
// This will be passed to each plugin action.
type RunContext struct {
	PlaywrightBrowser playwright.Browser
	PlaywrightPage    playwright.Page
	StorageService    storage.StorageService
	Logger            *slog.Logger
	EventCh           chan RunEvent
	StepName          string            // Current step name for context
	StepID            string            // Current step ID for context
	ActionID          string            // Current action ID for context
	ActionName        string            // Current action name for context
	ParentActionID    string            // Parent action ID for context
	LoopIndex         int               // Current loop index for multi-run context
	Runner            *Runner           // Reference to runner for variable resolution
	VariableContext   *VariableContext  // Variable context for resolution
	AutomationConfig  *AutomationConfig // Automation config for variable resolution
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

// VariableContext holds context variables for resolution
type VariableContext struct {
	LoopIndex      int
	LocalLoopIndex int
	Timestamp      string
	RunID          string
	UserID         string
	ProjectID      string
	AutomationID   string
	StaticVars     map[string]string
	RuntimeVars    map[string]interface{} // Variables set during execution (local to current loop)
	GlobalVars     map[string]interface{} // Variables set during execution (global across all loops)
}

// Variable represents a configuration variable
type Variable struct {
	Key         string `json:"key"`
	Type        string `json:"type"` // "static", "dynamic", "environment"
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// MultiRunConfig represents multi-run configuration
type MultiRunConfig struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"` // "sequential", "parallel"
	Count   int    `json:"count"`
	Delay   int    `json:"delay"` // delay in milliseconds
}

// ScreenshotConfig represents screenshot configuration
type ScreenshotConfig struct {
	Enabled   bool   `json:"enabled"`
	OnError   bool   `json:"onError"`
	OnSuccess bool   `json:"onSuccess"`
	Path      string `json:"path"`
}

// NotificationChannelConfig represents notification configuration
type NotificationChannelConfig struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"` // "slack", "email", "webhook"
	OnComplete bool           `json:"onComplete"`
	OnError    bool           `json:"onError"`
	Config     map[string]any `json:"config"`
}

// AutomationConfig represents the parsed automation configuration
type AutomationConfig struct {
	Variables     []Variable                  `json:"variables"`
	Multirun      MultiRunConfig              `json:"multirun"`
	Timeout       int                         `json:"timeout"` // in seconds
	Retries       int                         `json:"retries"`
	Screenshots   ScreenshotConfig            `json:"screenshots"`
	Notifications []NotificationChannelConfig `json:"notifications"`
}

// Automation represents an automation workflow
type Automation struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	ConfigJSON  string // JSON string containing variables, run settings, templates
	Steps       []*AutomationStep
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AutomationStep represents a step within an automation
type AutomationStep struct {
	ID           string
	AutomationID string
	Name         string
	StepOrder    int
	Actions      []*AutomationAction
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AutomationAction represents an action within a step
type AutomationAction struct {
	ID               string
	StepID           string
	Name             string // Optional human-readable name for the action
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