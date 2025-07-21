package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/playwright-community/playwright-go"
)

// Runner orchestrates the execution of automations.
type Runner struct {
	storageService LocalStorageService
	logger         *LocalLogger
}

// NewRunner creates a new Runner instance.
func NewRunner(storageService LocalStorageService, logger *LocalLogger) *Runner {
	return &Runner{
		storageService: storageService,
		logger:         logger,
	}
}

// RunAutomation executes a given automation from exported config.
func (r *Runner) RunAutomation(ctx context.Context, exportedConfig *ExportedAutomationConfig, runID string) error {
	automation := &exportedConfig.Automation
	automationConfig := &automation.Config

	r.logger.Info("Starting automation execution",
		"automation_name", automation.Name,
		"run_id", runID,
		"steps_count", len(exportedConfig.Steps))

	// Determine run count and mode
	runCount := 1
	runMode := "sequential"
	runDelay := time.Duration(1000) * time.Millisecond

	if automationConfig.Multirun.Enabled {
		runCount = automationConfig.Multirun.Count
		runMode = automationConfig.Multirun.Mode
		runDelay = time.Duration(automationConfig.Multirun.Delay) * time.Millisecond
	}

	r.logger.Info("Automation execution configuration",
		"run_count", runCount,
		"run_mode", runMode,
		"delay_ms", runDelay.Milliseconds())

	// Execute runs based on configuration
	var allLogs []map[string]interface{}
	var allOutputFiles []string
	var executionError error

	if runMode == "parallel" && runCount > 1 {
		// Parallel execution
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < runCount; i++ {
			wg.Add(1)
			go func(loopIndex int) {
				defer wg.Done()

				logs, outputFiles, err := r.executeSingleRun(ctx, exportedConfig, runID, loopIndex)

				mu.Lock()
				allLogs = append(allLogs, logs...)
				allOutputFiles = append(allOutputFiles, outputFiles...)
				if err != nil && executionError == nil {
					executionError = err // Capture first error
				}
				mu.Unlock()
			}(i)
		}
		wg.Wait()
	} else {
		// Sequential execution
		for i := 0; i < runCount; i++ {
			logs, outputFiles, err := r.executeSingleRun(ctx, exportedConfig, runID, i)
			allLogs = append(allLogs, logs...)
			allOutputFiles = append(allOutputFiles, outputFiles...)

			if err != nil {
				executionError = err
				break // Stop on first error in sequential mode
			}

			// Add delay between sequential runs (except for the last one)
			if i < runCount-1 && runDelay > 0 {
				r.logger.Info("Waiting between runs", "delay_ms", runDelay.Milliseconds())
				time.Sleep(runDelay)
			}
		}
	}

	// Save final logs to file
	if len(allLogs) > 0 {
		logsJSON, _ := json.MarshalIndent(allLogs, "", "  ")
		logsFile := fmt.Sprintf("logs/run-%s/final_logs.json", strings.Split(runID, "-")[1])
		r.storageService.UploadFile(ctx, logsFile, strings.NewReader(string(logsJSON)), "application/json")
	}

	if executionError != nil {
		r.logger.Error("Automation execution failed", "error", executionError)
		return executionError
	}

	r.logger.Info("Automation completed successfully",
		"total_runs", runCount,
		"total_output_files", len(allOutputFiles),
		"total_log_entries", len(allLogs))

	return nil
}

// executeSingleRun executes a single run of the automation
func (r *Runner) executeSingleRun(ctx context.Context, exportedConfig *ExportedAutomationConfig, runID string, loopIndex int) ([]map[string]interface{}, []string, error) {
	automation := &exportedConfig.Automation
	automationConfig := &automation.Config

	r.logger.Info("Starting single run execution", "loop_index", loopIndex)

	// Initialize Playwright for this run
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true), // Run headless for CLI
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()

	// Create new page
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		JavaScriptEnabled: playwright.Bool(true),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create page: %w", err)
	}

	// Create variable context for this run
	varContext := &VariableContext{
		LoopIndex:    loopIndex,
		Timestamp:    time.Now().Format("20060102-150405"),
		RunID:        runID,
		UserID:       "cli-user",
		ProjectID:    "cli-project",
		AutomationID: "cli-automation",
		StaticVars:   make(map[string]string),
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
		Logger:            r.logger.With("loop_index", loopIndex),
	}

	var logs []map[string]interface{}
	var outputFiles []string

	// Execute steps
	for stepIndex, step := range exportedConfig.Steps {
		// Check for cancellation before each step
		select {
		case <-ctx.Done():
			return logs, outputFiles, fmt.Errorf("automation cancelled")
		default:
		}

		r.logger.Info("Executing step", "step_name", step.Name, "step_order", step.StepOrder, "loop_index", loopIndex)

		// Execute actions for this step
		for _, action := range step.Actions {
			// Check for cancellation before each action
			select {
			case <-ctx.Done():
				return logs, outputFiles, fmt.Errorf("automation cancelled")
			default:
			}

			startTime := time.Now()

			// Resolve variables in action config
			resolvedActionConfig, resolveErr := r.resolveVariablesInConfig(action.ActionConfig, varContext, automationConfig)
			if resolveErr != nil {
				return nil, nil, fmt.Errorf("failed to resolve variables in action config: %w", resolveErr)
			}

			// Get plugin action
			pluginAction, getActionErr := GetAction(action.ActionType)
			if getActionErr != nil {
				return nil, nil, fmt.Errorf("unregistered plugin action type '%s': %w", action.ActionType, getActionErr)
			}

			// Execute action
			actionErr := pluginAction.Execute(ctx, resolvedActionConfig, runContext)
			duration := time.Since(startTime)

			// Create log entry
			logEntry := map[string]interface{}{
				"timestamp":   time.Now().Format(time.RFC3339),
				"step_name":   step.Name,
				"action_type": action.ActionType,
				"loop_index":  loopIndex,
				"duration_ms": duration.Milliseconds(),
				"status":      "success",
			}

			if actionErr != nil {
				logEntry["status"] = "failed"
				logEntry["error"] = actionErr.Error()
				logs = append(logs, logEntry)

				r.logger.Error("Action failed",
					"action_type", action.ActionType,
					"error", actionErr,
					"duration", duration,
					"loop_index", loopIndex)

				return logs, outputFiles, fmt.Errorf("action '%s' failed: %w", action.ActionType, actionErr)
			}

			// Check if this was a screenshot action that saved locally
			if action.ActionType == "playwright:screenshot" {
				if localPath, ok := resolvedActionConfig["local_path"].(string); ok && localPath != "" {
					outputFiles = append(outputFiles, localPath)
					logEntry["output_file"] = localPath
				}
			}

			logs = append(logs, logEntry)
			r.logger.Info("Action completed",
				"action_type", action.ActionType,
				"duration", duration,
				"loop_index", loopIndex)
		}
	}

	return logs, outputFiles, nil
}

// resolveVariablesInConfig resolves variables in action configuration
func (r *Runner) resolveVariablesInConfig(config map[string]interface{}, varContext *VariableContext, automationConfig *ExportedAutomationMeta) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	for key, value := range config {
		switch v := value.(type) {
		case string:
			resolvedValue, err := r.resolveVariablesInString(v, varContext, automationConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variables in field '%s': %w", key, err)
			}
			resolved[key] = resolvedValue
		case map[string]interface{}:
			// Recursively resolve nested objects
			nestedResolved, err := r.resolveVariablesInConfig(v, varContext, automationConfig)
			if err != nil {
				return nil, err
			}
			resolved[key] = nestedResolved
		default:
			// For non-string values, keep as-is
			resolved[key] = value
		}
	}

	return resolved, nil
}

// resolveVariablesInString resolves variables in a string value
func (r *Runner) resolveVariablesInString(input string, varContext *VariableContext, automationConfig *ExportedAutomationMeta) (string, error) {
	// Pattern to match {{variableName}} or {{faker.method}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.Trim(match, "{}")

		// Handle environment variables
		switch varName {
		case "loopIndex":
			return strconv.Itoa(varContext.LoopIndex)
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

		// Handle faker variables
		if strings.HasPrefix(varName, "faker.") {
			fakerMethod := strings.TrimPrefix(varName, "faker.")
			return r.generateFakerValue(fakerMethod)
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
					v, err := r.resolveVariablesInString(variable.Value, varContext, automationConfig)
					if err != nil {
						return ""
					}
					return v
				}
			}
		}

		// If no match found, return the original placeholder
		r.logger.Warn("Unresolved variable", "variable", varName)
		return match
	})

	return result, nil
}

// generateFakerValue generates a fake value based on the faker method
func (r *Runner) generateFakerValue(method string) string {
	gofakeit.Seed(time.Now().UnixNano()) // Ensure randomness

	switch method {
	case "name":
		return gofakeit.Name()
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
		r.logger.Warn("Unknown faker method", "method", method)
		return fmt.Sprintf("{{faker.%s}}", method)
	}
}