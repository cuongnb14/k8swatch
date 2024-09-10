package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordNotifier struct {
	WebhookURL string
}

// SendNotification sends a notification to Discord
func (d *DiscordNotifier) SendNotification(podName, namespace string, restartCount int32, reason string) error {
	embed := DiscordEmbed{
		Title:       fmt.Sprintf("Pod Restarted `%s`", podName),
		Description: "",
		Color:       16711680, // Red color for alert
		Fields: []EmbedField{
			{
				Name:   "Namespace",
				Value:  namespace,
				Inline: true,
			},
			{
				Name:   "Restart Count",
				Value:  fmt.Sprintf("%d", restartCount),
				Inline: true,
			},
			{
				Name:   "Reason",
				Value:  reason,
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339), // Current timestamp
	}

	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("received non-OK response from Discord: %s", resp.Status)
	}

	return nil
}

type DiscordEmbed struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Color       int          `json:"color"`
	Fields      []EmbedField `json:"fields"`
	Timestamp   string       `json:"timestamp"`
}

// EmbedField represents a field in the embed
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// DiscordWebhookPayload defines the payload to be sent to Discord
type DiscordWebhookPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}
