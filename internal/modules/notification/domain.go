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
}