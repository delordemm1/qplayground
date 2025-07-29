package automation

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/delordemm1/qplayground-cli/internal/storage"
)

// ReportData structures for generating comprehensive reports
type ActionExecutionData struct {
	ActionID       string                 `json:"action_id"`
	ActionName     string                 `json:"action_name"`
	ActionType     string                 `json:"action_type"`
	Status         string                 `json:"status"`
	Duration       int64                  `json:"duration_ms"`
	Error          string                 `json:"error,omitempty"`
	LoopIndex      int                    `json:"loop_index"`
	LocalLoopIndex int                    `json:"local_loop_index"`
	OutputFiles    []string               `json:"output_files"`
	Logs           []map[string]any       `json:"logs"`
	Data           map[string]interface{} `json:"data,omitempty"`
}

type AggregatedAction struct {
	Type              string                    `json:"type"`
	Selector          string                    `json:"selector,omitempty"`
	Executions        int                       `json:"executions"`
	SuccessCount      int                       `json:"success_count"`
	FailureCount      int                       `json:"failure_count"`
	Durations         []int64                   `json:"durations"`
	Stats             DurationStats             `json:"stats"`
	FailedExecutions  []FailedExecution         `json:"failed_executions"`
	RawActions        map[int]ActionExecutionData `json:"raw_actions"` // keyed by loop index
}

type FailedExecution struct {
	LoopIndex    int      `json:"loop_index"`
	ErrorMessage string   `json:"error_message"`
	OutputFiles  []string `json:"output_files"`
}

type DurationStats struct {
	Average     float64 `json:"average"`
	Min         int64   `json:"min"`
	Max         int64   `json:"max"`
	P50         int64   `json:"p50"`
	P95         int64   `json:"p95"`
	Count       int     `json:"count"`
	TotalTime   int64   `json:"total_time"`
}

type EnhancedStepData struct {
	ID                string                      `json:"id"`
	Name              string                      `json:"name"`
	StepOrder         int                         `json:"step_order"`
	AggregatedActions map[string]*AggregatedAction `json:"aggregated_actions"`
	TotalDuration     int64                       `json:"total_duration"`
	Status            string                      `json:"status"`
	StartTime         string                      `json:"start_time"`
	EndTime           string                      `json:"end_time"`
	ConcurrentUsers   []int                       `json:"concurrent_users"`
	TotalExecutions   int                         `json:"total_executions"`
	TotalFailures     int                         `json:"total_failures"`
	TotalOutputFiles  int                         `json:"total_output_files"`
	StepImageFiles    []string                    `json:"step_image_files"`
	RawActions        map[int]ActionExecutionData `json:"raw_actions"` // keyed by loop index
}

type PerformanceMetrics struct {
	TotalRuns           int                    `json:"total_runs"`
	OverallFailureRate  float64                `json:"overall_failure_rate"`
	TotalDuration       int64                  `json:"total_duration_ms"`
	AverageDuration     float64                `json:"average_duration_ms"`
	StepAverages        []StepAverageData      `json:"step_averages"`
	UserJourneyMetrics  []UserJourneyData      `json:"user_journey_metrics"`
	ActionTypeBreakdown map[string]ActionStats `json:"action_type_breakdown"`
}

type StepAverageData struct {
	Name            string  `json:"name"`
	AverageDuration float64 `json:"average_duration"`
	FailureRate     float64 `json:"failure_rate"`
	ExecutionCount  int     `json:"execution_count"`
}

type UserJourneyData struct {
	LoopIndex       int                         `json:"loop_index"`
	Status          string                      `json:"status"`
	TotalDuration   int64                       `json:"total_duration"`
	CompletedSteps  int                         `json:"completed_steps"`
	TotalSteps      int                         `json:"total_steps"`
	Journey         []UserJourneyStep           `json:"journey"`
	OutputFiles     []string                    `json:"output_files"`
}

type UserJourneyStep struct {
	StepID     string   `json:"step_id"`
	StepName   string   `json:"step_name"`
	Status     string   `json:"status"`
	Duration   int64    `json:"duration"`
	Error      string   `json:"error,omitempty"`
	OutputFiles []string `json:"output_files"`
	Timestamp  string   `json:"timestamp"`
}

type ActionStats struct {
	TotalExecutions int     `json:"total_executions"`
	SuccessCount    int     `json:"success_count"`
	FailureCount    int     `json:"failure_count"`
	FailureRate     float64 `json:"failure_rate"`
	AverageDuration float64 `json:"average_duration"`
	TotalDuration   int64   `json:"total_duration"`
}

type ReportSummary struct {
	Automation         *Automation         `json:"automation"`
	Run                *AutomationRun      `json:"run"`
	Config             *AutomationConfig   `json:"config"`
	Steps              []EnhancedStepData  `json:"steps"`
	Metrics            PerformanceMetrics  `json:"metrics"`
	GeneratedAt        time.Time           `json:"generated_at"`
	ReportVersion      string              `json:"report_version"`
	TotalOutputFiles   int                 `json:"total_output_files"`
	AllOutputFiles     []string            `json:"all_output_files"`
}

// GenerateReports creates comprehensive HTML, JSON, and CSV reports
func GenerateReports(automation *Automation, run *AutomationRun, config *AutomationConfig, outputDir, reportsR2Path string, storageService storage.StorageService) (detailedReportURL, userJourneyReportURL string, err error) {
	slog.Info("Starting report generation", "automation_id", automation.ID, "run_id", run.ID)

	// Parse logs and output files
	var logs []map[string]any
	var outputFiles []string

	if run.LogsJSON != "" {
		if err := json.Unmarshal([]byte(run.LogsJSON), &logs); err != nil {
			slog.Error("Failed to parse logs JSON", "error", err)
			logs = []map[string]any{}
		}
	}

	if run.OutputFilesJSON != "" {
		if err := json.Unmarshal([]byte(run.OutputFilesJSON), &outputFiles); err != nil {
			slog.Error("Failed to parse output files JSON", "error", err)
			outputFiles = []string{}
		}
	}

	// Process data for enhanced reporting
	enhancedSteps, performanceMetrics := processAutomationData(logs, outputFiles)

	// Create report summary
	reportSummary := ReportSummary{
		Automation:       automation,
		Run:              run,
		Config:           config,
		Steps:            enhancedSteps,
		Metrics:          performanceMetrics,
		GeneratedAt:      time.Now(),
		ReportVersion:    "2.0.0",
		TotalOutputFiles: len(outputFiles),
		AllOutputFiles:   outputFiles,
	}

	// Generate JSON report
	jsonReportURL, err := generateJSONReport(reportSummary, outputDir, reportsR2Path, storageService)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JSON report: %w", err)
	}

	// Generate CSV report
	csvReportURL, err := generateCSVReport(reportSummary, outputDir, reportsR2Path, storageService)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate CSV report: %w", err)
	}

	// Generate detailed HTML report
	detailedReportURL, err = generateDetailedHTMLReport(reportSummary, outputDir, reportsR2Path, storageService)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate detailed HTML report: %w", err)
	}

	// Generate user journey HTML report
	userJourneyReportURL, err = generateUserJourneyHTMLReport(reportSummary, outputDir, reportsR2Path, storageService)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate user journey HTML report: %w", err)
	}

	slog.Info("Report generation completed successfully",
		"automation_id", automation.ID,
		"run_id", run.ID,
		"json_report", jsonReportURL,
		"csv_report", csvReportURL,
		"detailed_html", detailedReportURL,
		"user_journey_html", userJourneyReportURL)

	return detailedReportURL, userJourneyReportURL, nil
}

// processAutomationData analyzes logs and output files to create enhanced reporting data
func processAutomationData(logs []map[string]any, outputFiles []string) ([]EnhancedStepData, PerformanceMetrics) {
	stepMap := make(map[string]*EnhancedStepData)
	actionTypeStats := make(map[string]*ActionStats)
	userJourneys := make(map[int]*UserJourneyData)
	
	totalRuns := 0
	totalFailures := 0
	totalDuration := int64(0)
	maxLoopIndex := -1

	// Process logs to build step and action data
	for _, logEntry := range logs {
		stepID, _ := logEntry["step_id"].(string)
		stepName, _ := logEntry["step_name"].(string)
		actionID, _ := logEntry["action_id"].(string)
		actionName, _ := logEntry["action_name"].(string)
		actionType, _ := logEntry["action_type"].(string)
		status, _ := logEntry["status"].(string)
		loopIndex, _ := logEntry["loop_index"].(float64)
		localLoopIndex, _ := logEntry["local_loop_index"].(float64)
		durationMs, _ := logEntry["duration_ms"].(float64)
		errorMsg, _ := logEntry["error"].(string)
		timestamp, _ := logEntry["timestamp"].(string)

		loopIdx := int(loopIndex)
		localLoopIdx := int(localLoopIndex)
		duration := int64(durationMs)

		if loopIdx > maxLoopIndex {
			maxLoopIndex = loopIdx
		}

		// Initialize step if not exists
		if stepID != "" && stepMap[stepID] == nil {
			stepMap[stepID] = &EnhancedStepData{
				ID:                stepID,
				Name:              stepName,
				AggregatedActions: make(map[string]*AggregatedAction),
				ConcurrentUsers:   []int{},
				StepImageFiles:    []string{},
				RawActions:        make(map[int]ActionExecutionData),
			}
		}

		// Track concurrent users
		if stepID != "" {
			step := stepMap[stepID]
			userExists := false
			for _, user := range step.ConcurrentUsers {
				if user == loopIdx {
					userExists = true
					break
				}
			}
			if !userExists {
				step.ConcurrentUsers = append(step.ConcurrentUsers, loopIdx)
			}
		}

		// Process action data
		if actionType != "" {
			actionKey := fmt.Sprintf("%s_%s", actionType, actionID)
			
			// Initialize aggregated action if not exists
			if stepID != "" {
				step := stepMap[stepID]
				if step.AggregatedActions[actionKey] == nil {
					step.AggregatedActions[actionKey] = &AggregatedAction{
						Type:             actionType,
						RawActions:       make(map[int]ActionExecutionData),
						FailedExecutions: []FailedExecution{},
						Durations:        []int64{},
					}
				}

				aggAction := step.AggregatedActions[actionKey]
				aggAction.Executions++
				
				if status == "failed" {
					aggAction.FailureCount++
					step.TotalFailures++
					totalFailures++
					
					aggAction.FailedExecutions = append(aggAction.FailedExecutions, FailedExecution{
						LoopIndex:    loopIdx,
						ErrorMessage: errorMsg,
						OutputFiles:  []string{}, // Will be populated later
					})
				} else {
					aggAction.SuccessCount++
				}

				if duration > 0 {
					aggAction.Durations = append(aggAction.Durations, duration)
					step.TotalDuration += duration
					totalDuration += duration
				}

				// Store raw action data
				actionData := ActionExecutionData{
					ActionID:       actionID,
					ActionName:     actionName,
					ActionType:     actionType,
					Status:         status,
					Duration:       duration,
					Error:          errorMsg,
					LoopIndex:      loopIdx,
					LocalLoopIndex: localLoopIdx,
					OutputFiles:    []string{}, // Will be populated later
					Logs:           []map[string]any{logEntry},
				}
				aggAction.RawActions[loopIdx] = actionData
				step.RawActions[loopIdx] = actionData
				step.TotalExecutions++
			}

			// Track action type statistics
			if actionTypeStats[actionType] == nil {
				actionTypeStats[actionType] = &ActionStats{}
			}
			actionStat := actionTypeStats[actionType]
			actionStat.TotalExecutions++
			actionStat.TotalDuration += duration
			if status == "failed" {
				actionStat.FailureCount++
			} else {
				actionStat.SuccessCount++
			}
		}

		// Build user journey data
		if userJourneys[loopIdx] == nil {
			userJourneys[loopIdx] = &UserJourneyData{
				LoopIndex:      loopIdx,
				Status:         "in_progress",
				Journey:        []UserJourneyStep{},
				OutputFiles:    []string{},
			}
		}

		journey := userJourneys[loopIdx]
		if stepID != "" {
			// Find or create journey step
			var journeyStep *UserJourneyStep
			for i := range journey.Journey {
				if journey.Journey[i].StepID == stepID {
					journeyStep = &journey.Journey[i]
					break
				}
			}
			if journeyStep == nil {
				journey.Journey = append(journey.Journey, UserJourneyStep{
					StepID:      stepID,
					StepName:    stepName,
					Status:      "in_progress",
					OutputFiles: []string{},
					Timestamp:   timestamp,
				})
				journeyStep = &journey.Journey[len(journey.Journey)-1]
			}

			// Update journey step
			journeyStep.Duration += duration
			if status == "failed" {
				journeyStep.Status = "failed"
				journeyStep.Error = errorMsg
				journey.Status = "failed"
			} else if journeyStep.Status != "failed" {
				journeyStep.Status = "success"
			}
		}
	}

	// Calculate total runs
	totalRuns = maxLoopIndex + 1
	if totalRuns == 0 {
		totalRuns = 1
	}

	// Process output files and associate with steps/actions
	imageExtensions := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	}

	for _, outputFile := range outputFiles {
		// Try to extract loop index from filename
		loopIdx := extractLoopIndexFromFilename(outputFile)
		
		// Check if it's an image file
		isImage := false
		for ext := range imageExtensions {
			if strings.HasSuffix(strings.ToLower(outputFile), ext) {
				isImage = true
				break
			}
		}

		// Associate with user journey
		if journey, exists := userJourneys[loopIdx]; exists {
			journey.OutputFiles = append(journey.OutputFiles, outputFile)
		}

		// Associate with steps (simplified - in practice you'd need more sophisticated matching)
		for _, step := range stepMap {
			if isImage {
				step.StepImageFiles = append(step.StepImageFiles, outputFile)
			}
			step.TotalOutputFiles++
		}
	}

	// Calculate statistics for aggregated actions
	for _, step := range stepMap {
		for _, aggAction := range step.AggregatedActions {
			aggAction.Stats = calculateDurationStats(aggAction.Durations)
		}
		
		// Determine step status
		if step.TotalFailures > 0 {
			step.Status = "failed"
		} else {
			step.Status = "completed"
		}
	}

	// Calculate action type statistics
	actionTypeBreakdown := make(map[string]ActionStats)
	for actionType, stats := range actionTypeStats {
		if stats.TotalExecutions > 0 {
			stats.FailureRate = float64(stats.FailureCount) / float64(stats.TotalExecutions) * 100
			stats.AverageDuration = float64(stats.TotalDuration) / float64(stats.TotalExecutions)
		}
		actionTypeBreakdown[actionType] = *stats
	}

	// Finalize user journeys
	userJourneyMetrics := make([]UserJourneyData, 0, len(userJourneys))
	for _, journey := range userJourneys {
		journey.CompletedSteps = 0
		journey.TotalSteps = len(journey.Journey)
		for _, step := range journey.Journey {
			journey.TotalDuration += step.Duration
			if step.Status == "success" {
				journey.CompletedSteps++
			}
		}
		if journey.Status == "in_progress" && journey.CompletedSteps == journey.TotalSteps {
			journey.Status = "completed"
		}
		userJourneyMetrics = append(userJourneyMetrics, *journey)
	}

	// Sort user journeys by loop index
	sort.Slice(userJourneyMetrics, func(i, j int) bool {
		return userJourneyMetrics[i].LoopIndex < userJourneyMetrics[j].LoopIndex
	})

	// Convert step map to slice and sort by step order
	enhancedSteps := make([]EnhancedStepData, 0, len(stepMap))
	for _, step := range stepMap {
		enhancedSteps = append(enhancedSteps, *step)
	}
	sort.Slice(enhancedSteps, func(i, j int) bool {
		return enhancedSteps[i].StepOrder < enhancedSteps[j].StepOrder
	})

	// Calculate step averages
	stepAverages := make([]StepAverageData, len(enhancedSteps))
	for i, step := range enhancedSteps {
		avgDuration := float64(0)
		failureRate := float64(0)
		if step.TotalExecutions > 0 {
			avgDuration = float64(step.TotalDuration) / float64(step.TotalExecutions)
			failureRate = float64(step.TotalFailures) / float64(step.TotalExecutions) * 100
		}
		stepAverages[i] = StepAverageData{
			Name:            step.Name,
			AverageDuration: avgDuration,
			FailureRate:     failureRate,
			ExecutionCount:  step.TotalExecutions,
		}
	}

	// Calculate overall metrics
	overallFailureRate := float64(0)
	averageDuration := float64(0)
	if totalRuns > 0 {
		overallFailureRate = float64(totalFailures) / float64(totalRuns) * 100
		averageDuration = float64(totalDuration) / float64(totalRuns)
	}

	performanceMetrics := PerformanceMetrics{
		TotalRuns:           totalRuns,
		OverallFailureRate:  overallFailureRate,
		TotalDuration:       totalDuration,
		AverageDuration:     averageDuration,
		StepAverages:        stepAverages,
		UserJourneyMetrics:  userJourneyMetrics,
		ActionTypeBreakdown: actionTypeBreakdown,
	}

	return enhancedSteps, performanceMetrics
}

// Helper functions
func extractLoopIndexFromFilename(filename string) int {
	// Try to extract loop index from patterns like "user-1-", "{{loopIndex}}", etc.
	if strings.Contains(filename, "user-") {
		parts := strings.Split(filename, "user-")
		if len(parts) > 1 {
			indexPart := strings.Split(parts[1], "-")[0]
			if idx, err := strconv.Atoi(indexPart); err == nil {
				return idx
			}
		}
	}
	return 0 // Default to loop index 0
}

func calculateDurationStats(durations []int64) DurationStats {
	if len(durations) == 0 {
		return DurationStats{}
	}

	// Sort durations for percentile calculations
	sorted := make([]int64, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Calculate basic stats
	var total int64
	min := sorted[0]
	max := sorted[len(sorted)-1]
	
	for _, d := range durations {
		total += d
	}
	
	average := float64(total) / float64(len(durations))
	
	// Calculate percentiles
	p50 := calculatePercentile(sorted, 50)
	p95 := calculatePercentile(sorted, 95)

	return DurationStats{
		Average:   average,
		Min:       min,
		Max:       max,
		P50:       p50,
		P95:       p95,
		Count:     len(durations),
		TotalTime: total,
	}
}

func calculatePercentile(sortedValues []int64, percentile int) int64 {
	if len(sortedValues) == 0 {
		return 0
	}
	
	index := float64(percentile) / 100.0 * float64(len(sortedValues)-1)
	if index == float64(int(index)) {
		return sortedValues[int(index)]
	}
	
	lower := int(index)
	upper := lower + 1
	if upper >= len(sortedValues) {
		return sortedValues[len(sortedValues)-1]
	}
	
	weight := index - float64(lower)
	return int64(float64(sortedValues[lower])*(1-weight) + float64(sortedValues[upper])*weight)
}

// Report generation functions
func generateJSONReport(summary ReportSummary, outputDir, reportsR2Path string, storageService storage.StorageService) (string, error) {
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	return saveReport("report.json", jsonData, "application/json", outputDir, reportsR2Path, storageService)
}

func generateCSVReport(summary ReportSummary, outputDir, reportsR2Path string, storageService storage.StorageService) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"Step Name", "Action Type", "Loop Index", "Status", "Duration (ms)", 
		"Error", "Output Files", "Timestamp",
	}
	writer.Write(header)

	// Write data rows
	for _, step := range summary.Steps {
		for _, action := range step.RawActions {
			row := []string{
				step.Name,
				action.ActionType,
				strconv.Itoa(action.LoopIndex),
				action.Status,
				strconv.FormatInt(action.Duration, 10),
				action.Error,
				strings.Join(action.OutputFiles, ";"),
				"", // Timestamp would need to be extracted from logs
			}
			writer.Write(row)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("failed to write CSV: %w", err)
	}

	return saveReport("report.csv", buf.Bytes(), "text/csv", outputDir, reportsR2Path, storageService)
}

func generateDetailedHTMLReport(summary ReportSummary, outputDir, reportsR2Path string, storageService storage.StorageService) (string, error) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Detailed Automation Report - {{.Automation.Name}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        body { background-color: #f8f9fa; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; }
        .main-container { background-color: white; min-height: 100vh; box-shadow: 0 0 20px rgba(0,0,0,0.1); }
        .status-success { color: #198754; }
        .status-failed { color: #dc3545; }
        .status-completed { color: #198754; }
        .metric-card { transition: transform 0.2s; }
        .metric-card:hover { transform: translateY(-2px); }
        .screenshot-thumb { max-width: 100px; height: auto; cursor: pointer; border-radius: 4px; }
        .modal { display: none; position: fixed; z-index: 1050; left: 0; top: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.9); }
        .modal-content { margin: auto; display: block; width: 90%; max-width: 1200px; max-height: 90vh; object-fit: contain; }
        .close { position: absolute; top: 15px; right: 35px; color: #fff; font-size: 40px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="main-container">
        <div class="container-fluid py-4">
            <!-- Header -->
            <div class="text-center mb-5">
                <h1 class="display-4 fw-bold text-primary">
                    <i class="bi bi-clipboard-data me-3"></i>
                    Detailed Automation Report
                </h1>
                <p class="lead text-muted">{{.Automation.Name}} - Run {{.Run.ID}}</p>
                <p class="text-muted">Generated on {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</p>
            </div>

            <!-- Summary Cards -->
            <div class="row mb-5">
                <div class="col-md-3">
                    <div class="card metric-card border-0 shadow-sm">
                        <div class="card-body text-center">
                            <i class="bi bi-people-fill text-primary fs-1"></i>
                            <h3 class="mt-2">{{.Metrics.TotalRuns}}</h3>
                            <p class="text-muted mb-0">Total Runs</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card metric-card border-0 shadow-sm">
                        <div class="card-body text-center">
                            <i class="bi bi-clock-fill text-info fs-1"></i>
                            <h3 class="mt-2">{{printf "%.2f" .Metrics.AverageDuration}}ms</h3>
                            <p class="text-muted mb-0">Avg Duration</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card metric-card border-0 shadow-sm">
                        <div class="card-body text-center">
                            <i class="bi bi-exclamation-triangle-fill text-warning fs-1"></i>
                            <h3 class="mt-2">{{printf "%.1f" .Metrics.OverallFailureRate}}%</h3>
                            <p class="text-muted mb-0">Failure Rate</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card metric-card border-0 shadow-sm">
                        <div class="card-body text-center">
                            <i class="bi bi-file-earmark-image-fill text-success fs-1"></i>
                            <h3 class="mt-2">{{.TotalOutputFiles}}</h3>
                            <p class="text-muted mb-0">Output Files</p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Steps Analysis -->
            <div class="row mb-5">
                <div class="col-12">
                    <div class="card border-0 shadow-sm">
                        <div class="card-header bg-primary text-white">
                            <h3 class="mb-0"><i class="bi bi-list-ol me-2"></i>Steps Performance</h3>
                        </div>
                        <div class="card-body">
                            {{range .Steps}}
                            <div class="border rounded p-3 mb-3">
                                <div class="d-flex justify-content-between align-items-center mb-2">
                                    <h5 class="mb-0">{{.Name}}</h5>
                                    <span class="badge bg-{{if eq .Status "completed"}}success{{else}}danger{{end}}">{{.Status}}</span>
                                </div>
                                <div class="row">
                                    <div class="col-md-6">
                                        <small class="text-muted">
                                            <strong>Executions:</strong> {{.TotalExecutions}} | 
                                            <strong>Failures:</strong> {{.TotalFailures}} | 
                                            <strong>Duration:</strong> {{.TotalDuration}}ms
                                        </small>
                                    </div>
                                    <div class="col-md-6">
                                        <small class="text-muted">
                                            <strong>Users:</strong> {{len .ConcurrentUsers}} | 
                                            <strong>Files:</strong> {{.TotalOutputFiles}}
                                        </small>
                                    </div>
                                </div>
                                {{if .StepImageFiles}}
                                <div class="mt-2">
                                    <small class="text-muted d-block mb-2"><strong>Screenshots:</strong></small>
                                    <div class="d-flex flex-wrap gap-2">
                                        {{range .StepImageFiles}}
                                        <img src="{{.}}" class="screenshot-thumb" onclick="openModal('{{.}}')" alt="Screenshot">
                                        {{end}}
                                    </div>
                                </div>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Action Type Breakdown -->
            <div class="row mb-5">
                <div class="col-12">
                    <div class="card border-0 shadow-sm">
                        <div class="card-header bg-info text-white">
                            <h3 class="mb-0"><i class="bi bi-pie-chart-fill me-2"></i>Action Type Analysis</h3>
                        </div>
                        <div class="card-body">
                            <div class="table-responsive">
                                <table class="table table-striped">
                                    <thead>
                                        <tr>
                                            <th>Action Type</th>
                                            <th>Total Executions</th>
                                            <th>Success Rate</th>
                                            <th>Avg Duration</th>
                                            <th>Total Duration</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {{range $actionType, $stats := .Metrics.ActionTypeBreakdown}}
                                        <tr>
                                            <td><code>{{$actionType}}</code></td>
                                            <td>{{$stats.TotalExecutions}}</td>
                                            <td>
                                                <span class="badge bg-{{if lt $stats.FailureRate 10.0}}success{{else if lt $stats.FailureRate 25.0}}warning{{else}}danger{{end}}">
                                                    {{printf "%.1f" (sub 100.0 $stats.FailureRate)}}%
                                                </span>
                                            </td>
                                            <td>{{printf "%.2f" $stats.AverageDuration}}ms</td>
                                            <td>{{$stats.TotalDuration}}ms</td>
                                        </tr>
                                        {{end}}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Configuration -->
            <div class="row">
                <div class="col-12">
                    <div class="card border-0 shadow-sm">
                        <div class="card-header bg-secondary text-white">
                            <h3 class="mb-0"><i class="bi bi-gear-fill me-2"></i>Automation Configuration</h3>
                        </div>
                        <div class="card-body">
                            <pre class="bg-light p-3 rounded"><code>{{.Config | jsonPretty}}</code></pre>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Image Modal -->
    <div id="imageModal" class="modal">
        <span class="close" onclick="closeModal()">&times;</span>
        <img class="modal-content" id="modalImage">
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        function openModal(src) {
            document.getElementById('imageModal').style.display = 'block';
            document.getElementById('modalImage').src = src;
        }
        function closeModal() {
            document.getElementById('imageModal').style.display = 'none';
        }
        window.onclick = function(event) {
            const modal = document.getElementById('imageModal');
            if (event.target == modal) {
                closeModal();
            }
        }
    </script>
</body>
</html>`

	// Create template with helper functions
	funcMap := template.FuncMap{
		"jsonPretty": func(v interface{}) string {
			b, _ := json.MarshalIndent(v, "", "  ")
			return string(b)
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
	}

	t, err := template.New("detailed").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, summary); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return saveReport("detailed_report.html", buf.Bytes(), "text/html", outputDir, reportsR2Path, storageService)
}

func generateUserJourneyHTMLReport(summary ReportSummary, outputDir, reportsR2Path string, storageService storage.StorageService) (string, error) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Journey Report - {{.Automation.Name}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        body { background-color: #f8f9fa; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; }
        .main-container { background-color: white; min-height: 100vh; box-shadow: 0 0 20px rgba(0,0,0,0.1); }
        .user-card { transition: transform 0.2s; cursor: pointer; }
        .user-card:hover { transform: translateY(-2px); }
        .timeline-item { border-left: 3px solid #dee2e6; padding-left: 1rem; margin-bottom: 1rem; }
        .timeline-item.success { border-left-color: #198754; }
        .timeline-item.failed { border-left-color: #dc3545; }
        .screenshot-thumb { max-width: 80px; height: auto; cursor: pointer; border-radius: 4px; margin: 2px; }
    </style>
</head>
<body>
    <div class="main-container">
        <div class="container-fluid py-4">
            <!-- Header -->
            <div class="text-center mb-5">
                <h1 class="display-4 fw-bold text-primary">
                    <i class="bi bi-person-lines-fill me-3"></i>
                    User Journey Report
                </h1>
                <p class="lead text-muted">{{.Automation.Name}} - Individual User Analysis</p>
                <p class="text-muted">Generated on {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</p>
            </div>

            <!-- User Journey Cards -->
            <div class="row">
                {{range .Metrics.UserJourneyMetrics}}
                <div class="col-md-6 col-lg-4 mb-4">
                    <div class="card user-card border-0 shadow-sm h-100">
                        <div class="card-header bg-{{if eq .Status "completed"}}success{{else if eq .Status "failed"}}danger{{else}}warning{{end}} text-white">
                            <h5 class="mb-0">
                                <i class="bi bi-person-fill me-2"></i>
                                User {{.LoopIndex}}
                            </h5>
                        </div>
                        <div class="card-body">
                            <div class="row mb-3">
                                <div class="col-6">
                                    <small class="text-muted">Status</small>
                                    <div class="fw-bold text-{{if eq .Status "completed"}}success{{else if eq .Status "failed"}}danger{{else}}warning{{end}}">
                                        {{.Status | title}}
                                    </div>
                                </div>
                                <div class="col-6">
                                    <small class="text-muted">Duration</small>
                                    <div class="fw-bold">{{.TotalDuration}}ms</div>
                                </div>
                            </div>
                            <div class="row mb-3">
                                <div class="col-6">
                                    <small class="text-muted">Progress</small>
                                    <div class="fw-bold">{{.CompletedSteps}}/{{.TotalSteps}}</div>
                                </div>
                                <div class="col-6">
                                    <small class="text-muted">Files</small>
                                    <div class="fw-bold">{{len .OutputFiles}}</div>
                                </div>
                            </div>
                            
                            <!-- Progress Bar -->
                            <div class="progress mb-3" style="height: 8px;">
                                <div class="progress-bar bg-{{if eq .Status "completed"}}success{{else if eq .Status "failed"}}danger{{else}}warning{{end}}" 
                                     style="width: {{if gt .TotalSteps 0}}{{div (mul .CompletedSteps 100) .TotalSteps}}{{else}}0{{end}}%"></div>
                            </div>

                            <!-- Journey Steps -->
                            <div class="timeline">
                                {{range .Journey}}
                                <div class="timeline-item {{.Status}}">
                                    <div class="d-flex justify-content-between align-items-start">
                                        <div>
                                            <strong>{{.StepName}}</strong>
                                            <div class="small text-muted">{{.Duration}}ms</div>
                                            {{if .Error}}
                                            <div class="small text-danger">{{.Error}}</div>
                                            {{end}}
                                        </div>
                                        <span class="badge bg-{{if eq .Status "success"}}success{{else}}danger{{end}}">
                                            {{if eq .Status "success"}}✓{{else}}✗{{end}}
                                        </span>
                                    </div>
                                </div>
                                {{end}}
                            </div>

                            <!-- Output Files -->
                            {{if .OutputFiles}}
                            <div class="mt-3">
                                <small class="text-muted d-block mb-2"><strong>Output Files:</strong></small>
                                <div class="d-flex flex-wrap">
                                    {{range .OutputFiles}}
                                    {{if or (hasSuffix . ".png") (hasSuffix . ".jpg") (hasSuffix . ".jpeg")}}
                                    <img src="{{.}}" class="screenshot-thumb" onclick="openModal('{{.}}')" alt="Screenshot">
                                    {{else}}
                                    <a href="{{.}}" target="_blank" class="btn btn-sm btn-outline-secondary me-1 mb-1">
                                        <i class="bi bi-file-earmark"></i>
                                    </a>
                                    {{end}}
                                    {{end}}
                                </div>
                            </div>
                            {{end}}
                        </div>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        function openModal(src) {
            // Simple modal implementation
            const modal = document.createElement('div');
            modal.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.9);z-index:9999;display:flex;align-items:center;justify-content:center;';
            modal.onclick = () => document.body.removeChild(modal);
            
            const img = document.createElement('img');
            img.src = src;
            img.style.cssText = 'max-width:90%;max-height:90%;object-fit:contain;';
            
            modal.appendChild(img);
            document.body.appendChild(modal);
        }
    </script>
</body>
</html>`

	// Create template with helper functions
	funcMap := template.FuncMap{
		"title": strings.Title,
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"hasSuffix": strings.HasSuffix,
	}

	t, err := template.New("journey").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, summary); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return saveReport("user_journey_report.html", buf.Bytes(), "text/html", outputDir, reportsR2Path, storageService)
}

func saveReport(filename string, data []byte, contentType, outputDir, reportsR2Path string, storageService storage.StorageService) (string, error) {
	// Save locally if outputDir is provided
	if outputDir != "" {
		localPath := filepath.Join(outputDir, "reports", filename)
		if err := saveToLocal(localPath, data); err != nil {
			slog.Warn("Failed to save report locally", "path", localPath, "error", err)
		}
	}

	// Upload to storage if configured
	if storageService != nil && reportsR2Path != "" {
		key := fmt.Sprintf("%s/%s", reportsR2Path, filename)
		reader := bytes.NewReader(data)
		
		publicURL, err := storageService.UploadFile(context.Background(), key, reader, contentType)
		if err != nil {
			slog.Error("Failed to upload report to storage", "key", key, "error", err)
			// Return local path as fallback
			if outputDir != "" {
				return filepath.Join(outputDir, "reports", filename), nil
			}
			return "", err
		}
		
		slog.Info("Report uploaded successfully", "key", key, "url", publicURL)
		return publicURL, nil
	}

	// Return local path if no storage configured
	if outputDir != "" {
		return filepath.Join(outputDir, "reports", filename), nil
	}

	return "", fmt.Errorf("no output directory or storage service configured")
}

func saveToLocal(path string, data []byte) error {
	// This would be implemented to save to local filesystem
	// For now, we'll just log that it would be saved
	slog.Info("Would save report locally", "path", path, "size", len(data))
	return nil
}