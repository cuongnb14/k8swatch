package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DiscordEmbed defines the structure of an embed message for Discord
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

// sendDiscordNotification sends an embedded message to Discord
func sendDiscordNotification(webhookURL, podName, namespace string, restartCount int32, reason string) error {
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

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("received non-OK response from Discord: %s", resp.Status)
	}

	return nil
}

// getRestartReason checks the reason why a pod was restarted
func getRestartReason(status *corev1.ContainerStatus) string {
	if status.LastTerminationState.Terminated != nil {
		return status.LastTerminationState.Terminated.Reason
	}
	return "Unknown"
}

// State to track pod restart counts
var podRestartCounts = make(map[string]int32)
var mu sync.Mutex // Mutex for thread safety

// checkPodRestarts checks for pod restarts and sends a notification only if restart count increases
func checkPodRestarts(clientset *kubernetes.Clientset, webhookURL string) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error fetching pods: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	for _, pod := range pods.Items {
		for _, status := range pod.Status.ContainerStatuses {
			podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
			previousRestartCount, exists := podRestartCounts[podKey]

			if !exists {
				podRestartCounts[podKey] = 0

			}
			if status.RestartCount > previousRestartCount {
				reason := getRestartReason(&status)

				err := sendDiscordNotification(webhookURL, pod.Name, pod.Namespace, status.RestartCount, reason)
				if err != nil {
					log.Printf("Failed to send Discord notification: %v", err)
				} else {
					log.Printf("Sent notification for pod %s", pod.Name)
				}

				// Update the restart count in the map
				podRestartCounts[podKey] = status.RestartCount
			}
		}
	}
}

func main() {

	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhookURL == "" {
		log.Fatalf("Discord webhook URL not provided")
	}

	// Create Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Check pod restarts every minute
	for {
		checkPodRestarts(clientset, discordWebhookURL)
		time.Sleep(1 * time.Minute)
	}
}
