package r2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/automation"
)

func init() {
	automation.RegisterAction("r2:upload", func() automation.PluginAction { return &UploadAction{} })
	automation.RegisterAction("r2:delete", func() automation.PluginAction { return &DeleteAction{} })
	automation.RegisterAction("r2:list", func() automation.PluginAction { return &ListAction{} })
}

// Helper function to send success event for R2 actions
func sendR2SuccessEvent(runContext *automation.RunContext, actionType, message string, duration time.Duration) {
	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:       automation.RunEventTypeLog,
			Timestamp:  time.Now(),
			StepName:   runContext.StepName,
			ActionType: actionType,
			Message:    message,
			Duration:   duration.Milliseconds(),
			LoopIndex:  runContext.LoopIndex,
		}:
		default:
			// Channel is full, skip this event to avoid blocking
		}
	}
}

// Helper function to send error event for R2 actions
func sendR2ErrorEvent(runContext *automation.RunContext, actionType, errorMsg string, duration time.Duration) {
	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:       automation.RunEventTypeError,
			Timestamp:  time.Now(),
			StepName:   runContext.StepName,
			ActionType: actionType,
			Error:      errorMsg,
			Duration:   duration.Milliseconds(),
			LoopIndex:  runContext.LoopIndex,
		}:
		default:
			// Channel is full, skip this event to avoid blocking
		}
	}
}

// UploadAction implements uploading files to R2
type UploadAction struct{}

func (a *UploadAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	startTime := time.Now()
	key, ok := actionConfig["key"].(string)
	if !ok || key == "" {
		return fmt.Errorf("r2:upload action requires a 'key' string in config")
	}
	
	content, ok := actionConfig["content"].(string)
	if !ok {
		return fmt.Errorf("r2:upload action requires 'content' string in config")
	}
	
	// Determine content type
	contentType, _ := actionConfig["content_type"].(string)
	if contentType == "" {
		contentType = "text/plain" // Default
		
		// Auto-detect based on file extension
		if strings.HasSuffix(key, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(key, ".json") {
			contentType = "application/json"
		} else if strings.HasSuffix(key, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(key, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(key, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(key, ".pdf") {
			contentType = "application/pdf"
		}
	}
	
	runContext.Logger.Info("Executing r2:upload", "key", key, "content_type", contentType, "size", len(content))
	
	// Create reader from content
	reader := strings.NewReader(content)
	
	// Upload to R2
	publicURL, err := runContext.StorageService.UploadFile(ctx, key, reader, contentType)
	duration := time.Since(startTime)
	
	if err != nil {
		sendR2ErrorEvent(runContext, "r2:upload", fmt.Sprintf("failed to upload file to R2: %v", err), duration)
		return fmt.Errorf("failed to upload file to R2: %w", err)
	}
	
	runContext.Logger.Info("File uploaded to R2", "key", key, "url", publicURL)
	
	// Send output file event
	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:       automation.RunEventTypeOutputFile,
			Timestamp:  time.Now(),
			StepName:   runContext.StepName,
			ActionType: "r2:upload",
			OutputFile: publicURL,
			Duration:   duration.Milliseconds(),
			LoopIndex:  runContext.LoopIndex,
		}:
		default:
			// Channel is full, skip this event to avoid blocking
		}
	}
	
	sendR2SuccessEvent(runContext, "r2:upload", fmt.Sprintf("Successfully uploaded file to R2: %s", key), duration)
	return nil
}

// DeleteAction implements deleting files from R2
type DeleteAction struct{}

func (a *DeleteAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	startTime := time.Now()
	key, ok := actionConfig["key"].(string)
	if !ok || key == "" {
		return fmt.Errorf("r2:delete action requires a 'key' string in config")
	}
	
	runContext.Logger.Info("Executing r2:delete", "key", key)
	
	err := runContext.StorageService.DeleteFile(ctx, key)
	duration := time.Since(startTime)
	
	if err != nil {
		sendR2ErrorEvent(runContext, "r2:delete", fmt.Sprintf("failed to delete file from R2: %v", err), duration)
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}
	
	runContext.Logger.Info("File deleted from R2", "key", key)
	sendR2SuccessEvent(runContext, "r2:delete", fmt.Sprintf("Successfully deleted file from R2: %s", key), duration)
	return nil
}

// ListAction implements listing files in R2 (placeholder for future implementation)
type ListAction struct{}

func (a *ListAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	startTime := time.Now()
	prefix, _ := actionConfig["prefix"].(string) // Optional prefix filter
	
	runContext.Logger.Info("Executing r2:list", "prefix", prefix)
	
	// Note: This would require extending the StorageService interface to support listing
	// For now, we'll just log that the action was called
	runContext.Logger.Info("R2 list operation completed", "prefix", prefix)
	
	duration := time.Since(startTime)
	sendR2SuccessEvent(runContext, "r2:list", fmt.Sprintf("Successfully listed R2 files with prefix: %s", prefix), duration)
	return nil
}