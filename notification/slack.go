package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackNotifier struct {
	WebhookURL string
}

// SendNotification sends a notification to Slack
func (s *SlackNotifier) SendNotification(podName, namespace string, restartCount int32, reason string) error {
	attachment := SlackAttachment{
		Color:     "danger", // Red color for alert
		Title:     fmt.Sprintf("Pod Restarted: %s", podName),
		Timestamp: time.Now().Unix(),
		Fields: []SlackField{
			{
				Title: "Namespace",
				Value: namespace,
				Short: true,
			},
			{
				Title: "Restart Count",
				Value: fmt.Sprintf("%d", restartCount),
				Short: true,
			},
			{
				Title: "Reason",
				Value: reason,
				Short: true,
			},
		},
	}

	payload := SlackWebhookPayload{
		Text:        fmt.Sprintf("🚨 Pod restart detected in namespace `%s`", namespace),
		Attachments: []SlackAttachment{attachment},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response from Slack: %s", resp.Status)
	}

	return nil
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackAttachment struct {
	Color     string       `json:"color"`
	Title     string       `json:"title"`
	Timestamp int64        `json:"ts"`
	Fields    []SlackField `json:"fields"`
}

type SlackWebhookPayload struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}