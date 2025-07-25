package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/automation"
)

func init() {
	automation.RegisterAction("api:get", func() automation.PluginAction { return &ApiGetAction{} })
	automation.RegisterAction("api:post", func() automation.PluginAction { return &ApiPostAction{} })
	automation.RegisterAction("api:put", func() automation.PluginAction { return &ApiPutAction{} })
	automation.RegisterAction("api:patch", func() automation.PluginAction { return &ApiPatchAction{} })
	automation.RegisterAction("api:delete", func() automation.PluginAction { return &ApiDeleteAction{} })
	automation.RegisterAction("api:if_else", func() automation.PluginAction { return &ApiIfElseAction{} })
	automation.RegisterAction("api:runtime_loop_until", func() automation.PluginAction { return &ApiRuntimeLoopUntilAction{} })
}

// Helper function to send success event for API actions
func sendApiSuccessEvent(runContext *automation.RunContext, actionType, message string, duration time.Duration, responseData ApiResponseData) {
	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:           automation.RunEventTypeLog,
			Timestamp:      time.Now(),
			StepName:       runContext.StepName,
			StepID:         runContext.StepID,
			ActionID:       runContext.ActionID,
			ActionName:     runContext.ActionName,
			ParentActionID: runContext.ParentActionID,
			ActionType:     actionType,
			Message:        message,
			Duration:       duration.Milliseconds(),
			LoopIndex:      runContext.LoopIndex,
			LocalLoopIndex: runContext.VariableContext.LocalLoopIndex,
			Data:           map[string]interface{}{"api_response": responseData},
		}:
		default:
			// Channel is full, skip this event to avoid blocking
		}
	}
}

// Helper function to send error event for API actions
func sendApiErrorEvent(runContext *automation.RunContext, actionType, errorMsg string, duration time.Duration, responseData *ApiResponseData) {
	eventData := map[string]interface{}{}
	if responseData != nil {
		eventData["api_response"] = *responseData
	}

	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:           automation.RunEventTypeError,
			Timestamp:      time.Now(),
			StepName:       runContext.StepName,
			StepID:         runContext.StepID,
			ActionID:       runContext.ActionID,
			ActionName:     runContext.ActionName,
			ParentActionID: runContext.ParentActionID,
			ActionType:     actionType,
			Error:          errorMsg,
			Duration:       duration.Milliseconds(),
			LoopIndex:      runContext.LoopIndex,
			LocalLoopIndex: runContext.VariableContext.LocalLoopIndex,
			Data:           eventData,
		}:
		default:
			// Channel is full, skip this event to avoid blocking
		}
	}
}

// BaseApiAction provides common functionality for all API actions
type BaseApiAction struct{}

func (b *BaseApiAction) executeApiRequest(ctx context.Context, method string, config ApiActionConfigBase, runContext *automation.RunContext) error {
	startTime := time.Now()

	// Validate required fields
	if config.URL == "" {
		return fmt.Errorf("%s action requires a 'url' string in config", method)
	}

	runContext.Logger.Info("Executing API request", "method", method, "url", config.URL)

	// Resolve variables in URL
	resolvedURL, err := runContext.Runner.ResolveVariablesInString(config.URL, runContext.VariableContext, runContext.AutomationConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve variables in URL: %w", err)
	}

	// Resolve variables in body
	resolvedBody := ""
	if config.Body != "" {
		resolvedBody, err = runContext.Runner.ResolveVariablesInString(config.Body, runContext.VariableContext, runContext.AutomationConfig)
		if err != nil {
			return fmt.Errorf("failed to resolve variables in body: %w", err)
		}
	}

	// Create request
	var bodyReader io.Reader
	if resolvedBody != "" {
		bodyReader = strings.NewReader(resolvedBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, resolvedURL, bodyReader)
	if err != nil {
		duration := time.Since(startTime)
		responseData := ApiResponseData{
			URL:    resolvedURL,
			Method: method,
			Error:  err.Error(),
		}
		sendApiErrorEvent(runContext, fmt.Sprintf("api:%s", strings.ToLower(method)), err.Error(), duration, &responseData)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default headers
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Resolve and set custom headers
	resolvedHeaders := make(map[string]string)
	for key, value := range config.Headers {
		resolvedKey, err := runContext.Runner.ResolveVariablesInString(key, runContext.VariableContext, runContext.AutomationConfig)
		if err != nil {
			return fmt.Errorf("failed to resolve variables in header key '%s': %w", key, err)
		}
		resolvedValue, err := runContext.Runner.ResolveVariablesInString(value, runContext.VariableContext, runContext.AutomationConfig)
		if err != nil {
			return fmt.Errorf("failed to resolve variables in header value '%s': %w", value, err)
		}
		resolvedHeaders[resolvedKey] = resolvedValue
		req.Header.Set(resolvedKey, resolvedValue)
	}

	// Handle authentication
	if config.Auth != nil {
		err := b.setAuthHeader(req, config.Auth, runContext)
		if err != nil {
			return fmt.Errorf("failed to set authentication header: %w", err)
		}
	} else {
		// Auto-detect common authentication tokens from runtime variables
		b.autoSetAuthHeaders(req, runContext)
	}

	// Set timeout
	timeout := 30 * time.Second // Default timeout
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Millisecond
	}

	client := &http.Client{Timeout: timeout}

	// Execute request
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	responseData := ApiResponseData{
		URL:            resolvedURL,
		Method:         method,
		RequestHeaders: resolvedHeaders,
		ResponseTime:   duration.Milliseconds(),
		ExtractedVars:  make(map[string]interface{}),
	}

	if err != nil {
		responseData.Error = err.Error()
		sendApiErrorEvent(runContext, fmt.Sprintf("api:%s", strings.ToLower(method)), err.Error(), duration, &responseData)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseData.StatusCode = resp.StatusCode

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		responseData.Error = "failed to read response body"
		sendApiErrorEvent(runContext, fmt.Sprintf("api:%s", strings.ToLower(method)), "failed to read response body", duration, &responseData)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	responseData.ResponseBody = string(responseBody)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		errorMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		responseData.Error = errorMsg
		sendApiErrorEvent(runContext, fmt.Sprintf("api:%s", strings.ToLower(method)), errorMsg, duration, &responseData)
		return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	// Parse response as JSON for after_hooks processing
	var responseJSON map[string]interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &responseJSON); err != nil {
			runContext.Logger.Warn("Failed to parse response as JSON, after_hooks will be skipped", "error", err)
		}
	}

	// Process after_hooks to extract runtime variables
	if responseJSON != nil && len(config.AfterHooks) > 0 {
		for _, hook := range config.AfterHooks {
			var extractedValue interface{}
			var err error

			if hook.Path == "" || hook.Path == "." {
				// Store the entire response
				extractedValue = responseJSON
			} else {
				extractedValue, err = b.extractJSONPath(responseJSON, hook.Path)
			}

			if err != nil {
				runContext.Logger.Warn("Failed to extract value from JSON path", "path", hook.Path, "error", err)
				continue
			}

			// Determine scope and save variable
			if hook.Scope == "global" {
				// Save to global vars (global scope)
				runContext.VariableContext.GlobalVars[hook.SaveAs] = extractedValue
			} else {
				// Save to runtime vars (local scope - default) as interface{}
				runContext.VariableContext.RuntimeVars[hook.SaveAs] = extractedValue
			}

			responseData.ExtractedVars[hook.SaveAs] = extractedValue
			runContext.Logger.Info("Extracted runtime variable",
				"path", hook.Path,
				"save_as", hook.SaveAs,
				"value_type", fmt.Sprintf("%T", extractedValue),
				"scope", hook.Scope)
		}
	}

	message := fmt.Sprintf("Successfully executed %s request to %s (HTTP %d)", method, resolvedURL, resp.StatusCode)
	sendApiSuccessEvent(runContext, fmt.Sprintf("api:%s", strings.ToLower(method)), message, duration, responseData)

	return nil
}

// setAuthHeader sets the authentication header based on auth configuration
func (b *BaseApiAction) setAuthHeader(req *http.Request, auth *AuthConfig, runContext *automation.RunContext) error {
	if auth.Token == "" {
		return nil // No token provided
	}

	// Resolve variables in token
	resolvedToken, err := runContext.Runner.ResolveVariablesInString(auth.Token, runContext.VariableContext, runContext.AutomationConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve variables in auth token: %w", err)
	}

	switch auth.Type {
	case "bearer":
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", resolvedToken))
	case "basic":
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", resolvedToken))
	case "api_key":
		headerName := auth.Header
		if headerName == "" {
			headerName = "X-API-Key" // Default header name
		}
		req.Header.Set(headerName, resolvedToken)
	case "custom":
		// For custom auth, the token should include the full header value
		req.Header.Set("Authorization", resolvedToken)
	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}

	return nil
}

// autoSetAuthHeaders automatically sets common authentication headers from runtime variables
func (b *BaseApiAction) autoSetAuthHeaders(req *http.Request, runContext *automation.RunContext) {
	// Check for common authentication tokens in runtime variables
	if accessToken, exists := runContext.VariableContext.RuntimeVars["access_token"]; exists {
		if tokenStr, ok := accessToken.(string); ok && tokenStr != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
			runContext.Logger.Debug("Auto-set Bearer token from runtime variable", "token_length", len(tokenStr))
		}
	} else if apiKey, exists := runContext.VariableContext.RuntimeVars["api_key"]; exists {
		if keyStr, ok := apiKey.(string); ok && keyStr != "" {
			req.Header.Set("X-API-Key", keyStr)
			runContext.Logger.Debug("Auto-set API key from runtime variable", "key_length", len(keyStr))
		}
	}
}

// extractJSONPath extracts a value from a JSON object using a dot-delimited path
func (b *BaseApiAction) extractJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("null value encountered at path segment '%s'", part)
		}

		// Handle array indices (e.g., "users[0]" or "users.0")
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Extract array name and index
			arrayName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			arrayValue, exists := current[arrayName]
			if !exists {
				return nil, fmt.Errorf("array '%s' not found", arrayName)
			}

			arraySlice, ok := arrayValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("'%s' is not an array", arrayName)
			}

			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s'", indexStr)
			}

			if index < 0 || index >= len(arraySlice) {
				return nil, fmt.Errorf("array index %d out of bounds for array '%s'", index, arrayName)
			}

			if i == len(parts)-1 {
				return arraySlice[index], nil
			}

			// Continue with the array element
			if nextMap, ok := arraySlice[index].(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, fmt.Errorf("array element at index %d is not an object", index)
			}
		} else {
			// Regular object property access
			value, exists := current[part]
			if !exists {
				return nil, fmt.Errorf("property '%s' not found", part)
			}

			if i == len(parts)-1 {
				return value, nil
			}

			// Continue traversing
			if nextMap, ok := value.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, fmt.Errorf("property '%s' is not an object, cannot traverse further", part)
			}
		}
	}

	return current, nil
}

// parseApiConfig parses the action config into ApiActionConfigBase
func (b *BaseApiAction) parseApiConfig(actionConfig map[string]interface{}) (ApiActionConfigBase, error) {
	var config ApiActionConfigBase

	// Parse URL
	if url, ok := actionConfig["url"].(string); ok {
		config.URL = url
	}

	// Parse headers
	if headers, ok := actionConfig["headers"].(map[string]interface{}); ok {
		config.Headers = make(map[string]string)
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				config.Headers[key] = strValue
			}
		}
	}

	// Parse body
	if body, ok := actionConfig["body"].(string); ok {
		config.Body = body
	}

	// Parse timeout
	if timeout, ok := actionConfig["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}

	// Parse auth configuration
	if authInterface, ok := actionConfig["auth"].(map[string]interface{}); ok {
		auth := &AuthConfig{}
		if authType, ok := authInterface["type"].(string); ok {
			auth.Type = authType
		}
		if token, ok := authInterface["token"].(string); ok {
			auth.Token = token
		}
		if header, ok := authInterface["header"].(string); ok {
			auth.Header = header
		}
		config.Auth = auth
	}

	// Parse after_hooks
	if hooksInterface, ok := actionConfig["after_hooks"].([]interface{}); ok {
		for _, hookInterface := range hooksInterface {
			if hookMap, ok := hookInterface.(map[string]interface{}); ok {
				hook := AfterHookConfig{}
				if path, ok := hookMap["path"].(string); ok {
					hook.Path = path
				}
				if saveAs, ok := hookMap["save_as"].(string); ok {
					hook.SaveAs = saveAs
				}
				if scope, ok := hookMap["scope"].(string); ok {
					hook.Scope = scope
				} else {
					hook.Scope = "local" // Default scope
				}
				config.AfterHooks = append(config.AfterHooks, hook)
			}
		}
	}

	return config, nil
}

// ApiGetAction implements HTTP GET requests
type ApiGetAction struct {
	BaseApiAction
}

func (a *ApiGetAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	config, err := a.parseApiConfig(actionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse API GET config: %w", err)
	}

	return a.executeApiRequest(ctx, "GET", config, runContext)
}

// ApiPostAction implements HTTP POST requests
type ApiPostAction struct {
	BaseApiAction
}

func (a *ApiPostAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	config, err := a.parseApiConfig(actionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse API POST config: %w", err)
	}

	return a.executeApiRequest(ctx, "POST", config, runContext)
}

// ApiPutAction implements HTTP PUT requests
type ApiPutAction struct {
	BaseApiAction
}

func (a *ApiPutAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	config, err := a.parseApiConfig(actionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse API PUT config: %w", err)
	}

	return a.executeApiRequest(ctx, "PUT", config, runContext)
}

// ApiPatchAction implements HTTP PATCH requests
type ApiPatchAction struct {
	BaseApiAction
}

func (a *ApiPatchAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	config, err := a.parseApiConfig(actionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse API PATCH config: %w", err)
	}

	return a.executeApiRequest(ctx, "PATCH", config, runContext)
}

// ApiDeleteAction implements HTTP DELETE requests
type ApiDeleteAction struct {
	BaseApiAction
}

func (a *ApiDeleteAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	config, err := a.parseApiConfig(actionConfig)
	if err != nil {
		return fmt.Errorf("failed to parse API DELETE config: %w", err)
	}

	return a.executeApiRequest(ctx, "DELETE", config, runContext)
}

// ApiIfElseAction implements conditional logic based on runtime variables
type ApiIfElseAction struct{}

func (a *ApiIfElseAction) Execute(ctx context.Context, actionConfig map[string]any, runContext *automation.RunContext) error {
	startTime := time.Now()

	variablePath, ok := actionConfig["variable_path"].(string)
	if !ok || variablePath == "" {
		return fmt.Errorf("api:if_else action requires a 'variable_path' string in config")
	}

	conditionType, ok := actionConfig["condition_type"].(string)
	if !ok || conditionType == "" {
		return fmt.Errorf("api:if_else action requires a 'condition_type' string in config")
	}

	expectedValue := actionConfig["expected_value"] // Can be any type

	runContext.Logger.Info("Executing api:if_else", "variable_path", variablePath, "condition_type", conditionType)

	var executionError error
	defer func() {
		// Always execute final actions regardless of the outcome
		if finalErr := a.executeFinalActions(ctx, actionConfig, runContext); finalErr != nil {
			runContext.Logger.Error("Failed to execute final actions", "error", finalErr)
			if executionError == nil {
				executionError = finalErr
			}
		}

		duration := time.Since(startTime)
		if executionError != nil {
			sendApiErrorEvent(runContext, "api:if_else", executionError.Error(), duration, nil)
		} else {
			sendApiSuccessEvent(runContext, "api:if_else", "Successfully completed conditional logic", duration, ApiResponseData{})
		}
	}()

	// Evaluate main condition
	conditionMet, err := a.evaluateApiCondition(variablePath, conditionType, expectedValue, runContext)
	if err != nil {
		executionError = fmt.Errorf("failed to evaluate main condition: %w", err)
		return executionError
	}

	if conditionMet {
		// Execute if_actions
		if ifActions, ok := actionConfig["if_actions"].([]interface{}); ok {
			runContext.Logger.Info("Main condition is true, executing if_actions", "count", len(ifActions))
			executionError = a.executeNestedActions(ctx, ifActions, runContext)
			return executionError
		}
		return executionError
	}

	// Check else_if_conditions
	if elseIfConditions, ok := actionConfig["else_if_conditions"].([]interface{}); ok {
		for i, elseIfCondition := range elseIfConditions {
			elseIfMap, ok := elseIfCondition.(map[string]interface{})
			if !ok {
				continue
			}

			elseIfVariablePath, ok := elseIfMap["variable_path"].(string)
			if !ok || elseIfVariablePath == "" {
				continue
			}

			elseIfConditionType, ok := elseIfMap["condition_type"].(string)
			if !ok || elseIfConditionType == "" {
				continue
			}

			elseIfExpectedValue := elseIfMap["expected_value"]

			runContext.Logger.Info("Evaluating else-if condition", "index", i, "variable_path", elseIfVariablePath, "condition_type", elseIfConditionType)

			elseIfConditionMet, err := a.evaluateApiCondition(elseIfVariablePath, elseIfConditionType, elseIfExpectedValue, runContext)
			if err != nil {
				runContext.Logger.Warn("Failed to evaluate else-if condition", "index", i, "error", err)
				continue
			}

			if elseIfConditionMet {
				// Execute this else-if's actions
				if elseIfActions, ok := elseIfMap["actions"].([]interface{}); ok {
					runContext.Logger.Info("Else-if condition is true, executing actions", "index", i, "count", len(elseIfActions))
					executionError = a.executeNestedActions(ctx, elseIfActions, runContext)
					return executionError
				}
				return executionError
			}
		}
	}

	// Execute else_actions if all conditions failed
	if elseActions, ok := actionConfig["else_actions"].([]interface{}); ok {
		runContext.Logger.Info("All conditions failed, executing else_actions", "count", len(elseActions))
		executionError = a.executeNestedActions(ctx, elseActions, runContext)
		return executionError
	}

	runContext.Logger.Info("No conditions met and no else actions defined")
	return executionError
}

func (a *ApiIfElseAction) evaluateApiCondition(variablePath, conditionType string, expectedValue interface{}, runContext *automation.RunContext) (bool, error) {
	// Resolve the variable value using enhanced resolution
	actualValue, err := a.resolveRuntimeVariable(variablePath, runContext)
	if err != nil {
		return false, fmt.Errorf("failed to resolve variable '%s': %w", variablePath, err)
	}

	runContext.Logger.Debug("Evaluating condition",
		"variable_path", variablePath,
		"actual_value", actualValue,
		"expected_value", expectedValue,
		"condition_type", conditionType)

	switch conditionType {
	case "equals":
		return fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", expectedValue), nil
	case "not_equals":
		return fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue), nil
	case "contains":
		actualStr := fmt.Sprintf("%v", actualValue)
		expectedStr := fmt.Sprintf("%v", expectedValue)
		return strings.Contains(actualStr, expectedStr), nil
	case "not_contains":
		actualStr := fmt.Sprintf("%v", actualValue)
		expectedStr := fmt.Sprintf("%v", expectedValue)
		return !strings.Contains(actualStr, expectedStr), nil
	case "is_null":
		return actualValue == nil, nil
	case "is_not_null":
		return actualValue != nil, nil
	case "is_true":
		if boolVal, ok := actualValue.(bool); ok {
			return boolVal, nil
		}
		return fmt.Sprintf("%v", actualValue) == "true", nil
	case "is_false":
		if boolVal, ok := actualValue.(bool); ok {
			return !boolVal, nil
		}
		return fmt.Sprintf("%v", actualValue) == "false", nil
	case "greater_than":
		return a.compareNumbers(actualValue, expectedValue, ">")
	case "less_than":
		return a.compareNumbers(actualValue, expectedValue, "<")
	case "greater_than_or_equal":
		return a.compareNumbers(actualValue, expectedValue, ">=")
	case "less_than_or_equal":
		return a.compareNumbers(actualValue, expectedValue, "<=")
	default:
		return false, fmt.Errorf("unsupported condition type: %s", conditionType)
	}
}

func (a *ApiIfElseAction) compareNumbers(actual, expected interface{}, operator string) (bool, error) {
	actualFloat, err1 := a.toFloat64(actual)
	expectedFloat, err2 := a.toFloat64(expected)

	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}

	switch operator {
	case ">":
		return actualFloat > expectedFloat, nil
	case "<":
		return actualFloat < expectedFloat, nil
	case ">=":
		return actualFloat >= expectedFloat, nil
	case "<=":
		return actualFloat <= expectedFloat, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}

func (a *ApiIfElseAction) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (a *ApiIfElseAction) resolveRuntimeVariable(variablePath string, runContext *automation.RunContext) (interface{}, error) {
	if !strings.HasPrefix(variablePath, "runtime.") {
		return nil, fmt.Errorf("variable path must start with 'runtime.'")
	}

	// Remove "runtime." prefix
	path := strings.TrimPrefix(variablePath, "runtime.")
	pathParts := strings.Split(path, ".")

	if len(pathParts) == 0 {
		return nil, fmt.Errorf("empty variable path")
	}

	// Get the base variable
	baseVarName := pathParts[0]
	var baseValue interface{}
	var exists bool

	// Check runtime vars first, then global vars
	if baseValue, exists = runContext.VariableContext.RuntimeVars[baseVarName]; !exists {
		if baseValue, exists = runContext.VariableContext.GlobalVars[baseVarName]; !exists {
			return nil, fmt.Errorf("runtime variable '%s' not found", baseVarName)
		}
	}

	// If only base variable requested, return it
	if len(pathParts) == 1 {
		return baseValue, nil
	}

	// Resolve nested path
	return a.resolveNestedPath(baseValue, pathParts[1:])
}

func (a *ApiIfElseAction) resolveNestedPath(base interface{}, pathParts []string) (interface{}, error) {
	current := base

	for i, part := range pathParts {
		_ = i
		if current == nil {
			return nil, fmt.Errorf("null value encountered at path segment '%s'", part)
		}

		// Handle array indices (e.g., "options[0]")
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			arrayName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			// Get the array from current object
			var arrayValue interface{}
			if arrayName == "" {
				// Direct array access like [0]
				arrayValue = current
			} else {
				// Named array access like options[0]
				if currentMap, ok := current.(map[string]interface{}); ok {
					var exists bool
					arrayValue, exists = currentMap[arrayName]
					if !exists {
						return nil, fmt.Errorf("array '%s' not found", arrayName)
					}
				} else {
					return nil, fmt.Errorf("cannot access property '%s' on non-object", arrayName)
				}
			}

			arraySlice, ok := arrayValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("'%s' is not an array", arrayName)
			}

			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s'", indexStr)
			}

			if index < 0 || index >= len(arraySlice) {
				return nil, fmt.Errorf("array index %d out of bounds for array '%s'", index, arrayName)
			}

			current = arraySlice[index]
		} else {
			// Regular object property access
			if currentMap, ok := current.(map[string]interface{}); ok {
				value, exists := currentMap[part]
				if !exists {
					return nil, fmt.Errorf("property '%s' not found", part)
				}
				current = value
			} else {
				return nil, fmt.Errorf("cannot access property '%s' on non-object", part)
			}
		}
	}

	return current, nil
}

func (a *ApiIfElseAction) executeNestedActions(ctx context.Context, actions []interface{}, runContext *automation.RunContext) error {
	for i, actionInterface := range actions {
		actionMap, ok := actionInterface.(map[string]interface{})
		if !ok {
			runContext.Logger.Warn("Invalid nested action format", "index", i)
			continue
		}

		actionType, ok := actionMap["action_type"].(string)
		if !ok || actionType == "" {
			runContext.Logger.Warn("Missing action_type in nested action", "index", i)
			continue
		}

		actionConfig, ok := actionMap["action_config"].(map[string]interface{})
		if !ok {
			actionConfig = make(map[string]interface{})
		}

		// Resolve variables in nested action config
		if runContext.Runner != nil {
			resolvedActionConfig, err := runContext.Runner.ResolveVariablesInConfig(actionConfig, runContext.VariableContext, runContext.AutomationConfig)
			if err != nil {
				runContext.Logger.Error("Failed to resolve variables in nested action config", "action_type", actionType, "error", err)
				return fmt.Errorf("failed to resolve variables in nested action '%s': %w", actionType, err)
			}
			actionConfig = resolvedActionConfig
		}

		runContext.Logger.Info("Executing nested action", "index", i, "action_type", actionType)

		// Store original action context
		originalActionID := runContext.ActionID
		originalParentActionID := runContext.ParentActionID

		// Set nested action context
		nestedActionID, _ := actionMap["id"].(string)
		if nestedActionID == "" {
			nestedActionID = fmt.Sprintf("%s-nested-%d", runContext.ActionID, i)
		}
		runContext.ParentActionID = runContext.ActionID
		runContext.ActionID = nestedActionID

		// Get the plugin action
		pluginAction, err := automation.GetAction(actionType)
		if err != nil {
			runContext.Logger.Error("Failed to get nested action", "action_type", actionType, "error", err)
			// Restore original context
			runContext.ActionID = originalActionID
			runContext.ParentActionID = originalParentActionID
			return fmt.Errorf("failed to get nested action '%s': %w", actionType, err)
		}

		// Execute the nested action
		err = pluginAction.Execute(ctx, actionConfig, runContext)
		if err != nil {
			runContext.Logger.Error("Nested action failed", "action_type", actionType, "error", err)
			// Restore original context
			runContext.ActionID = originalActionID
			runContext.ParentActionID = originalParentActionID
			return fmt.Errorf("nested action '%s' failed: %w", actionType, err)
		}

		// Restore original action context
		runContext.ActionID = originalActionID
		runContext.ParentActionID = originalParentActionID

		runContext.Logger.Info("Nested action completed", "action_type", actionType)
	}

	return nil
}

// executeFinalActions executes the final actions that should run regardless of condition outcomes
func (a *ApiIfElseAction) executeFinalActions(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	finalActions, ok := actionConfig["final_actions"].([]interface{})
	if !ok || len(finalActions) == 0 {
		runContext.Logger.Info("No final actions to execute")
		return nil
	}

	runContext.Logger.Info("Executing final actions", "count", len(finalActions))
	return a.executeNestedActions(ctx, finalActions, runContext)
}

// ApiRuntimeLoopUntilAction implements looping until a runtime variable condition is met
type ApiRuntimeLoopUntilAction struct{}

func (a *ApiRuntimeLoopUntilAction) Execute(ctx context.Context, actionConfig map[string]any, runContext *automation.RunContext) error {
	startTime := time.Now()
	runContext.Logger.Info("Executing api:runtime_loop_until")

	// Extract configuration
	variablePath, _ := actionConfig["variable_path"].(string)
	conditionType, _ := actionConfig["condition_type"].(string)
	expectedValue := actionConfig["expected_value"]
	maxLoops, _ := actionConfig["max_loops"].(float64)
	timeoutMs, _ := actionConfig["timeout_ms"].(float64)
	failOnForceStop, _ := actionConfig["fail_on_force_stop"].(bool)
	loopActionsInterface, _ := actionConfig["loop_actions"].([]any)

	// Validate required fields
	if variablePath == "" {
		return fmt.Errorf("api:runtime_loop_until requires a 'variable_path' string in config")
	}
	if conditionType == "" {
		return fmt.Errorf("api:runtime_loop_until requires a 'condition_type' string in config")
	}

	// Validate that at least one force stop condition is provided
	if maxLoops <= 0 && timeoutMs <= 0 {
		return fmt.Errorf("api:runtime_loop_until requires either max_loops or timeout_ms to prevent infinite loops")
	}

	// Convert loop actions to proper format
	var loopActions []map[string]any
	for _, actionInterface := range loopActionsInterface {
		if actionMap, ok := actionInterface.(map[string]any); ok {
			loopActions = append(loopActions, actionMap)
		}
	}

	if len(loopActions) == 0 {
		return fmt.Errorf("api:runtime_loop_until requires at least one loop action")
	}

	runContext.Logger.Info("Starting runtime variable loop",
		"variable_path", variablePath,
		"condition_type", conditionType,
		"expected_value", expectedValue,
		"max_loops", maxLoops,
		"timeout_ms", timeoutMs,
		"loop_actions_count", len(loopActions))

	var executionError error
	defer func() {
		duration := time.Since(startTime)
		if executionError != nil {
			sendApiErrorEvent(runContext, "api:runtime_loop_until", executionError.Error(), duration, nil)
		} else {
			sendApiSuccessEvent(runContext, "api:runtime_loop_until", "Successfully completed runtime variable loop", duration, ApiResponseData{})
		}
	}()

	// Initialize loop variables
	loopCount := 0
	loopStartTime := time.Now()
	var timeoutDuration time.Duration
	if timeoutMs > 0 {
		timeoutDuration = time.Duration(timeoutMs) * time.Millisecond
	}

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			executionError = fmt.Errorf("loop cancelled")
			return executionError
		default:
		}

		loopCount++
		runContext.Logger.Info("Runtime variable loop iteration", "count", loopCount)

		// Set the local loop index in the variable context
		runContext.VariableContext.LocalLoopIndex = loopCount

		// Check runtime variable condition
		conditionMet, err := a.evaluateRuntimeVariableCondition(variablePath, conditionType, expectedValue, runContext)
		if err != nil {
			runContext.Logger.Warn("Failed to evaluate runtime variable condition", "error", err)
		} else if conditionMet {
			runContext.Logger.Info("Runtime variable condition met, exiting loop", 
				"variable_path", variablePath, 
				"condition_type", conditionType, 
				"expected_value", expectedValue,
				"loops_completed", loopCount)
			break
		}

		// Check force stop conditions
		forceStop := false
		forceStopReason := ""

		if maxLoops > 0 && float64(loopCount) >= maxLoops {
			forceStop = true
			forceStopReason = fmt.Sprintf("reached maximum loops (%d)", int(maxLoops))
		}

		if timeoutMs > 0 && time.Since(loopStartTime) >= timeoutDuration {
			forceStop = true
			if forceStopReason != "" {
				forceStopReason += " and "
			}
			forceStopReason += fmt.Sprintf("reached timeout (%dms)", int(timeoutMs))
		}

		if forceStop {
			message := fmt.Sprintf("Runtime variable loop force stopped: %s", forceStopReason)
			if failOnForceStop {
				runContext.Logger.Error("Runtime variable loop force stopped", "reason", forceStopReason, "loops_completed", loopCount)
				executionError = fmt.Errorf(message)
				return executionError
			} else {
				runContext.Logger.Warn("Runtime variable loop force stopped", "reason", forceStopReason, "loops_completed", loopCount)
				break
			}
		}

		// Execute loop actions
		for actionIndex, actionData := range loopActions {
			// Check for cancellation before each action
			select {
			case <-ctx.Done():
				executionError = fmt.Errorf("loop cancelled during action execution")
				return executionError
			default:
			}

			actionType, ok := actionData["action_type"].(string)
			if !ok || actionType == "" {
				runContext.Logger.Warn("Skipping loop action with missing action_type", "action_index", actionIndex)
				continue
			}

			actionConfig, ok := actionData["action_config"].(map[string]any)
			if !ok {
				actionConfig = make(map[string]any)
			}

			// Resolve variables in loop action config
			if runContext.Runner != nil {
				resolvedActionConfig, err := runContext.Runner.ResolveVariablesInConfig(actionConfig, runContext.VariableContext, runContext.AutomationConfig)
				if err != nil {
					runContext.Logger.Error("Failed to resolve variables in loop action config", "action_type", actionType, "error", err)
					executionError = fmt.Errorf("failed to resolve variables in loop action '%s': %w", actionType, err)
					return executionError
				}
				actionConfig = resolvedActionConfig
			}

			runContext.Logger.Info("Executing runtime loop action", "loop_count", loopCount, "action_index", actionIndex, "action_type", actionType)

			// Store original action context
			originalActionID := runContext.ActionID
			originalParentActionID := runContext.ParentActionID

			// Set loop action context
			loopActionID, _ := actionData["id"].(string)
			if loopActionID == "" {
				loopActionID = fmt.Sprintf("%s-runtime-loop-%d-%d", runContext.ActionID, loopCount, actionIndex)
			}
			runContext.ParentActionID = runContext.ActionID
			runContext.ActionID = loopActionID

			// Get the plugin action
			pluginAction, err := automation.GetAction(actionType)
			if err != nil {
				runContext.Logger.Error("Failed to get runtime loop action", "action_type", actionType, "error", err)
				// Restore original context
				runContext.ActionID = originalActionID
				runContext.ParentActionID = originalParentActionID
				executionError = fmt.Errorf("failed to get runtime loop action '%s': %w", actionType, err)
				return executionError
			}

			// Execute the loop action
			err = pluginAction.Execute(ctx, actionConfig, runContext)
			if err != nil {
				runContext.Logger.Error("Runtime loop action failed", "action_type", actionType, "loop_count", loopCount, "error", err)
				// Restore original context
				runContext.ActionID = originalActionID
				runContext.ParentActionID = originalParentActionID
				executionError = fmt.Errorf("runtime loop action '%s' failed in iteration %d: %w", actionType, loopCount, err)
				return executionError
			}

			// Restore original action context
			runContext.ActionID = originalActionID
			runContext.ParentActionID = originalParentActionID

			runContext.Logger.Info("Runtime loop action completed", "action_type", actionType, "loop_count", loopCount)
		}

		// Small delay to prevent busy-waiting and allow for variable updates
		time.Sleep(100 * time.Millisecond)
	}

	runContext.Logger.Info("Runtime variable loop completed successfully", "total_loops", loopCount)
	return executionError
}

func (a *ApiRuntimeLoopUntilAction) evaluateRuntimeVariableCondition(variablePath, conditionType string, expectedValue interface{}, runContext *automation.RunContext) (bool, error) {
	// Resolve the runtime variable value
	actualValue, err := a.resolveRuntimeVariable(variablePath, runContext)
	if err != nil {
		return false, fmt.Errorf("failed to resolve runtime variable '%s': %w", variablePath, err)
	}

	runContext.Logger.Debug("Evaluating runtime variable condition",
		"variable_path", variablePath,
		"actual_value", actualValue,
		"expected_value", expectedValue,
		"condition_type", conditionType)

	switch conditionType {
	case "equals":
		return fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", expectedValue), nil
	case "not_equals":
		return fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue), nil
	case "contains":
		actualStr := fmt.Sprintf("%v", actualValue)
		expectedStr := fmt.Sprintf("%v", expectedValue)
		return strings.Contains(actualStr, expectedStr), nil
	case "not_contains":
		actualStr := fmt.Sprintf("%v", actualValue)
		expectedStr := fmt.Sprintf("%v", expectedValue)
		return !strings.Contains(actualStr, expectedStr), nil
	case "is_null":
		return actualValue == nil, nil
	case "is_not_null":
		return actualValue != nil, nil
	case "is_true":
		if boolVal, ok := actualValue.(bool); ok {
			return boolVal, nil
		}
		return fmt.Sprintf("%v", actualValue) == "true", nil
	case "is_false":
		if boolVal, ok := actualValue.(bool); ok {
			return !boolVal, nil
		}
		return fmt.Sprintf("%v", actualValue) == "false", nil
	case "greater_than":
		return a.compareNumbers(actualValue, expectedValue, ">")
	case "less_than":
		return a.compareNumbers(actualValue, expectedValue, "<")
	case "greater_than_or_equal":
		return a.compareNumbers(actualValue, expectedValue, ">=")
	case "less_than_or_equal":
		return a.compareNumbers(actualValue, expectedValue, "<=")
	default:
		return false, fmt.Errorf("unsupported condition type: %s", conditionType)
	}
}

func (a *ApiRuntimeLoopUntilAction) compareNumbers(actual, expected interface{}, operator string) (bool, error) {
	actualFloat, err1 := a.toFloat64(actual)
	expectedFloat, err2 := a.toFloat64(expected)

	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}

	switch operator {
	case ">":
		return actualFloat > expectedFloat, nil
	case "<":
		return actualFloat < expectedFloat, nil
	case ">=":
		return actualFloat >= expectedFloat, nil
	case "<=":
		return actualFloat <= expectedFloat, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}

func (a *ApiRuntimeLoopUntilAction) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (a *ApiRuntimeLoopUntilAction) resolveRuntimeVariable(variablePath string, runContext *automation.RunContext) (interface{}, error) {
	if !strings.HasPrefix(variablePath, "runtime.") {
		return nil, fmt.Errorf("variable path must start with 'runtime.'")
	}

	// Remove "runtime." prefix
	path := strings.TrimPrefix(variablePath, "runtime.")
	pathParts := strings.Split(path, ".")

	if len(pathParts) == 0 {
		return nil, fmt.Errorf("empty variable path")
	}

	// Get the base variable
	baseVarName := pathParts[0]
	var baseValue interface{}
	var exists bool

	// Check runtime vars first, then global vars
	if baseValue, exists = runContext.VariableContext.RuntimeVars[baseVarName]; !exists {
		if baseValue, exists = runContext.VariableContext.GlobalVars[baseVarName]; !exists {
			return nil, fmt.Errorf("runtime variable '%s' not found", baseVarName)
		}
	}

	// If only base variable requested, return it
	if len(pathParts) == 1 {
		return baseValue, nil
	}

	// Resolve nested path
	return a.resolveNestedPath(baseValue, pathParts[1:])
}

func (a *ApiRuntimeLoopUntilAction) resolveNestedPath(base interface{}, pathParts []string) (interface{}, error) {
	current := base

	for _, part := range pathParts {
		if current == nil {
			return nil, fmt.Errorf("null value encountered at path segment '%s'", part)
		}

		// Handle array indices (e.g., "options[0]")
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			arrayName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			// Get the array from current object
			var arrayValue interface{}
			if arrayName == "" {
				// Direct array access like [0]
				arrayValue = current
			} else {
				// Named array access like options[0]
				if currentMap, ok := current.(map[string]interface{}); ok {
					var exists bool
					arrayValue, exists = currentMap[arrayName]
					if !exists {
						return nil, fmt.Errorf("array '%s' not found", arrayName)
					}
				} else {
					return nil, fmt.Errorf("cannot access property '%s' on non-object", arrayName)
				}
			}

			arraySlice, ok := arrayValue.([]interface{})
			if !ok {
				return nil, fmt.Errorf("'%s' is not an array", arrayName)
			}

			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s'", indexStr)
			}

			if index < 0 || index >= len(arraySlice) {
				return nil, fmt.Errorf("array index %d out of bounds for array '%s'", index, arrayName)
			}

			current = arraySlice[index]
		} else {
			// Regular object property access
			if currentMap, ok := current.(map[string]interface{}); ok {
				value, exists := currentMap[part]
				if !exists {
					return nil, fmt.Errorf("property '%s' not found", part)
				}
				current = value
			} else {
				return nil, fmt.Errorf("cannot access property '%s' on non-object", part)
			}
		}
	}

	return current, nil
}