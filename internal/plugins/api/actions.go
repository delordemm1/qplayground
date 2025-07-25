package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
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
			extractedValue, err := b.extractJSONPath(responseJSON, hook.Path)
			if err != nil {
				runContext.Logger.Warn("Failed to extract value from JSON path", "path", hook.Path, "error", err)
				continue
			}

			// Determine scope and save variable
			if hook.Scope == "global" {
				// Save to static vars (global scope)
				runContext.VariableContext.StaticVars[hook.SaveAs] = fmt.Sprintf("%v", extractedValue)
			} else {
				// Save to runtime vars (local scope - default)
				runContext.VariableContext.RuntimeVars[hook.SaveAs] = extractedValue
			}

			responseData.ExtractedVars[hook.SaveAs] = extractedValue
			runContext.Logger.Info("Extracted runtime variable", 
				"path", hook.Path, 
				"save_as", hook.SaveAs, 
				"value", extractedValue,
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