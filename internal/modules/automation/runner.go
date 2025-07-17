package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/playwright-community/playwright-go"
)

// Runner orchestrates the execution of automations.
type Runner struct {
	automationRepo AutomationRepository
	storageService storage.StorageService
}

// NewRunner creates a new Runner instance.
func NewRunner(automationRepo AutomationRepository, storageService storage.StorageService) *Runner {
	return &Runner{
		automationRepo: automationRepo,
		storageService: storageService,
	}
}

// RunAutomation executes a given automation.
func (r *Runner) RunAutomation(ctx context.Context, automationID string) error {
	// 1. Fetch Automation details from DB
	automation, err := r.automationRepo.GetAutomationByID(ctx, automationID)
	if err != nil {
		return fmt.Errorf("failed to get automation: %w", err)
	}
	_ = automation
	// 2. Create a new AutomationRun record (status: running)
	run := &AutomationRun{
		ID:              platform.UtilGenerateUUID(),
		AutomationID:    automationID,
		Status:          "running",
		LogsJSON:        "[]",
		OutputFilesJSON: "[]",
	}

	// Set start time
	now := time.Now()
	run.StartTime = &now

	err = r.automationRepo.CreateRun(ctx, run)
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

	// 3. Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		err = fmt.Errorf("could not start playwright: %w", err)
		return err
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
		err = fmt.Errorf("could not launch browser: %w", err)
		return err
	}
	defer browser.Close()

	// Create new page
	page, err := browser.NewPage()
	if err != nil {
		err = fmt.Errorf("could not create page: %w", err)
		return err
	}

	// 4. Create RunContext
	runContext := &RunContext{
		PlaywrightBrowser: browser,
		PlaywrightPage:    page,
		StorageService:    r.storageService,
		Logger:            slog.Default().With("automation_id", automationID, "run_id", run.ID),
	}

	// 5. Fetch and execute steps
	steps, err := r.automationRepo.GetStepsByAutomationID(ctx, automationID)
	if err != nil {
		err = fmt.Errorf("failed to get automation steps: %w", err)
		return err
	}

	var logs []map[string]interface{}
	var outputFiles []string

	for _, step := range steps {
		runContext.Logger.Info("Executing step", "step_name", step.Name, "step_order", step.StepOrder)

		// Get actions for this step
		stepActions, err := r.automationRepo.GetActionsByStepID(ctx, step.ID)
		if err != nil {
			err = fmt.Errorf("failed to get actions for step %s: %w", step.Name, err)
			return err
		}

		for _, action := range stepActions {
			startTime := time.Now()

			// Parse action config
			actionConfigMap := make(map[string]interface{})
			if action.ActionConfigJSON != "" {
				if jsonErr := json.Unmarshal([]byte(action.ActionConfigJSON), &actionConfigMap); jsonErr != nil {
					err = fmt.Errorf("failed to parse action config JSON for action %s: %w", action.ActionType, jsonErr)
					return err
				}
			}

			// Get plugin action
			pluginAction, getActionErr := GetAction(action.ActionType)
			if getActionErr != nil {
				err = fmt.Errorf("unregistered plugin action type '%s': %w", action.ActionType, getActionErr)
				return err
			}

			// Execute action
			actionErr := pluginAction.Execute(ctx, actionConfigMap, runContext)
			duration := time.Since(startTime)

			// Create log entry
			logEntry := map[string]interface{}{
				"timestamp":   time.Now().Format(time.RFC3339),
				"step_id":     step.ID,
				"step_name":   step.Name,
				"action_id":   action.ID,
				"action_type": action.ActionType,
				"duration_ms": duration.Milliseconds(),
				"status":      "success",
			}

			if actionErr != nil {
				logEntry["status"] = "failed"
				logEntry["error"] = actionErr.Error()
				logs = append(logs, logEntry)

				// Update logs in DB immediately on failure
				logsBytes, _ := json.Marshal(logs)
				run.LogsJSON = string(logsBytes)
				r.automationRepo.UpdateRun(ctx, run)

				runContext.Logger.Error("Action failed",
					"action_type", action.ActionType,
					"error", actionErr,
					"duration", duration)

				err = fmt.Errorf("action '%s' failed: %w", action.ActionType, actionErr)
				return err // Stop execution on first action failure
			}

			// Check if this was a screenshot action that uploaded to R2
			if action.ActionType == "playwright:screenshot" {
				if uploadToR2, ok := actionConfigMap["upload_to_r2"].(bool); ok && uploadToR2 {
					if r2Key, ok := actionConfigMap["r2_key"].(string); ok && r2Key != "" {
						// Get the public URL for the uploaded screenshot
						publicURL := r.storageService.GetPublicURL(r2Key)
						outputFiles = append(outputFiles, publicURL)
						logEntry["output_file"] = publicURL
					}
				}
			}

			logs = append(logs, logEntry)
			runContext.Logger.Info("Action completed",
				"action_type", action.ActionType,
				"duration", duration)
		}
	}

	// 6. Update run record with final logs and output files
	logsBytes, _ := json.Marshal(logs)
	run.LogsJSON = string(logsBytes)

	outputFilesBytes, _ := json.Marshal(outputFiles)
	run.OutputFilesJSON = string(outputFilesBytes)

	runContext.Logger.Info("Automation completed successfully",
		"total_steps", len(steps),
		"total_output_files", len(outputFiles))

	return nil
}
