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
	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/playwright-community/playwright-go"
)

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

// NotificationConfig represents notification configuration
type NotificationChannelConfig struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "slack", "email", "webhook"
	OnComplete bool                   `json:"onComplete"`
	OnError    bool                   `json:"onError"`
	Config     map[string]interface{} `json:"config"`
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

// Runner orchestrates the execution of automations.
type Runner struct {
	automationRepo      AutomationRepository
	storageService      storage.StorageService
	notificationService notification.NotificationService
	sseManager          *SSEManager
}

// NewRunner creates a new Runner instance.
func NewRunner(automationRepo AutomationRepository, storageService storage.StorageService, notificationService notification.NotificationService, sseManager *SSEManager) *Runner {
	return &Runner{
		automationRepo:      automationRepo,
		storageService:      storageService,
		notificationService: notificationService,
		sseManager:          sseManager,
	}
}

// RunAutomation executes a given automation.
func (r *Runner) RunAutomation(ctx context.Context, projectID string, run *AutomationRun) error {
	// Create a ticker for polling cancellation status every 5 seconds
	statusTicker := time.NewTicker(5 * time.Second)
	defer statusTicker.Stop()

	// 1. Fetch Automation details from DB
	automation, err := r.automationRepo.GetAutomationByID(ctx, run.AutomationID)
	if err != nil {
		return fmt.Errorf("failed to get automation: %w", err)
	}

	// 2. Parse automation configuration
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

	// 2. Create a new AutomationRun record (status: running)
	// run := &AutomationRun{
	// 	ID:              platform.UtilGenerateUUID(),
	// 	AutomationID:    automationID,
	// 	Status:          "running",
	// 	LogsJSON:        "[]",
	// 	OutputFilesJSON: "[]",
	// }

	// Set start time
	now := time.Now()
	run.StartTime = &now

	// err = r.automationRepo.CreateRun(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to create automation run record: %w", err)
	}

	// Ensure run status is updated on exit
	defer func() {
		endTime := time.Now()
		run.EndTime = &endTime

		if rec := recover(); rec != nil {
			run.Status = "failed"
			run.ErrorMessage = fmt.Sprintf("panic: %v", rec)
			r.automationRepo.UpdateRun(ctx, run)
			panic(rec) // Re-throw panic
		}

		if err != nil {
			run.Status = "failed"
			run.ErrorMessage = err.Error()
		} else {
			run.Status = "completed"
		}

		r.automationRepo.UpdateRun(ctx, run)
	}()

	// 3. Determine run count and mode
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

	// Send initial status update via SSE
	if r.sseManager != nil {
		r.sseManager.SendRunStatusUpdate(projectID, run.AutomationID, run.ID, "running")
	}

	// 4. Execute runs based on configuration
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

				logs, outputFiles, err := r.executeSingleRun(ctx, automation, &automationConfig, run, loopIndex)

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
			logs, outputFiles, err := r.executeSingleRun(ctx, automation, &automationConfig, run, i)
			allLogs = append(allLogs, logs...)
			allOutputFiles = append(allOutputFiles, outputFiles...)

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

	// 5. Update run record with final logs and output files
	logsBytes, _ := json.Marshal(allLogs)
	run.LogsJSON = string(logsBytes)

	outputFilesBytes, _ := json.Marshal(allOutputFiles)
	run.OutputFilesJSON = string(outputFilesBytes)

	if executionError != nil {
		err = executionError
		// Send error notifications
		go r.sendNotifications(context.Background(), automation, run, &automationConfig, allLogs, allOutputFiles)
		return err
	}

	slog.Info("Automation completed successfully",
		"automation_id", run.AutomationID,
		"run_id", run.ID,
		"total_runs", runCount,
		"total_output_files", len(allOutputFiles))

	// Send completion update via SSE
	if r.sseManager != nil {
		totalDuration := int64(0)
		if run.StartTime != nil && run.EndTime != nil {
			totalDuration = run.EndTime.Sub(*run.StartTime).Milliseconds()
		}
		r.sseManager.SendRunComplete(projectID, run.AutomationID, run.ID, "completed", totalDuration, allOutputFiles)
	}

	// Send completion notifications
	go r.sendNotifications(context.Background(), automation, run, &automationConfig, allLogs, allOutputFiles)

	return nil
}

// executeSingleRun executes a single run of the automation
func (r *Runner) executeSingleRun(ctx context.Context, automation *Automation, automationConfig *AutomationConfig, run *AutomationRun, loopIndex int) ([]map[string]interface{}, []string, error) {
	// Create a ticker for polling cancellation status every 5 seconds
	statusTicker := time.NewTicker(5 * time.Second)
	defer statusTicker.Stop()

	// Start a goroutine to poll for cancellation
	cancelCh := make(chan struct{})
	go func() {
		defer close(cancelCh)
		for {
			select {
			case <-statusTicker.C:
				// Check if run was cancelled by polling Redis (via storage service if available)
				// For now, we'll rely on context cancellation
			case <-ctx.Done():
				return
			}
		}
	}()

	// Initialize Playwright for this run
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start playwright: %w", err)
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
		return nil, nil, fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()

	// Create new page with context
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create page: %w", err)
	}

	// Create variable context for this run
	varContext := &VariableContext{
		LoopIndex:    loopIndex,
		Timestamp:    time.Now().Format("20060102-150405"),
		RunID:        run.ID,
		UserID:       "", // TODO: Get from context if available
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
	}

	// Fetch and execute steps
	steps, err := r.automationRepo.GetStepsByAutomationID(ctx, automation.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get automation steps: %w", err)
	}

	var logs []map[string]interface{}
	var outputFiles []string

	totalSteps := len(steps)
	for stepIndex, step := range steps {
		// Check for cancellation before each step
		select {
		case <-ctx.Done():
			return logs, outputFiles, fmt.Errorf("automation cancelled")
		default:
		}

		runContext.Logger.Info("Executing step", "step_name", step.Name, "step_order", step.StepOrder, "loop_index", loopIndex)

		// Send step progress update via SSE
		if r.sseManager != nil {
			r.sseManager.SendRunStep(automation.ProjectID, run.AutomationID, run.ID, step.Name, stepIndex+1, totalSteps)
		}

		// Get actions for this step
		stepActions, err := r.automationRepo.GetActionsByStepID(ctx, step.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get actions for step %s: %w", step.Name, err)
		}

		for _, action := range stepActions {
			// Check for cancellation before each action
			select {
			case <-ctx.Done():
				return logs, outputFiles, fmt.Errorf("automation cancelled")
			default:
			}

			startTime := time.Now()

			// Parse action config
			actionConfigMap := make(map[string]interface{})
			if action.ActionConfigJSON != "" {
				if jsonErr := json.Unmarshal([]byte(action.ActionConfigJSON), &actionConfigMap); jsonErr != nil {
					return nil, nil, fmt.Errorf("failed to parse action config JSON for action %s: %w", action.ActionType, jsonErr)
				}
			}

			// Resolve variables in action config
			resolvedActionConfig, resolveErr := r.resolveVariablesInConfig(actionConfigMap, varContext, automationConfig)
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

			// Send real-time log update via SSE
			if r.sseManager != nil {
				if actionErr != nil {
					r.sseManager.SendRunError(automation.ProjectID, run.AutomationID, run.ID, step.Name, action.ActionType, actionErr.Error())
				} else {
					r.sseManager.SendRunLog(automation.ProjectID, run.AutomationID, run.ID, step.Name, action.ActionType, "Action completed successfully", duration.Milliseconds())
				}
			}

			// Create log entry
			logEntry := map[string]interface{}{
				"timestamp":   time.Now().Format(time.RFC3339),
				"step_id":     step.ID,
				"step_name":   step.Name,
				"action_id":   action.ID,
				"action_type": action.ActionType,
				"loop_index":  loopIndex,
				"duration_ms": duration.Milliseconds(),
				"status":      "success",
			}

			if actionErr != nil {
				logEntry["status"] = "failed"
				logEntry["error"] = actionErr.Error()
				logs = append(logs, logEntry)

				runContext.Logger.Error("Action failed",
					"action_type", action.ActionType,
					"error", actionErr,
					"duration", duration,
					"loop_index", loopIndex)

				return logs, outputFiles, fmt.Errorf("action '%s' failed: %w", action.ActionType, actionErr)
			}

			// Check if this was a screenshot action that uploaded to R2
			if action.ActionType == "playwright:screenshot" {
				if uploadToR2, ok := resolvedActionConfig["upload_to_r2"].(bool); ok && uploadToR2 {
					if r2Key, ok := resolvedActionConfig["r2_key"].(string); ok && r2Key != "" {
						// Get the public URL for the uploaded screenshot
						publicURL := r.storageService.GetPublicURL(r2Key)
						outputFiles = append(outputFiles, publicURL)
						logEntry["output_file"] = publicURL

						// Send output file update via SSE
						if r.sseManager != nil {
							r.sseManager.SendRunOutputFile(automation.ProjectID, run.AutomationID, run.ID, publicURL)
						}
					}
				}
			}

			logs = append(logs, logEntry)
			runContext.Logger.Info("Action completed",
				"action_type", action.ActionType,
				"duration", duration,
				"loop_index", loopIndex)
		}
	}

	return logs, outputFiles, nil
}

// resolveVariablesInConfig resolves variables in action configuration
func (r *Runner) resolveVariablesInConfig(config map[string]interface{}, varContext *VariableContext, automationConfig *AutomationConfig) (map[string]interface{}, error) {
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
func (r *Runner) resolveVariablesInString(input string, varContext *VariableContext, automationConfig *AutomationConfig) (string, error) {
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

// sendNotifications sends notifications based on the automation configuration
func (r *Runner) sendNotifications(ctx context.Context, automation *Automation, run *AutomationRun, automationConfig *AutomationConfig, logs []map[string]interface{}, outputFiles []string) {
	if len(automationConfig.Notifications) == 0 {
		return // No notifications configured
	}

	// Get project information (you might need to add this to the Runner or pass it in)
	// For now, we'll use the project ID from the automation
	projectName := "Unknown Project" // TODO: Fetch actual project name if needed

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
