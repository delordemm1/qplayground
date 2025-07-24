package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ReportData represents the complete report structure
type ReportData struct {
	Summary       RunSummary              `json:"summary"`
	Steps         []StepSummary           `json:"steps"`
	Actions       []ActionSummary         `json:"actions"`
	OutputFiles   []string                `json:"output_files"`
	RawLogs       []map[string]interface{} `json:"raw_logs"`
	GeneratedAt   time.Time               `json:"generated_at"`
	ExecutionTime time.Duration           `json:"execution_time"`
	Error         string                  `json:"error,omitempty"`
}

// RunSummary represents overall run statistics
type RunSummary struct {
	AutomationName    string        `json:"automation_name"`
	AutomationDesc    string        `json:"automation_description"`
	TotalSteps        int           `json:"total_steps"`
	TotalActions      int           `json:"total_actions"`
	TotalRuns         int           `json:"total_runs"`
	SuccessfulRuns    int           `json:"successful_runs"`
	FailedRuns        int           `json:"failed_runs"`
	TotalDuration     time.Duration `json:"total_duration"`
	AverageDuration   time.Duration `json:"average_duration"`
	OutputFilesCount  int           `json:"output_files_count"`
	SuccessRate       float64       `json:"success_rate"`
	Status            string        `json:"status"`
}

// StepSummary represents statistics for a specific step
type StepSummary struct {
	Name            string        `json:"name"`
	Order           int           `json:"order"`
	TotalExecutions int           `json:"total_executions"`
	SuccessCount    int           `json:"success_count"`
	FailureCount    int           `json:"failure_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	SuccessRate     float64       `json:"success_rate"`
	ActionsCount    int           `json:"actions_count"`
}

// ActionSummary represents statistics for a specific action
type ActionSummary struct {
	ID              string        `json:"id"`
	Type            string        `json:"type"`
	StepName        string        `json:"step_name"`
	TotalExecutions int           `json:"total_executions"`
	SuccessCount    int           `json:"success_count"`
	FailureCount    int           `json:"failure_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	SuccessRate     float64       `json:"success_rate"`
	LastError       string        `json:"last_error,omitempty"`
}

// Reporter handles generation of various report formats
type Reporter struct {
	reportsDir       string
	config           *ExportedAutomationConfig
	logs             []map[string]interface{}
	outputFiles      []string
	timestamp        string
	executionError   error
	reportData       *ReportData
}

// NewReporter creates a new Reporter instance
func NewReporter(reportsDir string, config *ExportedAutomationConfig, logs []map[string]interface{}, outputFiles []string, timestamp string, executionError error) *Reporter {
	reporter := &Reporter{
		reportsDir:     reportsDir,
		config:         config,
		logs:           logs,
		outputFiles:    outputFiles,
		timestamp:      timestamp,
		executionError: executionError,
	}
	
	reporter.analyzeData()
	return reporter
}

// analyzeData processes the raw logs and generates summary statistics
func (r *Reporter) analyzeData() {
	stepStats := make(map[string]*StepSummary)
	actionStats := make(map[string]*ActionSummary)
	runStats := make(map[int]bool) // track success/failure by loop index
	
	var totalDuration time.Duration
	var earliestTime, latestTime time.Time
	
	// Process each log entry
	for _, log := range r.logs {
		stepName, _ := log["step_name"].(string)
		actionType, _ := log["action_type"].(string)
		actionID, _ := log["action_id"].(string)
		loopIndex, _ := log["loop_index"].(int)
		durationMs, _ := log["duration_ms"].(float64)
		status, _ := log["status"].(string)
		errorMsg, _ := log["error"].(string)
		timestampStr, _ := log["timestamp"].(string)
		
		duration := time.Duration(durationMs) * time.Millisecond
		totalDuration += duration
		
		// Parse timestamp
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			if earliestTime.IsZero() || timestamp.Before(earliestTime) {
				earliestTime = timestamp
			}
			if latestTime.IsZero() || timestamp.After(latestTime) {
				latestTime = timestamp
			}
		}
		
		// Track run success/failure
		if status == "failed" {
			runStats[loopIndex] = false
		} else if _, exists := runStats[loopIndex]; !exists {
			runStats[loopIndex] = true
		}
		
		// Step statistics
		if stepName != "" {
			if _, exists := stepStats[stepName]; !exists {
				stepStats[stepName] = &StepSummary{
					Name:        stepName,
					MinDuration: duration,
					MaxDuration: duration,
				}
			}
			
			step := stepStats[stepName]
			step.TotalExecutions++
			step.TotalDuration += duration
			
			if status == "failed" {
				step.FailureCount++
			} else {
				step.SuccessCount++
			}
			
			if duration < step.MinDuration {
				step.MinDuration = duration
			}
			if duration > step.MaxDuration {
				step.MaxDuration = duration
			}
		}
		
		// Action statistics
		if actionID != "" && actionType != "" {
			if _, exists := actionStats[actionID]; !exists {
				actionStats[actionID] = &ActionSummary{
					ID:          actionID,
					Type:        actionType,
					StepName:    stepName,
					MinDuration: duration,
					MaxDuration: duration,
				}
			}
			
			action := actionStats[actionID]
			action.TotalExecutions++
			action.TotalDuration += duration
			
			if status == "failed" {
				action.FailureCount++
				action.LastError = errorMsg
			} else {
				action.SuccessCount++
			}
			
			if duration < action.MinDuration {
				action.MinDuration = duration
			}
			if duration > action.MaxDuration {
				action.MaxDuration = duration
			}
		}
	}
	
	// Calculate averages and success rates
	for _, step := range stepStats {
		if step.TotalExecutions > 0 {
			step.AverageDuration = step.TotalDuration / time.Duration(step.TotalExecutions)
			step.SuccessRate = float64(step.SuccessCount) / float64(step.TotalExecutions) * 100
		}
	}
	
	for _, action := range actionStats {
		if action.TotalExecutions > 0 {
			action.AverageDuration = action.TotalDuration / time.Duration(action.TotalExecutions)
			action.SuccessRate = float64(action.SuccessCount) / float64(action.TotalExecutions) * 100
		}
	}
	
	// Convert maps to slices and sort
	steps := make([]StepSummary, 0, len(stepStats))
	for _, step := range stepStats {
		steps = append(steps, *step)
	}
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Name < steps[j].Name
	})
	
	actions := make([]ActionSummary, 0, len(actionStats))
	for _, action := range actionStats {
		actions = append(actions, *action)
	}
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].StepName == actions[j].StepName {
			return actions[i].Type < actions[j].Type
		}
		return actions[i].StepName < actions[j].StepName
	})
	
	// Calculate overall statistics
	totalRuns := len(runStats)
	successfulRuns := 0
	for _, success := range runStats {
		if success {
			successfulRuns++
		}
	}
	
	var executionTime time.Duration
	if !earliestTime.IsZero() && !latestTime.IsZero() {
		executionTime = latestTime.Sub(earliestTime)
	}
	
	var averageDuration time.Duration
	if totalRuns > 0 {
		averageDuration = totalDuration / time.Duration(totalRuns)
	}
	
	var successRate float64
	if totalRuns > 0 {
		successRate = float64(successfulRuns) / float64(totalRuns) * 100
	}
	
	status := "completed"
	if r.executionError != nil {
		status = "failed"
	}
	
	errorMsg := ""
	if r.executionError != nil {
		errorMsg = r.executionError.Error()
	}
	
	// Build report data
	r.reportData = &ReportData{
		Summary: RunSummary{
			AutomationName:   r.config.Automation.Name,
			AutomationDesc:   r.config.Automation.Description,
			TotalSteps:       len(r.config.Steps),
			TotalActions:     r.countTotalActions(),
			TotalRuns:        totalRuns,
			SuccessfulRuns:   successfulRuns,
			FailedRuns:       totalRuns - successfulRuns,
			TotalDuration:    totalDuration,
			AverageDuration:  averageDuration,
			OutputFilesCount: len(r.outputFiles),
			SuccessRate:      successRate,
			Status:           status,
		},
		Steps:         steps,
		Actions:       actions,
		OutputFiles:   r.outputFiles,
		RawLogs:       r.logs,
		GeneratedAt:   time.Now(),
		ExecutionTime: executionTime,
		Error:         errorMsg,
	}
}

// countTotalActions counts all actions including nested ones
func (r *Reporter) countTotalActions() int {
	total := 0
	for _, step := range r.config.Steps {
		total += len(step.Actions)
		// Count nested actions
		for _, action := range step.Actions {
			total += r.countNestedActions(action.ActionConfig)
		}
	}
	return total
}

// countNestedActions recursively counts nested actions
func (r *Reporter) countNestedActions(config map[string]interface{}) int {
	count := 0
	
	// Check for if_else actions
	if ifActions, ok := config["if_actions"].([]interface{}); ok {
		count += len(ifActions)
		for _, action := range ifActions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				if actionConfig, ok := actionMap["action_config"].(map[string]interface{}); ok {
					count += r.countNestedActions(actionConfig)
				}
			}
		}
	}
	
	if elseActions, ok := config["else_actions"].([]interface{}); ok {
		count += len(elseActions)
		for _, action := range elseActions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				if actionConfig, ok := actionMap["action_config"].(map[string]interface{}); ok {
					count += r.countNestedActions(actionConfig)
				}
			}
		}
	}
	
	if finalActions, ok := config["final_actions"].([]interface{}); ok {
		count += len(finalActions)
		for _, action := range finalActions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				if actionConfig, ok := actionMap["action_config"].(map[string]interface{}); ok {
					count += r.countNestedActions(actionConfig)
				}
			}
		}
	}
	
	if loopActions, ok := config["loop_actions"].([]interface{}); ok {
		count += len(loopActions)
		for _, action := range loopActions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				if actionConfig, ok := actionMap["action_config"].(map[string]interface{}); ok {
					count += r.countNestedActions(actionConfig)
				}
			}
		}
	}
	
	if elseIfConditions, ok := config["else_if_conditions"].([]interface{}); ok {
		for _, condition := range elseIfConditions {
			if conditionMap, ok := condition.(map[string]interface{}); ok {
				if actions, ok := conditionMap["actions"].([]interface{}); ok {
					count += len(actions)
					for _, action := range actions {
						if actionMap, ok := action.(map[string]interface{}); ok {
							if actionConfig, ok := actionMap["action_config"].(map[string]interface{}); ok {
								count += r.countNestedActions(actionConfig)
							}
						}
					}
				}
			}
		}
	}
	
	return count
}

// GenerateCSVReport creates a CSV report
func (r *Reporter) GenerateCSVReport() error {
	csvFile := filepath.Join(r.reportsDir, "automation_report.csv")
	file, err := os.Create(csvFile)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write headers
	headers := []string{
		"Timestamp", "Step Name", "Step ID", "Action Type", "Action ID", 
		"Loop Index", "Duration (ms)", "Status", "Error", "Output File",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}
	
	// Write log entries
	for _, log := range r.logs {
		record := []string{
			getString(log, "timestamp"),
			getString(log, "step_name"),
			getString(log, "step_id"),
			getString(log, "action_type"),
			getString(log, "action_id"),
			fmt.Sprintf("%v", log["loop_index"]),
			fmt.Sprintf("%.0f", getFloat(log, "duration_ms")),
			getString(log, "status"),
			getString(log, "error"),
			getString(log, "output_file"),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}
	
	return nil
}

// GenerateJSONReport creates a JSON report
func (r *Reporter) GenerateJSONReport() error {
	jsonFile := filepath.Join(r.reportsDir, "automation_report.json")
	file, err := os.Create(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.reportData); err != nil {
		return fmt.Errorf("failed to encode JSON report: %w", err)
	}
	
	return nil
}

// GenerateHTMLReport creates an HTML report
func (r *Reporter) GenerateHTMLReport() error {
	htmlFile := filepath.Join(r.reportsDir, "automation_report.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()
	
	// Generate HTML content
	htmlContent := r.generateHTMLContent()
	
	if _, err := file.WriteString(htmlContent); err != nil {
		return fmt.Errorf("failed to write HTML content: %w", err)
	}
	
	return nil
}

// generateHTMLContent creates the HTML report content
func (r *Reporter) generateHTMLContent() string {
	var html strings.Builder
	
	// HTML header
	html.WriteString(htmlTemplate)
	
	// Replace placeholders with actual data
	content := html.String()
	content = strings.ReplaceAll(content, "{{AUTOMATION_NAME}}", r.config.Automation.Name)
	content = strings.ReplaceAll(content, "{{AUTOMATION_DESCRIPTION}}", r.config.Automation.Description)
	content = strings.ReplaceAll(content, "{{GENERATED_AT}}", r.reportData.GeneratedAt.Format("January 2, 2006 at 3:04 PM"))
	content = strings.ReplaceAll(content, "{{TIMESTAMP}}", r.timestamp)
	content = strings.ReplaceAll(content, "{{STATUS}}", r.reportData.Summary.Status)
	content = strings.ReplaceAll(content, "{{STATUS_CLASS}}", r.getStatusClass(r.reportData.Summary.Status))
	content = strings.ReplaceAll(content, "{{TOTAL_RUNS}}", fmt.Sprintf("%d", r.reportData.Summary.TotalRuns))
	content = strings.ReplaceAll(content, "{{SUCCESSFUL_RUNS}}", fmt.Sprintf("%d", r.reportData.Summary.SuccessfulRuns))
	content = strings.ReplaceAll(content, "{{FAILED_RUNS}}", fmt.Sprintf("%d", r.reportData.Summary.FailedRuns))
	content = strings.ReplaceAll(content, "{{SUCCESS_RATE}}", fmt.Sprintf("%.1f%%", r.reportData.Summary.SuccessRate))
	content = strings.ReplaceAll(content, "{{TOTAL_STEPS}}", fmt.Sprintf("%d", r.reportData.Summary.TotalSteps))
	content = strings.ReplaceAll(content, "{{TOTAL_ACTIONS}}", fmt.Sprintf("%d", r.reportData.Summary.TotalActions))
	content = strings.ReplaceAll(content, "{{TOTAL_DURATION}}", r.formatDuration(r.reportData.Summary.TotalDuration))
	content = strings.ReplaceAll(content, "{{AVERAGE_DURATION}}", r.formatDuration(r.reportData.Summary.AverageDuration))
	content = strings.ReplaceAll(content, "{{OUTPUT_FILES_COUNT}}", fmt.Sprintf("%d", r.reportData.Summary.OutputFilesCount))
	content = strings.ReplaceAll(content, "{{EXECUTION_TIME}}", r.formatDuration(r.reportData.ExecutionTime))
	
	// Error section
	errorSection := ""
	if r.reportData.Error != "" {
		errorSection = fmt.Sprintf(`
		<div class="error-section">
			<h3>❌ Execution Error</h3>
			<div class="error-message">%s</div>
		</div>`, r.reportData.Error)
	}
	content = strings.ReplaceAll(content, "{{ERROR_SECTION}}", errorSection)
	
	// Steps section
	stepsHTML := r.generateStepsHTML()
	content = strings.ReplaceAll(content, "{{STEPS_SECTION}}", stepsHTML)
	
	// Actions section
	actionsHTML := r.generateActionsHTML()
	content = strings.ReplaceAll(content, "{{ACTIONS_SECTION}}", actionsHTML)
	
	// Output files section
	outputFilesHTML := r.generateOutputFilesHTML()
	content = strings.ReplaceAll(content, "{{OUTPUT_FILES_SECTION}}", outputFilesHTML)
	
	return content
}

// generateStepsHTML creates the steps section of the HTML report
func (r *Reporter) generateStepsHTML() string {
	var html strings.Builder
	
	html.WriteString(`<div class="steps-section">
		<h3>📋 Step Performance</h3>
		<div class="table-container">
			<table>
				<thead>
					<tr>
						<th>Step Name</th>
						<th>Executions</th>
						<th>Success Rate</th>
						<th>Avg Duration</th>
						<th>Min Duration</th>
						<th>Max Duration</th>
						<th>Total Duration</th>
					</tr>
				</thead>
				<tbody>`)
	
	for _, step := range r.reportData.Steps {
		html.WriteString(fmt.Sprintf(`
					<tr>
						<td><strong>%s</strong></td>
						<td>%d</td>
						<td><span class="success-rate %s">%.1f%%</span></td>
						<td>%s</td>
						<td>%s</td>
						<td>%s</td>
						<td>%s</td>
					</tr>`,
			step.Name,
			step.TotalExecutions,
			r.getSuccessRateClass(step.SuccessRate),
			step.SuccessRate,
			r.formatDuration(step.AverageDuration),
			r.formatDuration(step.MinDuration),
			r.formatDuration(step.MaxDuration),
			r.formatDuration(step.TotalDuration),
		))
	}
	
	html.WriteString(`
				</tbody>
			</table>
		</div>
	</div>`)
	
	return html.String()
}

// generateActionsHTML creates the actions section of the HTML report
func (r *Reporter) generateActionsHTML() string {
	var html strings.Builder
	
	html.WriteString(`<div class="actions-section">
		<h3>⚡ Action Performance</h3>
		<div class="table-container">
			<table>
				<thead>
					<tr>
						<th>Action Type</th>
						<th>Step</th>
						<th>Executions</th>
						<th>Success Rate</th>
						<th>Avg Duration</th>
						<th>Total Duration</th>
						<th>Last Error</th>
					</tr>
				</thead>
				<tbody>`)
	
	for _, action := range r.reportData.Actions {
		lastError := action.LastError
		if len(lastError) > 50 {
			lastError = lastError[:47] + "..."
		}
		
		html.WriteString(fmt.Sprintf(`
					<tr>
						<td><code>%s</code></td>
						<td>%s</td>
						<td>%d</td>
						<td><span class="success-rate %s">%.1f%%</span></td>
						<td>%s</td>
						<td>%s</td>
						<td class="error-cell">%s</td>
					</tr>`,
			action.Type,
			action.StepName,
			action.TotalExecutions,
			r.getSuccessRateClass(action.SuccessRate),
			action.SuccessRate,
			r.formatDuration(action.AverageDuration),
			r.formatDuration(action.TotalDuration),
			lastError,
		))
	}
	
	html.WriteString(`
				</tbody>
			</table>
		</div>
	</div>`)
	
	return html.String()
}

// generateOutputFilesHTML creates the output files section
func (r *Reporter) generateOutputFilesHTML() string {
	var html strings.Builder
	
	html.WriteString(`<div class="output-files-section">
		<h3>📁 Output Files</h3>`)
	
	if len(r.outputFiles) == 0 {
		html.WriteString(`<p class="no-files">No output files generated.</p>`)
	} else {
		html.WriteString(`<div class="files-grid">`)
		
		for _, fileURL := range r.outputFiles {
			fileName := filepath.Base(fileURL)
			fileType := r.getFileType(fileURL)
			
			html.WriteString(fmt.Sprintf(`
				<div class="file-item">
					<div class="file-icon">%s</div>
					<div class="file-info">
						<div class="file-name">%s</div>
						<div class="file-type">%s</div>
						<a href="%s" target="_blank" class="file-link">View File →</a>
					</div>
				</div>`,
				r.getFileIcon(fileType),
				fileName,
				strings.ToUpper(fileType),
				fileURL,
			))
		}
		
		html.WriteString(`</div>`)
	}
	
	html.WriteString(`</div>`)
	return html.String()
}

// Helper functions
func (r *Reporter) getStatusClass(status string) string {
	switch strings.ToLower(status) {
	case "completed":
		return "status-success"
	case "failed":
		return "status-error"
	default:
		return "status-warning"
	}
}

func (r *Reporter) getSuccessRateClass(rate float64) string {
	if rate >= 95 {
		return "rate-excellent"
	} else if rate >= 80 {
		return "rate-good"
	} else if rate >= 60 {
		return "rate-warning"
	} else {
		return "rate-poor"
	}
}

func (r *Reporter) formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0ms"
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := d.Seconds() - float64(minutes*60)
		return fmt.Sprintf("%dm %.2fs", minutes, seconds)
	}
}

func (r *Reporter) getFileType(url string) string {
	ext := strings.ToLower(filepath.Ext(url))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return "image"
	case ".pdf":
		return "pdf"
	case ".json":
		return "json"
	case ".txt":
		return "text"
	case ".csv":
		return "csv"
	default:
		return "file"
	}
}

func (r *Reporter) getFileIcon(fileType string) string {
	switch fileType {
	case "image":
		return "🖼️"
	case "pdf":
		return "📄"
	case "json":
		return "📋"
	case "text":
		return "📝"
	case "csv":
		return "📊"
	default:
		return "📁"
	}
}

// Helper functions for type conversion
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}