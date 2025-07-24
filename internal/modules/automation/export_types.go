package automation

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
	Name      string                     `json:"name"`
	StepOrder int                        `json:"step_order"`
	Actions   []ExportedAutomationAction `json:"actions"`
}

// ExportedAutomationAction represents an action within a step for export
type ExportedAutomationAction struct {
	ID           string                 `json:"id"`
	ActionType   string                 `json:"action_type"`
	ActionConfig map[string]interface{} `json:"action_config"` // Direct map instead of JSON string
	ActionOrder  int                    `json:"action_order"`
}
