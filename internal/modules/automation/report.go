package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/storage"
)

// GenerateReports generates and uploads all report types to object storage
func GenerateReports(automation *Automation, run *AutomationRun, config *AutomationConfig, outputBaseDir string, reportsR2Path string, storageService storage.StorageService) (detailedReportURL, userJourneyReportURL string, err error) {
	// Parse logs and output files
	var logs []map[string]any
	if run.LogsJSON != "" {
		if err := json.Unmarshal([]byte(run.LogsJSON), &logs); err != nil {
			return "", "", fmt.Errorf("failed to parse logs JSON: %w", err)
		}
	}

	var outputFiles []string
	if run.OutputFilesJSON != "" {
		if err := json.Unmarshal([]byte(run.OutputFilesJSON), &outputFiles); err != nil {
			return "", "", fmt.Errorf("failed to parse output files JSON: %w", err)
		}
	}

	// Generate JSON report
	jsonReportURL, err := generateJSONReport(automation, run, logs, outputFiles, reportsR2Path, storageService)
	if err != nil {
		slog.Error("Failed to generate JSON report", "error", err)
	}

	// Generate CSV report
	csvReportURL, err := generateCSVReport(automation, run, logs, outputFiles, reportsR2Path, storageService)
	if err != nil {
		slog.Error("Failed to generate CSV report", "error", err)
	}

	// Generate detailed HTML report
	detailedReportURL, err = generateDetailedHTMLReport(automation, run, config, logs, outputFiles, reportsR2Path, storageService)
	if err != nil {
		slog.Error("Failed to generate detailed HTML report", "error", err)
		return "", "", fmt.Errorf("failed to generate detailed HTML report: %w", err)
	}

	// Generate user journey HTML report
	userJourneyReportURL, err = generateUserJourneyHTMLReport(automation, run, logs, outputFiles, reportsR2Path, storageService)
	if err != nil {
		slog.Error("Failed to generate user journey HTML report", "error", err)
		return "", "", fmt.Errorf("failed to generate user journey HTML report: %w", err)
	}

	slog.Info("All reports generated successfully",
		"automation_id", automation.ID,
		"run_id", run.ID,
		"json_report", jsonReportURL,
		"csv_report", csvReportURL,
		"detailed_report", detailedReportURL,
		"user_journey_report", userJourneyReportURL)

	return detailedReportURL, userJourneyReportURL, nil
}

// generateJSONReport generates and uploads a JSON report
func generateJSONReport(automation *Automation, run *AutomationRun, logs []map[string]any, outputFiles []string, reportsR2Path string, storageService storage.StorageService) (string, error) {
	reportData := map[string]interface{}{
		"run": map[string]interface{}{
			"id":            run.ID,
			"status":        run.Status,
			"start_time":    run.StartTime,
			"end_time":      run.EndTime,
			"error_message": run.ErrorMessage,
		},
		"automation": map[string]interface{}{
			"id":          automation.ID,
			"name":        automation.Name,
			"description": automation.Description,
		},
		"project": map[string]interface{}{
			"id":   automation.ProjectID,
			"name": automation.ProjectName,
		},
		"logs":         logs,
		"output_files": outputFiles,
		"summary": map[string]interface{}{
			"total_logs":         len(logs),
			"total_output_files": len(outputFiles),
			"duration_ms":        calculateDuration(run.StartTime, run.EndTime),
		},
	}

	jsonBytes, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	reader := bytes.NewReader(jsonBytes)
	publicURL, err := storageService.UploadFile(context.Background(), fmt.Sprintf("%s/report.json", reportsR2Path), reader, "application/json")
	if err != nil {
		return "", fmt.Errorf("failed to upload JSON report: %w", err)
	}

	return publicURL, nil
}

// generateCSVReport generates and uploads a CSV report
func generateCSVReport(automation *Automation, run *AutomationRun, logs []map[string]any, outputFiles []string, reportsR2Path string, storageService storage.StorageService) (string, error) {
	var csvRows []string

	// CSV Headers
	csvRows = append(csvRows, "Timestamp,Step Name,Step ID,Action ID,Action Name,Action Type,Message,Error,Duration (ms),Loop Index,Local Loop Index,Status,Output File")

	// Process logs
	for _, log := range logs {
		row := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%v,%v,%v,%s,%s",
			escapeCSV(getString(log, "timestamp")),
			escapeCSV(getString(log, "step_name")),
			escapeCSV(getString(log, "step_id")),
			escapeCSV(getString(log, "action_id")),
			escapeCSV(getString(log, "action_name")),
			escapeCSV(getString(log, "action_type")),
			escapeCSV(getString(log, "message")),
			escapeCSV(getString(log, "error")),
			getNumber(log, "duration_ms"),
			getNumber(log, "loop_index"),
			getNumber(log, "local_loop_index"),
			escapeCSV(getString(log, "status")),
			escapeCSV(getString(log, "output_file")),
		)
		csvRows = append(csvRows, row)
	}

	csvContent := strings.Join(csvRows, "\n")
	reader := bytes.NewReader([]byte(csvContent))
	publicURL, err := storageService.UploadFile(context.Background(), fmt.Sprintf("%s/logs.csv", reportsR2Path), reader, "text/csv")
	if err != nil {
		return "", fmt.Errorf("failed to upload CSV report: %w", err)
	}

	return publicURL, nil
}

// generateDetailedHTMLReport generates a Bootstrap-based HTML report similar to the current web UI
func generateDetailedHTMLReport(automation *Automation, run *AutomationRun, config *AutomationConfig, logs []map[string]any, outputFiles []string, reportsR2Path string, storageService storage.StorageService) (string, error) {
	// Process data similar to the web UI
	reportData := processReportData(logs, outputFiles)
	performanceMetrics := calculatePerformanceMetrics(logs)

	// Generate HTML content
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Detailed Report - %s</title>
    
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    
    <!-- Chart.js -->
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    
    <style>
        body {
            background-color: #f8f9fa;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        .main-container {
            background-color: white;
            min-height: 100vh;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .step-section {
            border: 1px solid #dee2e6;
            border-radius: 8px;
            margin-bottom: 1rem;
        }
        .step-header {
            background-color: #f8f9fa;
            padding: 1rem;
            border-bottom: 1px solid #dee2e6;
            cursor: pointer;
        }
        .step-content {
            padding: 1rem;
        }
        .action-item {
            border-left: 3px solid #007bff;
            padding: 0.5rem 1rem;
            margin-bottom: 0.5rem;
            background-color: #f8f9fa;
        }
        .status-success { color: #198754; }
        .status-failed { color: #dc3545; }
        .screenshot-img {
            max-width: 150px;
            height: auto;
            cursor: pointer;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .screenshot-img:hover {
            transform: scale(1.05);
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }
        
        /* Modal Styles */
        .modal {
            display: none;
            position: fixed;
            z-index: 1050;
            left: 0;
            top: 0;
            width: 100%%;
            height: 100%%;
            overflow: auto;
            background-color: rgba(0,0,0,0.9);
        }
        .modal-content {
            margin: auto;
            display: block;
            width: 90%%;
            max-width: 1200px;
            max-height: 90vh;
            object-fit: contain;
        }
        .close {
            position: absolute;
            top: 15px;
            right: 35px;
            color: #fff;
            font-size: 40px;
            font-weight: bold;
            cursor: pointer;
        }
        .close:hover {
            color: #ccc;
        }
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
                <p class="lead text-muted">%s - Run %s</p>
                <p class="text-muted">Generated: %s</p>
            </div>

            <!-- Summary Cards -->
            <div class="row mb-4">
                <div class="col-md-3">
                    <div class="card text-center">
                        <div class="card-body">
                            <i class="bi bi-list-check text-primary fs-1"></i>
                            <h5 class="card-title">Total Steps</h5>
                            <h3 class="text-primary">%d</h3>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card text-center">
                        <div class="card-body">
                            <i class="bi bi-play-circle text-info fs-1"></i>
                            <h5 class="card-title">Total Actions</h5>
                            <h3 class="text-info">%d</h3>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card text-center">
                        <div class="card-body">
                            <i class="bi bi-clock text-warning fs-1"></i>
                            <h5 class="card-title">Duration</h5>
                            <h3 class="text-warning">%s</h3>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card text-center">
                        <div class="card-body">
                            <i class="bi bi-file-earmark text-success fs-1"></i>
                            <h5 class="card-title">Output Files</h5>
                            <h3 class="text-success">%d</h3>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Performance Charts -->
            %s

            <!-- Step Details -->
            <div class="card">
                <div class="card-header">
                    <h3 class="mb-0">
                        <i class="bi bi-list-ol me-2"></i>
                        Step-by-Step Report
                    </h3>
                </div>
                <div class="card-body">
                    %s
                </div>
            </div>
        </div>
    </div>

    <!-- Image Modal -->
    <div id="imageModal" class="modal">
        <span class="close" onclick="closeModal()">&times;</span>
        <img class="modal-content" id="modalImage">
        <div id="caption" style="text-align: center; color: white; padding: 20px;"></div>
    </div>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        function openImage(src, caption) {
            document.getElementById('imageModal').style.display = 'block';
            document.getElementById('modalImage').src = src;
            document.getElementById('caption').innerHTML = caption;
        }
        
        function closeModal() {
            document.getElementById('imageModal').style.display = 'none';
        }
        
        // Close modal when clicking outside the image
        window.onclick = function(event) {
            const modal = document.getElementById('imageModal');
            if (event.target == modal) {
                closeModal();
            }
        }
        
        // Keyboard navigation
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                closeModal();
            }
        });

        %s
    </script>
</body>
</html>`,
		automation.Name,
		automation.Name, run.ID[:8],
		time.Now().Format("January 2, 2006 at 3:04 PM"),
		len(reportData.Steps),
		reportData.TotalActions,
		formatDuration(calculateDuration(run.StartTime, run.EndTime)),
		len(outputFiles),
		generateChartsSection(performanceMetrics),
		generateStepDetailsSection(reportData),
		generateChartsJavaScript(performanceMetrics),
	)

	// Upload to storage
	reader := bytes.NewReader([]byte(htmlContent))
	publicURL, err := storageService.UploadFile(context.Background(), fmt.Sprintf("%s/detailed_report.html", reportsR2Path), reader, "text/html")
	if err != nil {
		return "", fmt.Errorf("failed to upload detailed HTML report: %w", err)
	}

	return publicURL, nil
}

// generateUserJourneyHTMLReport generates an interactive user journey HTML report
func generateUserJourneyHTMLReport(automation *Automation, run *AutomationRun, logs []map[string]any, outputFiles []string, reportsR2Path string, storageService storage.StorageService) (string, error) {
	// Process data for user journey report
	groupedReports := groupReportsByUser(logs, outputFiles)
	uniqueFeatures := getUniqueFeatures(logs)
	allScreenshots := getAllScreenshots(outputFiles)
	userTimings := calculateUserTimings(logs)

	// Generate HTML content based on the provided snippet
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Journey Report - %s</title>
    
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" rel="stylesheet">
    
    %s
</head>
<body>
    <div class="main-container">
        <div class="py-4">
            <div class="text-center mb-5">
                <h1 class="display-4 fw-bold text-primary">
                    <i class="bi bi-people-fill me-3"></i>
                    User Journey Report
                </h1>
                <p class="lead text-muted">%s - Run %s</p>
                <p class="text-muted">Generated: %s</p>
            </div>
            
            %s
            %s
            %s
        </div>
    </div>
    
    %s
    
    <!-- Toast Container -->
    <div class="toast-container position-fixed bottom-0 end-0 p-3" id="toastContainer"></div>
    
    %s
</body>
</html>`,
		automation.Name,
		generateUserJourneyStyles(),
		automation.Name, run.ID[:8],
		time.Now().Format("January 2, 2006 at 3:04 PM"),
		generateSummarySection(groupedReports, userTimings),
		generateFiltersSection(uniqueFeatures, groupedReports),
		generateUserSections(groupedReports),
		generateCarouselModal(),
		generateUserJourneyScripts(allScreenshots, groupedReports),
	)

	// Upload to storage
	reader := bytes.NewReader([]byte(htmlContent))
	publicURL, err := storageService.UploadFile(context.Background(), fmt.Sprintf("%s/user_journey_report.html", reportsR2Path), reader, "text/html")
	if err != nil {
		return "", fmt.Errorf("failed to upload user journey HTML report: %w", err)
	}

	return publicURL, nil
}

// Helper types and functions for report generation
type ReportData struct {
	Steps        []StepData `json:"steps"`
	TotalActions int        `json:"total_actions"`
}

type StepData struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Status   string       `json:"status"`
	Duration int64        `json:"duration"`
	Actions  []ActionData `json:"actions"`
}

type ActionData struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Duration    int64    `json:"duration"`
	Error       string   `json:"error,omitempty"`
	OutputFiles []string `json:"output_files"`
}

type PerformanceMetrics struct {
	TotalRuns          int               `json:"total_runs"`
	OverallFailureRate float64           `json:"overall_failure_rate"`
	StepAverages       []StepPerformance `json:"step_averages"`
	RunData            []RunPerformance  `json:"run_data"`
}

type StepPerformance struct {
	Name            string  `json:"name"`
	AverageDuration float64 `json:"average_duration"`
	FailureRate     float64 `json:"failure_rate"`
	TotalRuns       int     `json:"total_runs"`
}

type RunPerformance struct {
	LoopIndex     int                   `json:"loop_index"`
	Steps         map[string]StepMetric `json:"steps"`
	TotalDuration int64                 `json:"total_duration"`
	Status        string                `json:"status"`
}

type StepMetric struct {
	Duration int64  `json:"duration"`
	Status   string `json:"status"`
}

type UserReport struct {
	UserID     int    `json:"userId"`
	Email      string `json:"email"`
	Feature    string `json:"feature"`
	Scenario   string `json:"scenario"`
	Status     string `json:"status"`
	Screenshot string `json:"screenshot"`
	Error      string `json:"error"`
	URL        string `json:"url"`
	Timestamp  string `json:"timestamp"`
}

type UserTiming struct {
	UserID      int    `json:"userId"`
	TimeSeconds string `json:"timeSeconds"`
	TimeMinutes string `json:"timeMinutes"`
}

// processReportData processes logs and output files into structured report data
func processReportData(logs []map[string]any, outputFiles []string) ReportData {
	stepMap := make(map[string]*StepData)
	totalActions := 0

	for _, log := range logs {
		stepID := getString(log, "step_id")
		stepName := getString(log, "step_name")
		actionID := getString(log, "action_id")
		actionType := getString(log, "action_type")
		actionName := getString(log, "action_name")
		status := getString(log, "status")
		duration := getNumber(log, "duration_ms")
		errorMsg := getString(log, "error")
		outputFile := getString(log, "output_file")

		if stepID == "" {
			continue
		}

		// Initialize step if not exists
		if _, exists := stepMap[stepID]; !exists {
			stepMap[stepID] = &StepData{
				ID:       stepID,
				Name:     stepName,
				Status:   "success",
				Duration: 0,
				Actions:  []ActionData{},
			}
		}

		step := stepMap[stepID]
		step.Duration += int64(duration)

		if status == "failed" {
			step.Status = "failed"
		}

		// Add action if actionID exists
		if actionID != "" {
			var actionFiles []string
			if outputFile != "" {
				actionFiles = append(actionFiles, outputFile)
			}

			action := ActionData{
				ID:          actionID,
				Name:        actionName,
				Type:        actionType,
				Status:      status,
				Duration:    int64(duration),
				Error:       errorMsg,
				OutputFiles: actionFiles,
			}

			step.Actions = append(step.Actions, action)
			totalActions++
		}
	}

	// Convert map to slice and sort by step name
	var steps []StepData
	for _, step := range stepMap {
		steps = append(steps, *step)
	}

	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Name < steps[j].Name
	})

	return ReportData{
		Steps:        steps,
		TotalActions: totalActions,
	}
}

// calculatePerformanceMetrics calculates performance metrics from logs
func calculatePerformanceMetrics(logs []map[string]any) PerformanceMetrics {
	stepMetrics := make(map[string]*StepPerformance)
	runMetrics := make(map[int]*RunPerformance)

	for _, log := range logs {
		stepName := getString(log, "step_name")
		loopIndex := int(getNumber(log, "loop_index"))
		duration := getNumber(log, "duration_ms")
		status := getString(log, "status")

		// Step-level metrics
		if stepName != "" {
			if _, exists := stepMetrics[stepName]; !exists {
				stepMetrics[stepName] = &StepPerformance{
					Name:        stepName,
					TotalRuns:   0,
					FailureRate: 0,
				}
			}

			stepMetric := stepMetrics[stepName]
			stepMetric.TotalRuns++
			stepMetric.AverageDuration = (stepMetric.AverageDuration*float64(stepMetric.TotalRuns-1) + duration) / float64(stepMetric.TotalRuns)

			if status == "failed" {
				stepMetric.FailureRate = (stepMetric.FailureRate*float64(stepMetric.TotalRuns-1) + 1) / float64(stepMetric.TotalRuns)
			}
		}

		// Run-level metrics
		if _, exists := runMetrics[loopIndex]; !exists {
			runMetrics[loopIndex] = &RunPerformance{
				LoopIndex:     loopIndex,
				Steps:         make(map[string]StepMetric),
				TotalDuration: 0,
				Status:        "success",
			}
		}

		runMetric := runMetrics[loopIndex]
		if stepName != "" {
			if _, exists := runMetric.Steps[stepName]; !exists {
				runMetric.Steps[stepName] = StepMetric{Duration: 0, Status: "success"}
			}

			stepMetric := runMetric.Steps[stepName]
			stepMetric.Duration += int64(duration)
			if status == "failed" {
				stepMetric.Status = "failed"
				runMetric.Status = "failed"
			}
			runMetric.Steps[stepName] = stepMetric
		}

		runMetric.TotalDuration += int64(duration)
	}

	// Convert maps to slices
	var stepAverages []StepPerformance
	for _, step := range stepMetrics {
		stepAverages = append(stepAverages, *step)
	}

	var runData []RunPerformance
	for _, run := range runMetrics {
		runData = append(runData, *run)
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

// groupReportsByUser groups logs by user (loop index) for user journey report
func groupReportsByUser(logs []map[string]any, outputFiles []string) map[int][]UserReport {
	grouped := make(map[int][]UserReport)

	for _, log := range logs {
		loopIndex := int(getNumber(log, "loop_index"))
		stepName := getString(log, "step_name")
		actionType := getString(log, "action_type")
		status := getString(log, "status")
		errorMsg := getString(log, "error")
		outputFile := getString(log, "output_file")
		timestamp := getString(log, "timestamp")

		if stepName == "" {
			continue
		}

		userReport := UserReport{
			UserID:     loopIndex,
			Email:      fmt.Sprintf("user%d@test.com", loopIndex),
			Feature:    stepName,
			Scenario:   actionType,
			Status:     status,
			Screenshot: outputFile,
			Error:      errorMsg,
			URL:        "", // Could be extracted from logs if available
			Timestamp:  timestamp,
		}

		grouped[loopIndex] = append(grouped[loopIndex], userReport)
	}

	return grouped
}

// getUniqueFeatures extracts unique features (step names) from logs
func getUniqueFeatures(logs []map[string]any) []string {
	featureSet := make(map[string]bool)
	for _, log := range logs {
		stepName := getString(log, "step_name")
		if stepName != "" {
			featureSet[stepName] = true
		}
	}

	var features []string
	for feature := range featureSet {
		features = append(features, feature)
	}

	sort.Strings(features)
	return features
}

// getAllScreenshots extracts all screenshot URLs from output files
func getAllScreenshots(outputFiles []string) []string {
	var screenshots []string
	for _, file := range outputFiles {
		if isImageFile(file) {
			screenshots = append(screenshots, file)
		}
	}
	return screenshots
}

// calculateUserTimings calculates timing information for each user
func calculateUserTimings(logs []map[string]any) []UserTiming {
	userDurations := make(map[int]int64)

	for _, log := range logs {
		loopIndex := int(getNumber(log, "loop_index"))
		duration := int64(getNumber(log, "duration_ms"))
		userDurations[loopIndex] += duration
	}

	var timings []UserTiming
	for userID, totalDuration := range userDurations {
		seconds := totalDuration / 1000
		minutes := float64(seconds) / 60

		timings = append(timings, UserTiming{
			UserID:      userID,
			TimeSeconds: fmt.Sprintf("%.2f", float64(totalDuration)/1000),
			TimeMinutes: fmt.Sprintf("%.2f", minutes),
		})
	}

	sort.Slice(timings, func(i, j int) bool {
		return timings[i].UserID < timings[j].UserID
	})

	return timings
}

// Helper functions for HTML generation
func generateChartsSection(metrics PerformanceMetrics) string {
	if metrics.TotalRuns <= 1 {
		return ""
	}

	return `
    <div class="card mb-4">
        <div class="card-header">
            <h3 class="mb-0">
                <i class="bi bi-bar-chart me-2"></i>
                Performance Analysis
            </h3>
        </div>
        <div class="card-body">
            <div class="row mb-4">
                <div class="col-md-3">
                    <div class="card text-center bg-primary text-white">
                        <div class="card-body">
                            <h4>` + strconv.Itoa(metrics.TotalRuns) + `</h4>
                            <p class="mb-0">Total Runs</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card text-center bg-success text-white">
                        <div class="card-body">
                            <h4>` + fmt.Sprintf("%.1f%%", 100-metrics.OverallFailureRate) + `</h4>
                            <p class="mb-0">Success Rate</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card text-center bg-danger text-white">
                        <div class="card-body">
                            <h4>` + fmt.Sprintf("%.1f%%", metrics.OverallFailureRate) + `</h4>
                            <p class="mb-0">Failure Rate</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card text-center bg-info text-white">
                        <div class="card-body">
                            <h4>` + strconv.Itoa(len(metrics.StepAverages)) + `</h4>
                            <p class="mb-0">Steps</p>
                        </div>
                    </div>
                </div>
            </div>
            <div style="height: 400px;">
                <canvas id="performanceChart"></canvas>
            </div>
        </div>
    </div>`
}

func generateStepDetailsSection(reportData ReportData) string {
	var stepsHTML strings.Builder

	for i, step := range reportData.Steps {
		statusClass := "text-success"
		statusIcon := "check-circle"
		if step.Status == "failed" {
			statusClass = "text-danger"
			statusIcon = "x-circle"
		}

		stepsHTML.WriteString(fmt.Sprintf(`
        <div class="step-section">
            <div class="step-header" data-bs-toggle="collapse" data-bs-target="#step%d" aria-expanded="%s">
                <div class="d-flex justify-content-between align-items-center">
                    <h5 class="mb-0">
                        <i class="bi bi-%s %s me-2"></i>
                        %s
                    </h5>
                    <div class="d-flex align-items-center">
                        <span class="badge bg-secondary me-2">%d actions</span>
                        <span class="text-muted">%s</span>
                        <i class="bi bi-chevron-down ms-2"></i>
                    </div>
                </div>
            </div>
            <div class="collapse %s" id="step%d">
                <div class="step-content">`,
			i, func() string {
				if i == 0 {
					return "true"
				} else {
					return "false"
				}
			}(),
			statusIcon, statusClass,
			step.Name,
			len(step.Actions),
			formatDuration(step.Duration),
			func() string {
				if i == 0 {
					return "show"
				} else {
					return ""
				}
			}(),
			i,
		))

		// Add actions
		for _, action := range step.Actions {
			actionStatusClass := "text-success"
			actionStatusIcon := "check-circle"
			if action.Status == "failed" {
				actionStatusClass = "text-danger"
				actionStatusIcon = "x-circle"
			}

			stepsHTML.WriteString(fmt.Sprintf(`
                    <div class="action-item">
                        <div class="d-flex justify-content-between align-items-start">
                            <div>
                                <h6 class="mb-1">
                                    <i class="bi bi-%s %s me-1"></i>
                                    %s
                                </h6>
                                <small class="text-muted">%s</small>`,
				actionStatusIcon, actionStatusClass,
				func() string {
					if action.Name != "" {
						return action.Name
					} else {
						return action.Type
					}
				}(),
				action.Type,
			))

			if action.Error != "" {
				stepsHTML.WriteString(fmt.Sprintf(`
                                <div class="text-danger mt-1">
                                    <i class="bi bi-exclamation-triangle me-1"></i>
                                    %s
                                </div>`, action.Error))
			}

			stepsHTML.WriteString(`
                            </div>
                            <div class="text-end">`)

			// Add screenshots
			for _, file := range action.OutputFiles {
				if isImageFile(file) {
					stepsHTML.WriteString(fmt.Sprintf(`
                                <img src="%s" alt="Screenshot" class="screenshot-img me-1" 
                                     onclick="openImage('%s', '%s - %s')">`,
						file, file, step.Name, action.Type))
				}
			}

			stepsHTML.WriteString(fmt.Sprintf(`
                                <div class="text-muted small">%s</div>
                            </div>
                        </div>
                    </div>`, formatDuration(action.Duration)))
		}

		stepsHTML.WriteString(`
                </div>
            </div>
        </div>`)
	}

	return stepsHTML.String()
}

func generateChartsJavaScript(metrics PerformanceMetrics) string {
	if metrics.TotalRuns <= 1 {
		return ""
	}

	// Prepare data for Chart.js
	stepNames := make([]string, len(metrics.StepAverages))
	stepDurations := make([]float64, len(metrics.StepAverages))
	stepFailureRates := make([]float64, len(metrics.StepAverages))

	for i, step := range metrics.StepAverages {
		stepNames[i] = step.Name
		stepDurations[i] = step.AverageDuration
		stepFailureRates[i] = step.FailureRate
	}

	stepNamesJSON, _ := json.Marshal(stepNames)
	stepDurationsJSON, _ := json.Marshal(stepDurations)
	stepFailureRatesJSON, _ := json.Marshal(stepFailureRates)

	return fmt.Sprintf(`
        // Initialize performance chart
        const ctx = document.getElementById('performanceChart').getContext('2d');
        new Chart(ctx, {
            type: 'bar',
            data: {
                labels: %s,
                datasets: [{
                    label: 'Average Duration (ms)',
                    data: %s,
                    backgroundColor: 'rgba(54, 162, 235, 0.6)',
                    borderColor: 'rgba(54, 162, 235, 1)',
                    borderWidth: 1,
                    yAxisID: 'y'
                }, {
                    label: 'Failure Rate (%%)',
                    data: %s,
                    backgroundColor: 'rgba(255, 99, 132, 0.6)',
                    borderColor: 'rgba(255, 99, 132, 1)',
                    borderWidth: 1,
                    type: 'line',
                    yAxisID: 'y1'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    title: {
                        display: true,
                        text: 'Step Performance Overview'
                    },
                    legend: {
                        display: true
                    }
                },
                scales: {
                    y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                        title: {
                            display: true,
                            text: 'Duration (ms)'
                        }
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                            display: true,
                            text: 'Failure Rate (%%)'
                        },
                        grid: {
                            drawOnChartArea: false,
                        },
                    }
                }
            }
        });`,
		string(stepNamesJSON),
		string(stepDurationsJSON),
		string(stepFailureRatesJSON),
	)
}

// User Journey Report HTML Generation Functions
func generateUserJourneyStyles() string {
	return `<style>
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
    </style>`
}

func generateSummarySection(groupedReports map[int][]UserReport, userTimings []UserTiming) string {
	totalUsers := len(groupedReports)
	totalTests := 0
	passedTests := 0
	failedTests := 0

	for _, reports := range groupedReports {
		for _, report := range reports {
			totalTests++
			if report.Status == "success" {
				passedTests++
			} else if report.Status == "failed" {
				failedTests++
			}
		}
	}

	userTimingsTable := ""
	if len(userTimings) > 0 {
		userTimingsTable = `
        <div class="col-12 mt-4">
            <div class="card">
                <div class="card-header">
                    <h5 class="mb-0">
                        <i class="bi bi-person-check me-2"></i>
                        Individual User Test Times
                    </h5>
                </div>
                <div class="card-body">
                    <div class="table-responsive">
                        <table class="table table-striped table-hover">
                            <thead class="table-dark">
                                <tr>
                                    <th>User ID</th>
                                    <th>Execution Time (seconds)</th>
                                    <th>Execution Time (minutes)</th>
                                </tr>
                            </thead>
                            <tbody>`

		for _, timing := range userTimings {
			userTimingsTable += fmt.Sprintf(`
                                <tr>
                                    <td><strong>User %d</strong></td>
                                    <td><span class="badge bg-info">%ss</span></td>
                                    <td><span class="badge bg-secondary">%sm</span></td>
                                </tr>`,
				timing.UserID, timing.TimeSeconds, timing.TimeMinutes)
		}

		userTimingsTable += `
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>`
	}

	return fmt.Sprintf(`
    <div class="stats-container">
        <div class="d-flex align-items-center mb-4">
            <i class="bi bi-graph-up text-primary fs-4 me-2"></i>
            <h3 class="mb-0">Test Summary</h3>
        </div>
        <div class="row g-4">
            <div class="col-6 col-md-3">
                <div class="stat-card bg-primary bg-opacity-10 border border-primary border-opacity-25">
                    <i class="bi bi-people-fill text-primary fs-3"></i>
                    <div class="stat-number text-primary" id="totalUsers">%d</div>
                    <div class="stat-label text-muted">Users</div>
                </div>
            </div>
            <div class="col-6 col-md-3">
                <div class="stat-card bg-info bg-opacity-10 border border-info border-opacity-25">
                    <i class="bi bi-list-check text-info fs-3"></i>
                    <div class="stat-number text-info" id="totalTests">%d</div>
                    <div class="stat-label text-muted">Total Tests</div>
                </div>
            </div>
            <div class="col-6 col-md-3">
                <div class="stat-card bg-success bg-opacity-10 border border-success border-opacity-25">
                    <i class="bi bi-check-circle-fill text-success fs-3"></i>
                    <div class="stat-number text-success" id="passedTests">%d</div>
                    <div class="stat-label text-muted">Passed</div>
                    <div class="progress mt-2">
                        <div class="progress-bar progress-bar-success bg-success" role="progressbar" style="width: %d%%" aria-valuenow="%d" aria-valuemin="0" aria-valuemax="100"></div>
                    </div>
                </div>
            </div>
            <div class="col-6 col-md-3">
                <div class="stat-card bg-danger bg-opacity-10 border border-danger border-opacity-25">
                    <i class="bi bi-x-circle-fill text-danger fs-3"></i>
                    <div class="stat-number text-danger" id="failedTests">%d</div>
                    <div class="stat-label text-muted">Failed</div>
                    <div class="progress mt-2">
                        <div class="progress-bar progress-bar-danger bg-danger" role="progressbar" style="width: %d%%" aria-valuenow="%d" aria-valuemin="0" aria-valuemax="100"></div>
                    </div>
                </div>
            </div>
            %s
        </div>
    </div>`,
		totalUsers, totalTests, passedTests,
		func() int {
			if totalTests > 0 {
				return (passedTests * 100) / totalTests
			} else {
				return 0
			}
		}(),
		func() int {
			if totalTests > 0 {
				return (passedTests * 100) / totalTests
			} else {
				return 0
			}
		}(),
		failedTests,
		func() int {
			if totalTests > 0 {
				return (failedTests * 100) / totalTests
			} else {
				return 0
			}
		}(),
		func() int {
			if totalTests > 0 {
				return (failedTests * 100) / totalTests
			} else {
				return 0
			}
		}(),
		userTimingsTable,
	)
}

func generateFiltersSection(uniqueFeatures []string, groupedReports map[int][]UserReport) string {
	featureOptions := ""
	for _, feature := range uniqueFeatures {
		featureOptions += fmt.Sprintf(`<option value="%s">%s</option>`, feature, feature)
	}

	userOptions := ""
	for userID := range groupedReports {
		userOptions += fmt.Sprintf(`<option value="%d">User %d</option>`, userID, userID)
	}

	return fmt.Sprintf(`
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
                    <option value="">All Users</option>
                    %s
                </select>
            </div>
            
            <div class="col-lg-2 col-md-3">
                <label for="featureFilter" class="form-label fw-bold">
                    <i class="bi bi-layers-fill me-1"></i>Feature
                </label>
                <select id="featureFilter" class="form-select" onchange="filterResults()">
                    <option value="">All Features</option>
                    %s
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
    </div>`, userOptions, featureOptions)
}

func generateUserSections(groupedReports map[int][]UserReport) string {
	var sectionsHTML strings.Builder
	index := 0

	// Sort user IDs for consistent output
	var userIDs []int
	for userID := range groupedReports {
		userIDs = append(userIDs, userID)
	}
	sort.Ints(userIDs)

	for _, userID := range userIDs {
		userReports := groupedReports[userID]
		successCount := 0
		failedCount := 0

		for _, report := range userReports {
			if report.Status == "success" {
				successCount++
			} else if report.Status == "failed" {
				failedCount++
			}
		}

		successRate := 0
		if len(userReports) > 0 {
			successRate = (successCount * 100) / len(userReports)
		}

		isExpanded := index == 0
		collapseClass := ""
		ariaExpanded := "false"
		if isExpanded {
			collapseClass = "show"
			ariaExpanded = "true"
		}

		sectionsHTML.WriteString(fmt.Sprintf(`
        <div class="user-section" data-user-id="%d">
            <div class="user-header" 
                 data-bs-toggle="collapse" 
                 data-bs-target="#userCollapse%d" 
                 aria-expanded="%s" 
                 aria-controls="userCollapse%d"
                 style="cursor: pointer;">
                <div class="d-flex justify-content-between align-items-center">
                    <div>
                        <h2 class="mb-2">
                            <i class="bi bi-person-circle user-icon"></i>
                            User %d - %s
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
                            <tbody>`,
			userID, userID, ariaExpanded, userID,
			userID, userReports[0].Email,
			len(userReports), successCount, failedCount, successRate,
			successRate, successRate, successRate,
			collapseClass, userID,
		))

		// Add test rows
		for _, report := range userReports {
			statusClass := "status-success"
			statusIcon := "check-circle"
			if report.Status == "failed" {
				statusClass = "status-failed"
				statusIcon = "x-circle"
			}

			screenshotHTML := `<span class="text-muted"><i class="bi bi-image-alt"></i> N/A</span>`
			if report.Screenshot != "" && isImageFile(report.Screenshot) {
				screenshotHTML = fmt.Sprintf(`<img src="%s" alt="Screenshot" class="screenshot-img" onclick="openCarousel('%s', %d)" data-bs-toggle="tooltip" title="Click to view user %d screenshots" />`,
					report.Screenshot, report.Screenshot, userID, userID)
			}

			errorHTML := `<span class="text-success"><i class="bi bi-check-circle me-1"></i>No errors</span>`
			if report.Error != "" {
				errorHTML = fmt.Sprintf(`<div class="text-danger small"><i class="bi bi-exclamation-triangle me-1"></i><span style="word-wrap: break-word;">%s</span></div>`, report.Error)
			}

			timestampHTML := `<span class="text-muted">N/A</span>`
			if report.Timestamp != "" {
				if parsedTime, err := time.Parse(time.RFC3339, report.Timestamp); err == nil {
					timestampHTML = fmt.Sprintf(`<small class="text-muted"><i class="bi bi-calendar3 me-1"></i>%s</small>`, parsedTime.Format("Jan 2, 2006 3:04 PM"))
				}
			}

			sectionsHTML.WriteString(fmt.Sprintf(`
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
                                    <td>%s</td>
                                    <td style="max-width: 300px;">%s</td>
                                    <td class="d-none d-lg-table-cell">%s</td>
                                </tr>`,
				report.Feature, report.Status,
				func() string {
					if report.Feature != "" {
						return report.Feature
					} else {
						return "N/A"
					}
				}(),
				func() string {
					if report.Scenario != "" {
						return report.Scenario
					} else {
						return "N/A"
					}
				}(),
				statusClass, statusIcon, report.Status,
				screenshotHTML, errorHTML, timestampHTML,
			))
		}

		sectionsHTML.WriteString(`
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>`)

		index++
	}

	return sectionsHTML.String()
}

func generateCarouselModal() string {
	return `
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
    </div>`
}

func generateUserJourneyScripts(allScreenshots []string, groupedReports map[int][]UserReport) string {
	// Prepare all images data for JavaScript
	var allImages []map[string]interface{}
	for userID, reports := range groupedReports {
		for i, report := range reports {
			if report.Screenshot != "" && isImageFile(report.Screenshot) {
				allImages = append(allImages, map[string]interface{}{
					"src":     report.Screenshot,
					"caption": fmt.Sprintf("User %d - %s - %s", userID, report.Feature, report.Scenario),
					"userId":  userID,
					"index":   i,
				})
			}
		}
	}

	allImagesJSON, _ := json.Marshal(allImages)

	return fmt.Sprintf(`
    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
    <script>
        // Store all images for carousel
        const allImages = %s;
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
                currentImageIndex = (currentImageIndex + 1) %% currentUserImages.length;
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
                
                $('.progress-bar-success').css('width', passPercent + '%%').attr('aria-valuenow', passPercent);
                $('.progress-bar-danger').css('width', failPercent + '%%').attr('aria-valuenow', failPercent);
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
            const toastHtml = `+"`"+`
        <div class="toast align-items-center text-white bg-${type} border-0" role="alert" aria-live="assertive" aria-atomic="true">
            <div class="d-flex">
                <div class="toast-body">${message}</div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        </div>
    `+"`"+`;
            
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
    </script>`, string(allImagesJSON))
}

// Utility helper functions
func getString(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getNumber(m map[string]any, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func calculateDuration(startTime, endTime *time.Time) int64 {
	if startTime == nil || endTime == nil {
		return 0
	}
	return endTime.Sub(*startTime).Milliseconds()
}

func formatDuration(durationMs int64) string {
	if durationMs < 1000 {
		return fmt.Sprintf("%dms", durationMs)
	} else if durationMs < 60000 {
		return fmt.Sprintf("%.2fs", float64(durationMs)/1000)
	} else {
		minutes := durationMs / 60000
		seconds := (durationMs % 60000) / 1000
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	return ext == "png" || ext == "jpg" || ext == "jpeg" || ext == "gif" || ext == "webp"
}

func escapeCSV(value string) string {
	if strings.Contains(value, ",") || strings.Contains(value, "\"") || strings.Contains(value, "\n") {
		value = strings.ReplaceAll(value, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", value)
	}
	return value
}
