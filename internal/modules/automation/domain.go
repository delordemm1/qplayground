package automation

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/playwright-community/playwright-go"
)

// RunEventType represents the type of event emitted during automation execution
type RunEventType string

const (
	RunEventTypeLog         RunEventType = "log"
	RunEventTypeError       RunEventType = "error"
	RunEventTypeOutputFile  RunEventType = "output_file"
	RunEventTypeStep        RunEventType = "step"
	RunEventTypeStepSummary RunEventType = "step_summary"
)

// RunEvent represents an event emitted during automation execution
type RunEvent struct {
	Type             RunEventType           `json:"type"`
	Timestamp        time.Time              `json:"timestamp"`
	StepID           string                 `json:"step_id,omitempty"`
	StepName         string                 `json:"step_name,omitempty"`
	ActionID         string                 `json:"action_id,omitempty"`
	ActionName       string                 `json:"action_name,omitempty"`
	ActionConfigJSON string                 `json:"action_config_json,omitempty"`
	ParentActionID   string                 `json:"parent_action_id,omitempty"`
	ActionType       string                 `json:"action_type,omitempty"`
	Message          string                 `json:"message,omitempty"`
	Error            string                 `json:"error,omitempty"`
	OutputFile       string                 `json:"output_file,omitempty"`
	Duration         int64                  `json:"duration_ms,omitempty"`
	LoopIndex        int                    `json:"loop_index,omitempty"`
	LocalLoopIndex   int                    `json:"local_loop_index,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
}

// RunOverrides allows overriding automation configuration at runtime
type RunOverrides struct {
	MaxConcurrentRuns *int    `json:"max_concurrent_runs,omitempty"`
	RunMode           *string `json:"run_mode,omitempty"` // "sequential" or "parallel"
	RunCount          *int    `json:"run_count,omitempty"`
	RunDelay          *int    `json:"run_delay,omitempty"` // delay in milliseconds
}

// Global action configuration structures
type GlobalGroupConfig struct {
	Actions []GroupAutomationAction `json:"actions"`
}

type ElseIfCondition struct {
	ConditionType   string                  `json:"condition_type"`
	ConditionConfig map[string]interface{}  `json:"condition_config"`
	Actions         []GroupAutomationAction `json:"actions"`
}

type GlobalIfElseConfig struct {
	ConditionType    string                  `json:"condition_type"`
	ConditionConfig  map[string]interface{}  `json:"condition_config"`
	IfActions        []GroupAutomationAction `json:"if_actions"`
	ElseIfConditions []ElseIfCondition       `json:"else_if_conditions"`
	ElseActions      []GroupAutomationAction `json:"else_actions"`
	FinalActions     []GroupAutomationAction `json:"final_actions"`
}

type GlobalLoopConfig struct {
	ConditionType   string                  `json:"condition_type"`
	ConditionConfig map[string]interface{}  `json:"condition_config"`
	MaxLoops        int                     `json:"max_loops"`
	TimeoutMs       int                     `json:"timeout_ms"`
	FailOnForceStop bool                    `json:"fail_on_force_stop"`
	LoopActions     []GroupAutomationAction `json:"loop_actions"`
}

// RunContext holds shared resources and state for a single automation run.
// This will be passed to each plugin action.
type RunContext struct {
	PlaywrightBrowserContext playwright.BrowserContext
	PlaywrightPage           playwright.Page
	StorageService           storage.StorageService
	Logger                   *slog.Logger
	EventCh                  chan RunEvent
	StepName                 string            // Current step name for context
	StepID                   string            // Current step ID for context
	ActionID                 string            // Current action ID for context
	ActionName               string            // Current action name for context
	ParentActionID           string            // Parent action ID for context
	LoopIndex                int               // Current loop index for multi-run context
	Runner                   *Runner           // Reference to runner for variable resolution
	VariableContext          *VariableContext  // Variable context for resolution
	AutomationConfig         *AutomationConfig // Automation config for variable resolution
	LastOutputFiles          []string          // Buffer for output files from nested actions
	ScreenshotsR2Path        string            // Path to store screenshots in R2
	ReportsR2Path            string            // Path to store reports in R2
}

// PluginAction defines the interface for any executable action provided by a plugin.
type PluginAction interface {
	// Execute performs the action.
	// actionConfig: The specific configuration for this action (from automation_actions.action_config_json).
	// runContext: Shared context for the entire automation run.
	Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error

	// EvaluateCondition evaluates a condition for this action type.
	// conditionConfig: The specific configuration for the condition.
	// runContext: Shared context for the entire automation run.
	EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *RunContext) (bool, error)
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

// StepConfig represents the parsed step configuration
type StepConfig struct {
	SkipCondition    string  `json:"skip_condition,omitempty"`     // e.g., "loop_index_is_even", "loop_index_is_odd", "loop_index_is_prime", "random"
	RunOnlyCondition string  `json:"run_only_condition,omitempty"` // alternative to skip_condition
	Probability      float64 `json:"probability,omitempty"`        // for random condition, defaults to 0.5
}

// Automation represents an automation workflow
type Automation struct {
	ID             string
	ProjectID      string
	ProjectName    string
	Name           string
	AutomationSlug string
	Description    string
	ConfigJSON     string // JSON string containing variables, run settings, templates
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// AutomationStep represents a step within an automation
type AutomationStep struct {
	ID           string
	AutomationID string
	Name         string
	StepOrder    int
	ConfigJSON   string // JSON string containing step-level configuration like skip conditions
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
	ActionConfig     map[string]any
	ActionOrder      int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
type GroupAutomationAction struct {
	ID           string         `json:"id"`
	StepID       string         `json:"step_id"`
	Name         string         `json:"name"`          // Optional human-readable name for the action
	ActionType   string         `json:"action_type"`   // e.g., "playwright:goto", "playwright:click"
	ActionConfig map[string]any `json:"action_config"` // JSON string containing action-specific parameters
	ActionOrder  int            `json:"action_order"`
}

// AutomationRun represents an execution of an automation
type AutomationRun struct {
	ID                   string
	AutomationID         string
	Status               string // pending, running, completed, failed, cancelled
	StartTime            *time.Time
	EndTime              *time.Time
	LogsJSON             string // JSON string containing execution logs
	OutputFilesJSON      string // JSON string containing file paths/URLs
	SubRunOutputsJSON    string // JSON string mapping subRunIndex to storage URLs
	TotalRunsExpected    *int   // Total number of sub-runs expected for multi-runner jobs
	RunsCompleted        *int   // Number of sub-runs that have completed
	DetailedReportURL    string // URL to the detailed HTML report
	UserJourneyReportURL string // URL to the user journey HTML report
	ErrorMessage         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// SubRunOutput represents the output of an individual sub-run
type SubRunOutput struct {
	LogsURL  string `json:"logs_url"`
	FilesURL string `json:"files_url"`
}

// AutomationRun status constants
const (
	AutomationRunStatusPending                = "pending"
	AutomationRunStatusRunning                = "running"
	AutomationRunStatusCompleted              = "completed"
	AutomationRunStatusFailed                 = "failed"
	AutomationRunStatusCancelled              = "cancelled"
	AutomationRunStatusQueued                 = "queued"
	AutomationRunStatusPartialCompleted       = "partial_completed"
	AutomationRunStatusAwaitingExternalRunner = "awaiting_external_runner"
	AutomationRunStatusConsolidating          = "consolidating"
)

// RunProgressMessage represents a progress update for an automation run
type RunProgressMessage struct {
	Type        string                 `json:"type"` // "status", "log", "step", "action", "error", "complete", "step_summary"
	RunID       string                 `json:"runId"`
	Status      string                 `json:"status,omitempty"`
	StepName    string                 `json:"stepName,omitempty"`
	StepID      string                 `json:"stepId,omitempty"`
	ActionType  string                 `json:"actionType,omitempty"`
	ActionName  string                 `json:"actionName,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Progress    int                    `json:"progress,omitempty"` // 0-100
	TotalSteps  int                    `json:"totalSteps,omitempty"`
	CurrentStep int                    `json:"currentStep,omitempty"`
	Duration    int64                  `json:"duration,omitempty"` // milliseconds
	OutputFile  string                 `json:"outputFile,omitempty"`
	FilesCount  int                    `json:"filesCount,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
	// Step summary specific fields
	CompletedCount    int   `json:"completedCount,omitempty"`
	InProgressCount   int   `json:"inProgressCount,omitempty"`
	FailedCount       int   `json:"failedCount,omitempty"`
	TotalUsersForStep int   `json:"totalUsersForStep,omitempty"`
	AverageDurationMs int64 `json:"averageDurationMs,omitempty"`
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

	// Multi-runner support
	UpdateSubRunProgress(ctx context.Context, runID string, subRunIndex int, logsURL, filesURL string) error

	// Order management
	GetStepByID(ctx context.Context, id string) (*AutomationStep, error)
	GetActionByID(ctx context.Context, id string) (*AutomationAction, error)
	GetStepByAutomationIDAndOrder(ctx context.Context, automationID string, order int) (*AutomationStep, error)
	GetActionByStepIDAndOrder(ctx context.Context, stepID string, order int) (*AutomationAction, error)
	GetMaxStepOrder(ctx context.Context, automationID string) (int, error)
	GetMaxActionOrder(ctx context.Context, stepID string) (int, error)
	ShiftStepOrders(ctx context.Context, automationID string, startOrder, endOrder int, increment bool) error
	ShiftActionOrders(ctx context.Context, stepID string, startOrder, endOrder int, increment bool) error
	ShiftActionOrdersAfterDelete(ctx context.Context, stepID string, deletedOrder int) error
}

// AutomationService defines the interface for automation business logic
type AutomationService interface {
	// Automation management
	CreateAutomation(ctx context.Context, projectID, name, description, configJSON string) (*Automation, error)
	GetAutomationsByProject(ctx context.Context, projectID string) ([]*Automation, error)
	GetAutomationByID(ctx context.Context, id string) (*Automation, error)
	GetFullAutomationConfig(ctx context.Context, automationID string) (*ExportedAutomationConfig, error)
	UpdateAutomation(ctx context.Context, automation *Automation) error
	DeleteAutomation(ctx context.Context, id string) error

	// Step management
	CreateStep(ctx context.Context, automationID, name string, stepOrder int, configJSON string) (*AutomationStep, error)
	GetStepsByAutomation(ctx context.Context, automationID string) ([]*AutomationStep, error)
	UpdateStep(ctx context.Context, step *AutomationStep) error
	DeleteStep(ctx context.Context, id string) error

	// Action management
	CreateAction(ctx context.Context, stepID, name, actionType, actionConfigJSON string, actionOrder int) (*AutomationAction, error)
	GetActionsByStep(ctx context.Context, stepID string) ([]*AutomationAction, error)
	UpdateAction(ctx context.Context, action *AutomationAction) error
	DeleteAction(ctx context.Context, id string) error

	// Run management
	TriggerRun(ctx context.Context, automationID string) (*AutomationRun, error)
	TriggerExternalRun(ctx context.Context, automationID string, expectedRunners int) (*AutomationRun, error)
	GetRunsByAutomation(ctx context.Context, automationID string) ([]*AutomationRun, error)
	GetRunByID(ctx context.Context, id string) (*AutomationRun, error)

	// Order management helpers
	GetMaxStepOrder(ctx context.Context, automationID string) (int, error)
	GetMaxActionOrder(ctx context.Context, stepID string) (int, error)

	// Run cache management
	UpdateRunStatus(ctx context.Context, runID, status string) error
}
