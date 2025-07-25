package api

// AfterHookConfig defines how to extract data from API responses and save as runtime variables
type AfterHookConfig struct {
	Path   string `json:"path"`    // Dot-delimited JSON path (e.g., "data.user.accessToken")
	SaveAs string `json:"save_as"` // Runtime variable name to save the extracted value
	Scope  string `json:"scope"`   // "local" (default) or "global" - determines variable scope
}

// ApiActionConfigBase contains common fields for all API requests
type ApiActionConfigBase struct {
	URL        string                 `json:"url"`
	Headers    map[string]string      `json:"headers"`
	Body       string                 `json:"body"`
	Timeout    int                    `json:"timeout"`     // Request timeout in milliseconds
	AfterHooks []AfterHookConfig      `json:"after_hooks"` // Data extraction hooks
	Auth       *AuthConfig            `json:"auth"`        // Authentication configuration
}

// AuthConfig defines authentication settings for API requests
type AuthConfig struct {
	Type   string `json:"type"`   // "bearer", "basic", "api_key", "custom"
	Token  string `json:"token"`  // Token value (can use runtime variables)
	Header string `json:"header"` // Custom header name for api_key type
}

// Specific configurations for each HTTP method
type ApiGetConfig struct {
	ApiActionConfigBase
}

type ApiPostConfig struct {
	ApiActionConfigBase
}

type ApiPutConfig struct {
	ApiActionConfigBase
}

type ApiPatchConfig struct {
	ApiActionConfigBase
}

type ApiDeleteConfig struct {
	ApiActionConfigBase
}

// ApiResponseData represents the structured response data for logging
type ApiResponseData struct {
	URL            string                 `json:"url"`
	Method         string                 `json:"method"`
	StatusCode     int                    `json:"status_code"`
	ResponseTime   int64                  `json:"response_time_ms"`
	RequestHeaders map[string]string      `json:"request_headers"`
	ResponseBody   string                 `json:"response_body"`
	ExtractedVars  map[string]interface{} `json:"extracted_vars"`
	Error          string                 `json:"error,omitempty"`
}

// ApiIfElseConfig represents configuration for conditional API logic
type ApiIfElseConfig struct {
	VariablePath     string                 `json:"variable_path"`      // e.g., "runtime.lastResponse.successCode"
	ConditionType    string                 `json:"condition_type"`     // "equals", "not_equals", "greater_than", etc.
	ExpectedValue    interface{}            `json:"expected_value"`     // Value to compare against
	IfActions        []NestedAction         `json:"if_actions"`         // Actions to execute if condition is true
	ElseIfConditions []ApiElseIfCondition   `json:"else_if_conditions"` // Alternative conditions
	ElseActions      []NestedAction         `json:"else_actions"`       // Actions to execute if no conditions match
	FinalActions     []NestedAction         `json:"final_actions"`      // Actions that always execute
}

// ApiElseIfCondition represents an else-if condition block
type ApiElseIfCondition struct {
	VariablePath  string         `json:"variable_path"`
	ConditionType string         `json:"condition_type"`
	ExpectedValue interface{}    `json:"expected_value"`
	Actions       []NestedAction `json:"actions"`
}

// NestedAction represents an action that can be nested within conditional blocks
type NestedAction struct {
	ID           string                 `json:"id"`
	ActionType   string                 `json:"action_type"`
	ActionConfig map[string]interface{} `json:"action_config"`
}