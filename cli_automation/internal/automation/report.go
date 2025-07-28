package automation

import (
	"context"
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
func GenerateReports(automation *Automation, run *AutomationRun, config *AutomationConfig, outputBaseDir, reportsR2Path string, storageService storage.StorageService) (detailedReportURL, userJourneyReportURL string, err error) {
	// Create unique directory for this run
	runTimestamp := time.Now().Format("20060102-150405")
	runDir := filepath.Join(outputBaseDir, fmt.Sprintf("%s-%s", runTimestamp, run.ID[:8]))
	reportsDir := filepath.Join(runDir, "reports")

	// Create directories
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create reports directory: %w", err)
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

	// Generate detailed HTML report
	detailedURL, err := generateDetailedHTMLReport(automation, run, config, logs, outputFiles, reportsR2Path, storageService)
	if err != nil {
		slog.Error("Failed to generate detailed HTML report", "error", err)
	} else {
		detailedReportURL = detailedURL
	}

	// Generate user journey HTML report
	userJourneyURL, err := generateUserJourneyHTMLReport(automation, run, logs, outputFiles, reportsR2Path, storageService)
	if err != nil {
		slog.Error("Failed to generate user journey HTML report", "error", err)
	} else {
		userJourneyReportURL = userJourneyURL
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
	return detailedReportURL, userJourneyReportURL, nil
}

// generateDetailedHTMLReport creates a comprehensive HTML report similar to the current web UI
func generateDetailedHTMLReport(automation *Automation, run *AutomationRun, config *AutomationConfig, logs []map[string]any, outputFiles []string, reportsR2Path string, storageService storage.StorageService) (string, error) {
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
					Name:           getString(log, "action_name"),
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
	htmlContent := generateDetailedHTMLContent(automation, run, config, stepMap, metrics, outputFiles)

	// Upload HTML file to storage
	htmlKey := fmt.Sprintf("%s/detailed_report.html", reportsR2Path)
	reader := strings.NewReader(htmlContent)
	publicURL, err := storageService.UploadFile(context.Background(), htmlKey, reader, "text/html")
	if err != nil {
		return "", fmt.Errorf("failed to upload detailed HTML report: %w", err)
	}

	return publicURL, nil
}

// generateUserJourneyHTMLReport creates a user journey HTML report based on the provided snippet
func generateUserJourneyHTMLReport(automation *Automation, run *AutomationRun, logs []map[string]any, outputFiles []string, reportsR2Path string, storageService storage.StorageService) (string, error) {
	// Process logs to create user journey data
	groupedReports := make(map[int][]map[string]interface{})
	uniqueFeatures := make(map[string]bool)
	allScreenshots := []string{}
	userTimings := make(map[int]map[string]interface{})

	for _, log := range logs {
		loopIndex := getInt(log, "loop_index")
		stepName := getString(log, "step_name")
		actionType := getString(log, "action_type")
		status := getString(log, "status")
		timestamp := getString(log, "timestamp")
		duration := getInt64(log, "duration_ms")
		errorMsg := getString(log, "error")
		outputFile := getString(log, "output_file")

		// Initialize user if not exists
		if _, exists := groupedReports[loopIndex]; !exists {
			groupedReports[loopIndex] = []map[string]interface{}{}
			userTimings[loopIndex] = map[string]interface{}{
				"userId": loopIndex,
				"totalDuration": int64(0),
			}
		}

		// Add to user timings
		if userTiming, exists := userTimings[loopIndex]; exists {
			if totalDuration, ok := userTiming["totalDuration"].(int64); ok {
				userTiming["totalDuration"] = totalDuration + duration
			}
		}

		// Create report item
		reportItem := map[string]interface{}{
			"userId":    loopIndex,
			"feature":   stepName,
			"scenario":  actionType,
			"status":    status,
			"timestamp": timestamp,
			"error":     errorMsg,
			"url":       "",
		}

		if outputFile != "" {
			reportItem["screenshot"] = outputFile
			allScreenshots = append(allScreenshots, outputFile)
		}

		groupedReports[loopIndex] = append(groupedReports[loopIndex], reportItem)
		
		if stepName != "" {
			uniqueFeatures[stepName] = true
		}
	}

	// Convert unique features to slice
	featuresSlice := make([]string, 0, len(uniqueFeatures))
	for feature := range uniqueFeatures {
		featuresSlice = append(featuresSlice, feature)
	}

	// Convert user timings to slice
	userTimingsSlice := make([]map[string]interface{}, 0, len(userTimings))
	for _, timing := range userTimings {
		totalDuration := timing["totalDuration"].(int64)
		timing["timeSeconds"] = fmt.Sprintf("%.2f", float64(totalDuration)/1000.0)
		timing["timeMinutes"] = fmt.Sprintf("%.2f", float64(totalDuration)/60000.0)
		userTimingsSlice = append(userTimingsSlice, timing)
	}

	// Calculate total time
	var totalDurationMs int64
	if run.StartTime != nil && run.EndTime != nil {
		totalDurationMs = run.EndTime.Sub(*run.StartTime).Milliseconds()
	}

	totalTime := map[string]interface{}{
		"seconds":      fmt.Sprintf("%.2f", float64(totalDurationMs)/1000.0),
		"minutes":      fmt.Sprintf("%.2f", float64(totalDurationMs)/60000.0),
		"milliseconds": totalDurationMs,
	}

	// Generate HTML content
	htmlContent := generateUserJourneyHTMLContent(automation, run, groupedReports, featuresSlice, allScreenshots, userTimingsSlice, totalTime)

	// Upload HTML file to storage
	htmlKey := fmt.Sprintf("%s/user_journey_report.html", reportsR2Path)
	reader := strings.NewReader(htmlContent)
	publicURL, err := storageService.UploadFile(context.Background(), htmlKey, reader, "text/html")
	if err != nil {
		return "", fmt.Errorf("failed to upload user journey HTML report: %w", err)
	}

	return publicURL, nil
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
					Name:           getString(log, "action_name"),
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

// generateDetailedHTMLContent creates the HTML content for the detailed report
func generateDetailedHTMLContent(automation *Automation, run *AutomationRun, config *AutomationConfig, stepMap map[string]*StepReport, metrics PerformanceMetrics, outputFiles []string) string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Detailed Automation Report - ` + automation.Name + `</title>
    
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    
    <!-- Chart.js -->
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            background: #f8f9fa; 
        }
        .main-container { 
            background-color: white; 
            min-height: 100vh; 
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .metric-card { 
            background: #f9fafb; 
            padding: 20px; 
            border-radius: 8px; 
            border-left: 4px solid #3b82f6; 
        }
        .metric-label { 
            font-size: 0.875rem; 
            color: #6b7280; 
            margin-bottom: 5px; 
        }
        .metric-value { 
            font-size: 1.5rem; 
            font-weight: bold; 
            color: #1f2937; 
        }
        .step-card { 
            border: 1px solid #e5e7eb; 
            border-radius: 8px; 
            margin-bottom: 15px; 
        }
        .step-header { 
            background: #f9fafb; 
            padding: 15px; 
            border-bottom: 1px solid #e5e7eb; 
        }
        .action-item { 
            background: #f8fafc; 
            padding: 10px; 
            border-radius: 6px; 
            margin-bottom: 10px; 
            border-left: 3px solid #10b981; 
        }
        .action-item.failed { 
            border-left-color: #ef4444; 
        }
        .status-badge { 
            display: inline-block; 
            padding: 2px 8px; 
            border-radius: 12px; 
            font-size: 0.75rem; 
            font-weight: 600; 
        }
        .status-success { 
            background: #dcfce7; 
            color: #166534; 
        }
        .status-failed { 
            background: #fecaca; 
            color: #991b1b; 
        }
        .chart-container {
            position: relative;
            height: 400px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="main-container">
        <div class="container-fluid py-4">`)

	// Header
	html.WriteString(fmt.Sprintf(`
            <div class="text-center mb-5">
                <h1 class="display-4 fw-bold text-primary">
                    <i class="bi bi-clipboard-data me-3"></i>
                    %s
                </h1>
                <p class="lead text-muted">Run ID: %s | Generated: %s</p>
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
            <div class="row mb-5">
                <div class="col-md-3">
                    <div class="metric-card">
                        <div class="metric-label">Total Steps</div>
                        <div class="metric-value">%d</div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="metric-card">
                        <div class="metric-label">Total Actions</div>
                        <div class="metric-value">%d</div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="metric-card">
                        <div class="metric-label">Concurrent Users</div>
                        <div class="metric-value">%d</div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="metric-card">
                        <div class="metric-label">Duration</div>
                        <div class="metric-value">%s</div>
                    </div>
                </div>
            </div>`, totalSteps, totalActions, metrics.TotalRuns, duration))

	// Performance Charts
	if metrics.TotalRuns > 1 {
		html.WriteString(`
            <div class="row mb-5">
                <div class="col-12">
                    <div class="card">
                        <div class="card-header">
                            <h3 class="mb-0">Performance Analysis</h3>
                        </div>
                        <div class="card-body">
                            <div class="chart-container">
                                <canvas id="performanceChart"></canvas>
                            </div>
                        </div>
                    </div>
                </div>
            </div>`)
	}

	// Steps section
	html.WriteString(`<div class="row"><div class="col-12"><h2 class="mb-4">Step-by-Step Report</h2>`)

	for _, step := range stepMap {
		statusClass := "status-success"
		if step.Status == "failed" {
			statusClass = "status-failed"
		}

		html.WriteString(fmt.Sprintf(`
            <div class="step-card">
                <div class="step-header">
                    <h4 class="mb-2">%s</h4>
                    <div>
                        <span class="status-badge %s">%s</span>
                        <span class="ms-3 text-muted">%d actions | %d concurrent users</span>
                    </div>
                </div>
                <div class="p-3">`, step.Name, statusClass, strings.ToUpper(step.Status), len(step.Actions), len(step.ConcurrentUsers)))

		for _, action := range step.Actions {
			actionStatusClass := "status-success"
			actionClass := "action-item"
			if action.Status == "failed" {
				actionStatusClass = "status-failed"
				actionClass += " failed"
			}

			actionDisplayName := action.Type
			if action.Name != "" {
				actionDisplayName = fmt.Sprintf("%s (%s)", action.Name, action.Type)
			}

			html.WriteString(fmt.Sprintf(`
                    <div class="%s">
                        <div class="fw-bold">%s</div>
                        <div class="small text-muted">
                            User %d | <span class="status-badge %s">%s</span> | Duration: %dms
                        </div>`, actionClass, actionDisplayName, action.LoopIndex, actionStatusClass, strings.ToUpper(action.Status), action.Duration))

			if action.Error != "" {
				html.WriteString(fmt.Sprintf(`<div class="text-danger mt-2">Error: %s</div>`, action.Error))
			}

			if len(action.OutputFiles) > 0 {
				html.WriteString(`<div class="mt-2">`)
				for _, file := range action.OutputFiles {
					fileName := filepath.Base(file)
					html.WriteString(fmt.Sprintf(`
                        <a href="%s" class="btn btn-sm btn-outline-primary me-2" target="_blank">%s</a>`, file, fileName))
				}
				html.WriteString(`</div>`)
			}

			html.WriteString(`</div>`)
		}
		html.WriteString(`</div></div>`)
	}

	html.WriteString(`</div></div>`)

	// Chart.js initialization
	if metrics.TotalRuns > 1 {
		html.WriteString(`
        <script>
            document.addEventListener('DOMContentLoaded', function() {
                const ctx = document.getElementById('performanceChart').getContext('2d');
                new Chart(ctx, {
                    type: 'bar',
                    data: {
                        labels: [`)

		// Add step names
		stepNames := make([]string, 0, len(stepMap))
		for _, step := range stepMap {
			stepNames = append(stepNames, step.Name)
		}
		for i, name := range stepNames {
			if i > 0 {
				html.WriteString(", ")
			}
			html.WriteString(fmt.Sprintf("'%s'", name))
		}

		html.WriteString(`],
                        datasets: [{
                            label: 'Average Duration (ms)',
                            data: [`)

		// Add duration data
		for i, step := range stepMap {
			if i > 0 {
				html.WriteString(", ")
			}
			// Calculate average duration for this step
			totalDuration := int64(0)
			actionCount := 0
			for _, action := range step.Actions {
				totalDuration += action.Duration
				actionCount++
			}
			avgDuration := int64(0)
			if actionCount > 0 {
				avgDuration = totalDuration / int64(actionCount)
			}
			html.WriteString(fmt.Sprintf("%d", avgDuration))
		}

		html.WriteString(`],
                            backgroundColor: 'rgba(59, 130, 246, 0.6)',
                            borderColor: 'rgba(59, 130, 246, 1)',
                            borderWidth: 1
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            title: {
                                display: true,
                                text: 'Step Performance Overview'
                            }
                        },
                        scales: {
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'Duration (ms)'
                                }
                            }
                        }
                    }
                });
            });
        </script>`)
	}

	html.WriteString(`
        </div>
    </div>
    
    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>`)

	return html.String()
}

// generateUserJourneyHTMLContent creates the HTML content for the user journey report
func generateUserJourneyHTMLContent(automation *Automation, run *AutomationRun, groupedReports map[int][]map[string]interface{}, uniqueFeatures, allScreenshots []string, userTimings []map[string]interface{}, totalTime map[string]interface{}) string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Journey Report - ` + automation.Name + `</title>
    
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    
    <style>
        body { 
            background-color: #f8f9fa; 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        .main-container { 
            background-color: white; 
            min-height: 100vh; 
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            max-width: 1600px;
            margin: 0 auto;
            padding: 1rem 1.5rem;
        }
        .success { color: #198754; font-weight: 600; }
        .failed { color: #dc3545; font-weight: 600; }
        .screenshot-img { 
            max-width: 150px; 
            height: auto; 
            cursor: pointer; 
            transition: all 0.3s ease; 
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .screenshot-img:hover { 
            transform: scale(1.05); 
            box-shadow: 0 4px 16px rgba(0,0,0,0.2);
        }
        .user-section { 
            margin-bottom: 40px; 
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 16px rgba(0,0,0,0.1);
        }
        .user-header { 
            background: linear-gradient(135deg, #007bff, #0056b3);
            color: white;
            padding: 20px;
            transition: all 0.3s ease;
            cursor: pointer;
        }
        .user-header:hover {
            background: linear-gradient(135deg, #0056b3, #004085);
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(0,0,0,0.2);
        }
        .user-header h2 {
            color: white;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
        }
        .user-header .user-icon {
            margin-right: 10px;
            font-size: 1.5em;
        }
        .collapse-icon {
            transition: transform 0.3s ease;
            font-size: 0.8em;
        }
        .user-header[aria-expanded="false"] .collapse-icon {
            transform: rotate(-90deg);
        }
        .stats-container {
            background: white;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 30px;
            box-shadow: 0 4px 16px rgba(0,0,0,0.1);
        }
        .stat-card {
            text-align: center;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 15px;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
        }
        .stat-label {
            font-size: 0.9em;
            opacity: 0.8;
        }
        .filter-section { 
            background: white;
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 30px;
            box-shadow: 0 4px 16px rgba(0,0,0,0.1);
        }
        .hidden { display: none !important; }
        .table-container {
            background: white;
            overflow: hidden;
        }
        .table th {
            background: linear-gradient(135deg, #6c757d, #495057);
            color: white;
            border: none;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85em;
            letter-spacing: 0.5px;
            padding: 14px 16px;
        }
        .table td {
            border-color: #e9ecef;
            vertical-align: middle;
            padding: 12px 16px;
        }
        .status-badge {
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: 600;
            text-transform: uppercase;
        }
        .status-success {
            background-color: #d1e7dd;
            color: #0f5132;
        }
        .status-failed {
            background-color: #f8d7da;
            color: #721c24;
        }
        .progress-bar {
            height: 6px;
            border-radius: 3px;
            margin-top: 5px;
        }
        
        /* Modal Styles */
        .modal { 
            display: none; 
            position: fixed; 
            z-index: 1050; 
            padding-top: 0;
            left: 0; 
            top: 0; 
            width: 100%; 
            height: 100%; 
            overflow: auto; 
            background-color: rgba(0,0,0,0.9); 
        }
        .modal-content { 
            margin: auto; 
            display: block; 
            width: 90%; 
            max-width: 1400px; 
            max-height: 90vh; 
            object-fit: contain; 
            border-radius: 12px;
        }
        .close { 
            position: absolute; 
            top: 20px; 
            right: 35px; 
            color: #fff; 
            font-size: 48px; 
            font-weight: 300; 
            transition: all 0.3s ease; 
            cursor: pointer;
            width: 60px;
            height: 60px;
            border-radius: 50%;
            background: rgba(0,0,0,0.5);
            display: flex;
            align-items: center;
            justify-content: center;
            backdrop-filter: blur(10px);
        }
        .close:hover, .close:focus { 
            background: rgba(220, 53, 69, 0.8);
            transform: scale(1.1);
        }
        .modal-caption { 
            margin: auto; 
            display: block; 
            width: 80%; 
            max-width: 700px; 
            text-align: center; 
            color: #fff; 
            padding: 20px 0; 
            font-size: 1.1em;
            background: rgba(0,0,0,0.7);
            border-radius: 8px;
            margin-top: 20px;
            backdrop-filter: blur(10px);
        }

        /* Enhanced Carousel Styles */
        .carousel-container {
            position: relative;
            width: 100%;
            max-width: 1400px;
            margin: 50px auto 0;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        
        .carousel-nav {
            position: absolute;
            top: 50%;
            transform: translateY(-50%);
            background: rgba(0,0,0,0.6);
            color: white;
            border: none;
            font-size: 28px;
            padding: 15px 20px;
            cursor: pointer;
            z-index: 1051;
            border-radius: 50%;
            transition: all 0.3s ease;
            width: 60px;
            height: 60px;
            display: flex;
            align-items: center;
            justify-content: center;
            backdrop-filter: blur(10px);
        }
        
        .carousel-nav:hover {
            background: rgba(0, 123, 255, 0.8);
            transform: translateY(-50%) scale(1.1);
        }
        
        .carousel-prev { left: 30px; }
        .carousel-next { right: 30px; }
        
        .carousel-indicators {
            position: fixed;
            bottom: 20px;
            left: 50%;
            transform: translateX(-50%);
            display: flex;
            gap: 12px;
            z-index: 1051;
            background: rgba(0,0,0,0.5);
            padding: 10px 20px;
            border-radius: 25px;
            backdrop-filter: blur(10px);
        }
        
        .carousel-indicator {
            width: 14px;
            height: 14px;
            border-radius: 50%;
            background: rgba(255,255,255,0.5);
            cursor: pointer;
            transition: all 0.3s ease;
            border: 2px solid transparent;
        }
        
        .carousel-indicator:hover {
            background: rgba(255,255,255,0.8);
            transform: scale(1.2);
        }
        
        .carousel-indicator.active {
            background: #007bff;
            border-color: white;
            transform: scale(1.3);
        }
        
        .carousel-counter {
            position: absolute;
            top: 30px;
            right: 30px;
            background: rgba(0,0,0,0.7);
            color: white;
            padding: 8px 16px;
            border-radius: 25px;
            font-size: 16px;
            font-weight: 600;
            z-index: 1051;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
        }
    </style>
</head>
<body>
    <div class="main-container">
        <div class="py-4">`)

	// Header
	html.WriteString(fmt.Sprintf(`
            <div class="text-center mb-5">
                <h1 class="display-4 fw-bold text-primary">
                    <i class="bi bi-person-lines-fill me-3"></i>
                    User Journey Report - %s
                </h1>
                <p class="lead text-muted">Interactive user journey analysis with filtering and carousel viewing</p>
            </div>`, automation.Name))

	// Summary section
	html.WriteString(`
            <div class="stats-container">
                <div class="d-flex align-items-center mb-4">
                    <i class="bi bi-graph-up text-primary fs-4 me-2"></i>
                    <h3 class="mb-0">Test Summary</h3>
                </div>
                <div class="row g-4">
                    <div class="col-6 col-md-3">
                        <div class="stat-card bg-primary bg-opacity-10 border border-primary border-opacity-25">
                            <i class="bi bi-people-fill text-primary fs-3"></i>
                            <div class="stat-number text-primary" id="totalUsers">0</div>
                            <div class="stat-label text-muted">Users</div>
                        </div>
                    </div>
                    <div class="col-6 col-md-3">
                        <div class="stat-card bg-info bg-opacity-10 border border-info border-opacity-25">
                            <i class="bi bi-list-check text-info fs-3"></i>
                            <div class="stat-number text-info" id="totalTests">0</div>
                            <div class="stat-label text-muted">Total Tests</div>
                        </div>
                    </div>
                    <div class="col-6 col-md-3">
                        <div class="stat-card bg-success bg-opacity-10 border border-success border-opacity-25">
                            <i class="bi bi-check-circle-fill text-success fs-3"></i>
                            <div class="stat-number text-success" id="passedTests">0</div>
                            <div class="stat-label text-muted">Passed</div>
                            <div class="progress mt-2">
                                <div class="progress-bar progress-bar-success bg-success" role="progressbar" style="width: 0%" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100"></div>
                            </div>
                        </div>
                    </div>
                    <div class="col-6 col-md-3">
                        <div class="stat-card bg-danger bg-opacity-10 border border-danger border-opacity-25">
                            <i class="bi bi-x-circle-fill text-danger fs-3"></i>
                            <div class="stat-number text-danger" id="failedTests">0</div>
                            <div class="stat-label text-muted">Failed</div>
                            <div class="progress mt-2">
                                <div class="progress-bar progress-bar-danger bg-danger" role="progressbar" style="width: 0%" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100"></div>
                            </div>
                        </div>
                    </div>`)

	// Add total time if available
	if totalTime != nil {
		html.WriteString(fmt.Sprintf(`
                    <div class="col-12">
                        <div class="stat-card bg-warning bg-opacity-10 border border-warning border-opacity-25">
                            <i class="bi bi-stopwatch text-warning fs-3"></i>
                            <div class="stat-number text-warning">%ss</div>
                            <div class="stat-label text-muted">Total Execution Time (%sm)</div>
                        </div>
                    </div>`, totalTime["seconds"], totalTime["minutes"]))
	}

	html.WriteString(`
                </div>
            </div>`)

	// Filters section
	html.WriteString(`
            <div class="filter-section">
                <div class="d-flex align-items-center mb-3">
                    <i class="bi bi-funnel-fill me-2 text-primary fs-4"></i>
                    <h3 class="mb-0">Filters</h3>
                </div>
                <div class="row g-3">
                    <div class="col-lg-2 col-md-3">
                        <label for="userFilter" class="form-label fw-bold">
                            <i class="bi bi-person-fill me-1"></i>User ID
                        </label>
                        <select id="userFilter" class="form-select" onchange="filterResults()">
                            <option value="">All Users</option>`)

	// Add user options
	for userId := range groupedReports {
		html.WriteString(fmt.Sprintf(`<option value="%d">User %d</option>`, userId, userId))
	}

	html.WriteString(`
                        </select>
                    </div>
                    
                    <div class="col-lg-2 col-md-3">
                        <label for="featureFilter" class="form-label fw-bold">
                            <i class="bi bi-layers-fill me-1"></i>Feature
                        </label>
                        <select id="featureFilter" class="form-select" onchange="filterResults()">
                            <option value="">All Features</option>`)

	// Add feature options
	for _, feature := range uniqueFeatures {
		html.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, feature, feature))
	}

	html.WriteString(`
                        </select>
                    </div>
                    
                    <div class="col-lg-2 col-md-3">
                        <label for="statusFilter" class="form-label fw-bold">
                            <i class="bi bi-check-circle-fill me-1"></i>Status
                        </label>
                        <select id="statusFilter" class="form-select" onchange="filterResults()">
                            <option value="">All Statuses</option>
                            <option value="success">Success</option>
                            <option value="failed">Failed</option>
                        </select>
                    </div>
                    
                    <div class="col-lg-3 col-md-3 d-flex align-items-end">
                        <button onclick="clearFilters()" class="btn btn-outline-secondary w-100">
                            <i class="bi bi-arrow-clockwise me-1"></i>Clear Filters
                        </button>
                    </div>
                    
                    <div class="col-lg-3 col-md-12">
                        <label class="form-label fw-bold">
                            <i class="bi bi-layout-sidebar me-1"></i>View Controls
                        </label>
                        <div class="btn-group w-100" role="group">
                            <button onclick="collapseAll()" class="btn btn-outline-primary">
                                <i class="bi bi-chevron-up me-1"></i>Collapse All
                            </button>
                            <button onclick="expandAll()" class="btn btn-outline-primary">
                                <i class="bi bi-chevron-down me-1"></i>Expand All
                            </button>
                        </div>
                    </div>
                </div>
            </div>`)

	// User sections
	userIndex := 0
	for userId, userReports := range groupedReports {
		successCount := 0
		failedCount := 0
		for _, report := range userReports {
			if status, ok := report["status"].(string); ok {
				if status == "success" {
					successCount++
				} else if status == "failed" {
					failedCount++
				}
			}
		}

		successRate := 0
		if len(userReports) > 0 {
			successRate = int(float64(successCount) / float64(len(userReports)) * 100)
		}

		isExpanded := userIndex == 0
		collapseClass := ""
		ariaExpanded := "false"
		if isExpanded {
			collapseClass = "show"
			ariaExpanded = "true"
		}

		html.WriteString(fmt.Sprintf(`
            <div class="user-section" data-user-id="%d">
                <div class="user-header" 
                     data-bs-toggle="collapse" 
                     data-bs-target="#userCollapse%d" 
                     aria-expanded="%s" 
                     aria-controls="userCollapse%d">
                    <div class="d-flex justify-content-between align-items-center">
                        <div>
                            <h2 class="mb-2">
                                <i class="bi bi-person-circle user-icon"></i>
                                User %d
                                <i class="bi bi-chevron-down ms-2 collapse-icon"></i>
                            </h2>
                            <div class="d-flex flex-wrap gap-3">
                                <span class="badge bg-light text-dark fs-6">
                                    <i class="bi bi-list-check me-1"></i>
                                    %d Tests
                                </span>
                                <span class="badge bg-success fs-6">
                                    <i class="bi bi-check-circle me-1"></i>
                                    %d Passed
                                </span>
                                <span class="badge bg-danger fs-6">
                                    <i class="bi bi-x-circle me-1"></i>
                                    %d Failed
                                </span>
                                <span class="badge bg-info fs-6">
                                    <i class="bi bi-percent me-1"></i>
                                    %d%% Success Rate
                                </span>
                            </div>
                        </div>
                        <div class="text-end">
                            <div class="progress" style="width: 120px; height: 8px;">
                                <div class="progress-bar bg-success" role="progressbar" 
                                     style="width: %d%%" 
                                     aria-valuenow="%d" 
                                     aria-valuemin="0" 
                                     aria-valuemax="100">
                                </div>
                            </div>
                            <small class="text-light mt-1 d-block">%d%% Success</small>
                        </div>
                    </div>
                </div>
                
                <div class="collapse %s" id="userCollapse%d">
                    <div class="table-container">
                        <div class="table-responsive">
                            <table class="table table-hover mb-0">
                                <thead>
                                    <tr>
                                        <th><i class="bi bi-layers me-1"></i>Feature</th>
                                        <th><i class="bi bi-play-circle me-1"></i>Scenario</th>
                                        <th><i class="bi bi-check-circle me-1"></i>Status</th>
                                        <th><i class="bi bi-camera me-1"></i>Screenshot</th>
                                        <th><i class="bi bi-exclamation-triangle me-1"></i>Error</th>
                                        <th class="d-none d-lg-table-cell"><i class="bi bi-clock me-1"></i>Timestamp</th>
                                    </tr>
                                </thead>
                                <tbody>`, userId, userId, ariaExpanded, userId, userId, len(userReports), successCount, failedCount, successRate, successRate, successRate, successRate, collapseClass, userId))

		for _, report := range userReports {
			feature := ""
			scenario := ""
			status := ""
			screenshot := ""
			errorMsg := ""
			timestamp := ""

			if f, ok := report["feature"].(string); ok {
				feature = f
			}
			if s, ok := report["scenario"].(string); ok {
				scenario = s
			}
			if st, ok := report["status"].(string); ok {
				status = st
			}
			if sc, ok := report["screenshot"].(string); ok {
				screenshot = sc
			}
			if e, ok := report["error"].(string); ok {
				errorMsg = e
			}
			if t, ok := report["timestamp"].(string); ok {
				timestamp = t
			}

			statusBadgeClass := "status-success"
			statusIcon := "check-circle"
			if status == "failed" {
				statusBadgeClass = "status-failed"
				statusIcon = "x-circle"
			}

			html.WriteString(fmt.Sprintf(`
                                    <tr class="test-row" data-feature="%s" data-status="%s">
                                        <td>
                                            <span class="badge bg-secondary">%s</span>
                                        </td>
                                        <td>%s</td>
                                        <td>
                                            <span class="status-badge %s">
                                                <i class="bi bi-%s me-1"></i>
                                                %s
                                            </span>
                                        </td>
                                        <td>`, feature, status, feature, scenario, statusBadgeClass, statusIcon, status))

			if screenshot != "" {
				html.WriteString(fmt.Sprintf(`
                                            <img src="%s" 
                                                 alt="Screenshot" 
                                                 class="screenshot-img" 
                                                 onclick="openCarousel('%s', %d)"
                                                 data-bs-toggle="tooltip" 
                                                 title="Click to view user %d screenshots" />`, screenshot, screenshot, userId, userId))
			} else {
				html.WriteString(`<span class="text-muted"><i class="bi bi-image-alt"></i> N/A</span>`)
			}

			html.WriteString(`</td><td style="max-width: 300px;">`)

			if errorMsg != "" {
				html.WriteString(fmt.Sprintf(`
                                            <div class="text-danger small">
                                                <i class="bi bi-exclamation-triangle me-1"></i>
                                                <span style="word-wrap: break-word;">%s</span>
                                            </div>`, errorMsg))
			} else {
				html.WriteString(`
                                            <span class="text-success">
                                                <i class="bi bi-check-circle me-1"></i>
                                                No errors
                                            </span>`)
			}

			html.WriteString(`</td><td class="d-none d-lg-table-cell">`)

			if timestamp != "" {
				if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
					html.WriteString(fmt.Sprintf(`
                                            <small class="text-muted">
                                                <i class="bi bi-calendar3 me-1"></i>
                                                %s
                                            </small>`, parsedTime.Format("2006-01-02 15:04:05")))
				} else {
					html.WriteString(`<span class="text-muted">N/A</span>`)
				}
			} else {
				html.WriteString(`<span class="text-muted">N/A</span>`)
			}

			html.WriteString(`</td></tr>`)
		}

		html.WriteString(`
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>`)

		userIndex++
	}

	// Enhanced Carousel Modal
	html.WriteString(`
        </div>
    </div>
    
    <!-- Enhanced Carousel Modal -->
    <div id="imageModal" class="modal">
        <span class="close" onclick="closeModal()" data-bs-toggle="tooltip" data-bs-placement="left" title="Close (ESC)">
            <i class="bi bi-x-lg"></i>
        </span>
        <div class="carousel-container">
            <img class="modal-content" id="modalImage" alt="Test Screenshot">
            <button class="carousel-nav carousel-prev" id="carouselPrev" data-bs-toggle="tooltip" data-bs-placement="right" title="Previous (←)">
                <i class="bi bi-chevron-left"></i>
            </button>
            <button class="carousel-nav carousel-next" id="carouselNext" data-bs-toggle="tooltip" data-bs-placement="left" title="Next (→)">
                <i class="bi bi-chevron-right"></i>
            </button>
            <div class="carousel-counter" id="carouselCounter">
                <i class="bi bi-images me-1"></i>
                1 / 1
            </div>
            <div class="carousel-indicators">
                <!-- Indicators will be generated dynamically -->
            </div>
        </div>
        <div id="caption" class="modal-caption"></div>
    </div>
    
    <!-- Toast Container -->
    <div class="toast-container position-fixed bottom-0 end-0 p-3" id="toastContainer"></div>`)

	// JavaScript
	html.WriteString(`
    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
    <script>
        // Store all images for carousel
        const allImages = [`)

	// Generate allImages array
	imageIndex := 0
	for userId, userReports := range groupedReports {
		for _, report := range userReports {
			if screenshot, ok := report["screenshot"].(string); ok && screenshot != "" {
				if imageIndex > 0 {
					html.WriteString(", ")
				}
				feature := ""
				scenario := ""
				if f, ok := report["feature"].(string); ok {
					feature = f
				}
				if s, ok := report["scenario"].(string); ok {
					scenario = s
				}
				html.WriteString(fmt.Sprintf(`{
                    "src": "%s",
                    "caption": "User %d - %s - %s",
                    "userId": %d,
                    "index": %d
                }`, screenshot, userId, feature, scenario, userId, imageIndex))
				imageIndex++
			}
		}
	}

	html.WriteString(`];
        let currentImageIndex = 0;
        let currentUserImages = [];
        
        function openCarousel(imageSrc, userId = null) {
            // Filter images by user if userId is provided
            if (userId) {
                currentUserImages = allImages.filter(img => img.userId == userId);
            } else {
                currentUserImages = allImages;
            }
            
            // Find the index of the clicked image within the filtered images
            currentImageIndex = currentUserImages.findIndex(img => img.src === imageSrc);
            if (currentImageIndex === -1) currentImageIndex = 0;
            
            const modal = $('#imageModal');
            modal.show();
            updateCarouselImage();
            updateCarouselIndicators();
            updateNavigationVisibility();
            
            // Add keyboard navigation
            $(document).on('keydown.carousel', function(e) {
                if (e.key === 'ArrowLeft') {
                    previousImage();
                } else if (e.key === 'ArrowRight') {
                    nextImage();
                } else if (e.key === 'Escape') {
                    closeModal();
                }
            });
        }
        
        function updateNavigationVisibility() {
            const hasMultipleImages = currentUserImages.length > 1;
            $('.carousel-nav').toggle(hasMultipleImages);
            $('.carousel-indicators').toggle(hasMultipleImages);
        }
        
        function updateCarouselImage() {
            const currentImage = currentUserImages[currentImageIndex];
            if (currentImage) {
                $('#modalImage').attr('src', currentImage.src);
                $('#caption').text(currentImage.caption);
                $('#carouselCounter').text((currentImageIndex + 1) + ' / ' + currentUserImages.length);
                
                // Update indicators
                $('.carousel-indicator').removeClass('active');
                $('.carousel-indicator').eq(currentImageIndex).addClass('active');
            }
        }
        
        function updateCarouselIndicators() {
            const indicatorsContainer = $('.carousel-indicators');
            indicatorsContainer.empty();
            
            currentUserImages.forEach((img, index) => {
                const indicator = $('<div class="carousel-indicator" data-index="' + index + '"></div>');
                if (index === currentImageIndex) {
                    indicator.addClass('active');
                }
                indicatorsContainer.append(indicator);
            });
        }
        
        function nextImage() {
            if (currentUserImages.length > 0) {
                currentImageIndex = (currentImageIndex + 1) % currentUserImages.length;
                updateCarouselImage();
            }
        }
        
        function previousImage() {
            if (currentUserImages.length > 0) {
                currentImageIndex = currentImageIndex === 0 ? currentUserImages.length - 1 : currentImageIndex - 1;
                updateCarouselImage();
            }
        }
        
        function goToImage(index) {
            if (index >= 0 && index < currentUserImages.length) {
                currentImageIndex = index;
                updateCarouselImage();
            }
        }
        
        function closeModal() {
            $('#imageModal').hide();
            $(document).off('keydown.carousel');
        }
        
        // Close modal when clicking outside the image
        $(window).on('click', function(event) {
            if (event.target.id === 'imageModal') {
                closeModal();
            }
        });
        
        function filterResults() {
            const featureFilter = $('#featureFilter').val();
            const statusFilter = $('#statusFilter').val();
            const userFilter = $('#userFilter').val();
            
            const userSections = $('.user-section');
            let visibleUsers = 0;
            let totalTests = 0;
            let passedTests = 0;
            let failedTests = 0;
            
            userSections.each(function() {
                const userId = $(this).data('user-id').toString();
                const rows = $(this).find('.test-row');
                let visibleRows = 0;
                
                rows.each(function() {
                    const feature = $(this).data('feature');
                    const status = $(this).data('status');
                    
                    let showRow = true;
                    
                    if (featureFilter && feature !== featureFilter) showRow = false;
                    if (statusFilter && status !== statusFilter) showRow = false;
                    if (userFilter && userId !== userFilter) showRow = false;
                    
                    if (showRow) {
                        $(this).removeClass('hidden');
                        visibleRows++;
                        totalTests++;
                        if (status === 'success') passedTests++;
                        if (status === 'failed') failedTests++;
                    } else {
                        $(this).addClass('hidden');
                    }
                });
                
                if (visibleRows > 0 && (!userFilter || userId === userFilter)) {
                    $(this).removeClass('hidden');
                    visibleUsers++;
                } else {
                    $(this).addClass('hidden');
                }
            });
            
            // Update summary with animations
            updateStatCard('#totalUsers', visibleUsers);
            updateStatCard('#totalTests', totalTests);
            updateStatCard('#passedTests', passedTests);
            updateStatCard('#failedTests', failedTests);
            
            // Update progress bars
            updateProgressBars(passedTests, failedTests, totalTests);
        }
        
        function updateStatCard(selector, value) {
            const element = $(selector);
            element.fadeOut(200, function() {
                element.text(value).fadeIn(200);
            });
        }
        
        function updateProgressBars(passed, failed, total) {
            if (total > 0) {
                const passPercent = (passed / total) * 100;
                const failPercent = (failed / total) * 100;
                
                $('.progress-bar-success').css('width', passPercent + '%').attr('aria-valuenow', passPercent);
                $('.progress-bar-danger').css('width', failPercent + '%').attr('aria-valuenow', failPercent);
            }
        }
        
        function clearFilters() {
            $('#featureFilter').val('');
            $('#statusFilter').val('');
            $('#userFilter').val('');
            filterResults();
            
            // Show success toast
            showToast('Filters cleared successfully!', 'success');
        }
        
        function collapseAll() {
            const collapses = document.querySelectorAll('.collapse.show');
            collapses.forEach(collapse => {
                const bsCollapse = new bootstrap.Collapse(collapse, {toggle: false});
                bsCollapse.hide();
            });
            showToast('All sections collapsed', 'info');
        }
        
        function expandAll() {
            const collapses = document.querySelectorAll('.collapse:not(.show)');
            collapses.forEach(collapse => {
                const bsCollapse = new bootstrap.Collapse(collapse, {toggle: false});
                bsCollapse.show();
            });
            showToast('All sections expanded', 'info');
        }
        
        function showToast(message, type = 'info') {
            const toastHtml = \`
                <div class="toast align-items-center text-white bg-\${type} border-0" role="alert" aria-live="assertive" aria-atomic="true">
                    <div class="d-flex">
                        <div class="toast-body">\${message}</div>
                        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
                    </div>
                </div>
            \`;
            
            const toastContainer = $('#toastContainer');
            const toastElement = $(toastHtml);
            toastContainer.append(toastElement);
            
            const toast = new bootstrap.Toast(toastElement[0]);
            toast.show();
            
            // Remove toast after it's hidden
            toastElement.on('hidden.bs.toast', function() {
                $(this).remove();
            });
        }
        
        $(document).ready(function() {
            filterResults();
            
            // Initialize carousel navigation
            $('#carouselPrev').on('click', previousImage);
            $('#carouselNext').on('click', nextImage);
            
            // Initialize carousel indicators
            $(document).on('click', '.carousel-indicator', function() {
                const index = parseInt($(this).data('index'));
                goToImage(index);
            });
            
            // Initialize tooltips
            $('[data-bs-toggle="tooltip"]').tooltip();
        });
    </script>
</body>
</html>`)

	return html.String()
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
	Name           string
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

			actionDisplayName := action.Type
			if action.Name != "" {
				actionDisplayName = fmt.Sprintf("%s (%s)", action.Name, action.Type)
			}

			html.WriteString(fmt.Sprintf(`
                <div class="%s">
                    <div class="action-title">%s%s</div>
                    <div class="action-meta">
                        User %d | <span class="status-badge %s">%s</span> | Duration: %dms
                    </div>`, actionClass, actionDisplayName, parentInfo, action.LoopIndex, actionStatusClass, strings.ToUpper(action.Status), action.Duration))

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