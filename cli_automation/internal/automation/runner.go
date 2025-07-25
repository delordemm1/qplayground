package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/delordemm1/qplayground-cli/internal/notification"
	"github.com/delordemm1/qplayground-cli/internal/storage"
	"github.com/delordemm1/qplayground-cli/internal/utils"
	"github.com/playwright-community/playwright-go"
)

// Runner orchestrates the execution of automations.
type Runner struct {
	storageService      storage.StorageService
	notificationService notification.NotificationService
	outputDir           string
}

// NewRunner creates a new Runner instance.
func NewRunner(storageService storage.StorageService, notificationService notification.NotificationService, outputDir string) *Runner {
	return &Runner{
		storageService:      storageService,
		notificationService: notificationService,
		outputDir:           outputDir,
	}
}

// RunAutomation executes a given automation.
func (r *Runner) RunAutomation(ctx context.Context, automation *Automation, run *AutomationRun) error {
	// Parse automation configuration
	var automationConfig AutomationConfig
	if automation.ConfigJSON != "" {
		if err := json.Unmarshal([]byte(automation.ConfigJSON), &automationConfig); err != nil {
			return fmt.Errorf("failed to parse automation config: %w", err)
		}
	} else {
		// Use default configuration if none provided
		automationConfig = AutomationConfig{
			Variables: []Variable{},
			Multirun: MultiRunConfig{
				Enabled: false,
				Mode:    "sequential",
				Count:   1,
				Delay:   1000,
			},
			Timeout:       300,
			Retries:       0,
			Screenshots:   ScreenshotConfig{Enabled: true, OnError: true, OnSuccess: false, Path: "screenshots/{{timestamp}}-{{loopIndex}}.png"},
			Notifications: []NotificationChannelConfig{},
		}
	}

	// Set start time
	now := time.Now()
	run.StartTime = &now

	// Ensure run status is updated on exit
	defer func() {
		endTime := time.Now()
		run.EndTime = &endTime

		if rec := recover(); rec != nil {
			run.Status = "failed"
			run.ErrorMessage = fmt.Sprintf("panic: %v", rec)
			panic(rec) // Re-throw panic
		}

		if run.Status == "" {
			if run.ErrorMessage != "" {
				run.Status = "failed"
			} else {
				run.Status = "completed"
			}
		}
	}()

	// Determine run count and mode
	runCount := 1
	runMode := "sequential"
	runDelay := time.Duration(1000) * time.Millisecond

	if automationConfig.Multirun.Enabled {
		runCount = automationConfig.Multirun.Count
		runMode = automationConfig.Multirun.Mode
		runDelay = time.Duration(automationConfig.Multirun.Delay) * time.Millisecond
	}

	slog.Info("Starting automation execution",
		"automation_id", run.AutomationID,
		"run_id", run.ID,
		"run_count", runCount,
		"run_mode", runMode)

	// Create shared event channel and data structures for all runs
	eventCh := make(chan RunEvent, 1000) // Large buffer for concurrent runs
	var allLogs []map[string]any
	var allOutputFiles []string
	var mu sync.Mutex // Protect shared data structures

	// Start single event processor for all runs
	eventProcessorDone := make(chan struct{})
	go r.processAllEvents(ctx, eventCh, &allLogs, &allOutputFiles, &mu, run, eventProcessorDone)

	// Execute runs based on configuration
	var executionError error

	if runMode == "parallel" && runCount > 1 {
		// Parallel execution
		var wg sync.WaitGroup

		for i := 0; i < runCount; i++ {
			wg.Add(1)
			go func(loopIndex int) {
				defer wg.Done()
				err := r.executeSingleRun(ctx, automation, &automationConfig, run, loopIndex, eventCh)

				if err != nil {
					// For parallel execution, we'll just log the error
					// The first error will be captured in executionError
					slog.Error("Parallel run failed", "loop_index", loopIndex, "error", err)
					mu.Lock()
					if executionError == nil {
						executionError = err // Capture first error
					}
					mu.Unlock()
				}
			}(i)
		}
		wg.Wait()
	} else {
		// Sequential execution
		for i := 0; i < runCount; i++ {
			err := r.executeSingleRun(ctx, automation, &automationConfig, run, i, eventCh)

			if err != nil {
				executionError = err
				break // Stop on first error in sequential mode
			}

			// Add delay between sequential runs (except for the last one)
			if i < runCount-1 && runDelay > 0 {
				time.Sleep(runDelay)
			}
		}
	}

	// Close event channel and wait for processor to finish
	close(eventCh)
	<-eventProcessorDone

	if executionError != nil {
		run.Status = "failed"
		run.ErrorMessage = executionError.Error()
		// Send error notifications
		go r.sendNotifications(context.Background(), automation, run, &automationConfig)
		return executionError
	}

	slog.Info("Automation completed successfully",
		"automation_id", run.AutomationID,
		"run_id", run.ID,
		"total_runs", runCount)

	// Send completion notifications
	go r.sendNotifications(context.Background(), automation, run, &automationConfig)

	return nil
}

// executeSingleRun executes a single run of the automation
func (r *Runner) executeSingleRun(ctx context.Context, automation *Automation, automationConfig *AutomationConfig, run *AutomationRun, loopIndex int, eventCh chan RunEvent) error {

	// Initialize Playwright for this run
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true), // Run headless for automation
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
		},
	})
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()

	// Create new page with context
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		JavaScriptEnabled: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	// Create variable context for this run
	varContext := &VariableContext{
		LoopIndex:      loopIndex,
		LocalLoopIndex: 0, // Will be updated by nested loops
		Timestamp:      time.Now().Format("20060102-150405"),
		RunID:          run.ID,
		UserID:         "", // TODO: Get from context if available
		ProjectID:      automation.ProjectID,
		AutomationID:   automation.ID,
		StaticVars:     make(map[string]string),
		RuntimeVars:    make(map[string]interface{}),
		GlobalVars:     make(map[string]interface{}),
	}

	// Build static variables map
	for _, variable := range automationConfig.Variables {
		if variable.Type == "static" {
			varContext.StaticVars[variable.Key] = variable.Value
		}
	}

	// Create RunContext
	runContext := &RunContext{
		PlaywrightBrowser: browser,
		PlaywrightPage:    page,
		StorageService:    r.storageService,
		Logger:            slog.Default().With("automation_id", automation.ID, "run_id", run.ID, "loop_index", loopIndex),
		EventCh:           eventCh,
		LoopIndex:         loopIndex,
		Runner:            r,
		VariableContext:   varContext,
		AutomationConfig:  automationConfig,
	}

	// Fetch and execute steps
	steps := automation.Steps
	if steps == nil {
		return fmt.Errorf("no steps found for automation")
	}

	totalSteps := len(steps)
	for stepIndex, step := range steps {
		// Check for cancellation before each step
		select {
		case <-ctx.Done():
			return fmt.Errorf("automation cancelled")
		default:
		}

		// Parse step configuration and check for skip conditions
		shouldSkipStep := false

		if step.ConfigJSON != "" {
			var stepConfigMap map[string]interface{}
			if err := json.Unmarshal([]byte(step.ConfigJSON), &stepConfigMap); err != nil {
				runContext.Logger.Warn("Failed to parse step config JSON", "step_id", step.ID, "error", err)
			} else {
				// Check for skip_condition
				if skipCondition, ok := stepConfigMap["skip_condition"].(string); ok && skipCondition != "" {
					probability := 0.5 // Default probability
					if prob, ok := stepConfigMap["probability"].(float64); ok {
						probability = prob
					}

					shouldSkip := evaluateLoopIndexCondition(skipCondition, loopIndex, probability)
					if shouldSkip {
						shouldSkipStep = true
						runContext.Logger.Info("Skipping step due to skip condition",
							"step_name", step.Name,
							"condition", skipCondition,
							"loop_index", loopIndex)
					}
				}

				// Check for run_only_condition
				if runOnlyCondition, ok := stepConfigMap["run_only_condition"].(string); ok && runOnlyCondition != "" {
					probability := 0.5 // Default probability
					if prob, ok := stepConfigMap["probability"].(float64); ok {
						probability = prob
					}

					shouldRun := evaluateLoopIndexCondition(runOnlyCondition, loopIndex, probability)
					if !shouldRun {
						shouldSkipStep = true
						runContext.Logger.Info("Skipping step due to run_only condition not met",
							"step_name", step.Name,
							"condition", runOnlyCondition,
							"loop_index", loopIndex)
					}
				}
			}
		}

		// Skip this step if conditions indicate so
		if shouldSkipStep {
			continue
		}

		// Update step context
		runContext.StepName = step.Name
		runContext.StepID = step.ID

		runContext.Logger.Info("Executing step", "step_name", step.Name, "step_order", step.StepOrder, "loop_index", loopIndex)

		// Get actions for this step
		stepActions := step.Actions
		if stepActions == nil {
			stepActions = []*AutomationAction{}
		}

		for _, action := range stepActions {
			// Check for cancellation before each action
			select {
			case <-ctx.Done():
				return fmt.Errorf("automation cancelled")
			default:
			}

			// Parse action config
			actionConfigMap := make(map[string]any)
			if action.ActionConfigJSON != "" {
				if jsonErr := json.Unmarshal([]byte(action.ActionConfigJSON), &actionConfigMap); jsonErr != nil {
					return fmt.Errorf("failed to parse action config JSON for action %s: %w", action.ActionType, jsonErr)
				}
			}

			// Resolve variables in action config
			resolvedActionConfig, resolveErr := r.ResolveVariablesInConfig(actionConfigMap, varContext, automationConfig)
			if resolveErr != nil {
				return fmt.Errorf("failed to resolve variables in action config: %w", resolveErr)
			}

			// Get plugin action
			pluginAction, getActionErr := GetAction(action.ActionType)
			if getActionErr != nil {
				return fmt.Errorf("unregistered plugin action type '%s': %w", action.ActionType, getActionErr)
			}

			runContext.ActionID = action.ID
			runContext.ActionName = action.Name
			runContext.ParentActionID = "" // Reset for top-level actions

			// Execute action
			actionErr := pluginAction.Execute(ctx, resolvedActionConfig, runContext)

			if actionErr != nil {
				runContext.Logger.Error("Action failed",
					"action_type", action.ActionType,
					"action_name", action.Name,
					"error", actionErr,
					"loop_index", loopIndex)

				return fmt.Errorf("action '%s' failed: %w", action.ActionType, actionErr)
			}

			runContext.Logger.Info("Action completed",
				"action_type", action.ActionType,
				"action_name", action.Name,
				"loop_index", loopIndex)
		}

		runContext.Logger.Info("Step completed", "step_name", step.Name, "step_index", stepIndex+1, "total_steps", totalSteps)
	}

	return nil
}

// processAllEvents handles events from the shared event channel and updates the run data periodically
func (r *Runner) processAllEvents(ctx context.Context, eventCh <-chan RunEvent, logs *[]map[string]any, outputFiles *[]string, mu *sync.Mutex, run *AutomationRun, done chan<- struct{}) {
	defer close(done)

	ticker := time.NewTicker(5 * time.Second) // Save to storage every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed, save final state and exit
				mu.Lock()
				r.saveRunProgress(ctx, run, *logs, *outputFiles)
				mu.Unlock()
				return
			}

			mu.Lock()
			// Process the event
			switch event.Type {
			case RunEventTypeLog:
				logEntry := map[string]any{
					"parent_action_id": event.ParentActionID,
					"local_loop_index": event.LocalLoopIndex,
					"timestamp":        event.Timestamp.Format(time.RFC3339),
					"step_name":        event.StepName,
					"step_id":          event.StepID,
					"action_id":        event.ActionID,
					"action_name":      event.ActionName,
					"action_type":      event.ActionType,
					"message":          event.Message,
					"loop_index":       event.LoopIndex,
					"duration_ms":      event.Duration,
					"status":           "success",
				}
				*logs = append(*logs, logEntry)

			case RunEventTypeError:
				logEntry := map[string]any{
					"parent_action_id": event.ParentActionID,
					"local_loop_index": event.LocalLoopIndex,
					"timestamp":        event.Timestamp.Format(time.RFC3339),
					"step_name":        event.StepName,
					"step_id":          event.StepID,
					"action_id":        event.ActionID,
					"action_name":      event.ActionName,
					"action_type":      event.ActionType,
					"error":            event.Error,
					"loop_index":       event.LoopIndex,
					"duration_ms":      event.Duration,
					"status":           "failed",
				}
				*logs = append(*logs, logEntry)

			case RunEventTypeOutputFile:
				*outputFiles = append(*outputFiles, event.OutputFile)

				// Also add to logs for completeness
				logEntry := map[string]any{
					"parent_action_id": event.ParentActionID,
					"local_loop_index": event.LocalLoopIndex,
					"timestamp":        event.Timestamp.Format(time.RFC3339),
					"step_name":        event.StepName,
					"step_id":          event.StepID,
					"action_id":        event.ActionID,
					"action_name":      event.ActionName,
					"action_type":      event.ActionType,
					"output_file":      event.OutputFile,
					"loop_index":       event.LoopIndex,
					"duration_ms":      event.Duration,
					"status":           "success",
				}
				*logs = append(*logs, logEntry)
			}
			mu.Unlock()

		case <-ticker.C:
			// Periodic save to storage
			mu.Lock()
			r.saveRunProgress(ctx, run, *logs, *outputFiles)
			mu.Unlock()

		case <-ctx.Done():
			// Context cancelled, save final state and exit
			mu.Lock()
			r.saveRunProgress(ctx, run, *logs, *outputFiles)
			mu.Unlock()
			return
		}
	}
}

// saveRunProgress saves the current logs and output files to the run
func (r *Runner) saveRunProgress(ctx context.Context, run *AutomationRun, logs []map[string]any, outputFiles []string) {
	// Update run with current logs and output files
	logsBytes, _ := json.Marshal(logs)
	run.LogsJSON = string(logsBytes)

	outputFilesBytes, _ := json.Marshal(outputFiles)
	run.OutputFilesJSON = string(outputFilesBytes)

	// Note: In CLI mode, we don't have a database to update, so we just keep the data in memory
	// The final report generation will use this data
}

// resolveVariablesInConfig resolves variables in action configuration
func (r *Runner) ResolveVariablesInConfig(config map[string]any, varContext *VariableContext, automationConfig *AutomationConfig) (map[string]any, error) {
	resolved := make(map[string]any)

	for key, value := range config {
		switch v := value.(type) {
		case string:
			resolvedValue, err := r.ResolveVariablesInString(v, varContext, automationConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variables in field '%s': %w", key, err)
			}
			resolved[key] = resolvedValue
		case map[string]any:
			// Recursively resolve nested objects
			nestedResolved, err := r.ResolveVariablesInConfig(v, varContext, automationConfig)
			if err != nil {
				return nil, err
			}
			resolved[key] = nestedResolved
		case []interface{}:
			// Handle arrays that might contain objects with variables
			resolvedArray := make([]interface{}, len(v))
			for i, item := range v {
				switch itemVal := item.(type) {
				case string:
					resolvedItem, err := r.ResolveVariablesInString(itemVal, varContext, automationConfig)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve variables in array item %d: %w", i, err)
					}
					resolvedArray[i] = resolvedItem
				case map[string]any:
					resolvedItem, err := r.ResolveVariablesInConfig(itemVal, varContext, automationConfig)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve variables in array item %d: %w", i, err)
					}
					resolvedArray[i] = resolvedItem
				default:
					resolvedArray[i] = item
				}
			}
			resolved[key] = resolvedArray
		default:
			// For non-string values, keep as-is
			resolved[key] = value
		}
	}

	return resolved, nil
}

// resolveVariablesInString resolves variables in a string value
func (r *Runner) ResolveVariablesInString(input string, varContext *VariableContext, automationConfig *AutomationConfig) (string, error) {
	// Pattern to match {{variableName}} or {{faker.method}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.Trim(match, "{}")

		// Handle environment variables
		switch varName {
		case "runtime":
			// This shouldn't happen as runtime variables should be accessed as {{runtime.varname}}
			return varName
		case "loopIndex":
			return strconv.Itoa(varContext.LoopIndex)
		case "localLoopIndex": // Add this case
			return strconv.Itoa(varContext.LocalLoopIndex)
		case "timestamp":
			return varContext.Timestamp
		case "runId":
			return varContext.RunID
		case "userId":
			return varContext.UserID
		case "projectId":
			return varContext.ProjectID
		case "automationId":
			return varContext.AutomationID
		}

		// Handle runtime variables ({{runtime.varname}})
		if strings.HasPrefix(varName, "runtime.") {
			// Enhanced runtime variable resolution with nested path support
			resolvedValue, err := r.resolveRuntimeVariable(varName, varContext)
			if err != nil {
				slog.Warn("Failed to resolve runtime variable", "variable", varName, "error", err)
				return ""
			}
			return fmt.Sprintf("%v", resolvedValue)
		}

		// Handle faker variables
		if strings.HasPrefix(varName, "faker.") {
			fakerMethod := strings.TrimPrefix(varName, "faker.")
			return r.generateFakerValue(fakerMethod)
		}

		// Handle function variables
		if strings.HasPrefix(varName, "function.") {
			fakerMethod := strings.TrimPrefix(varName, "function.")
			return r.generateFunctionValue(fakerMethod)
		}

		// Handle static variables
		if value, exists := varContext.StaticVars[varName]; exists {
			return value
		}

		// Handle dynamic variables from config
		for _, variable := range automationConfig.Variables {
			if variable.Key == varName {
				switch variable.Type {
				case "static":
					return variable.Value
				case "dynamic":
					// Variable.Value contains the faker method (e.g., "{{faker.email}}")
					if strings.HasPrefix(variable.Value, "{{faker.") && strings.HasSuffix(variable.Value, "}}") {
						fakerMethod := strings.TrimPrefix(strings.TrimSuffix(variable.Value, "}}"), "{{faker.")
						return r.generateFakerValue(fakerMethod)
					}
					return variable.Value
				case "environment":
					// Variable.Value contains the environment variable (e.g., "{{timestamp}}")
					v, err := r.ResolveVariablesInString(variable.Value, varContext, automationConfig)
					if err != nil {
						return ""
					}
					return v
				}
			}
		}

		// If no match found, return the original placeholder
		slog.Warn("Unresolved variable", "variable", varName)
		return match
	})

	return result, nil
}

// resolveRuntimeVariable resolves runtime variables with support for nested paths
func (r *Runner) resolveRuntimeVariable(variablePath string, varContext *VariableContext) (interface{}, error) {
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
	if baseValue, exists = varContext.RuntimeVars[baseVarName]; !exists {
		if baseValue, exists = varContext.GlobalVars[baseVarName]; !exists {
			return nil, fmt.Errorf("runtime variable '%s' not found", baseVarName)
		}
	}

	// If only base variable requested, return it
	if len(pathParts) == 1 {
		return baseValue, nil
	}

	// Resolve nested path
	return r.resolveNestedPath(baseValue, pathParts[1:])
}

// resolveNestedPath traverses nested objects and arrays to resolve complex paths
func (r *Runner) resolveNestedPath(base interface{}, pathParts []string) (interface{}, error) {
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

// evaluateLoopIndexCondition evaluates loop index based conditions
func evaluateLoopIndexCondition(conditionType string, loopIndex int, probability float64) bool {
	switch conditionType {
	case "loop_index_is_even":
		return loopIndex%2 == 0
	case "loop_index_is_odd":
		return loopIndex%2 != 0
	case "loop_index_is_prime":
		return isPrime(loopIndex)
	case "random":
		return rand.Float64() < probability
	default:
		return false
	}
}

// isPrime checks if a number is prime
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// generateFakerValue generates a fake value based on the faker method
func (r *Runner) generateFakerValue(method string) string {
	gofakeit.Seed(time.Now().UnixNano()) // Ensure randomness

	switch method {
	case "name":
		return gofakeit.Name()
	case "lastName":
		return gofakeit.LastName()
	case "firstName":
		return gofakeit.FirstName()
	case "email":
		return gofakeit.Email()
	case "phone":
		return gofakeit.Phone()
	case "address":
		return gofakeit.Address().Address
	case "company":
		return gofakeit.Company()
	case "username":
		return gofakeit.Username()
	case "password":
		return gofakeit.Password(true, true, true, true, false, 12)
	case "uuid":
		return gofakeit.UUID()
	case "number":
		return strconv.Itoa(gofakeit.Number(1, 1000))
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	default:
		slog.Warn("Unknown faker method", "method", method)
		return fmt.Sprintf("{{faker.%s}}", method)
	}
}

// generateFunctionValue generates a fake value based on custom functions
func (r *Runner) generateFunctionValue(method string) string {
	gofakeit.Seed(time.Now().UnixNano()) // Ensure randomness

	switch method {
	case "randomNumber.6":
		return strconv.Itoa(gofakeit.Number(111111, 999999))
	default:
		slog.Warn("Unknown function method", "method", method)
		return fmt.Sprintf("{{function.%s}}", method)
	}
}

// sendNotifications sends notifications based on the automation configuration
func (r *Runner) sendNotifications(ctx context.Context, automation *Automation, run *AutomationRun, automationConfig *AutomationConfig) {
	if len(automationConfig.Notifications) == 0 {
		return // No notifications configured
	}

	// Get project information (you might need to add this to the Runner or pass it in)
	// For now, we'll use the project ID from the automation
	projectName := "CLI Project" // TODO: Fetch actual project name if needed

	// Parse output files from run
	var outputFiles []string
	if run.OutputFilesJSON != "" {
		json.Unmarshal([]byte(run.OutputFilesJSON), &outputFiles)
	}

	// Parse logs from run
	var logs []map[string]any
	if run.LogsJSON != "" {
		json.Unmarshal([]byte(run.LogsJSON), &logs)
	}

	// Build notification message
	message := notification.NotificationMessage{
		AutomationID:   automation.ID,
		AutomationName: automation.Name,
		ProjectID:      automation.ProjectID,
		ProjectName:    projectName,
		RunID:          run.ID,
		Status:         run.Status,
		StartTime:      run.StartTime,
		EndTime:        run.EndTime,
		ErrorMessage:   run.ErrorMessage,
		OutputFiles:    outputFiles,
		LogsCount:      len(logs),
	}

	// Convert our config to the notification service format
	channels := make([]notification.NotificationChannelConfig, len(automationConfig.Notifications))
	for i, channel := range automationConfig.Notifications {
		channels[i] = notification.NotificationChannelConfig{
			ID:         channel.ID,
			Type:       channel.Type,
			OnComplete: channel.OnComplete,
			OnError:    channel.OnError,
			Config:     channel.Config,
		}
	}

	// Dispatch notifications
	err := r.notificationService.DispatchAutomationNotification(ctx, message, channels)
	if err != nil {
		slog.Error("Failed to dispatch automation notifications",
			"automation_id", automation.ID,
			"run_id", run.ID,
			"error", err)
	}
}