package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/playwright-community/playwright-go"
)

// ExportedAutomationConfig represents the complete automation configuration for export/import
type ExportedAutomationConfig struct {
	Automation ExportedAutomation       `json:"automation"`
	Steps      []ExportedAutomationStep `json:"steps"`
}

// ExportedAutomation represents the automation metadata for export
type ExportedAutomation struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      ExportedAutomationMeta `json:"config"`
}

// ExportedAutomationMeta represents the parsed automation configuration
type ExportedAutomationMeta struct {
	Variables     []ExportedVariable                  `json:"variables"`
	Multirun      ExportedMultiRunConfig              `json:"multirun"`
	Timeout       int                                 `json:"timeout"` // in seconds
	Retries       int                                 `json:"retries"`
	Screenshots   ExportedScreenshotConfig            `json:"screenshots"`
	Notifications []ExportedNotificationChannelConfig `json:"notifications"`
}

// ExportedVariable represents a configuration variable
type ExportedVariable struct {
	Key         string `json:"key"`
	Type        string `json:"type"` // "static", "dynamic", "environment"
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// ExportedMultiRunConfig represents multi-run configuration
type ExportedMultiRunConfig struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"` // "sequential", "parallel"
	Count   int    `json:"count"`
	Delay   int    `json:"delay"` // delay in milliseconds
}

// ExportedScreenshotConfig represents screenshot configuration
type ExportedScreenshotConfig struct {
	Enabled   bool   `json:"enabled"`
	OnError   bool   `json:"onError"`
	OnSuccess bool   `json:"onSuccess"`
	Path      string `json:"path"`
}

// ExportedNotificationChannelConfig represents notification configuration
type ExportedNotificationChannelConfig struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "slack", "email", "webhook"
	OnComplete bool                   `json:"onComplete"`
	OnError    bool                   `json:"onError"`
	Config     map[string]interface{} `json:"config"`
}

// ExportedAutomationStep represents a step within an automation for export
type ExportedAutomationStep struct {
	Name      string                    `json:"name"`
	StepOrder int                       `json:"step_order"`
	Actions   []ExportedAutomationAction `json:"actions"`
}

// ExportedAutomationAction represents an action within a step for export
type ExportedAutomationAction struct {
	ActionType   string                 `json:"action_type"`
	ActionConfig map[string]interface{} `json:"action_config"` // Direct map instead of JSON string
	ActionOrder  int                    `json:"action_order"`
}

// RunContext holds shared resources and state for a single automation run.
type RunContext struct {
	PlaywrightBrowser playwright.Browser
	PlaywrightPage    playwright.Page
	StorageService    LocalStorageService
	Logger            *LocalLogger
}

// PluginAction defines the interface for any executable action provided by a plugin.
type PluginAction interface {
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
	LoopIndex    int
	Timestamp    string
	RunID        string
	UserID       string
	ProjectID    string
	AutomationID string
	StaticVars   map[string]string
}