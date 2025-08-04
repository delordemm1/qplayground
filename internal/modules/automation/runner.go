package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/playwright-community/playwright-go"
)

// Runner handles the execution of automation workflows
type Runner struct {
	automationRepo      AutomationRepository
	storageService      storage.StorageService
	notificationService notification.NotificationService
	sseManager          *SSEManager
}

// NewRunner creates a new automation runner
func NewRunner(
	automationRepo AutomationRepository,
	storageService storage.StorageService,
	notificationService notification.NotificationService,
	sseManager *SSEManager,
) *Runner {
	return &Runner{
		automationRepo:      automationRepo,
		storageService:      storageService,
		notificationService: notificationService,
		sseManager:          sseManager,
	}
}

// RunAutomation executes an automation workflow
func (r *Runner) RunAutomation(ctx context.Context, projectID string, run *AutomationRun, isSubRun bool, subRunIndex int, overrides *RunOverrides) error {
	startTime := time.Now()
	
	// Get automation details
	automation, err := r.automationRepo.GetAutomationByID(ctx, run.AutomationID)
	if err != nil {
		return fmt.Errorf("failed to get automation: %w", err)
	}

	// Parse automation configuration
	var automationConfig AutomationConfig
	if automation.ConfigJSON != "" {
		if err := json.Unmarshal([]byte(automation.ConfigJSON), &automationConfig); err != nil {
			return fmt.Errorf("failed to parse automation config: %w", err)
		}
	} else {
		// Use default configuration
		automationConfig = AutomationConfig{
			Variables: []Variable{},
			Multirun: MultiRunConfig{
				Enabled: false,
				Mode:    "sequential",
				Count:   1,
				Delay:   1000,
			},
			Timeout:     300,
			Retries:     0,
			Screenshots: ScreenshotConfig{Enabled: true, OnError: true, OnSuccess: false, Path: "screenshots/{{timestamp}}-{{loopIndex}}.png"},
		}
	}

	// Apply overrides if provided
	if overrides != nil {
		r.applyRunOverrides(&automationConfig, overrides)
	}

	slog.Info("Starting automation execution",
		"automation_id", automation.ID,
		"run_id", run.ID,
		"is_sub_run", isSubRun,
		"sub_run_index", subRunIndex,
		"multirun_enabled", automationConfig.Multirun.Enabled)

	if isSubRun {
		// Execute as individual sub-run
		return r.executeSubRun(ctx, projectID, automation, run, subRunIndex, &automationConfig)
	} else {
		// Execute as complete automation (single run or multi-run orchestration)
		return r.executeCompleteAutomation(ctx, projectID, automation, run, &automationConfig)
	}
}

// applyRunOverrides applies runtime overrides to the automation configuration
func (r *Runner) applyRunOverrides(config *AutomationConfig, overrides *RunOverrides) {
	if overrides.MaxConcurrentRuns != nil {
		// This would be used by the scheduler, not directly in config
		slog.Info("Override: MaxConcurrentRuns", "value", *overrides.MaxConcurrentRuns)
	}
	
	if overrides.RunMode != nil {
		config.Multirun.Mode = *overrides.RunMode
		slog.Info("Override: RunMode", "value", *overrides.RunMode)
	}
	
	if overrides.RunCount != nil {
		config.Multirun.Count = *overrides.RunCount
		config.Multirun.Enabled = *overrides.RunCount > 1
		slog.Info("Override: RunCount", "value", *overrides.RunCount)
	}
	
	if overrides.RunDelay != nil {
		config.Multirun.Delay = *overrides.RunDelay
		slog.Info("Override: RunDelay", "value", *overrides.RunDelay)
	}
}

// executeSubRun executes a single sub-run and uploads individual results
func (r *Runner) executeSubRun(ctx context.Context, projectID string, automation *Automation, run *AutomationRun, subRunIndex int, config *AutomationConfig) error {
	slog.Info("Executing sub-run", "sub_run_index", subRunIndex, "run_id", run.ID)

	// Create event channel for this sub-run
	eventCh := make(chan RunEvent, 100)
	defer close(eventCh)

	// Collect events for this sub-run
	var subRunLogs []RunEvent
	var subRunOutputFiles []string

	// Start event collector goroutine
	go func() {
		for event := range eventCh {
			subRunLogs = append(subRunLogs, event)
			if event.Type == RunEventTypeOutputFile && event.OutputFile != "" {
				subRunOutputFiles = append(subRunOutputFiles, event.OutputFile)
			}
		}
	}()

	// Execute the automation workflow for this specific sub-run
	err := r.executeAutomationWorkflow(ctx, projectID, automation, run, config, subRunIndex, eventCh)

	// Wait for event collector to finish
	time.Sleep(100 * time.Millisecond)

	// Upload individual sub-run results to storage
	logsURL, filesURL, uploadErr := r.uploadSubRunResults(ctx, run.ID, subRunIndex, subRunLogs, subRunOutputFiles)
	if uploadErr != nil {
		slog.Error("Failed to upload sub-run results", "error", uploadErr, "sub_run_index", subRunIndex)
		// Continue with updating progress even if upload fails
		logsURL = ""
		filesURL = ""
	}

	// Update sub-run progress in the main run record
	progressErr := r.automationRepo.UpdateSubRunProgress(ctx, run.ID, subRunIndex, logsURL, filesURL)
	if progressErr != nil {
		slog.Error("Failed to update sub-run progress", "error", progressErr, "sub_run_index", subRunIndex)
	}

	if err != nil {
		slog.Error("Sub-run execution failed", "error", err, "sub_run_index", subRunIndex)
		return fmt.Errorf("sub-run %d failed: %w", subRunIndex, err)
	}

	slog.Info("Sub-run completed successfully", "sub_run_index", subRunIndex, "logs_count", len(subRunLogs), "files_count", len(subRunOutputFiles))
	return nil
}

// executeCompleteAutomation executes the automation as a complete workflow
func (r *Runner) executeCompleteAutomation(ctx context.Context, projectID string, automation *Automation, run *AutomationRun, config *AutomationConfig) error {
	if config.Multirun.Enabled && config.Multirun.Count > 1 {
		// Multi-run execution - orchestrate multiple sub-runs
		return r.executeMultiRun(ctx, projectID, automation, run, config)
	} else {
		// Single run execution
		return r.executeSingleRun(ctx, projectID, automation, run, config)
	}
}

// executeMultiRun orchestrates multiple sub-runs
func (r *Runner) executeMultiRun(ctx context.Context, projectID string, automation *Automation, run *AutomationRun, config *AutomationConfig) error {
	slog.Info("Starting multi-run execution", "run_count", config.Multirun.Count, "mode", config.Multirun.Mode)

	// For internal multi-runs, we would spawn multiple goroutines or processes
	// For external multi-runs, this would just set up the run record and wait for external runners
	
	// Update run record to track expected sub-runs
	run.TotalRunsExpected = &config.Multirun.Count
	runsCompleted := 0
	run.RunsCompleted = &runsCompleted
	run.Status = AutomationRunStatusRunning
	
	err := r.automationRepo.UpdateRun(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to update run for multi-run: %w", err)
	}

	// For now, we'll implement a simple sequential execution
	// In a production system, you might want to use a job queue or worker pool
	for i := 0; i < config.Multirun.Count; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("multi-run cancelled")
		default:
		}

		slog.Info("Starting sub-run", "index", i, "total", config.Multirun.Count)
		
		// Execute individual sub-run
		subRunErr := r.executeSubRun(ctx, projectID, automation, run, i, config)
		if subRunErr != nil {
			slog.Error("Sub-run failed", "index", i, "error", subRunErr)
			// Continue with other sub-runs even if one fails
		}

		// Add delay between runs if specified
		if config.Multirun.Delay > 0 && i < config.Multirun.Count-1 {
			time.Sleep(time.Duration(config.Multirun.Delay) * time.Millisecond)
		}
	}

	// After all sub-runs complete, trigger consolidation
	return r.ConsolidateSubRuns(ctx, run.ID)
}

// executeSingleRun executes a single automation run
func (r *Runner) executeSingleRun(ctx context.Context, projectID string, automation *Automation, run *AutomationRun, config *AutomationConfig) error {
	slog.Info("Executing single run", "run_id", run.ID)

	// Create event channel
	eventCh := make(chan RunEvent, 100)
	defer close(eventCh)

	// Collect events
	var allLogs []RunEvent
	var allOutputFiles []string

	// Start event collector goroutine
	go func() {
		for event := range eventCh {
			allLogs = append(allLogs, event)
			if event.Type == RunEventTypeOutputFile && event.OutputFile != "" {
				allOutputFiles = append(allOutputFiles, event.OutputFile)
			}
		}
	}()

	// Execute the automation workflow
	err := r.executeAutomationWorkflow(ctx, projectID, automation, run, config, 0, eventCh)

	// Wait for event collector to finish
	time.Sleep(100 * time.Millisecond)

	// Generate and upload reports for single run
	detailedReportURL, userJourneyReportURL, reportErr := r.generateAndUploadReports(ctx, run.ID, allLogs, allOutputFiles, false)
	if reportErr != nil {
		slog.Error("Failed to generate reports", "error", reportErr)
	}

	// Update run with final status and report URLs
	endTime := time.Now()
	run.EndTime = &endTime
	
	if err != nil {
		run.Status = AutomationRunStatusFailed
		run.ErrorMessage = err.Error()
	} else {
		run.Status = AutomationRunStatusCompleted
	}
	
	run.DetailedReportURL = detailedReportURL
	run.UserJourneyReportURL = userJourneyReportURL

	// Serialize logs and output files
	logsJSON, _ := json.Marshal(allLogs)
	outputFilesJSON, _ := json.Marshal(allOutputFiles)
	run.LogsJSON = string(logsJSON)
	run.OutputFilesJSON = string(outputFilesJSON)

	updateErr := r.automationRepo.UpdateRun(ctx, run)
	if updateErr != nil {
		slog.Error("Failed to update run", "error", updateErr)
	}

	// Send notifications
	r.sendNotifications(ctx, automation, run, config, allLogs, allOutputFiles)

	return err
}

// uploadSubRunResults uploads individual sub-run logs and files to storage
func (r *Runner) uploadSubRunResults(ctx context.Context, runID string, subRunIndex int, logs []RunEvent, outputFiles []string) (logsURL, filesURL string, err error) {
	// Create unique paths for this sub-run
	subRunPath := fmt.Sprintf("runs/%s/sub_runs/%d", runID, subRunIndex)
	
	// Upload logs as JSON
	logsJSON, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal logs: %w", err)
	}
	
	logsKey := fmt.Sprintf("%s/logs.json", subRunPath)
	logsURL, err = r.storageService.UploadFile(ctx, logsKey, strings.NewReader(string(logsJSON)), "application/json")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload logs: %w", err)
	}

	// Upload output files list as JSON
	outputFilesJSON, err := json.MarshalIndent(outputFiles, "", "  ")
	if err != nil {
		return logsURL, "", fmt.Errorf("failed to marshal output files: %w", err)
	}
	
	filesKey := fmt.Sprintf("%s/output_files.json", subRunPath)
	filesURL, err = r.storageService.UploadFile(ctx, filesKey, strings.NewReader(string(outputFilesJSON)), "application/json")
	if err != nil {
		return logsURL, "", fmt.Errorf("failed to upload output files list: %w", err)
	}

	slog.Info("Uploaded sub-run results", 
		"sub_run_index", subRunIndex,
		"logs_url", logsURL,
		"files_url", filesURL,
		"logs_count", len(logs),
		"files_count", len(outputFiles))

	return logsURL, filesURL, nil
}

// ConsolidateSubRuns consolidates all sub-run results into final reports
func (r *Runner) ConsolidateSubRuns(ctx context.Context, runID string) error {
	slog.Info("Starting consolidation", "run_id", runID)

	// Get the main run record
	run, err := r.automationRepo.GetRunByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	// Get automation details
	automation, err := r.automationRepo.GetAutomationByID(ctx, run.AutomationID)
	if err != nil {
		return fmt.Errorf("failed to get automation: %w", err)
	}

	// Parse automation configuration
	var automationConfig AutomationConfig
	if automation.ConfigJSON != "" {
		if err := json.Unmarshal([]byte(automation.ConfigJSON), &automationConfig); err != nil {
			return fmt.Errorf("failed to parse automation config: %w", err)
		}
	}

	// Update status to consolidating
	run.Status = AutomationRunStatusConsolidating
	err = r.automationRepo.UpdateRun(ctx, run)
	if err != nil {
		slog.Error("Failed to update run status to consolidating", "error", err)
	}

	// Parse sub-run outputs
	var subRunOutputs map[string]SubRunOutput
	if run.SubRunOutputsJSON != "" {
		if err := json.Unmarshal([]byte(run.SubRunOutputsJSON), &subRunOutputs); err != nil {
			return fmt.Errorf("failed to parse sub-run outputs: %w", err)
		}
	}

	if len(subRunOutputs) == 0 {
		return fmt.Errorf("no sub-run outputs found for consolidation")
	}

	slog.Info("Consolidating sub-runs", "sub_run_count", len(subRunOutputs))

	// Download and aggregate all sub-run data
	allLogs, allOutputFiles, err := r.downloadAndAggregateSubRuns(ctx, subRunOutputs)
	if err != nil {
		return fmt.Errorf("failed to aggregate sub-runs: %w", err)
	}

	// Generate consolidated reports
	detailedReportURL, userJourneyReportURL, err := r.generateAndUploadReports(ctx, runID, allLogs, allOutputFiles, true)
	if err != nil {
		slog.Error("Failed to generate consolidated reports", "error", err)
	}

	// Determine final status based on aggregated results
	finalStatus := r.determineFinalStatus(allLogs)

	// Update run with final consolidated results
	endTime := time.Now()
	run.EndTime = &endTime
	run.Status = finalStatus
	run.DetailedReportURL = detailedReportURL
	run.UserJourneyReportURL = userJourneyReportURL

	// Serialize aggregated data
	logsJSON, _ := json.Marshal(allLogs)
	outputFilesJSON, _ := json.Marshal(allOutputFiles)
	run.LogsJSON = string(logsJSON)
	run.OutputFilesJSON = string(outputFilesJSON)

	err = r.automationRepo.UpdateRun(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to update consolidated run: %w", err)
	}

	// Send notifications for consolidated results
	r.sendNotifications(ctx, automation, run, &automationConfig, allLogs, allOutputFiles)

	slog.Info("Consolidation completed successfully", 
		"run_id", runID,
		"final_status", finalStatus,
		"total_logs", len(allLogs),
		"total_files", len(allOutputFiles))

	return nil
}

// downloadAndAggregateSubRuns downloads and aggregates all sub-run data
func (r *Runner) downloadAndAggregateSubRuns(ctx context.Context, subRunOutputs map[string]SubRunOutput) ([]RunEvent, []string, error) {
	var allLogs []RunEvent
	var allOutputFiles []string

	// Sort sub-run indices for consistent processing
	var indices []int
	for indexStr := range subRunOutputs {
		if index, err := strconv.Atoi(indexStr); err == nil {
			indices = append(indices, index)
		}
	}
	sort.Ints(indices)

	for _, index := range indices {
		indexStr := strconv.Itoa(index)
		subRunOutput := subRunOutputs[indexStr]

		slog.Info("Processing sub-run", "index", index, "logs_url", subRunOutput.LogsURL, "files_url", subRunOutput.FilesURL)

		// Download and parse logs
		if subRunOutput.LogsURL != "" {
			logs, err := r.downloadAndParseJSON[[]RunEvent](ctx, subRunOutput.LogsURL)
			if err != nil {
				slog.Error("Failed to download sub-run logs", "index", index, "error", err)
				continue
			}
			allLogs = append(allLogs, logs...)
		}

		// Download and parse output files list
		if subRunOutput.FilesURL != "" {
			files, err := r.downloadAndParseJSON[[]string](ctx, subRunOutput.FilesURL)
			if err != nil {
				slog.Error("Failed to download sub-run files list", "index", index, "error", err)
				continue
			}
			allOutputFiles = append(allOutputFiles, files...)
		}
	}

	slog.Info("Aggregated sub-run data", "total_logs", len(allLogs), "total_files", len(allOutputFiles))
	return allLogs, allOutputFiles, nil
}

// downloadAndParseJSON downloads and parses JSON data from a URL
func (r *Runner) downloadAndParseJSON[T any](ctx context.Context, url string) (T, error) {
	var result T
	
	// For now, we'll assume the URL is accessible via HTTP
	// In a production system, you might want to use the storage service's download method
	// This is a simplified implementation
	
	// Since we're using object storage URLs, we can't directly download them here
	// This would need to be implemented based on your storage service interface
	// For now, return empty result
	slog.Warn("downloadAndParseJSON not fully implemented", "url", url)
	return result, nil
}

// determineFinalStatus determines the final status based on aggregated logs
func (r *Runner) determineFinalStatus(logs []RunEvent) string {
	hasErrors := false
	hasSuccess := false

	for _, log := range logs {
		if log.Type == RunEventTypeError {
			hasErrors = true
		} else if log.Type == RunEventTypeLog {
			hasSuccess = true
		}
	}

	if hasErrors {
		if hasSuccess {
			return AutomationRunStatusPartialCompleted
		}
		return AutomationRunStatusFailed
	}

	return AutomationRunStatusCompleted
}

// generateAndUploadReports generates and uploads consolidated reports
func (r *Runner) generateAndUploadReports(ctx context.Context, runID string, logs []RunEvent, outputFiles []string, isConsolidated bool) (detailedReportURL, userJourneyReportURL string, err error) {
	slog.Info("Generating reports", "run_id", runID, "is_consolidated", isConsolidated, "logs_count", len(logs), "files_count", len(outputFiles))

	// Generate detailed HTML report
	detailedHTML, err := r.generateDetailedHTMLReport(logs, outputFiles, isConsolidated)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate detailed report: %w", err)
	}

	// Generate user journey HTML report
	userJourneyHTML, err := r.generateUserJourneyHTMLReport(logs, outputFiles, isConsolidated)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate user journey report: %w", err)
	}

	// Upload reports to storage
	reportPath := fmt.Sprintf("runs/%s/reports", runID)
	if isConsolidated {
		reportPath = fmt.Sprintf("runs/%s/consolidated_reports", runID)
	}

	// Upload detailed report
	detailedKey := fmt.Sprintf("%s/detailed_report.html", reportPath)
	detailedReportURL, err = r.storageService.UploadFile(ctx, detailedKey, strings.NewReader(detailedHTML), "text/html")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload detailed report: %w", err)
	}

	// Upload user journey report
	userJourneyKey := fmt.Sprintf("%s/user_journey_report.html", reportPath)
	userJourneyReportURL, err = r.storageService.UploadFile(ctx, userJourneyKey, strings.NewReader(userJourneyHTML), "text/html")
	if err != nil {
		return detailedReportURL, "", fmt.Errorf("failed to upload user journey report: %w", err)
	}

	slog.Info("Reports generated and uploaded successfully",
		"detailed_url", detailedReportURL,
		"user_journey_url", userJourneyReportURL)

	return detailedReportURL, userJourneyReportURL, nil
}

// generateDetailedHTMLReport generates a detailed HTML report
func (r *Runner) generateDetailedHTMLReport(logs []RunEvent, outputFiles []string, isConsolidated bool) (string, error) {
	// This is a simplified implementation
	// In a production system, you would use a proper HTML template
	
	reportType := "Single Run"
	if isConsolidated {
		reportType = "Consolidated Multi-Run"
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s Detailed Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .log-entry { margin: 10px 0; padding: 10px; border-left: 3px solid #ccc; }
        .error { border-left-color: #ff0000; background: #ffe6e6; }
        .success { border-left-color: #00ff00; background: #e6ffe6; }
        .files { margin-top: 20px; }
        .file-link { display: block; margin: 5px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s Detailed Report</h1>
        <p>Generated: %s</p>
        <p>Total Log Entries: %d</p>
        <p>Total Output Files: %d</p>
    </div>
    
    <h2>Execution Logs</h2>
    <div class="logs">
`, reportType, reportType, time.Now().Format(time.RFC3339), len(logs), len(outputFiles))

	// Add log entries
	for _, log := range logs {
		cssClass := "log-entry"
		if log.Type == RunEventTypeError {
			cssClass += " error"
		} else if log.Type == RunEventTypeLog {
			cssClass += " success"
		}

		html += fmt.Sprintf(`
        <div class="%s">
            <strong>%s</strong> [%s] %s<br>
            <small>Step: %s | Action: %s | Duration: %dms</small>
        </div>
`, cssClass, log.Timestamp.Format(time.RFC3339), log.Type, log.Message, log.StepName, log.ActionType, log.Duration)
	}

	html += `
    </div>
    
    <h2>Output Files</h2>
    <div class="files">
`

	// Add output files
	for _, file := range outputFiles {
		html += fmt.Sprintf(`<a href="%s" class="file-link" target="_blank">%s</a>`, file, file)
	}

	html += `
    </div>
</body>
</html>
`

	return html, nil
}

// generateUserJourneyHTMLReport generates a user journey HTML report
func (r *Runner) generateUserJourneyHTMLReport(logs []RunEvent, outputFiles []string, isConsolidated bool) (string, error) {
	// This is a simplified implementation
	// In a production system, you would use a proper HTML template with charts and visualizations
	
	reportType := "Single Run"
	if isConsolidated {
		reportType = "Consolidated Multi-Run"
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s User Journey Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .journey-step { margin: 15px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .step-success { background: #e6ffe6; border-color: #00ff00; }
        .step-error { background: #ffe6e6; border-color: #ff0000; }
        .timeline { margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s User Journey Report</h1>
        <p>Generated: %s</p>
    </div>
    
    <h2>User Journey Timeline</h2>
    <div class="timeline">
`, reportType, reportType, time.Now().Format(time.RFC3339))

	// Group logs by step for journey visualization
	stepMap := make(map[string][]RunEvent)
	for _, log := range logs {
		if log.StepName != "" {
			stepMap[log.StepName] = append(stepMap[log.StepName], log)
		}
	}

	// Add journey steps
	for stepName, stepLogs := range stepMap {
		hasError := false
		totalDuration := int64(0)
		
		for _, log := range stepLogs {
			if log.Type == RunEventTypeError {
				hasError = true
			}
			totalDuration += log.Duration
		}

		cssClass := "journey-step step-success"
		if hasError {
			cssClass = "journey-step step-error"
		}

		html += fmt.Sprintf(`
        <div class="%s">
            <h3>%s</h3>
            <p>Duration: %dms | Actions: %d | Status: %s</p>
        </div>
`, cssClass, stepName, totalDuration, len(stepLogs), func() string {
			if hasError {
				return "Failed"
			}
			return "Success"
		}())
	}

	html += `
    </div>
</body>
</html>
`

	return html, nil
}

// executeAutomationWorkflow executes the core automation workflow
func (r *Runner) executeAutomationWorkflow(ctx context.Context, projectID string, automation *Automation, run *AutomationRun, config *AutomationConfig, loopIndex int, eventCh chan RunEvent) error {
	// Get steps for this automation
	steps, err := r.automationRepo.GetStepsByAutomationID(ctx, automation.ID)
	if err != nil {
		return fmt.Errorf("failed to get automation steps: %w", err)
	}

	if len(steps) == 0 {
		return fmt.Errorf("no steps defined for automation")
	}

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		return fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	// Create variable context
	variableContext := &VariableContext{
		LoopIndex:      loopIndex,
		LocalLoopIndex: 0,
		Timestamp:      time.Now().Format("20060102-150405"),
		RunID:          run.ID,
		ProjectID:      projectID,
		AutomationID:   automation.ID,
		StaticVars:     make(map[string]string),
		RuntimeVars:    make(map[string]interface{}),
		GlobalVars:     make(map[string]interface{}),
	}

	// Initialize static variables
	for _, variable := range config.Variables {
		if variable.Type == "static" {
			variableContext.StaticVars[variable.Key] = variable.Value
		}
	}

	// Create run context
	runContext := &RunContext{
		PlaywrightBrowserContext: context,
		PlaywrightPage:           page,
		StorageService:           r.storageService,
		Logger:                   slog.Default(),
		EventCh:                  eventCh,
		Runner:                   r,
		VariableContext:          variableContext,
		AutomationConfig:         config,
		ScreenshotsR2Path:        fmt.Sprintf("runs/%s/screenshots", run.ID),
		ReportsR2Path:            fmt.Sprintf("runs/%s/reports", run.ID),
	}

	// Execute steps in order
	for _, step := range steps {
		select {
		case <-ctx.Done():
			return fmt.Errorf("automation cancelled")
		default:
		}

		// Check step skip conditions
		if r.shouldSkipStep(step, loopIndex) {
			slog.Info("Skipping step due to condition", "step_name", step.Name, "loop_index", loopIndex)
			continue
		}

		runContext.StepName = step.Name
		runContext.StepID = step.ID

		// Send step start event
		if eventCh != nil {
			select {
			case eventCh <- RunEvent{
				Type:      RunEventTypeStep,
				Timestamp: time.Now(),
				StepID:    step.ID,
				StepName:  step.Name,
				Message:   fmt.Sprintf("Starting step: %s", step.Name),
				LoopIndex: loopIndex,
			}:
			default:
			}
		}

		// Execute step actions
		err := r.executeStepActions(ctx, step, runContext)
		if err != nil {
			// Send error event
			if eventCh != nil {
				select {
				case eventCh <- RunEvent{
					Type:      RunEventTypeError,
					Timestamp: time.Now(),
					StepID:    step.ID,
					StepName:  step.Name,
					Error:     err.Error(),
					LoopIndex: loopIndex,
				}:
				default:
				}
			}
			return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}

		// Send step completion event
		if eventCh != nil {
			select {
			case eventCh <- RunEvent{
				Type:      RunEventTypeStep,
				Timestamp: time.Now(),
				StepID:    step.ID,
				StepName:  step.Name,
				Message:   fmt.Sprintf("Completed step: %s", step.Name),
				LoopIndex: loopIndex,
			}:
			default:
			}
		}
	}

	return nil
}

// shouldSkipStep determines if a step should be skipped based on its configuration
func (r *Runner) shouldSkipStep(step *AutomationStep, loopIndex int) bool {
	if step.ConfigJSON == "" {
		return false
	}

	var stepConfig StepConfig
	if err := json.Unmarshal([]byte(step.ConfigJSON), &stepConfig); err != nil {
		slog.Warn("Failed to parse step config", "step_id", step.ID, "error", err)
		return false
	}

	// Check skip condition
	if stepConfig.SkipCondition != "" {
		return r.evaluateStepCondition(stepConfig.SkipCondition, loopIndex, stepConfig.Probability)
	}

	// Check run-only condition (inverse logic)
	if stepConfig.RunOnlyCondition != "" {
		return !r.evaluateStepCondition(stepConfig.RunOnlyCondition, loopIndex, stepConfig.Probability)
	}

	return false
}

// evaluateStepCondition evaluates a step condition
func (r *Runner) evaluateStepCondition(condition string, loopIndex int, probability float64) bool {
	switch condition {
	case "loop_index_is_even":
		return loopIndex%2 == 0
	case "loop_index_is_odd":
		return loopIndex%2 == 1
	case "loop_index_is_prime":
		return r.isPrime(loopIndex)
	case "random":
		if probability <= 0 {
			probability = 0.5 // Default 50% chance
		}
		// Simple random implementation
		return time.Now().UnixNano()%100 < int64(probability*100)
	default:
		return false
	}
}

// isPrime checks if a number is prime
func (r *Runner) isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// executeStepActions executes all actions in a step
func (r *Runner) executeStepActions(ctx context.Context, step *AutomationStep, runContext *RunContext) error {
	// Get actions for this step
	actions, err := r.automationRepo.GetActionsByStepID(ctx, step.ID)
	if err != nil {
		return fmt.Errorf("failed to get step actions: %w", err)
	}

	// Execute actions in order
	for _, action := range actions {
		select {
		case <-ctx.Done():
			return fmt.Errorf("step cancelled")
		default:
		}

		runContext.ActionID = action.ID
		runContext.ActionName = action.Name

		// Parse action config
		var actionConfig map[string]interface{}
		if action.ActionConfigJSON != "" {
			if err := json.Unmarshal([]byte(action.ActionConfigJSON), &actionConfig); err != nil {
				return fmt.Errorf("failed to parse action config for action %s: %w", action.ID, err)
			}
		}

		// Resolve variables in action config
		resolvedActionConfig, err := r.ResolveVariablesInConfig(actionConfig, runContext.VariableContext, runContext.AutomationConfig)
		if err != nil {
			return fmt.Errorf("failed to resolve variables in action config: %w", err)
		}

		// Get plugin action
		pluginAction, err := GetAction(action.ActionType)
		if err != nil {
			return fmt.Errorf("failed to get action plugin for type %s: %w", action.ActionType, err)
		}

		// Execute action
		actionStartTime := time.Now()
		err = pluginAction.Execute(ctx, resolvedActionConfig, runContext)
		actionDuration := time.Since(actionStartTime)

		if err != nil {
			// Send error event
			if runContext.EventCh != nil {
				select {
				case runContext.EventCh <- RunEvent{
					Type:       RunEventTypeError,
					Timestamp:  time.Now(),
					StepID:     step.ID,
					StepName:   step.Name,
					ActionID:   action.ID,
					ActionName: action.Name,
					ActionType: action.ActionType,
					Error:      err.Error(),
					Duration:   actionDuration.Milliseconds(),
					LoopIndex:  runContext.LoopIndex,
				}:
				default:
				}
			}
			return fmt.Errorf("action '%s' (%s) failed: %w", action.Name, action.ActionType, err)
		}

		// Send success event
		if runContext.EventCh != nil {
			select {
			case runContext.EventCh <- RunEvent{
				Type:       RunEventTypeLog,
				Timestamp:  time.Now(),
				StepID:     step.ID,
				StepName:   step.Name,
				ActionID:   action.ID,
				ActionName: action.Name,
				ActionType: action.ActionType,
				Message:    fmt.Sprintf("Action completed successfully"),
				Duration:   actionDuration.Milliseconds(),
				LoopIndex:  runContext.LoopIndex,
			}:
			default:
			}
		}
	}

	return nil
}

// sendNotifications sends notifications based on automation configuration
func (r *Runner) sendNotifications(ctx context.Context, automation *Automation, run *AutomationRun, config *AutomationConfig, logs []RunEvent, outputFiles []string) {
	if len(config.Notifications) == 0 {
		return
	}

	// Create notification message
	message := notification.NotificationMessage{
		AutomationID:   automation.ID,
		AutomationName: automation.Name,
		ProjectID:      automation.ProjectID,
		ProjectName:    automation.ProjectName,
		RunID:          run.ID,
		Status:         run.Status,
		StartTime:      run.StartTime,
		EndTime:        run.EndTime,
		ErrorMessage:   run.ErrorMessage,
		OutputFiles:    outputFiles,
		LogsCount:      len(logs),
	}

	// Send notifications asynchronously
	go func() {
		err := r.notificationService.DispatchAutomationNotification(context.Background(), message, config.Notifications)
		if err != nil {
			slog.Error("Failed to send notifications", "error", err, "run_id", run.ID)
		}
	}()
}

// ResolveVariablesInConfig resolves variables in action configuration
func (r *Runner) ResolveVariablesInConfig(config map[string]interface{}, varContext *VariableContext, automationConfig *AutomationConfig) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for key, value := range config {
		switch v := value.(type) {
		case string:
			resolved, err := r.ResolveVariablesInString(v, varContext, automationConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variables in key %s: %w", key, err)
			}
			result[key] = resolved
		case map[string]interface{}:
			resolved, err := r.ResolveVariablesInConfig(v, varContext, automationConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variables in nested config %s: %w", key, err)
			}
			result[key] = resolved
		case []interface{}:
			resolvedArray := make([]interface{}, len(v))
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					resolved, err := r.ResolveVariablesInConfig(itemMap, varContext, automationConfig)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve variables in array item %d of key %s: %w", i, key, err)
					}
					resolvedArray[i] = resolved
				} else if itemStr, ok := item.(string); ok {
					resolved, err := r.ResolveVariablesInString(itemStr, varContext, automationConfig)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve variables in array string %d of key %s: %w", i, key, err)
					}
					resolvedArray[i] = resolved
				} else {
					resolvedArray[i] = item
				}
			}
			result[key] = resolvedArray
		default:
			result[key] = value
		}
	}
	
	return result, nil
}

// ResolveVariablesInString resolves variables in a string template
func (r *Runner) ResolveVariablesInString(template string, varContext *VariableContext, automationConfig *AutomationConfig) (string, error) {
	result := template
	
	// Replace static variables
	for key, value := range varContext.StaticVars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Replace runtime variables
	for key, value := range varContext.RuntimeVars {
		placeholder := fmt.Sprintf("{{runtime.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	
	// Replace global variables
	for key, value := range varContext.GlobalVars {
		placeholder := fmt.Sprintf("{{runtime.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	
	// Replace environment variables
	envVars := map[string]string{
		"loopIndex":     strconv.Itoa(varContext.LoopIndex),
		"timestamp":     varContext.Timestamp,
		"runId":         varContext.RunID,
		"projectId":     varContext.ProjectID,
		"automationId":  varContext.AutomationID,
	}
	
	for key, value := range envVars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Replace faker variables (simplified implementation)
	fakerVars := map[string]func() string{
		"faker.name":     func() string { return "John Doe" },
		"faker.email":    func() string { return "test@example.com" },
		"faker.phone":    func() string { return "+1234567890" },
		"faker.uuid":     func() string { return platform.UtilGenerateUUID() },
		"faker.username": func() string { return "testuser" },
		"faker.password": func() string { return "password123" },
	}
	
	for placeholder, generator := range fakerVars {
		fullPlaceholder := fmt.Sprintf("{{%s}}", placeholder)
		if strings.Contains(result, fullPlaceholder) {
			result = strings.ReplaceAll(result, fullPlaceholder, generator())
		}
	}
	
	return result, nil
}