package main

import (
	"context"
	"fmt"
	"github/cuongnb14/k8swatch/notification"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Notifier interface {
	SendNotification(podName, namespace string, restartCount int32, reason string) error
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
func checkPodRestarts(clientset *kubernetes.Clientset, notifiers []Notifier) {
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

				for _, notifier := range notifiers {
					err := notifier.SendNotification(pod.Name, pod.Namespace, status.RestartCount, reason)
					if err != nil {
						log.Printf("Failed to send notification: %v", err)
					}
				}

				// Update the restart count in the map
				podRestartCounts[podKey] = status.RestartCount
			}
		}
	}
}

func main() {

	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	intervalStr := os.Getenv("CHECK_INTERVAL")
	iter := 60
	if intervalStr != "" {
		var err error
		iter, err = strconv.Atoi(intervalStr)
		if err != nil {
			log.Fatal("Error:", err)
		}
	}

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

	notifiers := []Notifier{
		&notification.DiscordNotifier{WebhookURL: discordWebhookURL},
		// Add Slack, Telegram notifier instances as needed
	}

	// Check pod restarts every minute
	for {
		checkPodRestarts(clientset, notifiers)
		time.Sleep(time.Duration(iter) * time.Second)
	}
}
