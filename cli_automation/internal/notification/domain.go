package notification

import (
	"context"
	"time"
)

// MailData represents the data needed for sending an email
type MailData struct {
	To      string
	Subject string
	Content string
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	SendMail(ctx context.Context, mailData MailData) error
	SendLoginCode(ctx context.Context, email string, code string) error
	SendOrganizationInvite(ctx context.Context, email, orgName, inviteURL string) error
	DispatchAutomationNotification(ctx context.Context, message NotificationMessage, channels []NotificationChannelConfig) error
}

// NotificationMessage represents the data for automation notifications
type NotificationMessage struct {
	AutomationID   string
	AutomationName string
	ProjectID      string
	ProjectName    string
	RunID          string
	Status         string // "completed", "failed"
	StartTime      *time.Time
	EndTime        *time.Time
	ErrorMessage   string
	OutputFiles    []string
	LogsCount      int
}

// NotificationChannelConfig represents a notification channel configuration
type NotificationChannelConfig struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "slack", "email", "webhook"
	OnComplete bool                   `json:"onComplete"`
	OnError    bool                   `json:"onError"`
	Config     map[string]interface{} `json:"config"`
}