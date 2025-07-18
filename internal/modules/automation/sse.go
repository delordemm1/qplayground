package automation

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexandrevicenzi/go-sse"
)

// SSEManager handles Server-Sent Events for automation runs
type SSEManager struct {
	server *sse.Server
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	server := sse.NewServer(&sse.Options{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization",
		},
	})

	return &SSEManager{
		server: server,
	}
}

// GetServer returns the underlying SSE server for mounting
func (s *SSEManager) GetServer() *sse.Server {
	return s.server
}

// Shutdown gracefully shuts down the SSE server
func (s *SSEManager) Shutdown() {
	s.server.Shutdown()
}

// RunProgressMessage represents a progress update for an automation run
type RunProgressMessage struct {
	Type        string                 `json:"type"` // "status", "log", "step", "action", "error", "complete"
	RunID       string                 `json:"runId"`
	Status      string                 `json:"status,omitempty"`
	StepName    string                 `json:"stepName,omitempty"`
	ActionType  string                 `json:"actionType,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Progress    int                    `json:"progress,omitempty"` // 0-100
	TotalSteps  int                    `json:"totalSteps,omitempty"`
	CurrentStep int                    `json:"currentStep,omitempty"`
	Duration    int64                  `json:"duration,omitempty"` // milliseconds
	OutputFile  string                 `json:"outputFile,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// SendRunProgress sends a progress update for a specific run
func (s *SSEManager) SendRunProgress(projectID, automationID, runID string, message RunProgressMessage) error {
	message.Timestamp = time.Now()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal progress message: %w", err)
	}

	channel := fmt.Sprintf("/projects/%s/automations/%s/runs/%s/events", projectID, automationID, runID)
	s.server.SendMessage(channel, sse.SimpleMessage(string(data)))

	slog.Debug("Sent SSE progress update",
		"run_id", runID,
		"type", message.Type,
		"status", message.Status)

	return nil
}

// SendRunStatusUpdate sends a status change update
func (s *SSEManager) SendRunStatusUpdate(projectID, automationID, runID, status string) error {
	return s.SendRunProgress(projectID, automationID, runID, RunProgressMessage{
		Type:   "status",
		RunID:  runID,
		Status: status,
	})
}

// SendRunLog sends a log entry update
func (s *SSEManager) SendRunLog(projectID, automationID, runID, stepName, actionType, message string, duration int64) error {
	return s.SendRunProgress(projectID, automationID, runID, RunProgressMessage{
		Type:       "log",
		RunID:      runID,
		StepName:   stepName,
		ActionType: actionType,
		Message:    message,
		Duration:   duration,
	})
}

// SendRunError sends an error update
func (s *SSEManager) SendRunError(projectID, automationID, runID, stepName, actionType, errorMsg string) error {
	return s.SendRunProgress(projectID, automationID, runID, RunProgressMessage{
		Type:       "error",
		RunID:      runID,
		StepName:   stepName,
		ActionType: actionType,
		Error:      errorMsg,
	})
}

// SendRunStep sends a step progress update
func (s *SSEManager) SendRunStep(projectID, automationID, runID, stepName string, currentStep, totalSteps int) error {
	progress := 0
	if totalSteps > 0 {
		progress = (currentStep * 100) / totalSteps
	}

	return s.SendRunProgress(projectID, automationID, runID, RunProgressMessage{
		Type:        "step",
		RunID:       runID,
		StepName:    stepName,
		Progress:    progress,
		CurrentStep: currentStep,
		TotalSteps:  totalSteps,
	})
}

// SendRunComplete sends a completion update
func (s *SSEManager) SendRunComplete(projectID, automationID, runID, status string, duration int64, outputFiles []string) error {
	return s.SendRunProgress(projectID, automationID, runID, RunProgressMessage{
		Type:     "complete",
		RunID:    runID,
		Status:   status,
		Duration: duration,
		Data: map[string]interface{}{
			"outputFiles": outputFiles,
		},
	})
}

// SendRunOutputFile sends an output file update
func (s *SSEManager) SendRunOutputFile(projectID, automationID, runID, outputFile string) error {
	return s.SendRunProgress(projectID, automationID, runID, RunProgressMessage{
		Type:       "output",
		RunID:      runID,
		OutputFile: outputFile,
	})
}
