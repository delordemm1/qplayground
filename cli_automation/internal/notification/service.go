package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/delordemm1/qplayground-cli/internal/platform"

	mail "github.com/xhit/go-simple-mail/v2"
)

type MailService struct {
	host     string
	port     int
	username string
	password string
	from     string
	
	// Notification channel implementations
	slackNotifier ChannelNotifier
}

func NewMailService() *MailService {
	return &MailService{
		host:     platform.ENV_SMTP_HOST,
		port:     platform.ENV_SMTP_PORT,
		username: platform.ENV_SMTP_USERNAME,
		password: platform.ENV_SMTP_PASSWORD,
		from:     platform.ENV_SMTP_FROM,
		
		slackNotifier: NewSlackNotifier(),
	}
}

func (s *MailService) SendMail(ctx context.Context, m MailData) error {
	if s.host == "" || s.username == "" {
		slog.Debug("SMTP not configured, skipping email send")
		return nil
	}

	server := mail.NewSMTPClient()
	server.Host = s.host
	server.Port = s.port
	server.Username = s.username
	server.Password = s.password
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		slog.Error("Failed to connect to SMTP server", "error", err)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	email := mail.NewMSG()
	email.SetFrom(s.from).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	err = email.Send(client)
	if err != nil {
		slog.Error("Failed to send email", "error", err, "to", m.To)
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("Email sent successfully", "to", m.To, "subject", m.Subject)
	return nil
}

func (s *MailService) SendLoginCode(ctx context.Context, email string, code string) error {
	m := MailData{
		To:      email,
		Subject: "Your QPlayground Login Code",
		Content: fmt.Sprintf(`
			<html>
			<body>
				<h2>Your QPlayground Login Code</h2>
				<p>Your temporary login code is: <strong>%s</strong></p>
				<p>This code will expire in 10 minutes.</p>
				<p>If you didn't request this code, please ignore this email.</p>
			</body>
			</html>
		`, code),
	}

	return s.SendMail(ctx, m)
}

func (s *MailService) SendOrganizationInvite(ctx context.Context, email, orgName, inviteURL string) error {
	m := MailData{
		To:      email,
		Subject: fmt.Sprintf("You're invited to join %s on QPlayground", orgName),
		Content: fmt.Sprintf(`
			<html>
			<body>
				<h2>You're invited to join %s</h2>
				<p>You've been invited to join <strong>%s</strong> on QPlayground.</p>
				<p>Click the link below to accept the invitation:</p>
				<p><a href="%s" style="background-color: #007BFF; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Accept Invitation</a></p>
				<p>If the button doesn't work, copy and paste this link into your browser:</p>
				<p>%s</p>
				<p>This invitation will expire in 7 days.</p>
				<p>If you didn't expect this invitation, please ignore this email.</p>
			</body>
			</html>
		`, orgName, orgName, inviteURL, inviteURL),
	}

	return s.SendMail(ctx, m)
}

func (s *MailService) DispatchAutomationNotification(ctx context.Context, message NotificationMessage, channels []NotificationChannelConfig) error {
	for _, channel := range channels {
		// Check if this channel should be triggered based on the message status
		shouldSend := false
		if message.Status == "completed" && channel.OnComplete {
			shouldSend = true
		} else if message.Status == "failed" && channel.OnError {
			shouldSend = true
		}

		if !shouldSend {
			continue
		}

		// Send notification based on channel type
		var err error
		switch channel.Type {
		case "slack":
			err = s.slackNotifier.Send(ctx, message, channel.Config)
		case "email":
			// TODO: Implement email notifications
			slog.Warn("Email notifications not yet implemented", "channel_id", channel.ID)
			continue
		case "webhook":
			// TODO: Implement generic webhook notifications
			slog.Warn("Generic webhook notifications not yet implemented", "channel_id", channel.ID)
			continue
		default:
			slog.Warn("Unknown notification channel type", "type", channel.Type, "channel_id", channel.ID)
			continue
		}

		if err != nil {
			slog.Error("Failed to send notification", 
				"channel_type", channel.Type,
				"channel_id", channel.ID,
				"automation_id", message.AutomationID,
				"error", err)
			// Continue with other channels even if one fails
		} else {
			slog.Info("Notification sent successfully",
				"channel_type", channel.Type,
				"channel_id", channel.ID,
				"automation_id", message.AutomationID)
		}
	}

	return nil
}