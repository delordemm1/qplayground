package platform

import (
	"context"
	"encoding/json"

	"log/slog"

	"github.com/alexedwards/scs/v2"
	inertia "github.com/romsar/gonertia/v2"
)

// SCSFlashProvider implements gonertia.FlashProvider using SCS session manager
type SCSFlashProvider struct {
	sessionManager *scs.SessionManager
}

// NewSCSFlashProvider creates a new SCS-based flash provider
func NewSCSFlashProvider(sessionManager *scs.SessionManager) *SCSFlashProvider {
	return &SCSFlashProvider{
		sessionManager: sessionManager,
	}
}

// FlashErrors stores validation errors in the session
func (p *SCSFlashProvider) FlashErrors(ctx context.Context, errors inertia.ValidationErrors) error {
	p.sessionManager.Put(ctx, "validation_errors", errors)
	return nil
}

// GetErrors retrieves and removes validation errors from the session
func (p *SCSFlashProvider) GetErrors(ctx context.Context) (inertia.ValidationErrors, error) {
	errors := p.sessionManager.Get(ctx, "validation_errors")
	if errors == nil {
		return nil, nil
	}

	// Remove the errors from session after retrieving
	p.sessionManager.Remove(ctx, "validation_errors")

	if validationErrors, ok := errors.(inertia.ValidationErrors); ok {
		return validationErrors, nil
	}

	return nil, nil
}

// FlashClearHistory sets a flag to clear history on next request
func (p *SCSFlashProvider) FlashClearHistory(ctx context.Context) error {
	p.sessionManager.Put(ctx, "clear_history", true)
	return nil
}

// ShouldClearHistory checks and removes the clear history flag
func (p *SCSFlashProvider) ShouldClearHistory(ctx context.Context) (bool, error) {
	shouldClear := p.sessionManager.GetBool(ctx, "clear_history")
	if shouldClear {
		p.sessionManager.Remove(ctx, "clear_history")
	}
	return shouldClear, nil
}

type FlashMessageType string

const (
	FlashMessageTypeSuccess FlashMessageType = "success"
	FlashMessageTypeError   FlashMessageType = "error"
	FlashMessageTypeInfo    FlashMessageType = "info"
	FlashMessageTypeWarning FlashMessageType = "warning"
)

// FlashMessage represents a general flash message
type FlashMessage struct {
	Type    FlashMessageType `json:"type"` // success, error, info, warning
	Message string           `json:"message"`
}

// GetFlashMessage retrieves and removes a general flash message from the session
func GetFlashMessage(ctx context.Context, sessionManager *scs.SessionManager) *FlashMessage {
	// Pop does a Get and a Remove in one atomic operation.
	// It returns a byte slice, so we need to unmarshal it.
	poppedData, ok := sessionManager.Pop(ctx, "flash_message").([]byte)
	if !ok || poppedData == nil {
		return nil
	}
	slog.Debug("Popped flash message", "data", string(poppedData))
	var flashMessage FlashMessage
	if err := json.Unmarshal(poppedData, &flashMessage); err != nil {
		// Handle error, maybe log it. The data in the session was malformed.
		slog.Error("Error unmarshaling flash message", "error", err)
		return nil
	}
	slog.Debug("Unmarshalled flash message", "flashMessage", flashMessage)
	checkFlash, ok := sessionManager.Get(ctx, "flash_message").([]byte)
	slog.Debug("Check flash message", "checkFlash", string(checkFlash), "ok", ok)
	return &flashMessage
}

func SetFlashMessage(ctx context.Context, sessionManager *scs.SessionManager, messageType FlashMessageType, message string) {
	flashMessage := FlashMessage{
		Type:    messageType,
		Message: message,
	}
	// Marshal the struct to JSON before storing
	jsonData, err := json.Marshal(flashMessage)
	if err != nil {
		// Handle error, maybe log it and don't set the message
		slog.Error("Error marshaling flash message", "error", err)
		return
	}
	sessionManager.Put(ctx, "flash_message", jsonData)
}

// SetFlashSuccess is a convenience function for success messages
func SetFlashSuccess(ctx context.Context, sessionManager *scs.SessionManager, message string) {
	SetFlashMessage(ctx, sessionManager, FlashMessageTypeSuccess, message)
}

// SetFlashError is a convenience function for error messages
func SetFlashError(ctx context.Context, sessionManager *scs.SessionManager, message string) {
	SetFlashMessage(ctx, sessionManager, FlashMessageTypeError, message)
}

// SetFlashInfo is a convenience function for info messages
func SetFlashInfo(ctx context.Context, sessionManager *scs.SessionManager, message string) {
	SetFlashMessage(ctx, sessionManager, FlashMessageTypeInfo, message)
}

// SetFlashWarning is a convenience function for warning messages
func SetFlashWarning(ctx context.Context, sessionManager *scs.SessionManager, message string) {
	SetFlashMessage(ctx, sessionManager, FlashMessageTypeWarning, message)
}
