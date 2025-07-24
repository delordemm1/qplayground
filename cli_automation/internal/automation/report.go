package automation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log/slog"
)

// GenerateReports generates HTML, JSON, and CSV reports for an automation run
func GenerateReports(automation *Automation, run *AutomationRun, config *AutomationConfig, outputBaseDir string) error {
	// Create unique directory for this run
	runTimestamp := time.Now().Format("20060102-150405")
	runDir := filepath.Join(outputBaseDir, fmt.Sprintf("%s-%s", runTimestamp, run.ID[:8]))
	reportsDir := filepath.Join(runDir, "reports")

	// Create directories
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Parse logs and output files
	var logs []map[string]any
	if run.LogsJSON != "" {
		if err := json.Unmarshal([]byte(run.LogsJSON), &logs); err != nil {
			slog.Warn("Failed to parse logs JSON", "error", err)
			logs = []map[string]any{}
		}
	}

	var outputFiles []string
	if run.OutputFilesJSON != "" {
		if err := json.Unmarshal([]byte(run.OutputFilesJSON), &outputFiles); err != nil {
			slog.Warn("Failed to parse output files JSON", "error", err)
			outputFiles = []string{}
		}
	}

	// Generate HTML report
	if err := generateHTMLReport(automation, run, config, logs, outputFiles, reportsDir); err != nil {
		slog.Error("Failed to generate HTML report", "error", err)
	}

	// Generate JSON report
	if err := generateJSONReport(automation, run, config, logs, outputFiles, reportsDir); err != nil {
		slog.Error("Failed to generate JSON report", "error", err)
	}

	// Generate CSV report
	if err := generateCSVReport(automation, run, logs, reportsDir); err != nil {
		slog.Error("Failed to generate CSV report", "error", err)
	}

	slog.Info("Reports generated successfully", "run_id", run.ID, "output_dir", runDir)
	return nil
}

// generateHTMLReport creates a comprehensive HTML report
func generateHTMLReport(automation *Automation, run *AutomationRun, config *AutomationConfig, logs []map[string]any, outputFiles []string, reportsDir string) error {
	// Organize data by steps
	stepMap := make(map[string]*StepReport)
	
	for _, log := range logs {
		stepID, _ := log["step_id"].(string)
		if stepID == "" {
			continue
		}

		if _, exists := stepMap[stepID]; !exists {
			stepMap[stepID] = &StepReport{
				ID:              stepID,
				Name:            getString(log, "step_name"),
				Actions:         make(map[string]*ActionReport),
				ConcurrentUsers: make(map[int]bool),
				Status:          "success",
				StartTime:       getString(log, "timestamp"),
				EndTime:         getString(log, "timestamp"),
			}
		}

		step := stepMap[stepID]
		step.EndTime = getString(log, "timestamp")
		
		loopIndex := getInt(log, "loop_index")
		step.ConcurrentUsers[loopIndex] = true

		if getString(log, "status") == "failed" {
			step.Status = "failed"
		}

		// Process action
		actionID := getString(log, "action_id")
		if actionID != "" {
			actionKey := fmt.Sprintf("%s-%d", actionID, loopIndex)
			if _, exists := step.Actions[actionKey]; !exists {
				step.Actions[actionKey] = &ActionReport{
					ID:             actionID,
					Type:           getString(log, "action_type"),
					ParentActionID: getString(log, "parent_action_id"),
					LoopIndex:      loopIndex,
					Status:         getString(log, "status"),
					Duration:       getInt64(log, "duration_ms"),
					OutputFiles:    []string{},
				}
			}

			action := step.Actions[actionKey]
			if outputFile := getString(log, "output_file"); outputFile != "" {
				action.OutputFiles = append(action.OutputFiles, outputFile)
			}
			if getString(log, "status") == "failed" {
				action.Status = "failed"
				action.Error = getString(log, "error")
			}
		}
	}

	// Calculate performance metrics
	metrics := calculatePerformanceMetrics(logs)

	// Generate HTML content
	htmlContent := generateHTMLContent(automation, run, config, stepMap, metrics, outputFiles)

	// Write HTML file
	htmlPath := filepath.Join(reportsDir, "report.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	return nil
}

// generateJSONReport creates a detailed JSON report
func generateJSONReport(automation *Automation, run *AutomationRun, config *AutomationConfig, logs []map[string]any, outputFiles []string, reportsDir string) error {
	reportData := map[string]interface{}{
		"run": map[string]interface{}{
			"id":           run.ID,
			"status":       run.Status,
			"startTime":    run.StartTime,
			"endTime":      run.EndTime,
			"errorMessage": run.ErrorMessage,
		},
		"automation": map[string]interface{}{
			"id":          automation.ID,
			"name":        automation.Name,
			"description": automation.Description,
		},
		"config": config,
		"logs":   logs,
		"outputFiles": outputFiles,
		"metrics": calculatePerformanceMetrics(logs),
		"generatedAt": time.Now().Format(time.RFC3339),
	}

	jsonBytes, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	jsonPath := filepath.Join(reportsDir, "report.json")
	if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	return nil
}

// generateCSVReport creates a CSV export of the logs
func generateCSVReport(automation *Automation, run *AutomationRun, logs []map[string]any, reportsDir string) error {
	csvPath := filepath.Join(reportsDir, "logs.csv")
	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := []string{
		"Timestamp", "Step ID", "Step Name", "Action ID", "Parent Action ID",
		"Action Type", "Status", "Duration (ms)", "Loop Index", "Local Loop Index",
		"Message", "Error", "Output File",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, log := range logs {
		row := []string{
			getString(log, "timestamp"),
			getString(log, "step_id"),
			getString(log, "step_name"),
			getString(log, "action_id"),
			getString(log, "parent_action_id"),
			getString(log, "action_type"),
			getString(log, "status"),
			fmt.Sprintf("%d", getInt64(log, "duration_ms")),
			fmt.Sprintf("%d", getInt(log, "loop_index")),
			fmt.Sprintf("%d", getInt(log, "local_loop_index")),
			getString(log, "message"),
			getString(log, "error"),
			getString(log, "output_file"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// Helper types for report generation
type StepReport struct {
	ID              string
	Name            string
	Actions         map[string]*ActionReport
	ConcurrentUsers map[int]bool
	Status          string
	StartTime       string
	EndTime         string
}

type ActionReport struct {
	ID             string
	Type           string
	ParentActionID string
	LoopIndex      int
	Status         string
	Duration       int64
	Error          string
	OutputFiles    []string
}

type PerformanceMetrics struct {
	TotalRuns        int                    `json:"totalRuns"`
	OverallFailureRate float64             `json:"overallFailureRate"`
	StepAverages     []StepMetric          `json:"stepAverages"`
	RunData          []RunMetric           `json:"runData"`
}

type StepMetric struct {
	Name            string  `json:"name"`
	AverageDuration float64 `json:"averageDuration"`
	FailureRate     float64 `json:"failureRate"`
	TotalRuns       int     `json:"totalRuns"`
}

type RunMetric struct {
	LoopIndex     int                    `json:"loopIndex"`
	Steps         map[string]StepRunData `json:"steps"`
	TotalDuration int64                  `json:"totalDuration"`
	Status        string                 `json:"status"`
}

type StepRunData struct {
	Duration int64  `json:"duration"`
	Status   string `json:"status"`
}

// calculatePerformanceMetrics analyzes logs to generate performance insights
func calculatePerformanceMetrics(logs []map[string]any) PerformanceMetrics {
	stepMetrics := make(map[string]*StepMetric)
	runMetrics := make(map[int]*RunMetric)

	for _, log := range logs {
		stepName := getString(log, "step_name")
		loopIndex := getInt(log, "loop_index")
		duration := getInt64(log, "duration_ms")
		status := getString(log, "status")

		if stepName == "" {
			continue
		}

		// Step-level metrics
		if _, exists := stepMetrics[stepName]; !exists {
			stepMetrics[stepName] = &StepMetric{
				Name:      stepName,
				TotalRuns: 0,
			}
		}

		stepMetric := stepMetrics[stepName]
		stepMetric.TotalRuns++
		stepMetric.AverageDuration = (stepMetric.AverageDuration*float64(stepMetric.TotalRuns-1) + float64(duration)) / float64(stepMetric.TotalRuns)
		
		if status == "failed" {
			stepMetric.FailureRate = (stepMetric.FailureRate*float64(stepMetric.TotalRuns-1) + 100) / float64(stepMetric.TotalRuns)
		} else {
			stepMetric.FailureRate = (stepMetric.FailureRate * float64(stepMetric.TotalRuns-1)) / float64(stepMetric.TotalRuns)
		}

		// Run-level metrics
		if _, exists := runMetrics[loopIndex]; !exists {
			runMetrics[loopIndex] = &RunMetric{
				LoopIndex: loopIndex,
				Steps:     make(map[string]StepRunData),
				Status:    "success",
			}
		}

		runMetric := runMetrics[loopIndex]
		if _, exists := runMetric.Steps[stepName]; !exists {
			runMetric.Steps[stepName] = StepRunData{Status: "success"}
		}

		stepData := runMetric.Steps[stepName]
		stepData.Duration += duration
		if status == "failed" {
			stepData.Status = "failed"
			runMetric.Status = "failed"
		}
		runMetric.Steps[stepName] = stepData
		runMetric.TotalDuration += duration
	}

	// Convert maps to slices
	var stepAverages []StepMetric
	for _, metric := range stepMetrics {
		stepAverages = append(stepAverages, *metric)
	}

	var runData []RunMetric
	for _, metric := range runMetrics {
		runData = append(runData, *metric)
	}

	// Calculate overall failure rate
	failedRuns := 0
	for _, run := range runData {
		if run.Status == "failed" {
			failedRuns++
		}
	}

	overallFailureRate := 0.0
	if len(runData) > 0 {
		overallFailureRate = float64(failedRuns) / float64(len(runData)) * 100
	}

	return PerformanceMetrics{
		TotalRuns:          len(runData),
		OverallFailureRate: overallFailureRate,
		StepAverages:       stepAverages,
		RunData:            runData,
	}
}

// generateHTMLContent creates the HTML report content
func generateHTMLContent(automation *Automation, run *AutomationRun, config *AutomationConfig, stepMap map[string]*StepReport, metrics PerformanceMetrics, outputFiles []string) string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Automation Report - ` + automation.Name + `</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { border-bottom: 2px solid #e5e7eb; padding-bottom: 20px; margin-bottom: 30px; }
        .title { font-size: 2rem; font-weight: bold; color: #1f2937; margin: 0; }
        .subtitle { color: #6b7280; margin: 10px 0 0 0; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric { background: #f9fafb; padding: 20px; border-radius: 8px; border-left: 4px solid #3b82f6; }
        .metric-label { font-size: 0.875rem; color: #6b7280; margin-bottom: 5px; }
        .metric-value { font-size: 1.5rem; font-weight: bold; color: #1f2937; }
        .section { margin-bottom: 30px; }
        .section-title { font-size: 1.25rem; font-weight: bold; color: #1f2937; margin-bottom: 15px; }
        .step { border: 1px solid #e5e7eb; border-radius: 8px; margin-bottom: 15px; }
        .step-header { background: #f9fafb; padding: 15px; border-bottom: 1px solid #e5e7eb; }
        .step-title { font-weight: bold; color: #1f2937; margin: 0; }
        .step-meta { color: #6b7280; font-size: 0.875rem; margin-top: 5px; }
        .actions { padding: 15px; }
        .action { background: #f8fafc; padding: 10px; border-radius: 6px; margin-bottom: 10px; border-left: 3px solid #10b981; }
        .action.failed { border-left-color: #ef4444; }
        .action-title { font-weight: 600; color: #1f2937; }
        .action-meta { color: #6b7280; font-size: 0.875rem; margin-top: 3px; }
        .status-badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; }
        .status-success { background: #dcfce7; color: #166534; }
        .status-failed { background: #fecaca; color: #991b1b; }
        .files-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px; margin-top: 10px; }
        .file-item { background: white; border: 1px solid #e5e7eb; border-radius: 6px; padding: 10px; }
        .file-link { color: #3b82f6; text-decoration: none; font-size: 0.875rem; }
        .file-link:hover { text-decoration: underline; }
        .concurrent-users { background: #dbeafe; padding: 10px; border-radius: 6px; margin-bottom: 10px; }
        .concurrent-users-title { font-weight: 600; color: #1e40af; margin-bottom: 5px; }
        .concurrent-users-list { color: #1e40af; font-size: 0.875rem; }
    </style>
</head>
<body>
    <div class="container">`)

	// Header
	html.WriteString(fmt.Sprintf(`
        <div class="header">
            <h1 class="title">%s</h1>
            <p class="subtitle">Run ID: %s | Generated: %s</p>
        </div>`, automation.Name, run.ID, time.Now().Format("2006-01-02 15:04:05")))

	// Summary metrics
	totalSteps := len(stepMap)
	totalActions := 0
	for _, step := range stepMap {
		totalActions += len(step.Actions)
	}

	duration := "N/A"
	if run.StartTime != nil && run.EndTime != nil {
		d := run.EndTime.Sub(*run.StartTime)
		duration = fmt.Sprintf("%.2fs", d.Seconds())
	}

	html.WriteString(fmt.Sprintf(`
        <div class="summary">
            <div class="metric">
                <div class="metric-label">Total Steps</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric">
                <div class="metric-label">Total Actions</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric">
                <div class="metric-label">Concurrent Users</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric">
                <div class="metric-label">Duration</div>
                <div class="metric-value">%s</div>
            </div>
            <div class="metric">
                <div class="metric-label">Success Rate</div>
                <div class="metric-value">%.1f%%</div>
            </div>
        </div>`, totalSteps, totalActions, metrics.TotalRuns, duration, 100-metrics.OverallFailureRate))

	// Steps section
	html.WriteString(`<div class="section"><h2 class="section-title">Step-by-Step Report</h2>`)

	for _, step := range stepMap {
		concurrentUsersList := make([]string, 0, len(step.ConcurrentUsers))
		for userIndex := range step.ConcurrentUsers {
			concurrentUsersList = append(concurrentUsersList, fmt.Sprintf("User %d", userIndex))
		}

		statusClass := "status-success"
		if step.Status == "failed" {
			statusClass = "status-failed"
		}

		html.WriteString(fmt.Sprintf(`
            <div class="step">
                <div class="step-header">
                    <h3 class="step-title">%s</h3>
                    <div class="step-meta">
                        <span class="status-badge %s">%s</span> | 
                        %d actions | %d concurrent users
                    </div>
                </div>`, step.Name, statusClass, strings.ToUpper(step.Status), len(step.Actions), len(step.ConcurrentUsers)))

		if len(step.ConcurrentUsers) > 1 {
			html.WriteString(fmt.Sprintf(`
                <div class="concurrent-users">
                    <div class="concurrent-users-title">Concurrent Execution</div>
                    <div class="concurrent-users-list">%s</div>
                </div>`, strings.Join(concurrentUsersList, ", ")))
		}

		html.WriteString(`<div class="actions">`)
		for _, action := range step.Actions {
			actionStatusClass := "status-success"
			if action.Status == "failed" {
				actionStatusClass = "status-failed"
			}

			actionClass := "action"
			if action.Status == "failed" {
				actionClass += " failed"
			}

			parentInfo := ""
			if action.ParentActionID != "" {
				parentInfo = fmt.Sprintf(" (nested under %s)", action.ParentActionID[:8])
			}

			html.WriteString(fmt.Sprintf(`
                <div class="%s">
                    <div class="action-title">%s%s</div>
                    <div class="action-meta">
                        User %d | <span class="status-badge %s">%s</span> | Duration: %dms
                    </div>`, actionClass, action.Type, parentInfo, action.LoopIndex, actionStatusClass, strings.ToUpper(action.Status), action.Duration))

			if action.Error != "" {
				html.WriteString(fmt.Sprintf(`<div style="color: #dc2626; margin-top: 5px;">Error: %s</div>`, action.Error))
			}

			if len(action.OutputFiles) > 0 {
				html.WriteString(`<div class="files-grid">`)
				for _, file := range action.OutputFiles {
					fileName := filepath.Base(file)
					html.WriteString(fmt.Sprintf(`
                        <div class="file-item">
                            <a href="%s" class="file-link" target="_blank">%s</a>
                        </div>`, file, fileName))
				}
				html.WriteString(`</div>`)
			}

			html.WriteString(`</div>`)
		}
		html.WriteString(`</div></div>`)
	}

	html.WriteString(`</div></div></body></html>`)

	return html.String()
}

// Helper functions for type conversion
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func getInt64(m map[string]any, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	if v, ok := m[key].(int64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return int64(v)
	}
	return 0
}