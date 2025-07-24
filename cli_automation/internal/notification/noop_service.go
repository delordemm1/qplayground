package notification

import (
	"context"
	"log/slog"
)

type NoOpMailService struct{}

func NewNoOpMailService() *NoOpMailService {
	return &NoOpMailService{}
}

func (s *NoOpMailService) SendMail(ctx context.Context, mailData MailData) error {
	slog.Debug("NoOp: SendMail called", "to", mailData.To, "subject", mailData.Subject)
	return nil
}

func (s *NoOpMailService) SendLoginCode(ctx context.Context, email string, code string) error {
	slog.Debug("NoOp: SendLoginCode called", "email", email, "code", code)
	return nil
}

func (s *NoOpMailService) SendOrganizationInvite(ctx context.Context, email, orgName, inviteURL string) error {
	slog.Debug("NoOp: SendOrganizationInvite called", "email", email, "orgName", orgName)
	return nil
}

func (s *NoOpMailService) DispatchAutomationNotification(ctx context.Context, message NotificationMessage, channels []NotificationChannelConfig) error {
	slog.Debug("NoOp: DispatchAutomationNotification called", "automation_id", message.AutomationID, "status", message.Status)
	return nil
}