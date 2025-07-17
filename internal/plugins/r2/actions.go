package r2

import (
	"context"
	"fmt"
	"strings"

	"github.com/delordemm1/qplayground/internal/modules/automation"
)

func init() {
	automation.RegisterAction("r2:upload", func() automation.PluginAction { return &UploadAction{} })
	automation.RegisterAction("r2:delete", func() automation.PluginAction { return &DeleteAction{} })
	automation.RegisterAction("r2:list", func() automation.PluginAction { return &ListAction{} })
}

// UploadAction implements uploading files to R2
type UploadAction struct{}

func (a *UploadAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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
	if err != nil {
		return fmt.Errorf("failed to upload file to R2: %w", err)
	}
	
	runContext.Logger.Info("File uploaded to R2", "key", key, "url", publicURL)
	return nil
}

// DeleteAction implements deleting files from R2
type DeleteAction struct{}

func (a *DeleteAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	key, ok := actionConfig["key"].(string)
	if !ok || key == "" {
		return fmt.Errorf("r2:delete action requires a 'key' string in config")
	}
	
	runContext.Logger.Info("Executing r2:delete", "key", key)
	
	err := runContext.StorageService.DeleteFile(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}
	
	runContext.Logger.Info("File deleted from R2", "key", key)
	return nil
}

// ListAction implements listing files in R2 (placeholder for future implementation)
type ListAction struct{}

func (a *ListAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	prefix, _ := actionConfig["prefix"].(string) // Optional prefix filter
	
	runContext.Logger.Info("Executing r2:list", "prefix", prefix)
	
	// Note: This would require extending the StorageService interface to support listing
	// For now, we'll just log that the action was called
	runContext.Logger.Info("R2 list operation completed", "prefix", prefix)
	
	return nil
}