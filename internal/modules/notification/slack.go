package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// SlackNotifier implements ChannelNotifier for Slack incoming webhooks
type SlackNotifier struct{}

// SlackMessage represents the structure of a Slack webhook message
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackNotifier creates a new SlackNotifier instance
func NewSlackNotifier() ChannelNotifier {
	return &SlackNotifier{}
}

// Send sends a notification to Slack via incoming webhook
func (s *SlackNotifier) Send(ctx context.Context, message NotificationMessage, channelConfig map[string]interface{}) error {
	webhookURL, ok := channelConfig["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("slack webhook URL is required")
	}

	// Build Slack message
	slackMessage := s.buildSlackMessage(message, channelConfig)

	// Marshal to JSON
	payload, err := json.Marshal(slackMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	// Send HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	slog.Info("Slack notification sent successfully", 
		"automation_id", message.AutomationID,
		"run_id", message.RunID,
		"status", message.Status)

	return nil
}

// buildSlackMessage constructs the Slack message from the notification data
func (s *SlackNotifier) buildSlackMessage(message NotificationMessage, channelConfig map[string]interface{}) SlackMessage {
	// Extract optional config values
	username, _ := channelConfig["username"].(string)
	if username == "" {
		username = "QPlayground Bot"
	}

	iconEmoji, _ := channelConfig["icon_emoji"].(string)
	if iconEmoji == "" {
		iconEmoji = ":robot_face:"
	}

	channel, _ := channelConfig["channel"].(string)

	// Determine color and main text based on status
	var color, statusEmoji, mainText string
	switch message.Status {
	case "completed":
		color = "good"
		statusEmoji = ":white_check_mark:"
		mainText = fmt.Sprintf("%s Automation *%s* completed successfully!", statusEmoji, message.AutomationName)
	case "failed":
		color = "danger"
		statusEmoji = ":x:"
		mainText = fmt.Sprintf("%s Automation *%s* failed!", statusEmoji, message.AutomationName)
	default:
		color = "warning"
		statusEmoji = ":warning:"
		mainText = fmt.Sprintf("%s Automation *%s* finished with status: %s", statusEmoji, message.AutomationName, message.Status)
	}

	// Build fields
	fields := []SlackField{
		{
			Title: "Project",
			Value: message.ProjectName,
			Short: true,
		},
		{
			Title: "Run ID",
			Value: message.RunID[:8] + "...", // Shortened for display
			Short: true,
		},
	}

	// Add duration if both start and end times are available
	if message.StartTime != nil && message.EndTime != nil {
		duration := message.EndTime.Sub(*message.StartTime)
		fields = append(fields, SlackField{
			Title: "Duration",
			Value: s.formatDuration(duration),
			Short: true,
		})
	}

	// Add logs count
	if message.LogsCount > 0 {
		fields = append(fields, SlackField{
			Title: "Log Entries",
			Value: fmt.Sprintf("%d", message.LogsCount),
			Short: true,
		})
	}

	// Add output files count
	if len(message.OutputFiles) > 0 {
		fields = append(fields, SlackField{
			Title: "Output Files",
			Value: fmt.Sprintf("%d files generated", len(message.OutputFiles)),
			Short: true,
		})
	}

	// Add error message if present
	if message.ErrorMessage != "" {
		fields = append(fields, SlackField{
			Title: "Error",
			Value: message.ErrorMessage,
			Short: false,
		})
	}

	attachment := SlackAttachment{
		Color:  color,
		Title:  fmt.Sprintf("Automation Run Details"),
		Text:   fmt.Sprintf("Run ID: `%s`", message.RunID),
		Fields: fields,
	}

	// Add timestamp if available
	if message.EndTime != nil {
		attachment.Timestamp = message.EndTime.Unix()
	} else if message.StartTime != nil {
		attachment.Timestamp = message.StartTime.Unix()
	}

	return SlackMessage{
		Text:        mainText,
		Username:    username,
		IconEmoji:   iconEmoji,
		Channel:     channel,
		Attachments: []SlackAttachment{attachment},
	}
}

// formatDuration formats a duration into a human-readable string
func (s *SlackNotifier) formatDuration(duration time.Duration) string {
	if duration < time.Second {
		return fmt.Sprintf("%dms", duration.Milliseconds())
	} else if duration < time.Minute {
		return fmt.Sprintf("%.2fs", duration.Seconds())
	} else {
		minutes := int(duration.Minutes())
		seconds := duration.Seconds() - float64(minutes*60)
		return fmt.Sprintf("%dm %.2fs", minutes, seconds)
	}
}