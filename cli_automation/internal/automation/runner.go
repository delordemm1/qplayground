package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	outputBaseDir       string
}

// NewRunner creates a new Runner instance.
func NewRunner(storageService storage.StorageService, notificationService notification.NotificationService, outputBaseDir string) *Runner {
	return &Runner{
		storageService:      storageService,
		notificationService: notificationService,
		outputBaseDir:       outputBaseDir,
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

		if err != nil {
			run.Status = "failed"
			run.ErrorMessage = err.Error()
		} else {
			run.Status = "completed"
		}

		// Generate final report
		r.generateFinalReport(automation, run, &automationConfig)
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
					executionError = err // Capture first error
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
		err = executionError
		return err
	}

	slog.Info("Automation completed successfully",
		"automation_id", run.AutomationID,
		"run_id", run.ID,
		"total_runs", runCount)

	return nil
}

// executeSingleRun executes a single run of the automation
func (r *Runner) executeSingleRun(ctx context.Context, automation *Automation, automationConfig *AutomationConfig, run *AutomationRun, loopIndex int, eventCh chan<- RunEvent) error {
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
		LoopIndex:    loopIndex,
		Timestamp:    time.Now().Format("20060102-150405"),
		RunID:        run.ID,
		UserID:       "", // Not applicable for CLI
		ProjectID:    automation.ProjectID,
		AutomationID: automation.ID,
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
		Logger:            slog.Default().With("automation_id", automation.ID, "run_id", run.ID, "loop_index", loopIndex),
		EventCh:           eventCh,
		LoopIndex:         loopIndex,
		Runner:            r,
		VariableContext:   varContext,
		AutomationConfig:  automationConfig,
	}

	totalSteps := len(automation.Steps)
	for stepIndex, step := range automation.Steps {
		// Check for cancellation before each step
		select {
		case <-ctx.Done():
			return fmt.Errorf("automation cancelled")
		default:
		}

		// Update step context
		runContext.StepName = step.Name
		runContext.StepID = step.ID

		runContext.Logger.Info("Executing step", "step_name", step.Name, "step_order", step.StepOrder, "loop_index", loopIndex)

		for _, action := range step.Actions {
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
	}

	return nil
}

// processAllEvents handles events from the shared event channel
func (r *Runner) processAllEvents(ctx context.Context, eventCh <-chan RunEvent, logs *[]map[string]any, outputFiles *[]string, mu *sync.Mutex, run *AutomationRun, done chan<- struct{}) {
	defer close(done)

	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed, save final state and exit
				mu.Lock()
				r.saveRunProgress(run, *logs, *outputFiles)
				mu.Unlock()
				return
			}

			mu.Lock()
			// Process the event
			switch event.Type {
			case RunEventTypeLog:
				logEntry := map[string]any{
					"timestamp":        event.Timestamp.Format(time.RFC3339),
					"step_name":        event.StepName,
					"step_id":          event.StepID,
					"action_id":        event.ActionID,
					"action_name":      event.ActionName,
					"parent_action_id": event.ParentActionID,
					"action_type":      event.ActionType,
					"message":          event.Message,
					"loop_index":       event.LoopIndex,
					"local_loop_index": event.LocalLoopIndex,
					"duration_ms":      event.Duration,
					"status":           "success",
				}
				*logs = append(*logs, logEntry)

			case RunEventTypeError:
				logEntry := map[string]any{
					"timestamp":        event.Timestamp.Format(time.RFC3339),
					"step_name":        event.StepName,
					"step_id":          event.StepID,
					"action_id":        event.ActionID,
					"action_name":      event.ActionName,
					"parent_action_id": event.ParentActionID,
					"action_type":      event.ActionType,
					"error":            event.Error,
					"loop_index":       event.LoopIndex,
					"local_loop_index": event.LocalLoopIndex,
					"duration_ms":      event.Duration,
					"status":           "failed",
				}
				*logs = append(*logs, logEntry)

			case RunEventTypeOutputFile:
				*outputFiles = append(*outputFiles, event.OutputFile)

				// Also add to logs for completeness
				logEntry := map[string]any{
					"timestamp":        event.Timestamp.Format(time.RFC3339),
					"step_name":        event.StepName,
					"step_id":          event.StepID,
					"action_id":        event.ActionID,
					"action_name":      event.ActionName,
					"parent_action_id": event.ParentActionID,
					"action_type":      event.ActionType,
					"output_file":      event.OutputFile,
					"loop_index":       event.LoopIndex,
					"local_loop_index": event.LocalLoopIndex,
					"duration_ms":      event.Duration,
					"status":           "success",
				}
				*logs = append(*logs, logEntry)
			}
			mu.Unlock()

		case <-ctx.Done():
			// Context cancelled, save final state and exit
			mu.Lock()
			r.saveRunProgress(run, *logs, *outputFiles)
			mu.Unlock()
			return
		}
	}
}

// saveRunProgress saves the current logs and output files to the run object
func (r *Runner) saveRunProgress(run *AutomationRun, logs []map[string]any, outputFiles []string) {
	// Update run with current logs and output files
	logsBytes, _ := json.Marshal(logs)
	run.LogsJSON = string(logsBytes)

	outputFilesBytes, _ := json.Marshal(outputFiles)
	run.OutputFilesJSON = string(outputFilesBytes)
}

// generateFinalReport generates the final report after automation completion
func (r *Runner) generateFinalReport(automation *Automation, run *AutomationRun, automationConfig *AutomationConfig) {
	if r.outputBaseDir == "" {
		return // No output directory specified
	}

	err := GenerateReports(automation, run, automationConfig, r.outputBaseDir)
	if err != nil {
		slog.Error("Failed to generate final report", "error", err, "run_id", run.ID)
	}
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
		case "loopIndex":
			return strconv.Itoa(varContext.LoopIndex)
		case "localLoopIndex":
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
		return utils.UtilGenerateUUID()
	case "number":
		return strconv.Itoa(gofakeit.Number(1, 1000))
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	default:
		slog.Warn("Unknown faker method", "method", method)
		return fmt.Sprintf("{{faker.%s}}", method)
	}
}