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

// NotificationTask represents a notification task to be processed
type NotificationTask struct {
	ID        string
	Channel   NotificationChannel
	Priority  int
	Data      map[string]interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NotificationChannel represents the type of notification
type NotificationChannel string

const (
	NotificationChannelEmail NotificationChannel = "EMAIL"
	NotificationChannelSMS   NotificationChannel = "SMS"
	NotificationChannelPush  NotificationChannel = "PUSH"
)

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	// Future: Store notification tasks for background processing
	InsertNotificationTask(ctx context.Context, task *NotificationTask) error
	GetPendingTasks(ctx context.Context, limit int) ([]*NotificationTask, error)
	DeleteTask(ctx context.Context, id string) error
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

// ChannelNotifier defines the interface for specific notification channels
type ChannelNotifier interface {
	Send(ctx context.Context, message NotificationMessage, channelConfig map[string]interface{}) error
}