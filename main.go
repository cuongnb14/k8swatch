package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DiscordWebhookPayload defines the payload to be sent to Discord
type DiscordWebhookPayload struct {
	Content string `json:"content"`
}

func sendDiscordNotification(webhookURL, message string) error {
	payload := DiscordWebhookPayload{
		Content: message,
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

func checkPodRestarts(clientset *kubernetes.Clientset, webhookURL string) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error fetching pods: %v", err)
	}

	for _, pod := range pods.Items {
		for _, status := range pod.Status.ContainerStatuses {
			if status.RestartCount > 0 {
				message := fmt.Sprintf("Pod %s in namespace %s has restarted %d times.", pod.Name, pod.Namespace, status.RestartCount)
				err := sendDiscordNotification(webhookURL, message)
				if err != nil {
					log.Printf("Failed to send Discord notification: %v", err)
				} else {
					log.Printf("Sent notification for pod %s", pod.Name)
				}
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
